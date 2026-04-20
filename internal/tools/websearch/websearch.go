// Package websearch implements a web search tool using DuckDuckGo HTML.
package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/atom-yt/claude-code-go/internal/tools"
	"github.com/atom-yt/claude-code-go/internal/urlutil"
)

// Tool implements the WebSearch tool.
type Tool struct {
	validator *urlutil.URLValidator
}

var _ tools.Tool = (*Tool)(nil)

// NewTool creates a new WebSearch tool with URL validation enabled.
func NewTool() *Tool {
	return &Tool{
		validator: urlutil.NewURLValidator(),
	}
}

func (t *Tool) Name() string            { return "WebSearch" }
func (t *Tool) IsReadOnly() bool        { return true }
func (t *Tool) IsConcurrencySafe() bool { return true }

func (t *Tool) Description() string {
	return "Search the web for real-time information using a query string. Returns titles, URLs, and snippets from search results."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results to return (default 5, max 10)",
			},
		},
		"required": []string{"query"},
	}
}

func (t *Tool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	query, _ := input["query"].(string)
	if query == "" {
		return tools.ToolResult{Output: "Error: query parameter is required", IsError: true}, nil
	}

	maxResults := 5
	if mr, ok := input["max_results"].(float64); ok && int(mr) > 0 {
		maxResults = int(mr)
		if maxResults > 10 {
			maxResults = 10
		}
	}

	results, err := t.search(ctx, query, maxResults)
	if err != nil {
		return tools.ToolResult{Output: fmt.Sprintf("Search error: %v", err), IsError: true}, nil
	}

	if len(results) == 0 {
		return tools.ToolResult{Output: "No results found for: " + query}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n\n", i+1, r.title, r.url, r.snippet))
	}
	return tools.ToolResult{Output: sb.String()}, nil
}

type searchResult struct {
	title   string
	url     string
	snippet string
}

// search performs an HTML search via DuckDuckGo and parses the results.
func (t *Tool) search(ctx context.Context, query string, maxResults int) ([]searchResult, error) {
	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)

	// Validate the search URL
	if err := t.validator.Validate(searchURL); err != nil {
		return nil, fmt.Errorf("search URL validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; claude-code-go/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // 512KB limit
	if err != nil {
		return nil, err
	}

	return t.parseResults(string(body), maxResults), nil
}

// Regex patterns for DuckDuckGo HTML result parsing.
var (
	// Each result is in a <div class="result ..."> block containing an <a class="result__a"> and <a class="result__snippet">.
	reResultBlock = regexp.MustCompile(`(?s)<div[^>]*class="[^"]*result[^"]*results_links[^"]*"[^>]*>(.+?)</div>\s*</div>`)
	reTitle       = regexp.MustCompile(`<a[^>]*class="result__a"[^>]*>([^<]+)</a>`)
	reURL         = regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]+)"`)
	reSnippet     = regexp.MustCompile(`(?s)<a[^>]*class="result__snippet"[^>]*>(.*?)</a>`)
	reHTMLTag     = regexp.MustCompile(`<[^>]+>`)
)

func (t *Tool) parseResults(html string, maxResults int) []searchResult {
	blocks := reResultBlock.FindAllStringSubmatch(html, maxResults*2)
	var results []searchResult

	for _, block := range blocks {
		if len(results) >= maxResults {
			break
		}
		content := block[1]

		titleMatch := reTitle.FindStringSubmatch(content)
		urlMatch := reURL.FindStringSubmatch(content)
		snippetMatch := reSnippet.FindStringSubmatch(content)

		if titleMatch == nil || urlMatch == nil {
			continue
		}

		title := strings.TrimSpace(titleMatch[1])
		link := strings.TrimSpace(urlMatch[1])
		snippet := ""
		if snippetMatch != nil {
			snippet = strings.TrimSpace(reHTMLTag.ReplaceAllString(snippetMatch[1], ""))
		}

		// DuckDuckGo sometimes wraps URLs in a redirect; extract actual URL.
		if strings.Contains(link, "uddg=") {
			if u, err := url.Parse(link); err == nil {
				if actual := u.Query().Get("uddg"); actual != "" {
					link = actual
				}
			}
		}

		// Skip results that point to internal addresses
		if err := t.validator.Validate(link); err != nil {
			continue
		}

		results = append(results, searchResult{
			title:   title,
			url:     link,
			snippet: snippet,
		})
	}
	return results
}

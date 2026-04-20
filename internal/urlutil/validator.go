package urlutil

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type URLValidator struct {
	allowPatterns  []*regexp.Regexp
	denyPatterns   []*regexp.Regexp
	denySchemes    []string
	blockInternal   bool
}

func NewURLValidator() *URLValidator {
	return &URLValidator{
		denySchemes:  []string{"file://", "ftp://", "data:", "javascript:", "vbscript:"},
		blockInternal: true,
	}
}

func (v *URLValidator) SetAllowPatterns(patterns []string) error {
	v.allowPatterns = make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid allow pattern %q: %w", p, err)
		}
		v.allowPatterns[i] = re
	}
	return nil
}

func (v *URLValidator) SetDenyPatterns(patterns []string) error {
	v.denyPatterns = make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid deny pattern %q: %w", p, err)
		}
		v.denyPatterns[i] = re
	}
	return nil
}

func (v *URLValidator) SetBlockInternal(block bool) {
	v.blockInternal = block
}

func (v *URLValidator) Validate(inputURL string) error {
	if inputURL == "" {
		return fmt.Errorf("URL is empty")
	}

	parsed, err := url.Parse(inputURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	for _, scheme := range v.denySchemes {
		if strings.HasPrefix(inputURL, scheme) {
			return fmt.Errorf("URL scheme %q is not allowed", scheme)
		}
	}

	if v.blockInternal {
		if host := parsed.Host; host != "" && isInternalHost(host) {
			return fmt.Errorf("URL %q points to internal address", inputURL)
		}
	}

	for _, denyRe := range v.denyPatterns {
		if denyRe.MatchString(inputURL) || denyRe.MatchString(parsed.Host) {
			return fmt.Errorf("URL %q is blocked by security policy", inputURL)
		}
	}

	if len(v.allowPatterns) > 0 {
		allowed := false
		for _, allowRe := range v.allowPatterns {
			if allowRe.MatchString(inputURL) || allowRe.MatchString(parsed.Host) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("URL %q is not on allowlist", inputURL)
		}
	}

	return nil
}

func isInternalHost(host string) bool {
	if strings.HasPrefix(host, "[") {
		if idx := strings.LastIndex(host, "]"); idx > 0 {
			host = host[1:idx]
		}
	} else if strings.Count(host, ":") == 1 {
		host = strings.Split(host, ":")[0]
	}

	host = strings.ToLower(host)
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0":
		return true
	}

	return isPrivateIP(host)
}

func isPrivateIP(host string) bool {
	if host == "::1" || host == "::" || host == "0.0.0.0" {
		return true
	}

	if strings.HasPrefix(host, "fc00:") || strings.HasPrefix(host, "fe80:") {
		return true
	}

	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}

	first, _ := strconv.Atoi(parts[0])
	second, _ := strconv.Atoi(parts[1])

	if first == 10 {
		return true
	}
	if first == 172 && second >= 16 && second <= 31 {
		return true
	}
	if first == 192 && second == 168 {
		return true
	}

	return false
}
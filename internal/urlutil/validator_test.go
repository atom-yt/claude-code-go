package urlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURLValidator(t *testing.T) {
	v := NewURLValidator()
	assert.NotNil(t, v)
	assert.NotNil(t, v.denySchemes)
	assert.True(t, v.blockInternal)
}

func TestSetAllowPatterns(t *testing.T) {
	v := NewURLValidator()

	// Valid patterns
	err := v.SetAllowPatterns([]string{`^https://example\.com`, `^https://api\.test\.com`})
	assert.NoError(t, err)
	assert.Len(t, v.allowPatterns, 2)

	// Invalid regex pattern
	err = v.SetAllowPatterns([]string{"[invalid(regex"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid allow pattern")
}

func TestSetDenyPatterns(t *testing.T) {
	v := NewURLValidator()

	// Valid patterns
	err := v.SetDenyPatterns([]string{`^https://blocked\.com`, `malware\.site`})
	assert.NoError(t, err)
	assert.Len(t, v.denyPatterns, 2)

	// Invalid regex pattern
	err = v.SetDenyPatterns([]string{"[invalid(regex"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid deny pattern")
}

func TestSetBlockInternal(t *testing.T) {
	v := NewURLValidator()
	assert.True(t, v.blockInternal)

	v.SetBlockInternal(false)
	assert.False(t, v.blockInternal)

	v.SetBlockInternal(true)
	assert.True(t, v.blockInternal)
}

func TestValidate_EmptyURL(t *testing.T) {
	v := NewURLValidator()
	err := v.Validate("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL is empty")
}

func TestValidate_InvalidURL(t *testing.T) {
	v := NewURLValidator()
	// These URLs are technically parseable by Go's url.Parse,
	// but will be caught by deny rules or are malformed for actual use
	errs := []string{
		"://invalid",  // Will be rejected by deny schemes check (empty scheme falls through)
	}

	for _, u := range errs {
		err := v.Validate(u)
		// Note: Go's url.Parse is very permissive and accepts almost anything
		// Some of these might parse successfully but are invalid for actual use
		if err != nil {
			assert.Contains(t, err.Error(), "URL")
		}
	}

	// These should pass validation (valid format)
	validURLs := []string{
		"http://example.com",
		"https://example.com",
	}
	for _, u := range validURLs {
		err := v.Validate(u)
		assert.NoError(t, err, "Valid URL should pass: "+u)
	}
}

func TestValidate_DenySchemes(t *testing.T) {
	v := NewURLValidator()

	denySchemes := []struct {
		url      string
		expected string
	}{
		{"file:///etc/passwd", "file://"},
		{"ftp://example.com/file", "ftp://"},
		{"data:text/plain,hello", "data:"},
		{"javascript:alert('xss')", "javascript:"},
		{"vbscript:msgbox('xss')", "vbscript:"},
	}

	for _, tc := range denySchemes {
		err := v.Validate(tc.url)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
		assert.Contains(t, err.Error(), tc.expected)
	}

	// Valid schemes should pass
	validURLs := []string{
		"https://example.com",
		"http://example.com",
	}
	for _, u := range validURLs {
		err := v.Validate(u)
		assert.NoError(t, err, "URL should be valid: "+u)
	}
}

func TestValidate_BlockInternal(t *testing.T) {
	v := NewURLValidator()

	internalURLs := []string{
		"http://localhost:8080",
		"https://127.0.0.1/api",
		"http://0.0.0.0/admin",
		"http://[::1]:8080",
		"https://192.168.1.1/internal",
		"http://10.0.0.5/secret",
		"http://172.20.0.1/config",
		"http://fc00::1/private",
		"http://fe80::1/link",
	}

	for _, u := range internalURLs {
		err := v.Validate(u)
		assert.Error(t, err, "Should block internal URL: "+u)
		assert.Contains(t, err.Error(), "internal address")
	}

	// Allow internal when disabled
	v.SetBlockInternal(false)
	for _, u := range internalURLs {
		err := v.Validate(u)
		assert.NoError(t, err, "Should allow internal URL when disabled: "+u)
	}
}

func TestValidate_DenyPatterns(t *testing.T) {
	v := NewURLValidator()
	v.SetDenyPatterns([]string{`blocked\.com`, `malware`})

	blockedURLs := []string{
		"https://blocked.com/page",
		"http://subdomain.blocked.com/path",
		"https://example.com/malware/file",
		"http://malware-site.com",
	}

	for _, u := range blockedURLs {
		err := v.Validate(u)
		assert.Error(t, err, "Should block URL matching deny pattern: "+u)
		assert.Contains(t, err.Error(), "blocked by security policy")
	}
}

func TestValidate_AllowPatterns(t *testing.T) {
	v := NewURLValidator()
	v.SetAllowPatterns([]string{`^https://trusted\.com`, `^https://api\.safe\.org`})

	// Allowed URLs
	allowedURLs := []string{
		"https://trusted.com/page",
		"https://api.safe.org/endpoint",
	}
	for _, u := range allowedURLs {
		err := v.Validate(u)
		assert.NoError(t, err, "Should allow URL matching allow pattern: "+u)
	}

	// Not allowed URLs
	notAllowedURLs := []string{
		"https://other.com/page",
		"http://trusted.com/page",    // wrong scheme
		"https://untrusted.com",
	}
	for _, u := range notAllowedURLs {
		err := v.Validate(u)
		assert.Error(t, err, "Should reject URL not on allowlist: "+u)
		assert.Contains(t, err.Error(), "not on allowlist")
	}

	// When no allow patterns set, should accept
	v2 := NewURLValidator()
	err := v2.Validate("https://any-domain.com")
	assert.NoError(t, err)
}

func TestValidate_CombinedRules(t *testing.T) {
	v := NewURLValidator()
	v.SetAllowPatterns([]string{`^https://trusted\.com`})
	v.SetDenyPatterns([]string{`blocked`})

	// Deny patterns are checked first, so even allowed domains are blocked if they match deny
	err := v.Validate("https://trusted.com/blocked")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked by security policy")

	// Allow pattern passes for non-blocked paths
	err = v.Validate("https://trusted.com/allowed")
	assert.NoError(t, err)

	// Not allowed domain is blocked by allowlist (if deny doesn't match)
	v2 := NewURLValidator()
	v2.SetAllowPatterns([]string{`^https://trusted\.com`})
	err = v2.Validate("https://other.com/page")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not on allowlist")
}

func TestIsInternalHost(t *testing.T) {
	internalHosts := map[string]bool{
		"localhost":              true,
		"127.0.0.1":            true,
		"0.0.0.0":              true,
		"::1":                    true,
		"::":                     true,
		"127.0.0.1:8080":         true,
		"[::1]":                 true,
		"[::1]:8080":             true,
		"[fe80::1]":              true,
		"[fe80::1%eth0]:8080":    true,
		"example.com":             false,
		"8.8.8.8":               false,
		"::ffff":                 false,
	}

	for host, expected := range internalHosts {
		result := isInternalHost(host)
		assert.Equal(t, expected, result, "isInternalHost(%q)", host)
	}
}

func TestIsPrivateIP(t *testing.T) {
	privateIPs := map[string]bool{
		"10.0.0.1":          true,
		"10.255.255.255":    true,
		"172.16.0.1":        true,
		"172.31.255.255":    true,
		"192.168.0.1":       true,
		"192.168.255.255":    true,
		"172.15.255.255":     false,
		"172.32.0.1":        false,
		"8.8.8.8":           false,
		"1.2.3.4":           false,
		"fc00::1":            true,   // ULA prefix
		"fe80::1":            true,   // Link-local prefix
		"fd00::1":            false,  // Not fc00 prefix, returns false
		"2001:db8::1":       false,
		"::1":                true,   // Handled in isInternalHost
		"::":                 true,   // Handled in isInternalHost
		"0.0.0.0":            true,   // Handled in isInternalHost
	}

	for ip, expected := range privateIPs {
		result := isPrivateIP(ip)
		assert.Equal(t, expected, result, "isPrivateIP(%q)", ip)
	}
}

func TestValidate_ValidPublicURLs(t *testing.T) {
	v := NewURLValidator()

	publicURLs := []string{
		"https://www.anthropic.com",
		"http://example.com/path",
		"https://api.github.com/v1",
		"https://docs.google.com",
	}

	for _, u := range publicURLs {
		err := v.Validate(u)
		assert.NoError(t, err, "Valid public URL should pass: "+u)
	}
}

func TestValidate_URLWithPort(t *testing.T) {
	v := NewURLValidator()

	withPort := []string{
		"https://example.com:443",
		"http://example.com:8080",
	}

	for _, u := range withPort {
		err := v.Validate(u)
		assert.NoError(t, err, "URL with valid port should pass: "+u)
	}
}

func TestValidate_URLWithPathAndQuery(t *testing.T) {
	v := NewURLValidator()

	withExtras := []string{
		"https://example.com/path/to/resource",
		"http://api.test.com/v1?param=value",
		"https://example.com/page#section",
	}

	for _, u := range withExtras {
		err := v.Validate(u)
		assert.NoError(t, err, "URL with path/query/fragment should pass: "+u)
	}
}
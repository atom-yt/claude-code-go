package feishu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// SignatureVerifier verifies Feishu webhook signatures.
type SignatureVerifier struct {
	verificationToken string
	encryptKey       string
}

// NewSignatureVerifier creates a new signature verifier.
func NewSignatureVerifier(verificationToken, encryptKey string) *SignatureVerifier {
	return &SignatureVerifier{
		verificationToken: verificationToken,
		encryptKey:       encryptKey,
	}
}

// VerifyToken verifies the token from Feishu URL verification challenge.
func (v *SignatureVerifier) VerifyToken(token string) bool {
	return token == v.verificationToken
}

// VerifySignature verifies the HMAC-SHA256 signature of a webhook request.
func (v *SignatureVerifier) VerifySignature(req *http.Request) (bool, error) {
	if v.encryptKey == "" {
		// No encryption configured, skip verification
		return true, nil
	}

	// Get timestamp and signature from headers
	timestamp := req.Header.Get("X-Lark-Request-Timestamp")
	nonce := req.Header.Get("X-Lark-Request-Nonce")
	signature := req.Header.Get("X-Lark-Signature")

	if timestamp == "" || nonce == "" || signature == "" {
		return false, ErrMissingSignatureHeaders
	}

	// Read request body
	if req.Body == nil {
		return false, ErrMissingRequestBody
	}

	body := make([]byte, 1024) // Read up to 1KB for signature
	n, err := req.Body.Read(body)
	if err != nil {
		return false, fmt.Errorf("failed to read request body: %w", err)
	}
	body = body[:n]

	// Verify timestamp (within 5 minutes)
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	if time.Since(ts) > 5*time.Minute {
		return false, ErrTimestampExpired
	}

	// Compute expected signature
	expectedSig := v.computeSignature(timestamp, nonce, body)

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSig)), nil
}

// computeSignature computes HMAC-SHA256 signature.
func (v *SignatureVerifier) computeSignature(timestamp, nonce string, body []byte) string {
	// Concatenate: timestamp + nonce + body
	data := strings.Join([]string{timestamp, nonce, string(body)}, "")

	// Create HMAC-SHA256
	h := hmac.New(sha256.New, []byte(v.encryptKey))
	h.Write([]byte(data))

	// Base64 encode
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyEvent verifies the signature of an event payload.
func (v *SignatureVerifier) VerifyEvent(event *Event, signature string) bool {
	if v.encryptKey == "" {
		return true
	}

	// Serialize event to JSON (sorted keys for consistency)
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return false
	}

	// Compute signature
	h := hmac.New(sha256.New, []byte(v.encryptKey))
	h.Write(eventJSON)

	expectedSig := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// SignPayload creates a signature for outgoing requests.
func (v *SignatureVerifier) SignPayload(timestamp string, payload []byte) string {
	if v.encryptKey == "" {
		return ""
	}

	// Sort keys for consistent signature
	var sortedKeys []string
	if m, ok := payloadToMap(payload); ok {
		for k := range m {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Build string in key order
		var parts []string
		for _, k := range sortedKeys {
			parts = append(parts, fmt.Sprintf("%s=%v", k, m[k]))
		}
		payload = []byte(strings.Join(parts, "&"))
	}

	h := hmac.New(sha256.New, []byte(v.encryptKey))
	h.Write([]byte(timestamp))
	h.Write(payload)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// payloadToMap converts JSON payload to sorted map.
func payloadToMap(payload []byte) (map[string]any, bool) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return nil, false
	}
	return m, true
}

// GetSignedHeaders returns the signature headers for outgoing requests.
func (v *SignatureVerifier) GetSignedHeaders(payload []byte) map[string]string {
	if v.encryptKey == "" {
		return nil
	}

	timestamp := time.Now().Format(time.RFC3339)
	nonce := generateNonce()
	signature := v.computeSignature(timestamp, nonce, payload)

	return map[string]string{
		"X-Lark-Request-Timestamp": timestamp,
		"X-Lark-Request-Nonce":     nonce,
		"X-Lark-Signature":          signature,
	}
}

// generateNonce generates a random nonce for signature.
func generateNonce() string {
	// Simple nonce implementation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Signature verification errors
var (
	ErrMissingSignatureHeaders = fmt.Errorf("missing signature headers")
	ErrMissingRequestBody    = fmt.Errorf("missing request body")
	ErrTimestampExpired      = fmt.Errorf("timestamp expired")
	ErrInvalidSignature      = fmt.Errorf("invalid signature")
)
package xss

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kokuroshesh/bugvay/internal/httpclient"
	"github.com/kokuroshesh/bugvay/internal/scanners"
)

type XSSScanner struct {
	client *httpclient.Scanner
}

func New(client *httpclient.Scanner) *XSSScanner {

	return &XSSScanner{client: client}
}

func (s *XSSScanner) Name() string {
	return "xss"
}

var xssPayloads = []string{
	"<script>alert(1)</script>",
	"<img src=x onerror=alert(1)>",
	"'><script>alert(1)</script>",
	"\"><script>alert(1)</script>",
	"javascript:alert(1)",
	"<svg/onload=alert(1)>",
	"<iframe src=javascript:alert(1)>",
}

func (s *XSSScanner) Scan(ctx context.Context, input *scanners.ScanInput) (*scanners.ScanResult, error) {
	u, err := url.Parse(input.URL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	params := u.Query()
	if len(params) == 0 {
		return &scanners.ScanResult{Vulnerable: false}, nil
	}

	// Test each parameter with each payload
	for param := range params {
		for _, payload := range xssPayloads {
			testURL := buildTestURL(u, param, payload)

			req, _ := http.NewRequestWithContext(ctx, "GET", testURL, nil)
			status, body, err := s.client.DoRequest(ctx, req)
			if err != nil {
				continue
			}

			// Check for reflection (both raw and HTML-decoded)
			bodyStr := string(body)
			bodyLower := strings.ToLower(bodyStr)

			// Skip if payload is HTML-encoded (likely not exploitable)
			encodedPayload := strings.ReplaceAll(payload, "<", "&lt;")
			encodedPayload = strings.ReplaceAll(encodedPayload, ">", "&gt;")
			encodedPayload = strings.ReplaceAll(encodedPayload, "\"", "&quot;")

			isRawReflected := strings.Contains(bodyStr, payload)
			isEncodedOnly := !isRawReflected && strings.Contains(bodyStr, encodedPayload)

			// Skip if only encoded version found (safe)
			if isEncodedOnly {
				continue
			}

			// Check if in executable context (not in comments or JSON)
			if isRawReflected {
				// Skip if inside HTML comment
				if strings.Contains(bodyStr, "<!--"+payload) || strings.Contains(bodyStr, payload+"-->") {
					continue
				}

				// Skip if inside JSON response (Content-Type check would be better)
				if strings.HasPrefix(strings.TrimSpace(bodyStr), "{") && strings.Contains(bodyLower, "application/json") {
					continue
				}

				return &scanners.ScanResult{
					Vulnerable: true,
					Severity:   "medium",
					CWE:        79,
					Evidence: map[string]interface{}{
						"param":     param,
						"payload":   payload,
						"url":       testURL,
						"reflected": true,
					},
					Proof: fmt.Sprintf("XSS payload reflected in response:\nURL: %s\nPayload: %s\nStatus: %d",
						testURL, payload, status),
					Confidence: 0.8,
				}, nil
			}
		}
	}

	return &scanners.ScanResult{Vulnerable: false}, nil
}

func buildTestURL(u *url.URL, param, payload string) string {
	q := u.Query()
	q.Set(param, payload)
	u.RawQuery = q.Encode()
	return u.String()
}

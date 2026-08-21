// Package hibp implements a privacy-preserving client for the Have I Been
// Pwned (HIBP) Pwned Passwords range API.
//
// It uses the k-anonymity protocol: only the first 5 characters of the
// uppercase SHA-1 hash of a password are sent to the API, so the password
// itself (and its full hash) never leave the server. The API is free and
// requires no API key.
package hibp

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/swayrider/swlib/logger"
)

const (
	// DefaultBaseURL is the Pwned Passwords range endpoint base URL.
	DefaultBaseURL = "https://api.pwnedpasswords.com"

	// userAgent is sent with every request; the HIBP API rejects requests
	// without a User-Agent header with HTTP 403.
	userAgent = "swayrider-authservice"
)

// Client checks passwords against the Pwned Passwords API using the
// k-anonymity range protocol.
type Client struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
	minCount   int
	l          *log.Logger
}

// New builds a Client. When enabled is false the client short-circuits every
// check to (false, 0, nil) without making a network call, so deployments can
// switch the feature off entirely without code changes.
func New(enabled bool, timeout time.Duration, minCount int, l *log.Logger) *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		enabled:  enabled,
		minCount: minCount,
		l:        l.Derive(log.WithComponent("hibp")),
	}
}

// IsBreached reports whether password has appeared in a known breach at least
// minCount times, along with the exact count.
//
// Only the first 5 characters of the uppercase SHA-1 hash leave the server;
// the API returns every matching suffix and the client compares locally. Any
// API error (timeout, rate limit, non-200 status) is returned to the caller,
// which is expected to fail open (allow the password) so an HIBP outage never
// blocks users.
func (c *Client) IsBreached(ctx context.Context, password string) (bool, int, error) {
	if !c.enabled {
		return false, 0, nil
	}

	hash := fmt.Sprintf("%X", sha1.Sum([]byte(password)))
	prefix := hash[:5]
	suffix := hash[5:]

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.baseURL+"/range/"+prefix, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	// Add-Padding appends dummy entries so the response size does not reveal
	// whether the suffix matched (defeats response-size side channels).
	req.Header.Set("Add-Padding", "true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.l.Warnf("pwned passwords API request failed: %v", err)
		return false, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("pwned passwords API returned status %d", resp.StatusCode)
		c.l.Warnf("%v", err)
		return false, 0, err
	}

	count, err := c.countForSuffix(resp.Body, suffix)
	if err != nil {
		c.l.Warnf("failed to parse pwned passwords response: %v", err)
		return false, 0, err
	}

	return count >= c.minCount, count, nil
}

// countForSuffix scans a range response body (one "SUFFIX:COUNT" pair per
// line, possibly interleaved with padding entries) and returns the count for
// the given suffix, or 0 when the suffix does not appear. Suffix comparison is
// case-insensitive.
func (c *Client) countForSuffix(body io.Reader, suffix string) (int, error) {
	scanner := bufio.NewScanner(body)
	// Range responses contain thousands of lines; grow the scanner buffer
	// well beyond the 64 KiB default.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		// Padding entries (Add-Padding) may be arbitrary suffixes without a
		// count, so a line without a colon is simply skipped.
		colon := strings.LastIndex(line, ":")
		if colon == -1 {
			continue
		}
		if !strings.EqualFold(line[:colon], suffix) {
			continue
		}
		return strconv.Atoi(line[colon+1:])
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}

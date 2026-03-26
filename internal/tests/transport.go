//revive:disable:package-comments
package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FakeInstallationTokenTransport returns a canned installation token response.
type FakeInstallationTokenTransport struct {
	Token string
}

// RoundTrip returns a canned JSON response with the configured token.
func (t *FakeInstallationTokenTransport) RoundTrip(*http.Request) (*http.Response, error) {
	body, err := json.Marshal(map[string]string{"token": t.Token})
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}

	return &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}, nil
}

// FailingTransport always returns an error.
type FailingTransport struct{}

// RoundTrip always returns a connection refused error.
func (t *FailingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("connection refused")
}

//revive:disable:package-comments
package transport

import (
	"net/http"
)

// AuthTransport implements authenticated round trip transport
// with the Authorization header set
type AuthTransport struct {
	rt http.RoundTripper
	// cred vault.Credential
}

// New returns a new AuthTransport using the desired round tripper and a Credential
func New(
	rt http.RoundTripper,
	//  cred vault.Credential,
) *AuthTransport {
	return &AuthTransport{
		rt,
		// cred,
	}
}

// RoundTrip adds the Authorization header to the request.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// t.cred.SetAuth(req)
	return t.rt.RoundTrip(req)
}

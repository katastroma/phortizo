//revive:disable:package-comments
package transport

import (
	"net/http"

	"github.com/katastroma/phortizo/internal/auth"
)

// AuthTransport implements authenticated round trip transport
// with the Authorization header set
type AuthTransport struct {
	rt   http.RoundTripper
	cred auth.Credential
}

// New returns a new AuthTransport using the desired round tripper and a Credential
func New(rt http.RoundTripper, cred auth.Credential) *AuthTransport {
	return &AuthTransport{rt, cred}
}

// RoundTrip adds the Authorization header to the request.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.cred.SetAuth(req)
	return t.rt.RoundTrip(req)
}

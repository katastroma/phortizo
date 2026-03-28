//revive:disable:package-comments
package tests

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/object"
)

// MockAuthenticator mocks credential.Authenticator for testing
type MockAuthenticator struct {
	authMethod transport.AuthMethod
	Err        error
}

// Authenticate mocks authenticating
func (m *MockAuthenticator) Authenticate(context.Context, *http.Client) (transport.AuthMethod, error) {
	return m.authMethod, m.Err
}

// MockCredentialReader mocks credential.Reader for testing
type MockCredentialReader struct {
	Cred credential.Authenticator
	Err  error
}

// Get mocks getting a credential for a mock credential reader for testing
func (m *MockCredentialReader) Get(
	context.Context, object.Store[[]byte], string,
) (credential.Authenticator, error) {
	return m.Cred, m.Err
}

package credential_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/katastroma/phortizo/internal/credential"
	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/tests"
)

func TestResolveFunc_EmptySecret(t *testing.T) {
	newStore := func(_ string) object.Store[[]byte] { return tests.NewMockStore[[]byte]() }
	resolve := credential.ResolveFunc(&tests.MockCredentialReader{}, http.DefaultClient, newStore)

	auth, err := resolve(t.Context(), "tenant-a", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auth != nil {
		t.Errorf("expected nil auth for empty secret, got %v", auth)
	}
}

func TestResolveFunc(t *testing.T) {
	reader := &tests.MockCredentialReader{Cred: &tests.MockAuthenticator{}}
	newStore := func(_ string) object.Store[[]byte] { return tests.NewMockStore[[]byte]() }
	resolve := credential.ResolveFunc(reader, http.DefaultClient, newStore)

	auth, err := resolve(t.Context(), "tenant-a", "my-cred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auth != nil {
		t.Errorf("expected nil auth from mock, got %v", auth)
	}
}

func TestResolveFunc_ReaderError(t *testing.T) {
	reader := &tests.MockCredentialReader{Err: fmt.Errorf("not found")}
	newStore := func(_ string) object.Store[[]byte] { return tests.NewMockStore[[]byte]() }
	resolve := credential.ResolveFunc(reader, http.DefaultClient, newStore)

	if _, err := resolve(t.Context(), "tenant-a", "my-cred"); err == nil {
		t.Fatal("expected error from failing reader")
	}
}

func TestResolveFunc_AuthenticateError(t *testing.T) {
	reader := &tests.MockCredentialReader{
		Cred: &tests.MockAuthenticator{Err: fmt.Errorf("auth failed")},
	}
	newStore := func(_ string) object.Store[[]byte] { return tests.NewMockStore[[]byte]() }
	resolve := credential.ResolveFunc(reader, http.DefaultClient, newStore)

	if _, err := resolve(t.Context(), "tenant-a", "my-cred"); err == nil {
		t.Fatal("expected error from failing authenticate")
	}
}

func TestResolveFunc_PassesNamespace(t *testing.T) {
	var calledNS string
	reader := &tests.MockCredentialReader{Cred: &tests.MockAuthenticator{}}
	newStore := func(ns string) object.Store[[]byte] {
		calledNS = ns
		return tests.NewMockStore[[]byte]()
	}
	resolve := credential.ResolveFunc(reader, http.DefaultClient, newStore)

	if _, err := resolve(t.Context(), "tenant-a", "my-cred"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calledNS != "tenant-a" {
		t.Errorf("namespace = %q, want %q", calledNS, "tenant-a")
	}
}

package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func sign(payload, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func validPayload() []byte {
	return []byte(`{
		"ref": "refs/heads/main",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [{"added": ["deploy/values.yaml"], "removed": [], "modified": []}]
	}`)
}

func TestServeHTTP_MissingID(t *testing.T) {
	handler := New(slog.Default(), nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_RegistrationNotFound(t *testing.T) {
	handler := New(slog.Default(), nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", nil)
	req.SetPathValue("registration_id", "missing")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_SignatureFailure(t *testing.T) {
	handler := New(slog.Default(), nil)

	body := validPayload()
	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", bytes.NewReader(body))
	req.SetPathValue("registration_id", "reg-1")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_InvalidPayload(t *testing.T) {
	handler := New(slog.Default(), nil)

	body := []byte("not json")
	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", bytes.NewReader(body))
	req.SetPathValue("registration_id", "reg-1")
	req.Header.Set("X-Hub-Signature-256", sign(body, []byte("test-secret")))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_NoMatch(t *testing.T) {
	handler := New(slog.Default(), nil)

	body := []byte(`{
		"ref": "refs/heads/develop",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [{"added": ["src/main.go"], "removed": [], "modified": []}]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", bytes.NewReader(body))
	req.SetPathValue("registration_id", "reg-1")
	req.Header.Set("X-Hub-Signature-256", sign(body, []byte("test-secret")))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServeHTTP_Match(t *testing.T) {
	handler := New(slog.Default(), nil)

	body := validPayload()
	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", bytes.NewReader(body))
	req.SetPathValue("registration_id", "reg-1")
	req.Header.Set("X-Hub-Signature-256", sign(body, []byte("test-secret")))
	req.Header.Set("X-GitHub-Delivery", "delivery-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestServeHTTP_BodyReadError(t *testing.T) {
	handler := New(slog.Default(), nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook/{id}", &errorReader{})
	req.SetPathValue("registration_id", "reg-1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

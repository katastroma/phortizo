package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/k8s/secret"
	"github.com/katastroma/phortizo/internal/match"
)

const testNamespace = "tenant-a"
const testSecret = "test-webhook-secret"

func sign(payload []byte, webhookSecret string) string {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
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

func webhookSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.WebhookSecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			secret.WebhookSecretKey: []byte(testSecret),
		},
	}
}

func watchTargetConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wt-1",
			Namespace: testNamespace,
			Labels:    map[string]string{configmap.TypeLabel: configmap.WatchTargetType},
		},
		Data: map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		},
	}
}

type mockHandler struct {
	calls []match.Result
}

func (m *mockHandler) HandleMatch(_ context.Context, _ string, r match.Result) {
	m.calls = append(m.calls, r)
}

func TestServeHTTP_MissingNamespace(t *testing.T) {
	handler := New(slog.Default(), nil, fake.NewSimpleClientset())

	req := httptest.NewRequest(http.MethodPost, "/webhook/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_WatchTargetsError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor("list", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("list denied")
	})
	handler := New(slog.Default(), nil, k8s)

	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", nil)
	req.SetPathValue("namespace", testNamespace)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_WebhookSecretNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetConfigMap())
	handler := New(slog.Default(), nil, k8s)

	body := validPayload()
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_BodyReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := New(slog.Default(), nil, k8s)

	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", &errorReader{})
	req.SetPathValue("namespace", testNamespace)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_SignatureFailure(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := New(slog.Default(), nil, k8s)

	body := validPayload()
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_InvalidPayload(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := New(slog.Default(), nil, k8s)

	body := []byte("not json")
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("X-Hub-Signature-256", sign(body, testSecret))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_NoMatch(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := New(slog.Default(), nil, k8s)

	body := []byte(`{
		"ref": "refs/heads/develop",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [{"added": ["src/main.go"], "removed": [], "modified": []}]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("X-Hub-Signature-256", sign(body, testSecret))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServeHTTP_Match(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	runner := &mockHandler{}
	handler := New(slog.Default(), runner, k8s)

	body := validPayload()
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("X-Hub-Signature-256", sign(body, testSecret))
	req.Header.Set("X-GitHub-Delivery", "delivery-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 HandleMatch call, got %d", len(runner.calls))
	}

	if runner.calls[0].Target.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", runner.calls[0].Target.RepoURL, "https://github.com/acme/app.git")
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

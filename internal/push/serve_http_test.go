package push_test

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

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5/plumbing/transport"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/push"
	"github.com/katastroma/phortizo/internal/source"
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

func pushRequest(body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", bytes.NewReader(body))
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", sign(body, testSecret))
	return req
}

func webhookSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      push.SecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			push.SecretKey: []byte(testSecret),
		},
	}
}

func watchTargetConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wt-1",
			Namespace: testNamespace,
			Labels:    map[string]string{object.TypeLabel: source.TypeLabel},
		},
		Data: map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		},
	}
}

func helmFS(t *testing.T) billy.Filesystem {
	t.Helper()

	fs := memfs.New()
	if err := util.WriteFile(fs, "deploy/Chart.yaml", []byte("name: test"), 0o644); err != nil {
		t.Fatalf("writing Chart.yaml: %v", err)
	}

	return fs
}

func noopHandler(t *testing.T) (
	func(context.Context, string, string, string, int) error,
	func(context.Context, string, string) (transport.AuthMethod, error),
	func(context.Context, string, string, transport.AuthMethod) (billy.Filesystem, error),
	func(billy.Filesystem, string) (string, error),
	func(context.Context, string, string, string) (bool, error),
	func(context.Context, billy.Filesystem, string, string) error,
) {
	t.Helper()

	fs := helmFS(t)

	return func(_ context.Context, _, _, _ string, _ int) error { return nil },
		func(_ context.Context, _, _ string) (transport.AuthMethod, error) { return nil, nil },
		func(_ context.Context, _, _ string, _ transport.AuthMethod) (billy.Filesystem, error) { return fs, nil },
		func(_ billy.Filesystem, _ string) (string, error) { return "helm-renderer:8080", nil },
		func(_ context.Context, _, _, _ string) (bool, error) { return true, nil },
		func(_ context.Context, _ billy.Filesystem, _, _ string) error { return nil }
}

func newTestHandler(t *testing.T, k8s *fake.Clientset) *push.Handler {
	t.Helper()

	acquire, resolve, clone, lookup, verify, stream := noopHandler(t)
	return push.New(slog.Default(), k8s, acquire, resolve, clone, lookup, verify, stream)
}

func TestServeHTTP_MissingNamespace(t *testing.T) {
	handler := newTestHandler(t, fake.NewSimpleClientset())

	req := httptest.NewRequest(http.MethodPost, "/webhook/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_WatchTargetsError(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret())
	k8s.PrependReactor(
		"list", "configmaps",
		func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, fmt.Errorf("list denied")
		},
	)
	handler := newTestHandler(t, k8s)

	req := pushRequest(validPayload())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_WebhookSecretNotFound(t *testing.T) {
	k8s := fake.NewSimpleClientset(watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	req := pushRequest(validPayload())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_WebhookSecretMissingKey(t *testing.T) {
	secretWithoutKey := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      push.SecretName,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{"wrong-key": []byte("value")},
	}

	k8s := fake.NewSimpleClientset(secretWithoutKey)
	handler := newTestHandler(t, k8s)

	req := pushRequest(validPayload())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeHTTP_ValidationFailure_BodyReadError(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	req := httptest.NewRequest(http.MethodPost, "/webhook/{namespace}", &errorReader{})
	req.SetPathValue("namespace", testNamespace)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_ValidationFailure_BadSignature(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	req := pushRequest(validPayload())
	req.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestServeHTTP_InvalidPayload(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	body := []byte("not json")
	req := pushRequest(body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeHTTP_NoMatch(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	body := []byte(`{
		"ref": "refs/heads/develop",
		"repository": {"clone_url": "https://github.com/acme/app.git"},
		"commits": [{"added": ["src/main.go"], "removed": [], "modified": []}]
	}`)
	req := pushRequest(body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServeHTTP_NonPushEvent(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())
	handler := newTestHandler(t, k8s)

	body := []byte(`{"action": "opened"}`)
	req := pushRequest(body)
	req.Header.Set("X-GitHub-Event", "issues")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServeHTTP_Match(t *testing.T) {
	k8s := fake.NewSimpleClientset(webhookSecret(), watchTargetConfigMap())

	var processed int
	acquire, resolve, clone, lookup, verify, _ := noopHandler(t)
	countingStream := func(_ context.Context, _ billy.Filesystem, _, _ string) error {
		processed++
		return nil
	}

	handler := push.New(slog.Default(), k8s, acquire, resolve, clone, lookup, verify, countingStream)

	body := validPayload()
	req := pushRequest(body)
	req.Header.Set("X-GitHub-Delivery", "delivery-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	if processed != 1 {
		t.Errorf("expected 1 Process call, got %d", processed)
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

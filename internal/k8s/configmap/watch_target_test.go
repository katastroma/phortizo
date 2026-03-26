package configmap_test

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/katastroma/phortizo/internal/k8s/configmap"
	"github.com/katastroma/phortizo/internal/registration"
)

func watchTargetConfigMap(name, namespace, credentialSecret string, data map[string]string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{configmap.TypeLabel: configmap.WatchTargetType},
		},
		Data: data,
	}

	if credentialSecret != "" {
		cm.Annotations = map[string]string{configmap.CredentialSecretAnnotation: credentialSecret}
	}

	return cm
}

func TestListWatchTargets(t *testing.T) {
	k8s := fake.NewSimpleClientset(
		watchTargetConfigMap("wt-1", "tenant-a", "my-cred", map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		}),
		watchTargetConfigMap("wt-2", "tenant-a", "", map[string]string{
			"repo-url": "https://github.com/acme/other.git",
			"ref":      "refs/heads/main",
			"path":     "k8s/",
		}),
	)

	results, err := configmap.ListWatchTargets(t.Context(), k8s, "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	var withCred, withoutCred registration.WatchTarget
	for _, r := range results {
		if r.CredentialSecret != "" {
			withCred = r
		} else {
			withoutCred = r
		}
	}

	if withCred.RepoURL != "https://github.com/acme/app.git" {
		t.Errorf("RepoURL = %q, want %q", withCred.RepoURL, "https://github.com/acme/app.git")
	}

	if withCred.CredentialSecret != "my-cred" {
		t.Errorf("CredentialSecret = %q, want %q", withCred.CredentialSecret, "my-cred")
	}

	if withoutCred.Path != "k8s/" {
		t.Errorf("Path = %q, want %q", withoutCred.Path, "k8s/")
	}

	if withoutCred.CredentialSecret != "" {
		t.Errorf("CredentialSecret = %q, want empty", withoutCred.CredentialSecret)
	}
}

func TestListWatchTargets_WithOverrides(t *testing.T) {
	k8s := fake.NewSimpleClientset(
		watchTargetConfigMap("wt-1", "tenant-a", "", map[string]string{
			"repo-url":  "https://github.com/acme/app.git",
			"ref":       "refs/heads/main",
			"path":      "deploy/",
			"overrides": "image:\n  tag: v1.2.3\n",
		}),
	)

	results, err := configmap.ListWatchTargets(t.Context(), k8s, "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if string(results[0].Overrides) != "image:\n  tag: v1.2.3\n" {
		t.Errorf("Overrides = %q, want %q", results[0].Overrides, "image:\n  tag: v1.2.3\n")
	}
}

func TestListWatchTargets_Empty(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	results, err := configmap.ListWatchTargets(t.Context(), k8s, "tenant-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestListWatchTargets_ListError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor("list", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("list denied")
	})

	_, err := configmap.ListWatchTargets(t.Context(), k8s, "tenant-a")
	if err == nil {
		t.Fatal("expected error from failing list")
	}
}

func TestListWatchTargets_DeserializeError(t *testing.T) {
	k8s := fake.NewSimpleClientset(
		watchTargetConfigMap("wt-bad", "tenant-a", "", map[string]string{
			"ref":  "refs/heads/main",
			"path": "deploy/",
		}),
	)

	_, err := configmap.ListWatchTargets(t.Context(), k8s, "tenant-a")
	if err == nil {
		t.Fatal("expected error for missing repo-url")
	}
}

func TestPutWatchTarget_Create(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	target := registration.WatchTarget{
		RepoURL:          "https://github.com/acme/app.git",
		Ref:              "refs/heads/main",
		Path:             "deploy/",
		CredentialSecret: "my-cred",
	}

	err := configmap.PutWatchTarget(t.Context(), k8s, "tenant-a", "wt-1", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm, err := k8s.CoreV1().ConfigMaps("tenant-a").Get(t.Context(), "wt-1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("configmap not created: %v", err)
	}

	if cm.Data["repo-url"] != "https://github.com/acme/app.git" {
		t.Errorf("repo-url = %q, want %q", cm.Data["repo-url"], "https://github.com/acme/app.git")
	}

	if cm.Labels[configmap.TypeLabel] != configmap.WatchTargetType {
		t.Errorf("label = %q, want %q", cm.Labels[configmap.TypeLabel], configmap.WatchTargetType)
	}

	if cm.Annotations[configmap.CredentialSecretAnnotation] != "my-cred" {
		t.Errorf("annotation = %q, want %q", cm.Annotations[configmap.CredentialSecretAnnotation], "my-cred")
	}
}

func TestPutWatchTarget_CreateNoCredential(t *testing.T) {
	k8s := fake.NewSimpleClientset()

	target := registration.WatchTarget{
		RepoURL: "https://github.com/acme/app.git",
		Ref:     "refs/heads/main",
		Path:    "deploy/",
	}

	err := configmap.PutWatchTarget(t.Context(), k8s, "tenant-a", "wt-1", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm, err := k8s.CoreV1().ConfigMaps("tenant-a").Get(t.Context(), "wt-1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("configmap not created: %v", err)
	}

	if len(cm.Annotations) != 0 {
		t.Errorf("expected no annotations, got %v", cm.Annotations)
	}
}

func TestPutWatchTarget_Update(t *testing.T) {
	k8s := fake.NewSimpleClientset(
		watchTargetConfigMap("wt-1", "tenant-a", "old-cred", map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		}),
	)

	target := registration.WatchTarget{
		RepoURL:          "https://github.com/acme/app.git",
		Ref:              "refs/heads/main",
		Path:             "deploy/prod/",
		CredentialSecret: "new-cred",
	}

	err := configmap.PutWatchTarget(t.Context(), k8s, "tenant-a", "wt-1", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm, err := k8s.CoreV1().ConfigMaps("tenant-a").Get(t.Context(), "wt-1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("reading configmap: %v", err)
	}

	if cm.Data["path"] != "deploy/prod/" {
		t.Errorf("path = %q, want %q", cm.Data["path"], "deploy/prod/")
	}

	if cm.Annotations[configmap.CredentialSecretAnnotation] != "new-cred" {
		t.Errorf("annotation = %q, want %q", cm.Annotations[configmap.CredentialSecretAnnotation], "new-cred")
	}
}

func TestPutWatchTarget_UpdateError(t *testing.T) {
	k8s := fake.NewSimpleClientset(
		watchTargetConfigMap("wt-1", "tenant-a", "", map[string]string{
			"repo-url": "https://github.com/acme/app.git",
			"ref":      "refs/heads/main",
			"path":     "deploy/",
		}),
	)
	k8s.PrependReactor("update", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("update denied")
	})

	target := registration.WatchTarget{RepoURL: "x", Ref: "y", Path: "z"}
	err := configmap.PutWatchTarget(t.Context(), k8s, "tenant-a", "wt-1", target)
	if err == nil {
		t.Fatal("expected error from failing update")
	}
}

func TestPutWatchTarget_CreateError(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	k8s.PrependReactor("create", "configmaps", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("create denied")
	})

	target := registration.WatchTarget{RepoURL: "x", Ref: "y", Path: "z"}
	err := configmap.PutWatchTarget(t.Context(), k8s, "tenant-a", "wt-1", target)
	if err == nil {
		t.Fatal("expected error from failing create")
	}
}

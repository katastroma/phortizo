//revive:disable:package-comments
package configmap

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/lease"
)

// Store implements lease.ObjectStore for ConfigMaps in a namespace.
type Store struct {
	client    kubernetes.Interface
	namespace string
}

// NewStore returns a ConfigMap-backed ObjectStore.
func NewStore(client kubernetes.Interface, namespace string) *Store {
	return &Store{client: client, namespace: namespace}
}

// Get returns a ConfigMap as a lease.Annotatable.
func (s *Store) Get(ctx context.Context, name string) (lease.Annotatable, error) {
	return s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, name, metav1.GetOptions{})
}

// Update persists annotation changes on a ConfigMap.
func (s *Store) Update(ctx context.Context, obj lease.Annotatable) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("expected *corev1.ConfigMap")
	}

	_, err := s.client.CoreV1().ConfigMaps(s.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

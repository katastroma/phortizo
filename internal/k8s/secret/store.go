//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/object"
)

// Store implements object.Store[[]byte] for Secrets in a namespace.
type Store struct {
	client    kubernetes.Interface
	namespace string
}

// NewStore returns a Secret-backed object.Store[[]byte].
func NewStore(client kubernetes.Interface, namespace string) *Store {
	return &Store{client: client, namespace: namespace}
}

// New returns an empty Secret Resource with the given name.
func (s *Store) New(name string) object.Resource[[]byte] {
	return NewResource(name)
}

// Get returns a Secret wrapped as an object.Resource[[]byte].
func (s *Store) Get(ctx context.Context, name string) (object.Resource[[]byte], error) {
	sec, err := s.client.CoreV1().Secrets(s.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &Resource{sec}, nil
}

// List returns Secrets matching the given labels.
func (s *Store) List(ctx context.Context, labels map[string]string) ([]object.Resource[[]byte], error) {
	opts := metav1.ListOptions{}
	if len(labels) > 0 {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: labels})
	}

	list, err := s.client.CoreV1().Secrets(s.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	results := make([]object.Resource[[]byte], 0, len(list.Items))
	for _, sec := range list.Items {
		results = append(results, &Resource{sec.DeepCopy()})
	}

	return results, nil
}

// Update persists annotation changes on a Secret.
func (s *Store) Update(ctx context.Context, obj object.Annotatable) error {
	r, ok := obj.(*Resource)
	if !ok {
		return fmt.Errorf("expected *secret.Resource")
	}

	_, err := s.client.CoreV1().Secrets(s.namespace).Update(ctx, r.Secret, metav1.UpdateOptions{})
	return err
}

// Put creates or updates a Secret from a Resource.
func (s *Store) Put(ctx context.Context, obj object.Resource[[]byte]) error {
	r, ok := obj.(*Resource)
	if !ok {
		return fmt.Errorf("expected *secret.Resource")
	}

	secrets := s.client.CoreV1().Secrets(s.namespace)

	existing, err := secrets.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err == nil {
		existing.Data = r.GetData()
		existing.Labels = r.GetLabels()
		existing.Annotations = r.GetAnnotations()
		_, err = secrets.Update(ctx, existing, metav1.UpdateOptions{})
		return err
	}

	r.Secret.Namespace = s.namespace
	_, err = secrets.Create(ctx, r.Secret, metav1.CreateOptions{})
	return err
}

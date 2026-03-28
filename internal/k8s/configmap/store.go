//revive:disable:package-comments
package configmap

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/katastroma/phortizo/internal/object"
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

// New returns an empty ConfigMap Resource with the given name.
func (s *Store) New(name string) object.Resource {
	return NewResource(name)
}

// Get returns a ConfigMap wrapped as an object.Resource.
func (s *Store) Get(ctx context.Context, name string) (object.Resource, error) {
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &Resource{cm}, nil
}

// List returns ConfigMaps matching the given labels.
func (s *Store) List(ctx context.Context, labels map[string]string) ([]object.Resource, error) {
	opts := metav1.ListOptions{}
	if len(labels) > 0 {
		opts.LabelSelector = metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: labels})
	}

	list, err := s.client.CoreV1().ConfigMaps(s.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	results := make([]object.Resource, 0, len(list.Items))
	for _, cm := range list.Items {
		results = append(results, &Resource{cm.DeepCopy()})
	}

	return results, nil
}

// Update persists annotation changes on a ConfigMap.
func (s *Store) Update(ctx context.Context, obj object.Annotatable) error {
	r, ok := obj.(*Resource)
	if !ok {
		return fmt.Errorf("expected *configmap.Resource")
	}

	_, err := s.client.CoreV1().ConfigMaps(s.namespace).Update(ctx, r.ConfigMap, metav1.UpdateOptions{})
	return err
}

// Put creates or updates a ConfigMap from a Resource.
func (s *Store) Put(ctx context.Context, obj object.Resource) error {
	r, ok := obj.(*Resource)
	if !ok {
		return fmt.Errorf("expected *configmap.Resource")
	}

	configmaps := s.client.CoreV1().ConfigMaps(s.namespace)

	existing, err := configmaps.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err == nil {
		existing.Data = r.GetData()
		existing.Labels = r.GetLabels()
		existing.Annotations = r.GetAnnotations()
		_, err = configmaps.Update(ctx, existing, metav1.UpdateOptions{})
		return err
	}

	r.ConfigMap.Namespace = s.namespace
	_, err = configmaps.Create(ctx, r.ConfigMap, metav1.CreateOptions{})
	return err
}

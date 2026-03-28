//revive:disable:package-comments
package configmap

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource wraps a ConfigMap to satisfy object.Resource.
type Resource struct {
	*corev1.ConfigMap
}

// NewResource returns a Resource with the given name.
func NewResource(name string) *Resource {
	return &Resource{&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name}}}
}

// GetData returns the ConfigMap data.
func (r *Resource) GetData() map[string]string {
	return r.Data
}

// SetData sets the ConfigMap data.
func (r *Resource) SetData(data map[string]string) {
	r.Data = data
}

//revive:disable:package-comments
package secret

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource wraps a Secret to satisfy object.Resource[[]byte].
type Resource struct {
	*corev1.Secret
}

// NewResource returns a Resource with the given name.
func NewResource(name string) *Resource {
	return &Resource{&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name}}}
}

// GetData returns the Secret data.
func (r *Resource) GetData() map[string][]byte {
	return r.Data
}

// SetData sets the Secret data.
func (r *Resource) SetData(data map[string][]byte) {
	r.Data = data
}

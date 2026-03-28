//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Writer writes Secrets to tenant namespaces.
type Writer struct {
	client kubernetes.Interface
}

// NewWriter returns a Writer.
func NewWriter(client kubernetes.Interface) *Writer {
	return &Writer{client: client}
}

// Put creates or updates a Secret in the given namespace.
func (w *Writer) Put(ctx context.Context, namespace, name string, data map[string][]byte) error {
	secrets := w.client.CoreV1().Secrets(namespace)

	existing, err := secrets.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		existing.Data = data
		_, err = secrets.Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("updating secret %s/%s: %w", namespace, name, err)
		}

		return nil
	}

	_, err = secrets.Create(
		ctx,
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}, Data: data},
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("creating secret %s/%s: %w", namespace, name, err)
	}

	return nil
}

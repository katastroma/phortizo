//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Read returns the full data map from a Secret.
func Read(ctx context.Context, client kubernetes.Interface, namespace, name string) (map[string][]byte, error) {
	s, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading secret %s/%s: %w", namespace, name, err)
	}

	return s.Data, nil
}

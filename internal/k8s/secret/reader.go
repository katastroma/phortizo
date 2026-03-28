//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Read reads the webhook HMAC secret from a Secret in the tenant namespace.
func Read(ctx context.Context, client kubernetes.Interface, namespace, name, key string) ([]byte, error) {
	s, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading secret %s/%s: %w", namespace, name, err)
	}

	secret, ok := s.Data[key]
	if !ok {
		return nil, fmt.Errorf("missing key %q in secret %s/%s", key, namespace, name)
	}

	return secret, nil
}

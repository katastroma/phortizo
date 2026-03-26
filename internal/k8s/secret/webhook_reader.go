//revive:disable:package-comments
package secret

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/onboarding"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ReadWebhookSecret reads the webhook HMAC secret from a Secret in the tenant namespace.
func ReadWebhookSecret(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
) (onboarding.WebhookSecret, error) {
	s, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("reading secret %s/%s: %w", namespace, name, err)
	}

	secret, ok := s.Data[WebhookSecretKey]
	if !ok {
		return nil, fmt.Errorf("missing key %q in secret %s/%s", WebhookSecretKey, namespace, name)
	}

	return onboarding.WebhookSecret(secret), nil
}

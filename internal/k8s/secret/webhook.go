//revive:disable:package-comments
package secret

const (
	// WebhookSecretKey is the data key for the HMAC webhook secret.
	WebhookSecretKey = "webhook-secret"

	// WebhookSecretName is the name of the k8s Secret in the tenant namespace
	WebhookSecretName = "github-webhook-secret"
)

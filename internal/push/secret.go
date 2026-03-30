//revive:disable:package-comments
package push

const (
	// SecretKey is the data key for the HMAC webhook secret.
	SecretKey = "webhook-secret"

	// SecretName is the name of the k8s Secret in the tenant namespace
	SecretName = "github-webhook-secret"
)

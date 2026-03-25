//revive:disable:package-comments
package secret

// TODO Implement vault.Withdrawer backed by Kubernetes Secrets.
// Reads all secret material from Secrets in tenant namespaces:
// repo credentials (auth.Credential) and webhook secrets.

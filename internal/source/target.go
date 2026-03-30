//revive:disable:package-comments
package source

import "fmt"

const (
	// TypeLabel is the label value for source target ConfigMaps.
	TypeLabel = "source-target"

	// CredentialSecretAnnotation is the annotation key pointing to the
	// credential Secret name in the same namespace.
	CredentialSecretAnnotation = "katastroma.org/credential-secret"
)

// Target represents a tracked location within a repository.
type Target struct {
	RepoURL   string
	Ref       string
	Path      string
	Overrides []byte

	// Name is the ConfigMap resource name in the tenant namespace. Set by the
	// k8s/configmap package during read, used for lease operations.
	Name string

	// CredentialSecret is the name of the k8s Secret holding repo credentials.
	// Empty for public repositories. Stored as a ConfigMap annotation
	// (katastroma.org/credential-secret) rather than a data field, and set by
	// the k8s/configmap package during read/write.
	CredentialSecret string
}

// TargetFromResourceData deserializes a SourceTarget from ConfigMap data.
func TargetFromResourceData(data map[string]string) (Target, error) {
	repoURL, ok := data["repo-url"]
	if !ok {
		return Target{}, fmt.Errorf("missing key %q", "repo-url")
	}

	ref, ok := data["ref"]
	if !ok {
		return Target{}, fmt.Errorf("missing key %q", "ref")
	}

	path, ok := data["path"]
	if !ok {
		return Target{}, fmt.Errorf("missing key %q", "path")
	}

	var overrides []byte
	if raw, ok := data["overrides"]; ok {
		overrides = []byte(raw)
	}

	return Target{RepoURL: repoURL, Ref: ref, Path: path, Overrides: overrides}, nil
}

// MarshalData serializes the Target to Resource data.
func (t Target) MarshalData() map[string]string {
	data := map[string]string{"repo-url": t.RepoURL, "ref": t.Ref, "path": t.Path}

	if len(t.Overrides) > 0 {
		data["overrides"] = string(t.Overrides)
	}

	return data
}

//revive:disable:package-comments
package object

// Annotatable is something with annotations.
type Annotatable interface {
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

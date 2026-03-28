//revive:disable:package-comments
package object

// Resource is a named object with annotations, labels, and data.
type Resource[T string | []byte] interface {
	Annotatable
	GetName() string
	SetName(string)
	GetLabels() map[string]string
	SetLabels(map[string]string)
	GetData() map[string]T
	SetData(map[string]T)
}

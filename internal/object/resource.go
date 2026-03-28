//revive:disable:package-comments
package object

// Resource is a named object with annotations, labels, and data.
type Resource interface {
	Annotatable
	GetName() string
	SetName(string)
	GetLabels() map[string]string
	SetLabels(map[string]string)
	GetData() map[string]string
	SetData(map[string]string)
}

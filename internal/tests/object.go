//revive:disable:package-comments
package tests

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/object"
)

// MockObject is an Object that can be stored for mocking.
type MockObject[T string | []byte] struct {
	name        string
	annotations map[string]string
	labels      map[string]string
	data        map[string]T
}

// NewMockObject returns a new object for mocking.
func NewMockObject[T string | []byte](annotations map[string]string) *MockObject[T] {
	return &MockObject[T]{annotations: annotations}
}

// GetName returns the object name.
func (m *MockObject[T]) GetName() string { return m.name }

// SetName sets the object name.
func (m *MockObject[T]) SetName(n string) { m.name = n }

// GetAnnotations returns annotations of the mock object.
func (m *MockObject[T]) GetAnnotations() map[string]string { return m.annotations }

// SetAnnotations sets annotations on the mock object.
func (m *MockObject[T]) SetAnnotations(a map[string]string) { m.annotations = a }

// GetLabels returns labels of the mock object.
func (m *MockObject[T]) GetLabels() map[string]string { return m.labels }

// SetLabels sets labels on the mock object.
func (m *MockObject[T]) SetLabels(l map[string]string) { m.labels = l }

// GetData returns data of the mock object.
func (m *MockObject[T]) GetData() map[string]T { return m.data }

// SetData sets data on the mock object.
func (m *MockObject[T]) SetData(d map[string]T) { m.data = d }

// MockStore allows mocking an Object store.
type MockStore[T string | []byte] struct {
	objects   map[string]*MockObject[T]
	ListErr   error
	UpdateErr error
	PutErr    error
}

// NewMockStore returns a mock object store.
func NewMockStore[T string | []byte]() *MockStore[T] {
	return &MockStore[T]{objects: make(map[string]*MockObject[T])}
}

// New returns an empty MockObject with the given name.
func (s *MockStore[T]) New(name string) object.Resource[T] {
	return &MockObject[T]{name: name}
}

// Add seeds an object in the store for testing.
func (s *MockStore[T]) Add(name string, obj *MockObject[T]) {
	s.objects[name] = obj
}

// Get an object from the store.
func (s *MockStore[T]) Get(_ context.Context, name string) (object.Resource[T], error) {
	obj, ok := s.objects[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}

	return obj, nil
}

// List returns objects matching the given labels.
func (s *MockStore[T]) List(_ context.Context, labels map[string]string) ([]object.Resource[T], error) {
	if s.ListErr != nil {
		return nil, s.ListErr
	}

	var results []object.Resource[T]
	for _, obj := range s.objects {
		match := true
		for k, v := range labels {
			if obj.GetLabels()[k] != v {
				match = false
				break
			}
		}

		if match {
			results = append(results, obj)
		}
	}

	return results, nil
}

// Update an object in the store.
func (s *MockStore[T]) Update(_ context.Context, _ object.Annotatable) error {
	return s.UpdateErr
}

// Put creates or updates a resource in the store.
func (s *MockStore[T]) Put(_ context.Context, obj object.Resource[T]) error {
	if s.PutErr != nil {
		return s.PutErr
	}

	s.objects[obj.GetName()] = &MockObject[T]{
		name:        obj.GetName(),
		annotations: obj.GetAnnotations(),
		labels:      obj.GetLabels(),
		data:        obj.GetData(),
	}
	return nil
}

//revive:disable:package-comments
package tests

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/object"
)

// MockObject is an Object that can be stored for mocking
type MockObject struct {
	name        string
	annotations map[string]string
	labels      map[string]string
	data        map[string]string
}

// NewMockObject returns a new object for mocking
func NewMockObject(annotations map[string]string) *MockObject {
	return &MockObject{annotations: annotations}
}

// GetName returns the object name.
func (m *MockObject) GetName() string { return m.name }

// SetName sets the object name.
func (m *MockObject) SetName(n string) { m.name = n }

// GetAnnotations returns annotations of the mock object
func (m *MockObject) GetAnnotations() map[string]string { return m.annotations }

// SetAnnotations sets annotations on the mock object
func (m *MockObject) SetAnnotations(a map[string]string) { m.annotations = a }

// GetLabels returns labels of the mock object.
func (m *MockObject) GetLabels() map[string]string { return m.labels }

// SetLabels sets labels on the mock object.
func (m *MockObject) SetLabels(l map[string]string) { m.labels = l }

// GetData returns data of the mock object.
func (m *MockObject) GetData() map[string]string { return m.data }

// SetData sets data on the mock object.
func (m *MockObject) SetData(d map[string]string) { m.data = d }

// MockStore allows mocking an Object store
type MockStore struct {
	objects   map[string]*MockObject
	UpdateErr error
	PutErr    error
}

// NewMockStore returns a mock object store
func NewMockStore() *MockStore {
	return &MockStore{objects: make(map[string]*MockObject)}
}

// Add seeds an object in the store for testing.
func (s *MockStore) Add(name string, obj *MockObject) {
	s.objects[name] = obj
}

// Get an object from the store
func (s *MockStore) Get(_ context.Context, name string) (object.Resource, error) {
	obj, ok := s.objects[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}

	return obj, nil
}

// List returns objects matching the given labels.
func (s *MockStore) List(_ context.Context, labels map[string]string) ([]object.Resource, error) {
	var results []object.Resource
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

// Update an object in the store
func (s *MockStore) Update(_ context.Context, _ object.Annotatable) error {
	return s.UpdateErr
}

// Put creates or updates a resource in the store
func (s *MockStore) Put(_ context.Context, obj object.Resource) error {
	if s.PutErr != nil {
		return s.PutErr
	}

	s.objects[obj.GetName()] = &MockObject{
		name:        obj.GetName(),
		annotations: obj.GetAnnotations(),
		labels:      obj.GetLabels(),
		data:        obj.GetData(),
	}
	return nil
}

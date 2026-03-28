//revive:disable:package-comments
package source

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/object"
	"github.com/katastroma/phortizo/internal/transform"
)

// TargetFromResource converts an object.Resource to a *Target.
func TargetFromResource(r object.Resource[string]) (*Target, error) {
	target, err := TargetFromResourceData(r.GetData())
	if err != nil {
		return nil, err
	}

	target.Name = r.GetName()
	target.CredentialSecret = r.GetAnnotations()[CredentialSecretAnnotation]
	return &target, nil
}

// Get reads a single target by name from the store.
func Get(ctx context.Context, store object.Store[string], name string) (*Target, error) {
	r, err := store.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}

	return TargetFromResource(r)
}

// List returns all targets from the store.
func List(ctx context.Context, store object.Store[string]) ([]*Target, error) {
	resources, err := store.List(ctx, map[string]string{object.TypeLabel: TypeLabel})
	if err != nil {
		return nil, err
	}

	return transform.Map(resources, TargetFromResource)
}

// Put writes a target to the store.
func Put(ctx context.Context, store object.Store[string], target *Target) error {
	r, err := store.Get(ctx, target.Name)
	if err != nil {
		r = store.New(target.Name)
	}

	cm := target.MarshalData()
	r.SetData(cm)
	r.SetLabels(map[string]string{object.TypeLabel: TypeLabel})

	annotations := r.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	if target.CredentialSecret != "" {
		annotations[CredentialSecretAnnotation] = target.CredentialSecret
	} else {
		delete(annotations, CredentialSecretAnnotation)
	}

	r.SetAnnotations(annotations)
	return store.Put(ctx, r)
}

//revive:disable:package-comments
package lease

import (
	"context"
	"fmt"

	"github.com/katastroma/phortizo/internal/object"
)

// Release clears the lease annotations on an object.
func Release(ctx context.Context, s object.Store[string], name string) error {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}

	annotations := obj.GetAnnotations()
	delete(annotations, IDAnnotation)
	delete(annotations, StartedAnnotation)
	delete(annotations, ReplayCountAnnotation)
	obj.SetAnnotations(annotations)

	if err = s.Update(ctx, obj); err != nil {
		return fmt.Errorf("releasing lease on %s: %w", name, err)
	}

	return nil
}

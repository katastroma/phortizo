//revive:disable:package-comments
package lease

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/katastroma/phortizo/internal/object"
)

// Read reads the current lease state from an object.
func Read(ctx context.Context, s object.Store[string], name string) (*State, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", name, err)
	}

	id := obj.GetAnnotations()[IDAnnotation]
	if id == "" {
		return nil, nil
	}

	started, err := time.Parse(time.RFC3339, obj.GetAnnotations()[StartedAnnotation])
	if err != nil {
		return nil, err
	}

	replayCount, _ := strconv.Atoi(obj.GetAnnotations()[ReplayCountAnnotation])

	return &State{ID: id, Started: started, replayCount: replayCount}, nil
}

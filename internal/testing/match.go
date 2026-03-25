//revive:disable:package-comments
package testing

import (
	"context"
	"sync"

	"github.com/katastroma/phortizo/internal/match"
)

// CaptureMatch records match results dispatched during tests.
type CaptureMatch struct {
	mu      sync.Mutex
	results []match.Result
}

// Handle records a match result.
func (c *CaptureMatch) Handle(_ context.Context, m match.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, m)
}

// Results returns the captured match results.
func (c *CaptureMatch) Results() []match.Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]match.Result, len(c.results))
	copy(cp, c.results)
	return cp
}

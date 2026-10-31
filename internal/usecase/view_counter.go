package usecase

import (
	"context"
	"log"
	"sync"
	"time"
)

// viewStore persists batched view counts.
type viewStore interface {
	IncrementViews(ctx context.Context, counts map[int64]int) error
}

// ViewCounter buffers listing views in memory and writes them in batches,
// so reading a listing doesn't cost a database write per request.
type ViewCounter struct {
	store   viewStore
	mu      sync.Mutex
	pending map[int64]int
}

// NewViewCounter creates a view counter that flushes to store.
func NewViewCounter(store viewStore) *ViewCounter {
	return &ViewCounter{store: store, pending: map[int64]int{}}
}

// Record counts one view of a listing.
func (v *ViewCounter) Record(listingID int64) {
	v.mu.Lock()
	v.pending[listingID]++
	v.mu.Unlock()
}

// Pending returns the views of a listing not yet written to the store.
func (v *ViewCounter) Pending(listingID int64) int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.pending[listingID]
}

// Flush writes buffered views to the store. On failure the views are kept
// for the next flush.
func (v *ViewCounter) Flush(ctx context.Context) error {
	v.mu.Lock()
	batch := v.pending
	v.pending = map[int64]int{}
	v.mu.Unlock()

	if err := v.store.IncrementViews(ctx, batch); err != nil {
		v.mu.Lock()
		for id, n := range batch {
			v.pending[id] += n
		}
		v.mu.Unlock()
		return err
	}
	return nil
}

// Run flushes every interval until ctx is cancelled, then flushes once more.
func (v *ViewCounter) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := v.Flush(ctx); err != nil {
				log.Printf("failed to flush listing views: %v", err)
			}
		case <-ctx.Done():
			if err := v.Flush(context.Background()); err != nil {
				log.Printf("failed to flush listing views on shutdown: %v", err)
			}
			return
		}
	}
}

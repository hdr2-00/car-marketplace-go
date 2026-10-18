package usecase

import (
	"context"
	"errors"
	"testing"
)

type fakeViewStore struct {
	got  map[int64]int
	fail bool
}

func (f *fakeViewStore) IncrementViews(_ context.Context, counts map[int64]int) error {
	if f.fail {
		return errors.New("db down")
	}
	for id, n := range counts {
		f.got[id] += n
	}
	return nil
}

func TestViewCounterBatchesViewsUntilFlush(t *testing.T) {
	store := &fakeViewStore{got: map[int64]int{}}
	vc := NewViewCounter(store)

	vc.Record(1)
	vc.Record(1)
	vc.Record(2)
	if got := vc.Pending(1); got != 2 {
		t.Fatalf("Pending(1) = %d, want 2", got)
	}

	if err := vc.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.got[1] != 2 || store.got[2] != 1 {
		t.Fatalf("flushed %v, want map[1:2 2:1]", store.got)
	}
	if vc.Pending(1) != 0 {
		t.Fatal("pending views not cleared after flush")
	}
}

func TestViewCounterKeepsViewsWhenFlushFails(t *testing.T) {
	store := &fakeViewStore{got: map[int64]int{}, fail: true}
	vc := NewViewCounter(store)
	vc.Record(7)

	if err := vc.Flush(context.Background()); err == nil {
		t.Fatal("expected flush error")
	}
	vc.Record(7)
	if got := vc.Pending(7); got != 2 {
		t.Fatalf("Pending(7) = %d after failed flush, want 2", got)
	}
}

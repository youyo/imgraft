package runtime_test

import (
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/runtime"
)

// Compile-time interface compliance checks.
var (
	_ runtime.Clock = runtime.SystemClock{}
	_ runtime.Clock = runtime.FixedClock{}
)

func TestSystemClock_Now(t *testing.T) {
	c := runtime.SystemClock{}
	before := time.Now()
	got := c.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("SystemClock.Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestFixedClock_Now(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c := runtime.NewFixedClock(fixed)

	got := c.Now()
	if !got.Equal(fixed) {
		t.Errorf("FixedClock.Now() = %v, want %v", got, fixed)
	}
}

func TestFixedClock_Now_IsIdempotent(t *testing.T) {
	fixed := time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC)
	c := runtime.NewFixedClock(fixed)

	first := c.Now()
	second := c.Now()

	if !first.Equal(second) {
		t.Errorf("FixedClock.Now() not idempotent: first=%v, second=%v", first, second)
	}
}

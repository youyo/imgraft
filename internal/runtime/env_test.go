package runtime_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/runtime"
)

func TestGetWithDefault_Unset(t *testing.T) {
	// Ensure the variable is not set.
	t.Setenv("IMGRAFT_TEST_UNSET", "")
	// Unsetenv by setting to empty then unsetting via OS-level trick is not
	// directly possible with t.Setenv. Instead we use a unique key.
	got := runtime.GetWithDefault("IMGRAFT_TEST_TRULY_UNSET_XYZ", "flash")
	if got != "flash" {
		t.Errorf("GetWithDefault() = %q, want %q", got, "flash")
	}
}

func TestGetWithDefault_Set(t *testing.T) {
	t.Setenv("IMGRAFT_TEST_SET", "pro")

	got := runtime.GetWithDefault("IMGRAFT_TEST_SET", "flash")
	if got != "pro" {
		t.Errorf("GetWithDefault() = %q, want %q", got, "pro")
	}
}

func TestGetWithDefault_EmptyStringIsNotDefault(t *testing.T) {
	t.Setenv("IMGRAFT_TEST_EMPTY", "")

	got := runtime.GetWithDefault("IMGRAFT_TEST_EMPTY", "flash")
	if got != "" {
		t.Errorf("GetWithDefault() = %q, want empty string (not default)", got)
	}
}

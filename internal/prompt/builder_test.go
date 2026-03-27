package prompt

import (
	"context"
	"errors"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
)

func TestBuild_TransparentOn(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	parts, err := Build(ctx, "blue robot mascot", true)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	if len(parts) != 2 {
		t.Fatalf("len(parts) = %d, want 2", len(parts))
	}

	if parts[0].Role != RoleSystem {
		t.Errorf("parts[0].Role = %q, want %q", parts[0].Role, RoleSystem)
	}
	if parts[0].Text != SystemPrompt(true) {
		t.Error("parts[0].Text does not match SystemPrompt(true)")
	}

	if parts[1].Role != RoleUser {
		t.Errorf("parts[1].Role = %q, want %q", parts[1].Role, RoleUser)
	}
	if parts[1].Text != "blue robot mascot" {
		t.Errorf("parts[1].Text = %q, want %q", parts[1].Text, "blue robot mascot")
	}
}

func TestBuild_TransparentOff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	parts, err := Build(ctx, "cute cat", false)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	// transparent OFF => SystemPrompt returns "" => system Part is omitted
	if len(parts) != 1 {
		t.Fatalf("len(parts) = %d, want 1", len(parts))
	}

	if parts[0].Role != RoleUser {
		t.Errorf("parts[0].Role = %q, want %q", parts[0].Role, RoleUser)
	}
	if parts[0].Text != "cute cat" {
		t.Errorf("parts[0].Text = %q, want %q", parts[0].Text, "cute cat")
	}
}

func TestBuild_TransparentOn_SystemFirst(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	parts, err := Build(ctx, "test prompt", true)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	if len(parts) < 1 {
		t.Fatal("parts is empty")
	}
	if parts[0].Role != RoleSystem {
		t.Errorf("first part Role = %q, want %q", parts[0].Role, RoleSystem)
	}
}

func TestBuild_EmptyPrompt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := Build(ctx, "", true)
	if err == nil {
		t.Fatal("Build with empty prompt must return error")
	}

	var coded *errs.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("error type = %T, want *errs.CodedError", err)
	}
	if coded.Code != errs.ErrInvalidArgument {
		t.Errorf("error code = %q, want %q", coded.Code, errs.ErrInvalidArgument)
	}
}

func TestBuild_WhitespaceOnlyPrompt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := Build(ctx, "   \t\n  ", true)
	if err == nil {
		t.Fatal("Build with whitespace-only prompt must return error")
	}

	var coded *errs.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("error type = %T, want *errs.CodedError", err)
	}
	if coded.Code != errs.ErrInvalidArgument {
		t.Errorf("error code = %q, want %q", coded.Code, errs.ErrInvalidArgument)
	}
}

func TestBuild_PromptWithNewlines(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := "hello\nworld\ttab"
	parts, err := Build(ctx, input, true)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	found := false
	for _, p := range parts {
		if p.Role == RoleUser && p.Text == input {
			found = true
			break
		}
	}
	if !found {
		t.Error("user prompt with newlines/tabs must be preserved as-is")
	}
}

func TestBuild_PromptWithJapanese(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := "青いロボットマスコット"
	parts, err := Build(ctx, input, true)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	found := false
	for _, p := range parts {
		if p.Role == RoleUser && p.Text == input {
			found = true
			break
		}
	}
	if !found {
		t.Error("user prompt with Japanese must be preserved as-is")
	}
}

func TestBuild_PromptContainingGreenColor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := "robot with #00FF00 background"
	parts, err := Build(ctx, input, true)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	// System prompt should still be present and unchanged
	if len(parts) != 2 {
		t.Fatalf("len(parts) = %d, want 2", len(parts))
	}
	if parts[0].Role != RoleSystem {
		t.Errorf("parts[0].Role = %q, want %q", parts[0].Role, RoleSystem)
	}
	if parts[1].Text != input {
		t.Errorf("user text = %q, want %q", parts[1].Text, input)
	}
}

func TestBuild_TransparentOff_PartsLength(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	parts, err := Build(ctx, "test", false)
	if err != nil {
		t.Fatalf("Build returned unexpected error: %v", err)
	}

	if len(parts) != 1 {
		t.Fatalf("len(parts) = %d, want 1 (system omitted when transparent=false)", len(parts))
	}
	if parts[0].Role != RoleUser {
		t.Errorf("parts[0].Role = %q, want %q", parts[0].Role, RoleUser)
	}
}

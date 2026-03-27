package prompt

import (
	"strings"
	"testing"
)

func TestSystemPrompt_TransparentOn(t *testing.T) {
	t.Parallel()

	got := SystemPrompt(true)

	if got == "" {
		t.Fatal("SystemPrompt(true) must return non-empty string")
	}

	// SPEC §11.2: must contain key phrases
	required := []string{
		"solid pure green background",
		"Center the subject",
		"Do not use gradients",
		"Do not use shadows on the background",
		"Do not include background objects",
		"Keep the full silhouette visible",
		"Ensure strong color contrast",
	}
	for _, phrase := range required {
		if !strings.Contains(got, phrase) {
			t.Errorf("SystemPrompt(true) missing required phrase: %q", phrase)
		}
	}
}

func TestSystemPrompt_TransparentOff(t *testing.T) {
	t.Parallel()

	got := SystemPrompt(false)

	if got != "" {
		t.Errorf("SystemPrompt(false) = %q, want empty string", got)
	}
}

func TestSystemPrompt_TransparentOn_NoGreenKeyword(t *testing.T) {
	t.Parallel()

	got := SystemPrompt(true)

	// SPEC §12.3: background color is #00FF00 (pure green).
	// The system prompt uses "green" as the color keyword.
	if !strings.Contains(got, "green") {
		t.Error("SystemPrompt(true) must mention green background")
	}
}

func TestSystemPrompt_TransparentOff_NoBackgroundForce(t *testing.T) {
	t.Parallel()

	got := SystemPrompt(false)

	forbidden := []string{
		"solid pure green background",
		"silhouette",
		"Do not include background objects",
	}
	for _, phrase := range forbidden {
		if strings.Contains(got, phrase) {
			t.Errorf("SystemPrompt(false) must not contain %q", phrase)
		}
	}
}

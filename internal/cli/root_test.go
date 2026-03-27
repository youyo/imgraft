package cli_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/cli"
)

func TestCLI_ParsePrompt(t *testing.T) {
	c, _, err := cli.Parse([]string{"blue robot mascot"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Generate.Prompt != "blue robot mascot" {
		t.Errorf("expected prompt %q, got %q", "blue robot mascot", c.Generate.Prompt)
	}
}

func TestCLI_ParseFlags(t *testing.T) {
	c, _, err := cli.Parse([]string{
		"test prompt",
		"--model", "pro",
		"--output", "./out.png",
		"--dir", "./outdir",
		"--no-transparent",
		"--pretty",
		"--verbose",
		"--debug",
		"--profile", "myprofile",
		"--config", "/tmp/config.toml",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Generate.Model != "pro" {
		t.Errorf("expected model %q, got %q", "pro", c.Generate.Model)
	}
	if c.Generate.Output != "./out.png" {
		t.Errorf("expected output %q, got %q", "./out.png", c.Generate.Output)
	}
	if c.Generate.Dir != "./outdir" {
		t.Errorf("expected dir %q, got %q", "./outdir", c.Generate.Dir)
	}
	if !c.Generate.NoTransparent {
		t.Error("expected no-transparent to be true")
	}
	if !c.Generate.Pretty {
		t.Error("expected pretty to be true")
	}
	if !c.Generate.Verbose {
		t.Error("expected verbose to be true")
	}
	if !c.Generate.Debug {
		t.Error("expected debug to be true")
	}
	if c.Generate.Profile != "myprofile" {
		t.Errorf("expected profile %q, got %q", "myprofile", c.Generate.Profile)
	}
	if c.Generate.ConfigPath != "/tmp/config.toml" {
		t.Errorf("expected config %q, got %q", "/tmp/config.toml", c.Generate.ConfigPath)
	}
}

func TestCLI_EmptyPrompt(t *testing.T) {
	c, _, err := cli.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Generate.Prompt != "" {
		t.Errorf("expected empty prompt, got %q", c.Generate.Prompt)
	}
}

func TestCLI_RefFlag(t *testing.T) {
	c, _, err := cli.Parse([]string{
		"test prompt",
		"--ref", "./img1.png",
		"--ref", "https://example.com/img2.png",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Generate.Ref) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(c.Generate.Ref))
	}
	if c.Generate.Ref[0] != "./img1.png" {
		t.Errorf("expected ref[0] %q, got %q", "./img1.png", c.Generate.Ref[0])
	}
	if c.Generate.Ref[1] != "https://example.com/img2.png" {
		t.Errorf("expected ref[1] %q, got %q", "https://example.com/img2.png", c.Generate.Ref[1])
	}
}

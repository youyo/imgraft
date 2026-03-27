package output

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/errs"
)

func TestGenerateFilename_Basic(t *testing.T) {
	dir := t.TempDir()
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	path, err := GenerateFilename(dir, tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(dir, "imgraft-20260324-153012-001.png")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestGenerateFilename_Collision(t *testing.T) {
	dir := t.TempDir()
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	// 001 を先に作成
	first := filepath.Join(dir, "imgraft-20260324-153012-001.png")
	if err := os.WriteFile(first, []byte{}, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	path, err := GenerateFilename(dir, tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(dir, "imgraft-20260324-153012-002.png")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestGenerateFilename_MultipleCollisions(t *testing.T) {
	dir := t.TempDir()
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	// 001〜005 を作成
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("imgraft-20260324-153012-%03d.png", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte{}, 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	path, err := GenerateFilename(dir, tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(dir, "imgraft-20260324-153012-006.png")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestGenerateFilename_AllSlotsFull(t *testing.T) {
	dir := t.TempDir()
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	// 001〜999 を全て作成
	for i := 1; i <= 999; i++ {
		name := fmt.Sprintf("imgraft-20260324-153012-%03d.png", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte{}, 0o644); err != nil {
			t.Fatalf("setup i=%d: %v", i, err)
		}
	}

	_, err := GenerateFilename(dir, tm)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	var coded *errs.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("expected *errs.CodedError, got %T: %v", err, err)
	}
	if coded.Code != errs.ErrFileAlreadyExists {
		t.Errorf("got code %q, want %q", coded.Code, errs.ErrFileAlreadyExists)
	}
}

func TestGenerateFilename_DirNotExist(t *testing.T) {
	// ディレクトリが存在しなくてもファイル名生成は成功する
	// （ディレクトリ作成は save.go の責務）
	dir := filepath.Join(t.TempDir(), "nonexistent")
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	path, err := GenerateFilename(dir, tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(dir, "imgraft-20260324-153012-001.png")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}

func TestGenerateFilename_Timestamp(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name     string
		t        time.Time
		wantName string
	}{
		{
			name:     "midnight",
			t:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantName: "imgraft-20260101-000000-001.png",
		},
		{
			name:     "end of day",
			t:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			wantName: "imgraft-20261231-235959-001.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GenerateFilename(dir, tt.t)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if filepath.Base(path) != tt.wantName {
				t.Errorf("got %q, want %q", filepath.Base(path), tt.wantName)
			}
		})
	}
}


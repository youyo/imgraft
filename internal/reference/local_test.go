package reference_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/webp"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/reference"
)

// mustMakePNG はテスト用のPNG bytesを生成する。
func mustMakePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		t.Fatalf("failed to create test PNG: %v", err)
	}
	return buf.Bytes()
}

// mustMakeJPEG はテスト用のJPEG bytesを生成する。
func mustMakeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			rgba.SetRGBA(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("failed to create test JPEG: %v", err)
	}
	return buf.Bytes()
}

// mustWriteTempFile はtempディレクトリにファイルを書き込み、パスを返す。
func mustWriteTempFile(t *testing.T, data []byte, filename string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestLoadLocalFile_PNG(t *testing.T) {
	data := mustMakePNG(t, 10, 10)
	path := mustWriteTempFile(t, data, "test.png")

	ref, err := reference.LoadLocalFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.SourceType != "file" {
		t.Errorf("SourceType = %q; want %q", ref.SourceType, "file")
	}
	if ref.OriginalInput != path {
		t.Errorf("OriginalInput = %q; want %q", ref.OriginalInput, path)
	}
	if ref.LocalCachedPath != path {
		t.Errorf("LocalCachedPath = %q; want %q", ref.LocalCachedPath, path)
	}
	if ref.Filename != "test.png" {
		t.Errorf("Filename = %q; want %q", ref.Filename, "test.png")
	}
	if ref.MimeType != "image/png" {
		t.Errorf("MimeType = %q; want %q", ref.MimeType, "image/png")
	}
	if ref.Width != 10 {
		t.Errorf("Width = %d; want %d", ref.Width, 10)
	}
	if ref.Height != 10 {
		t.Errorf("Height = %d; want %d", ref.Height, 10)
	}
	if ref.SizeBytes != int64(len(data)) {
		t.Errorf("SizeBytes = %d; want %d", ref.SizeBytes, len(data))
	}
	if !bytes.Equal(ref.Data, data) {
		t.Error("Data does not match original")
	}
}

func TestLoadLocalFile_JPEG(t *testing.T) {
	data := mustMakeJPEG(t, 20, 15)
	path := mustWriteTempFile(t, data, "test.jpg")

	ref, err := reference.LoadLocalFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.MimeType != "image/jpeg" {
		t.Errorf("MimeType = %q; want %q", ref.MimeType, "image/jpeg")
	}
	if ref.Width != 20 {
		t.Errorf("Width = %d; want %d", ref.Width, 20)
	}
	if ref.Height != 15 {
		t.Errorf("Height = %d; want %d", ref.Height, 15)
	}
}

func TestLoadLocalFile_NotFound(t *testing.T) {
	_, err := reference.LoadLocalFile("/nonexistent/path/image.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrFileNotFound {
		t.Errorf("error code = %q; want %q", code, errs.ErrFileNotFound)
	}
}

func TestLoadLocalFile_InvalidImage(t *testing.T) {
	path := mustWriteTempFile(t, []byte("not an image"), "fake.png")

	_, err := reference.LoadLocalFile(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrInvalidImage && code != errs.ErrUnsupportedImageFormat {
		t.Errorf("error code = %q; want ErrInvalidImage or ErrUnsupportedImageFormat", code)
	}
}

func TestLoadLocalFile_UnsupportedFormat(t *testing.T) {
	// GIF マジックバイト
	gifData := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
	path := mustWriteTempFile(t, gifData, "test.gif")

	_, err := reference.LoadLocalFile(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrUnsupportedImageFormat && code != errs.ErrInvalidImage {
		t.Errorf("error code = %q; want ErrUnsupportedImageFormat or ErrInvalidImage", code)
	}
}

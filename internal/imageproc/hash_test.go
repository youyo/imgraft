package imageproc_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

func TestSHA256OfBytes_Hello(t *testing.T) {
	// SHA256("hello") は既知の値
	got := imageproc.SHA256OfBytes([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("SHA256OfBytes(\"hello\") = %q, want %q", got, want)
	}
}

func TestSHA256OfBytes_Empty(t *testing.T) {
	got := imageproc.SHA256OfBytes([]byte{})
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("SHA256OfBytes(empty) = %q, want %q", got, want)
	}
}

func TestSHA256OfFile(t *testing.T) {
	content := []byte("hello")
	path := mustWriteTempFile(t, content, ".bin")

	got, err := imageproc.SHA256OfFile(path)
	if err != nil {
		t.Fatalf("SHA256OfFile returned error: %v", err)
	}
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("SHA256OfFile = %q, want %q", got, want)
	}
}

func TestSHA256OfFile_NotExists(t *testing.T) {
	_, err := imageproc.SHA256OfFile(filepath.Join(t.TempDir(), "nonexistent.bin"))
	if err == nil {
		t.Fatal("SHA256OfFile(nonexistent) returned nil error")
	}
	var coded *errs.CodedError
	if !errors.As(err, &coded) || coded.Code != errs.ErrFileNotFound {
		t.Errorf("error code = %v, want %v", errs.CodeOf(err), errs.ErrFileNotFound)
	}
}

func TestSHA256OfFile_Unreadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noperm.bin")
	if err := os.WriteFile(path, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	// パーミッションを削除して読み込み不能にする
	if err := os.Chmod(path, 0000); err != nil {
		t.Skip("cannot change file permissions on this platform")
	}
	t.Cleanup(func() { os.Chmod(path, 0644) })

	_, err := imageproc.SHA256OfFile(path)
	if err == nil {
		t.Fatal("SHA256OfFile(unreadable) returned nil error")
	}
	code := errs.CodeOf(err)
	if code != errs.ErrFileReadFailed {
		t.Errorf("error code = %v, want %v", code, errs.ErrFileReadFailed)
	}
}

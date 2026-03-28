package reference_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/reference"
)

// makeRef はテスト用の ReferenceImage を生成するヘルパー。
func makeRef(width, height int, sizeBytes int64, mimeType string) reference.ReferenceImage {
	return reference.ReferenceImage{
		SourceType:    "file",
		OriginalInput: "test.png",
		MimeType:      mimeType,
		Width:         width,
		Height:        height,
		SizeBytes:     sizeBytes,
	}
}

func TestValidate_OK(t *testing.T) {
	refs := []reference.ReferenceImage{
		makeRef(100, 100, 1024, "image/png"),
		makeRef(200, 200, 2048, "image/jpeg"),
	}

	if err := reference.Validate(refs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_TooMany(t *testing.T) {
	refs := make([]reference.ReferenceImage, 9)
	for i := range refs {
		refs[i] = makeRef(100, 100, 1024, "image/png")
	}

	err := reference.Validate(refs)
	if err == nil {
		t.Fatal("expected error for too many refs, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrInvalidArgument {
		t.Errorf("error code = %q; want %q", code, errs.ErrInvalidArgument)
	}
}

func TestValidate_MaxAllowed(t *testing.T) {
	refs := make([]reference.ReferenceImage, 8)
	for i := range refs {
		refs[i] = makeRef(100, 100, 1024, "image/png")
	}

	if err := reference.Validate(refs); err != nil {
		t.Fatalf("unexpected error for 8 refs: %v", err)
	}
}

func TestValidate_FileTooLarge(t *testing.T) {
	// 20MB + 1 byte
	refs := []reference.ReferenceImage{
		makeRef(100, 100, 20*1024*1024+1, "image/png"),
	}

	err := reference.Validate(refs)
	if err == nil {
		t.Fatal("expected error for file too large, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrImageTooLarge {
		t.Errorf("error code = %q; want %q", code, errs.ErrImageTooLarge)
	}
}

func TestValidate_MaxFileSizeOK(t *testing.T) {
	// ちょうど 20MB は OK
	refs := []reference.ReferenceImage{
		makeRef(100, 100, 20*1024*1024, "image/png"),
	}

	if err := reference.Validate(refs); err != nil {
		t.Fatalf("unexpected error for 20MB file: %v", err)
	}
}

func TestValidate_ResolutionTooLarge(t *testing.T) {
	// 4097 x 100 はNG
	refs := []reference.ReferenceImage{
		makeRef(4097, 100, 1024, "image/png"),
	}

	err := reference.Validate(refs)
	if err == nil {
		t.Fatal("expected error for resolution too large, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrImageTooLarge {
		t.Errorf("error code = %q; want %q", code, errs.ErrImageTooLarge)
	}
}

func TestValidate_ResolutionMaxOK(t *testing.T) {
	// ちょうど 4096 x 4096 は OK
	refs := []reference.ReferenceImage{
		makeRef(4096, 4096, 1024, "image/png"),
	}

	if err := reference.Validate(refs); err != nil {
		t.Fatalf("unexpected error for 4096x4096: %v", err)
	}
}

func TestValidate_UnsupportedMIME(t *testing.T) {
	refs := []reference.ReferenceImage{
		makeRef(100, 100, 1024, "image/gif"),
	}

	err := reference.Validate(refs)
	if err == nil {
		t.Fatal("expected error for unsupported MIME type, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrUnsupportedImageFormat {
		t.Errorf("error code = %q; want %q", code, errs.ErrUnsupportedImageFormat)
	}
}

func TestValidate_EmptyList(t *testing.T) {
	// 空リストは OK（参照画像なし）
	if err := reference.Validate([]reference.ReferenceImage{}); err != nil {
		t.Fatalf("unexpected error for empty list: %v", err)
	}
}

func TestValidate_HeightTooLarge(t *testing.T) {
	refs := []reference.ReferenceImage{
		makeRef(100, 4097, 1024, "image/png"),
	}

	err := reference.Validate(refs)
	if err == nil {
		t.Fatal("expected error for height too large, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrImageTooLarge {
		t.Errorf("error code = %q; want %q", code, errs.ErrImageTooLarge)
	}
}

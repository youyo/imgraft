package reference_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/reference"
)

// makeRemoteTransport はモック HTTP トランスポートを返す。
func makeRemoteTransport(data []byte, contentType string, statusCode int) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		h := make(http.Header)
		if contentType != "" {
			h.Set("Content-Type", contentType)
		}
		return &http.Response{
			StatusCode: statusCode,
			Status:     http.StatusText(statusCode),
			Header:     h,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})
}

func TestLoadReferences_LocalOnly(t *testing.T) {
	data := mustMakePNG(t, 4, 4)
	path := mustWriteTempFile(t, data, "local.png")

	ctx := context.Background()
	refs, err := reference.LoadReferences(ctx, []string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 1 {
		t.Fatalf("len(refs) = %d; want 1", len(refs))
	}
	if refs[0].SourceType != "file" {
		t.Errorf("SourceType = %q; want file", refs[0].SourceType)
	}
}

func TestLoadReferences_RemoteOnly(t *testing.T) {
	data := mustMakePNG(t, 6, 6)
	reference.SetHTTPTransport(makeRemoteTransport(data, "image/png", 200))
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	refs, err := reference.LoadReferences(ctx, []string{"https://example.com/img.png"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 1 {
		t.Fatalf("len(refs) = %d; want 1", len(refs))
	}
	if refs[0].SourceType != "url" {
		t.Errorf("SourceType = %q; want url", refs[0].SourceType)
	}
}

func TestLoadReferences_Mixed_OrderPreserved(t *testing.T) {
	pngData := mustMakePNG(t, 4, 4)
	jpegData := mustMakeJPEG(t, 6, 6)

	localPath := mustWriteTempFile(t, pngData, "local.png")

	reference.SetHTTPTransport(makeRemoteTransport(jpegData, "image/jpeg", 200))
	defer reference.SetHTTPTransport(nil)

	// ローカル → リモート → ローカルの順序
	paths := []string{localPath, "https://example.com/img.jpg", localPath}

	ctx := context.Background()
	refs, err := reference.LoadReferences(ctx, paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 3 {
		t.Fatalf("len(refs) = %d; want 3", len(refs))
	}

	// 順序保持の確認
	if refs[0].SourceType != "file" {
		t.Errorf("refs[0].SourceType = %q; want file", refs[0].SourceType)
	}
	if refs[1].SourceType != "url" {
		t.Errorf("refs[1].SourceType = %q; want url", refs[1].SourceType)
	}
	if refs[2].SourceType != "file" {
		t.Errorf("refs[2].SourceType = %q; want file", refs[2].SourceType)
	}
}

func TestLoadReferences_Empty(t *testing.T) {
	ctx := context.Background()
	refs, err := reference.LoadReferences(ctx, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 0 {
		t.Errorf("len(refs) = %d; want 0", len(refs))
	}
}

func TestLoadReferences_FailFastOnError(t *testing.T) {
	data := mustMakePNG(t, 4, 4)
	goodPath := mustWriteTempFile(t, data, "good.png")

	ctx := context.Background()
	_, err := reference.LoadReferences(ctx, []string{goodPath, "/nonexistent/bad.png"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrFileNotFound {
		t.Errorf("error code = %q; want %q", code, errs.ErrFileNotFound)
	}
}

func TestLoadReferences_TooMany(t *testing.T) {
	data := mustMakePNG(t, 4, 4)
	path := mustWriteTempFile(t, data, "img.png")

	// 9枚（最大8枚を超える）
	paths := make([]string, 9)
	for i := range paths {
		paths[i] = path
	}

	ctx := context.Background()
	_, err := reference.LoadReferences(ctx, paths)
	if err == nil {
		t.Fatal("expected error for too many references, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrInvalidArgument {
		t.Errorf("error code = %q; want %q", code, errs.ErrInvalidArgument)
	}
}

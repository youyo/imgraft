package reference_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/reference"
)

// roundTripFunc は http.RoundTripper インターフェースを実装するテスト用ヘルパー。
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// newMockResponse はテスト用の HTTP レスポンスを生成する。
func newMockResponse(statusCode int, body []byte, headers map[string]string) *http.Response {
	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d OK", statusCode),
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func TestLoadRemoteFile_PNG(t *testing.T) {
	data := mustMakePNG(t, 8, 8)
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return newMockResponse(200, data, map[string]string{"Content-Type": "image/png"}), nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	ref, err := reference.LoadRemoteFile(ctx, "https://example.com/test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.SourceType != "url" {
		t.Errorf("SourceType = %q; want %q", ref.SourceType, "url")
	}
	if ref.MimeType != "image/png" {
		t.Errorf("MimeType = %q; want %q", ref.MimeType, "image/png")
	}
	if ref.Width != 8 || ref.Height != 8 {
		t.Errorf("dimensions = %dx%d; want 8x8", ref.Width, ref.Height)
	}
	if ref.SizeBytes != int64(len(data)) {
		t.Errorf("SizeBytes = %d; want %d", ref.SizeBytes, len(data))
	}
}

func TestLoadRemoteFile_JPEG(t *testing.T) {
	data := mustMakeJPEG(t, 12, 10)
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return newMockResponse(200, data, map[string]string{"Content-Type": "image/jpeg"}), nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	ref, err := reference.LoadRemoteFile(ctx, "https://example.com/image.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.MimeType != "image/jpeg" {
		t.Errorf("MimeType = %q; want %q", ref.MimeType, "image/jpeg")
	}
}

func TestLoadRemoteFile_404(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Status:     "404 Not Found",
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader([]byte("not found"))),
		}, nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	_, err := reference.LoadRemoteFile(ctx, "https://example.com/missing.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrReferenceFetchFailed {
		t.Errorf("error code = %q; want %q", code, errs.ErrReferenceFetchFailed)
	}
}

func TestLoadRemoteFile_TooLarge_ContentLength(t *testing.T) {
	// Content-Length が 20MB 超を示す
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		h := make(http.Header)
		h.Set("Content-Type", "image/png")
		return &http.Response{
			StatusCode:    200,
			Status:        "200 OK",
			Header:        h,
			ContentLength: 21 * 1024 * 1024, // 21MB
			Body:          io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	_, err := reference.LoadRemoteFile(ctx, "https://example.com/huge.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrImageTooLarge {
		t.Errorf("error code = %q; want %q", code, errs.ErrImageTooLarge)
	}
}

func TestLoadRemoteFile_TooManyRedirects(t *testing.T) {
	// 毎回リダイレクトを返す
	callCount := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		callCount++
		h := make(http.Header)
		h.Set("Location", "https://example.com/redirect")
		return &http.Response{
			StatusCode: 302,
			Status:     "302 Found",
			Header:     h,
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}, nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	_, err := reference.LoadRemoteFile(ctx, "https://example.com/loop")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrReferenceRedirectLimitExceeded && code != errs.ErrReferenceFetchFailed {
		t.Errorf("error code = %q; want ErrReferenceRedirectLimitExceeded or ErrReferenceFetchFailed", code)
	}
}

func TestLoadRemoteFile_Timeout(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// コンテキストがキャンセルされるまで待機
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("unexpected timeout")
		}
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := reference.LoadRemoteFile(ctx, "https://example.com/slow")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrReferenceTimeout && code != errs.ErrReferenceFetchFailed {
		t.Errorf("error code = %q; want ErrReferenceTimeout or ErrReferenceFetchFailed", code)
	}
}

func TestLoadRemoteFile_ForbiddenURL(t *testing.T) {
	ctx := context.Background()

	// localhost は forbidden（ValidateURL で弾かれるので transport は呼ばれない）
	_, err := reference.LoadRemoteFile(ctx, "http://localhost/image.png")
	if err == nil {
		t.Fatal("expected error for localhost, got nil")
	}

	code := errs.CodeOf(err)
	if code != errs.ErrReferenceURLForbidden {
		t.Errorf("error code = %q; want %q", code, errs.ErrReferenceURLForbidden)
	}
}

func TestLoadRemoteFile_FilenameFromURL(t *testing.T) {
	data := mustMakePNG(t, 4, 4)
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return newMockResponse(200, data, map[string]string{"Content-Type": "image/png"}), nil
	})
	reference.SetHTTPTransport(transport)
	defer reference.SetHTTPTransport(nil)

	ctx := context.Background()
	ref, err := reference.LoadRemoteFile(ctx, "https://example.com/assets/myimage.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ref.Filename != "myimage.png" {
		t.Errorf("Filename = %q; want %q", ref.Filename, "myimage.png")
	}
	if ref.OriginalInput != "https://example.com/assets/myimage.png" {
		t.Errorf("OriginalInput = %q; want original URL", ref.OriginalInput)
	}
}

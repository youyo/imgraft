package imageproc_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// mustMakePNG はテスト用のPNG bytesを生成する。
// withAlphaがtrueの場合、半透明ピクセルを含むNRGBA画像を生成する。
func mustMakePNG(t *testing.T, w, h int, withAlpha bool) []byte {
	t.Helper()
	var img image.Image
	if withAlpha {
		nrgba := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := range h {
			for x := range w {
				nrgba.SetNRGBA(x, y, color.NRGBA{R: 100, G: 150, B: 200, A: 128})
			}
		}
		img = nrgba
	} else {
		rgba := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := range h {
			for x := range w {
				rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			}
		}
		img = rgba
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
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

// mustReadWebP はtestdata/sample.webpを読み込んで返す。
func mustReadWebP(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sample.webp"))
	if err != nil {
		t.Fatalf("failed to read testdata/sample.webp: %v", err)
	}
	return data
}

// mustWriteTempFile はtempディレクトリにファイルを書き込み、パスを返す。
func mustWriteTempFile(t *testing.T, data []byte, ext string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test"+ext)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// makeGIFBytes はGIFのマジックバイトを持つ最小限のバイト列を返す。
func makeGIFBytes() []byte {
	return []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
}

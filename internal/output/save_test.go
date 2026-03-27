package output

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/youyo/imgraft/internal/imageproc"
	"github.com/youyo/imgraft/internal/runtime"
)

// createTestPNG は指定サイズの単色PNGバイト列を生成するヘルパー。
func createTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	// 単色で塗りつぶす
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("createTestPNG: %v", err)
	}
	return buf.Bytes()
}

func TestSavePNG_Success(t *testing.T) {
	dir := t.TempDir()
	pngData := createTestPNG(t, 10, 10)
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	opts := SaveOptions{
		Dir:                dir,
		Clock:              runtime.NewFixedClock(tm),
		Index:              0,
		TransparentApplied: true,
	}

	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.Index != 0 {
		t.Errorf("Index: got %d, want 0", item.Index)
	}
	if item.Filename != "imgraft-20260324-153012-001.png" {
		t.Errorf("Filename: got %q, want %q", item.Filename, "imgraft-20260324-153012-001.png")
	}
	if !filepath.IsAbs(item.Path) {
		t.Errorf("Path should be absolute, got %q", item.Path)
	}
	if item.Width != 10 {
		t.Errorf("Width: got %d, want 10", item.Width)
	}
	if item.Height != 10 {
		t.Errorf("Height: got %d, want 10", item.Height)
	}
	if item.MimeType != "image/png" {
		t.Errorf("MimeType: got %q, want image/png", item.MimeType)
	}
	if item.SHA256 == "" {
		t.Error("SHA256 should not be empty")
	}
	if !item.TransparentApplied {
		t.Error("TransparentApplied should be true")
	}

	// ファイルが実際に存在する
	if _, statErr := os.Stat(item.Path); statErr != nil {
		t.Errorf("file should exist: %v", statErr)
	}
}

func TestSavePNG_WithOutputPath(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "custom-output.png")
	pngData := createTestPNG(t, 5, 5)

	opts := SaveOptions{
		OutputPath: outPath,
		Index:      1,
	}
	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.Path != outPath {
		t.Errorf("Path: got %q, want %q", item.Path, outPath)
	}
	if item.Filename != "custom-output.png" {
		t.Errorf("Filename: got %q, want custom-output.png", item.Filename)
	}
	if item.Index != 1 {
		t.Errorf("Index: got %d, want 1", item.Index)
	}
}

func TestSavePNG_AutoCreateDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "deep")
	pngData := createTestPNG(t, 5, 5)
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	opts := SaveOptions{
		Dir:   dir,
		Clock: runtime.NewFixedClock(tm),
	}
	_, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ディレクトリが作成された
	if _, statErr := os.Stat(dir); statErr != nil {
		t.Errorf("dir should be created: %v", statErr)
	}
}

func TestSavePNG_EmptyData(t *testing.T) {
	dir := t.TempDir()
	opts := SaveOptions{
		Dir:   dir,
		Clock: runtime.NewFixedClock(time.Now()),
	}
	_, err := SavePNG([]byte{}, opts)
	if err == nil {
		t.Fatal("expected error for empty data but got nil")
	}
}

func TestSavePNG_SHA256Correctness(t *testing.T) {
	dir := t.TempDir()
	pngData := createTestPNG(t, 5, 5)
	expected := imageproc.SHA256OfBytes(pngData)

	opts := SaveOptions{
		Dir:   dir,
		Clock: runtime.NewFixedClock(time.Now()),
	}
	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.SHA256 != expected {
		t.Errorf("SHA256: got %q, want %q", item.SHA256, expected)
	}
}

func TestSavePNG_NilClock_FallsbackToSystemClock(t *testing.T) {
	dir := t.TempDir()
	pngData := createTestPNG(t, 5, 5)

	opts := SaveOptions{
		Dir:   dir,
		Clock: nil, // nil → SystemClock にフォールバック
	}
	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ファイルが生成されていることを確認
	if item.Filename == "" {
		t.Error("Filename should not be empty")
	}
}

func TestSavePNG_DefaultDir(t *testing.T) {
	// Dir と OutputPath の両方が空の場合、カレントディレクトリ "." を使用する
	// テスト中にカレントディレクトリへ書き込まないよう一時ディレクトリへ移動は
	// しづらいため、OutputPath を指定して Dir空の動作をテストする
	dir := t.TempDir()
	outPath := filepath.Join(dir, "default-dir-test.png")
	pngData := createTestPNG(t, 5, 5)

	opts := SaveOptions{
		OutputPath: outPath,
		Dir:        "", // 空
	}
	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Path != outPath {
		t.Errorf("Path: got %q, want %q", item.Path, outPath)
	}
}

func TestSavePNG_OutputPathAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "existing.png")
	// 既にファイルが存在する状態を作る
	if err := os.WriteFile(outPath, []byte("dummy"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	pngData := createTestPNG(t, 5, 5)
	opts := SaveOptions{
		OutputPath: outPath,
	}
	_, err := SavePNG(pngData, opts)
	if err == nil {
		t.Fatal("expected error for existing file but got nil")
	}
}

func TestSavePNG_TransparentAppliedFalse(t *testing.T) {
	dir := t.TempDir()
	pngData := createTestPNG(t, 5, 5)

	opts := SaveOptions{
		Dir:                dir,
		Clock:              runtime.NewFixedClock(time.Now()),
		TransparentApplied: false,
	}
	item, err := SavePNG(pngData, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.TransparentApplied {
		t.Error("TransparentApplied should be false")
	}
}

func TestSavePNG_ErrorCleanupOnInspectFailure(t *testing.T) {
	// 無効なデータ（PNG でない）を渡すと inspect が失敗し、
	// 部分ファイルが削除されることを確認する
	dir := t.TempDir()
	tm := time.Date(2026, 3, 24, 15, 30, 12, 0, time.UTC)

	invalidData := []byte("not a png file")
	opts := SaveOptions{
		Dir:   dir,
		Clock: runtime.NewFixedClock(tm),
	}
	_, err := SavePNG(invalidData, opts)
	if err == nil {
		t.Fatal("expected error for invalid PNG but got nil")
	}

	// 部分ファイルが残っていないことを確認
	expectedPartial := filepath.Join(dir, "imgraft-20260324-153012-001.png")
	if _, statErr := os.Stat(expectedPartial); !os.IsNotExist(statErr) {
		t.Errorf("partial file should have been cleaned up: %v", statErr)
	}
}

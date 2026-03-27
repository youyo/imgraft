package imageproc

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	_ "golang.org/x/image/webp"

	"github.com/youyo/imgraft/internal/errs"
)

// ImageMeta は画像のメタデータを表す。
type ImageMeta struct {
	Width    int
	Height   int
	MimeType string
}

// formatToMIME はimage.Decodeが返すフォーマット文字列をMIMEタイプに変換する。
var formatToMIME = map[string]string{
	"png":  "image/png",
	"jpeg": "image/jpeg",
	"webp": "image/webp",
}

// InspectFile はファイルパスからImageMetaを取得する。
// image.DecodeConfigを使用して全デコードせずに軽量に読み込む。
func InspectFile(path string) (ImageMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ImageMeta{}, errs.Wrap(errs.ErrFileNotFound, err)
		}
		return ImageMeta{}, errs.Wrap(errs.ErrFileReadFailed, err)
	}
	defer f.Close()

	// io.ReadAll でファイル全体を読み込み、bytes.NewReaderで複数回処理可能にする
	data, err := io.ReadAll(f)
	if err != nil {
		return ImageMeta{}, errs.Wrap(errs.ErrFileReadFailed, err)
	}

	return InspectBytes(data)
}

// InspectBytes はbytesからImageMetaを取得する。
func InspectBytes(data []byte) (ImageMeta, error) {
	if len(data) == 0 {
		return ImageMeta{}, errs.New(errs.ErrInvalidImage, "image data is empty")
	}

	r := bytes.NewReader(data)
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		// DetectFormatで判定してエラーを切り分ける
		detected := DetectFormat(data)
		if detected != "" && !supportedFormats[detected] {
			return ImageMeta{}, errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+detected)
		}
		return ImageMeta{}, errs.Wrap(errs.ErrInvalidImage, err)
	}

	if !supportedFormats[format] {
		return ImageMeta{}, errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+format)
	}

	mime, ok := formatToMIME[format]
	if !ok {
		return ImageMeta{}, errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+format)
	}

	return ImageMeta{
		Width:    cfg.Width,
		Height:   cfg.Height,
		MimeType: mime,
	}, nil
}

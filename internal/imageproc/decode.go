// Package imageproc はPNG/JPEG/WebPの画像デコード・エンコード・ハッシュ・メタデータ取得を提供する。
package imageproc

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/youyo/imgraft/internal/errs"
)

// supportedFormats は対応フォーマットのセット。
var supportedFormats = map[string]bool{
	"png":  true,
	"jpeg": true,
	"webp": true,
}

// Decode はPNG/JPEG/WebPのbytesをimage.Imageにデコードする。
// フォーマット不明・デコード失敗時は errs.CodedError を返す。
func Decode(data []byte) (image.Image, string, error) {
	if len(data) == 0 {
		return nil, "", errs.New(errs.ErrInvalidImage, "image data is empty")
	}

	r := bytes.NewReader(data)
	img, format, err := image.Decode(r)
	if err != nil {
		// image.Decode が失敗した場合、フォーマット検出だけ試みてエラーを切り分ける
		r.Reset(data)
		_, detectedFormat, cfgErr := image.DecodeConfig(r)
		if cfgErr == nil && !supportedFormats[detectedFormat] {
			return nil, "", errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+detectedFormat)
		}
		// DetectFormat でフォーマット推定
		detected := DetectFormat(data)
		if detected != "" && !supportedFormats[detected] {
			return nil, "", errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+detected)
		}
		return nil, "", errs.Wrap(errs.ErrInvalidImage, err)
	}

	if !supportedFormats[format] {
		return nil, "", errs.New(errs.ErrUnsupportedImageFormat, "unsupported image format: "+format)
	}

	return img, format, nil
}

// DetectFormat はbytesのマジックバイトからフォーマットを判定する。
// 戻り値: "png" | "jpeg" | "webp" | ""（判定不能時）
func DetectFormat(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// PNG: 89 50 4E 47
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg"
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "webp"
	}

	// GIF: GIF87a or GIF89a
	if len(data) >= 6 && string(data[0:3]) == "GIF" {
		return "gif"
	}

	return ""
}

package reference

import (
	"fmt"

	"github.com/youyo/imgraft/internal/errs"
)

const (
	// MaxReferenceCount は参照画像の最大枚数。
	MaxReferenceCount = 8

	// MaxFileSizeBytes は参照画像の最大ファイルサイズ（20MB）。
	MaxFileSizeBytes = 20 * 1024 * 1024

	// MaxDimension は参照画像の最大解像度（幅・高さそれぞれ）。
	MaxDimension = 4096
)

// supportedMIMETypes は参照画像として対応する MIME タイプのセット。
var supportedMIMETypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

// Validate は参照画像リストのバリデーションを行う。
// 以下の条件を検証する:
//   - 最大 8 枚
//   - 各 20MB 以下
//   - 各 4096x4096 以下
//   - PNG/JPEG/WebP のみ
//
// 1 枚でも不正なら fail-fast で最初のエラーを返す。
func Validate(refs []ReferenceImage) error {
	if len(refs) > MaxReferenceCount {
		return errs.New(errs.ErrInvalidArgument,
			fmt.Sprintf("too many reference images: %d (max %d)", len(refs), MaxReferenceCount))
	}

	for i, ref := range refs {
		if !supportedMIMETypes[ref.MimeType] {
			return errs.New(errs.ErrUnsupportedImageFormat,
				fmt.Sprintf("reference[%d]: unsupported MIME type: %s (supported: image/png, image/jpeg, image/webp)", i, ref.MimeType))
		}

		if ref.SizeBytes > MaxFileSizeBytes {
			return errs.New(errs.ErrImageTooLarge,
				fmt.Sprintf("reference[%d]: file size %d bytes exceeds limit of %d bytes (20MB)", i, ref.SizeBytes, MaxFileSizeBytes))
		}

		if ref.Width > MaxDimension || ref.Height > MaxDimension {
			return errs.New(errs.ErrImageTooLarge,
				fmt.Sprintf("reference[%d]: resolution %dx%d exceeds limit of %dx%d", i, ref.Width, ref.Height, MaxDimension, MaxDimension))
		}
	}

	return nil
}

package reference

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/imageproc"
)

// LoadLocalFile はローカルファイルパスから ReferenceImage を読み込む。
// ファイルが存在しない場合は ErrFileNotFound、
// デコード失敗時は ErrInvalidImage または ErrUnsupportedImageFormat を返す。
func LoadLocalFile(path string) (ReferenceImage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ReferenceImage{}, errs.Wrap(errs.ErrFileNotFound, err)
		}
		return ReferenceImage{}, errs.Wrap(errs.ErrFileReadFailed, err)
	}

	meta, err := imageproc.InspectBytes(data)
	if err != nil {
		return ReferenceImage{}, err
	}

	return ReferenceImage{
		SourceType:      "file",
		OriginalInput:   path,
		LocalCachedPath: path,
		Filename:        filepath.Base(path),
		MimeType:        meta.MimeType,
		Width:           meta.Width,
		Height:          meta.Height,
		SizeBytes:       int64(len(data)),
		Data:            data,
	}, nil
}

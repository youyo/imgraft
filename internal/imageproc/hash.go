package imageproc

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/youyo/imgraft/internal/errs"
)

// SHA256OfBytes はbytesのSHA256をhex stringで返す。
func SHA256OfBytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// SHA256OfFile はファイルパスを受け取り、SHA256 hex stringを返す。
// ファイル読み込みエラーは errs.CodedError を返す。
func SHA256OfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errs.Wrap(errs.ErrFileNotFound, err)
		}
		return "", errs.Wrap(errs.ErrFileReadFailed, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errs.Wrap(errs.ErrFileReadFailed, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

package imageproc

import (
	"bytes"
	"image"
	"image/png"

	"github.com/youyo/imgraft/internal/errs"
)

// EncodePNG はimage.ImageをPNG bytesにエンコードする。
// alphaチャネルを保持する（NRGBA/RGBAはそのままPNGに収まる）。
func EncodePNG(img image.Image) ([]byte, error) {
	if img == nil {
		return nil, errs.New(errs.ErrInternal, "cannot encode nil image")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errs.Wrap(errs.ErrInternal, err)
	}

	return buf.Bytes(), nil
}

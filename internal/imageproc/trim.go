package imageproc

import (
	"image"
	"image/color"

	"github.com/youyo/imgraft/internal/errs"
)

// TrimTransparent は alpha > 0 の bounding box でクロップした新しい *image.NRGBA を返す。
// 全ピクセルが透明（alpha=0）の場合は INTERNAL_ERROR を返す。
func TrimTransparent(img *image.NRGBA) (*image.NRGBA, error) {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	// alpha > 0 のピクセルの bounding box を計算する
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	// 全ピクセルが透明の場合はエラーを返す
	if minX > maxX || minY > maxY {
		return nil, errs.New(errs.ErrInternal, "image is fully transparent after background removal")
	}

	// bounding box でクロップする（maxX, maxY は inclusive なので +1 する）
	cropRect := image.Rect(minX, minY, maxX+1, maxY+1)
	w := cropRect.Dx()
	h := cropRect.Dy()

	result := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := cropRect.Min.Y; y < cropRect.Max.Y; y++ {
		for x := cropRect.Min.X; x < cropRect.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			result.SetNRGBA(x-cropRect.Min.X, y-cropRect.Min.Y, color.NRGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: c.A,
			})
		}
	}

	return result, nil
}

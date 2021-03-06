package imaging

import (
	"image"
	"math"
)

// Crop cuts out a rectangular region with the specified bounds
// from the image and returns the cropped image.
func Crop(img image.Image, rect image.Rectangle) *image.NRGBA {
	src := toNRGBA(img)
	srcRect := rect.Sub(img.Bounds().Min)
	sub := src.SubImage(srcRect)
	return Clone(sub) // New image Bounds().Min point will be (0, 0)
}

// CropCenter cuts out a rectangular region with the specified size
// from the center of the image and returns the cropped image.
func CropCenter(img image.Image, width, height int) *image.NRGBA {
	cropW, cropH := width, height

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y

	centerX := srcMinX + srcW/2
	centerY := srcMinY + srcH/2

	x0 := centerX - cropW/2
	y0 := centerY - cropH/2
	x1 := x0 + cropW
	y1 := y0 + cropH

	return Crop(img, image.Rect(x0, y0, x1, y1))
}

func CropTop(img image.Image, width, height int) *image.NRGBA {
	cropW, cropH := width, height

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcMinX := srcBounds.Min.X

	centerX := srcMinX + srcW/2

	x0 := centerX - cropW/2
	y0 := 0
	x1 := x0 + cropW
	y1 := y0 + cropH

	return Crop(img, image.Rect(x0, y0, x1, y1))
}

// Paste pastes the img image to the background image at the specified position and returns the combined image.
func Paste(background, img image.Image, pos image.Point) *image.NRGBA {
	src := toNRGBA(img)
	dst := Clone(background)                    // cloned image bounds start at (0, 0)
	startPt := pos.Sub(background.Bounds().Min) // so we should translate start point
	endPt := startPt.Add(src.Bounds().Size())
	pasteBounds := image.Rectangle{startPt, endPt}

	if dst.Bounds().Overlaps(pasteBounds) {
		intersectBounds := dst.Bounds().Intersect(pasteBounds)

		rowSize := intersectBounds.Dx() * 4
		numRows := intersectBounds.Dy()

		srcStartX := intersectBounds.Min.X - pasteBounds.Min.X
		srcStartY := intersectBounds.Min.Y - pasteBounds.Min.Y

		i0 := dst.PixOffset(intersectBounds.Min.X, intersectBounds.Min.Y)
		j0 := src.PixOffset(srcStartX, srcStartY)

		di := dst.Stride
		dj := src.Stride

		for row := 0; row < numRows; row++ {
			copy(dst.Pix[i0:i0+rowSize], src.Pix[j0:j0+rowSize])
			i0 += di
			j0 += dj
		}
	}

	return dst
}

// PasteCenter pastes the img image to the center of the background image and returns the combined image.
func PasteCenter(background, img image.Image) *image.NRGBA {
	bgBounds := background.Bounds()
	bgW := bgBounds.Dx()
	bgH := bgBounds.Dy()
	bgMinX := bgBounds.Min.X
	bgMinY := bgBounds.Min.Y

	centerX := bgMinX + bgW/2
	centerY := bgMinY + bgH/2

	x0 := centerX - img.Bounds().Dx()/2
	y0 := centerY - img.Bounds().Dy()/2

	return Paste(background, img, image.Pt(x0, y0))
}

// Overlay draws the img image over the background image at given position
// and returns the combined image. Opacity parameter is the opacity of the img
// image layer, used to compose the images, it must be from 0.0 to 1.0.
//
// Usage examples:
//
//		// draw the sprite over the background at position (50, 50)
//		dstImage := imaging.Overlay(backgroundImage, spriteImage, image.Pt(50, 50), 1.0)
//
//		// blend two opaque images of the same size
//		dstImage := imaging.Overlay(imageOne, imageTwo, image.Pt(0, 0), 0.5)
//
func Overlay(background, img image.Image, pos image.Point, opacity float64) *image.NRGBA {
	opacity = math.Min(math.Max(opacity, 0.0), 1.0) // check: 0.0 <= opacity <= 1.0

	src := toNRGBA(img)
	dst := Clone(background)                    // cloned image bounds start at (0, 0)
	startPt := pos.Sub(background.Bounds().Min) // so we should translate start point
	endPt := startPt.Add(src.Bounds().Size())
	pasteBounds := image.Rectangle{startPt, endPt}

	if dst.Bounds().Overlaps(pasteBounds) {
		intersectBounds := dst.Bounds().Intersect(pasteBounds)

		for y := intersectBounds.Min.Y; y < intersectBounds.Max.Y; y++ {
			for x := intersectBounds.Min.X; x < intersectBounds.Max.X; x++ {
				i := y*dst.Stride + x*4

				srcX := x - pasteBounds.Min.X
				srcY := y - pasteBounds.Min.Y
				j := srcY*src.Stride + srcX*4

				a1 := float64(dst.Pix[i+3])
				a2 := float64(src.Pix[j+3])

				coef2 := opacity * a2 / 255.0
				coef1 := (1 - coef2) * a1 / 255.0
				coefSum := coef1 + coef2
				coef1 /= coefSum
				coef2 /= coefSum

				dst.Pix[i+0] = uint8(float64(dst.Pix[i+0])*coef1 + float64(src.Pix[j+0])*coef2)
				dst.Pix[i+1] = uint8(float64(dst.Pix[i+1])*coef1 + float64(src.Pix[j+1])*coef2)
				dst.Pix[i+2] = uint8(float64(dst.Pix[i+2])*coef1 + float64(src.Pix[j+2])*coef2)
				dst.Pix[i+3] = uint8(math.Min(a1+a2*opacity*(255.0-a1)/255.0, 255.0))
			}
		}
	}

	return dst
}

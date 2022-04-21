package image

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

func DrawRect(img *image.RGBA, x, y, width, height int, color color.Color) {
	for i := x; i < width+x; i++ {
		for j := y; j < height+y; j++ {
			(*img).Set(i, j, color)
		}
	}
}

// Draw an image on top of another image. Transparancy is handled with A over B.
func DrawImage(img *image.RGBA, other image.Image, x, y int) {
	for i := 0; i < other.Bounds().Max.X; i++ {
		for j := 0; j < other.Bounds().Max.Y; j++ {
			// new colors
			var nr, ng, nb, na float32

			// Overlay colors
			or, og, ob, oa := other.At(i, j).RGBA()

			// Convert tot 8 bit representation
			or, og, ob, oa = or/257, og/257, ob/257, oa/257

			// Convert to float
			var orf, ogf, obf, oaf float32 = float32(or) / 255, float32(og) / 255, float32(ob) / 255, float32(oa) / 255

			if oa != 255 {
				// Current colors
				cr, cg, cb, ca := img.At(i+x, j+y).RGBA()

				// convert to 8 bit representation
				cr, cg, cb, ca = cr/257, cg/257, cb/257, ca/257

				// Convert to float
				var crf, cgf, cbf, caf float32 = float32(cr) / 255, float32(cg) / 255, float32(cb) / 255, float32(ca) / 255

				if oa != 0 {
					na = oaf + caf*(1-oaf)
					nr = (orf*oaf + crf*caf*(1-oaf)) / na
					ng = (ogf*oaf + cgf*caf*(1-oaf)) / na
					nb = (obf*oaf + cbf*caf*(1-oaf)) / na
				} else {
					nr, ng, nb, na = crf, cgf, cbf, caf
				}
			} else {
				nr, ng, nb, na = orf, ogf, obf, oaf
			}

			(*img).Set(i+x, j+y, color.RGBA{uint8(nr * 255), uint8(ng * 255), uint8(nb * 255), uint8(na * 255)})
		}
	}
}

// Export as png
func ExportImage(img image.Image, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	png.Encode(f, img)
}

// Only works for scaling down, not scaling up
func ResizeImage(ogImg image.Image, width, height int) image.Image {
	return resize.Resize(uint(width), uint(height), ogImg, resize.NearestNeighbor)
}

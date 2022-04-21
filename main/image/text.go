package image

import (
	"image"
	"image/color"

	"github.com/golang/freetype"
)

// Draw text onto an image
//  - `ctx`: the freetype context, can be created using `freetype.NewContext()`
func AddLabel(img *image.RGBA, ctx *freetype.Context, x, y int, text string, fontSize float64, color color.Color) error {
	ctx.SetDst(img)
	ctx.SetClip(img.Bounds())
	fg := image.NewRGBA(img.Bounds())
	DrawRect(fg, fg.Bounds().Min.X, fg.Bounds().Min.Y, fg.Bounds().Max.X, fg.Bounds().Max.Y, color)
	ctx.SetSrc(fg)
	ctx.SetFontSize(fontSize)
	pt := freetype.Pt(x, y+int(ctx.PointToFixed(fontSize)>>6)) // y

	if _, err := ctx.DrawString(text, pt); err != nil {
		return err
	} else {
		return nil
	}
}

package sharecard

import (
	"image"
	"image/color"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// textFace is the one and only font this v1 engine ships with: a built-in
// bitmap face from golang.org/x/image, chosen specifically so this pass
// doesn't have to source/embed/license a TTF (that belongs to the design
// pass). It renders small (7x13px per glyph) so every draw call below
// upscales it with nearest-neighbor sampling — the resulting pixelated
// look is intentional placeholder texture, not a bug.
var textFace = basicfont.Face7x13

// measureWidth returns the pixel width `text` would occupy at the given
// integer scale, without drawing anything.
func measureWidth(text string, scale int) int {
	d := font.Drawer{Face: textFace}
	return d.MeasureString(asciiFold(text)).Ceil() * scale
}

// drawLine draws a single line of text with its top-left corner at
// (x, y) and returns the y coordinate immediately below the drawn glyphs,
// so callers can stack lines without recomputing line heights by hand.
//
// text is folded to basicfont's ASCII-only glyph set via asciiFold before
// measuring or drawing (see ascii_fold.go), so callers can pass raw
// Catalan strings (torró names, etc.) directly.
func drawLine(dst *image.RGBA, text string, x, y int, col color.Color, scale int) int {
	text = asciiFold(text)
	if text == "" {
		return y
	}

	d := &font.Drawer{
		Src:  image.NewUniform(col),
		Face: textFace,
	}
	width := d.MeasureString(text).Ceil()
	height := textFace.Height
	if width <= 0 {
		return y
	}

	// Render at native (small) size onto a tightly-cropped scratch image,
	// then blit it onto dst scaled up. Drawing directly at native size and
	// scaling the destination rectangle would require re-measuring glyph
	// positions per scale factor; this is simpler and keeps basicfont's
	// metrics authoritative.
	scratch := image.NewRGBA(image.Rect(0, 0, width, height))
	d.Dst = scratch
	d.Dot = fixed.Point26_6{X: 0, Y: fixed.I(textFace.Ascent)}
	d.DrawString(text)

	dstRect := image.Rect(x, y, x+width*scale, y+height*scale)
	draw.NearestNeighbor.Scale(dst, dstRect, scratch, scratch.Bounds(), draw.Over, nil)

	return y + height*scale
}

// drawCenteredLine behaves like drawLine but horizontally centers the
// text within the full canvas width.
func drawCenteredLine(dst *image.RGBA, text string, y int, col color.Color, scale int) int {
	width := measureWidth(text, scale)
	x := (CanvasWidth - width) / 2
	if x < 0 {
		x = 0
	}
	return drawLine(dst, text, x, y, col, scale)
}

// wrapText greedily wraps text into lines that fit within maxWidthPx at
// the given scale, breaking on word boundaries. It never splits a single
// word, so a word wider than maxWidthPx will still overflow on its own
// line — acceptable for a v1 placeholder card with short torró names.
func wrapText(text string, maxWidthPx int, scale int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		if measureWidth(candidate, scale) > maxWidthPx {
			lines = append(lines, current)
			current = w
			continue
		}
		current = candidate
	}
	lines = append(lines, current)

	return lines
}

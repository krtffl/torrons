package sharecard

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// This package used to ship with a built-in 7x13 bitmap face
// (golang.org/x/image/font/basicfont) upscaled with nearest-neighbor
// sampling, plus an accent-stripping fallback (asciiFold) for Catalan
// diacritics that fall outside basicfont's ASCII-only glyph set - see git
// history for that v1 approach.
//
// This is the real fix: two embedded outline (TrueType) fonts, parsed once
// via golang.org/x/image/font/sfnt and rasterized on demand via
// golang.org/x/image/font/opentype, giving proper anti-aliased glyphs at
// exact sizes with full Latin Extended coverage (à, è, é, í, ï, ò, ó, ú, ü,
// ç, the "l·l" interpunct, ...). No folding/stripping of any kind.
//
// See assets/fonts/README.md for exactly which fonts these are and why.
var (
	//go:embed assets/fonts/LiberationSans-Bold.ttf
	sansBoldTTF []byte
	//go:embed assets/fonts/LiberationSerif-Italic.ttf
	serifItalicTTF []byte

	sansBoldFont    *sfnt.Font
	serifItalicFont *sfnt.Font
)

func init() {
	var err error
	sansBoldFont, err = opentype.Parse(sansBoldTTF)
	if err != nil {
		panic(fmt.Sprintf("sharecard: parse embedded sans-bold font: %v", err))
	}
	serifItalicFont, err = opentype.Parse(serifItalicTTF)
	if err != nil {
		panic(fmt.Sprintf("sharecard: parse embedded serif-italic font: %v", err))
	}
}

// fontStyle picks one of the two embedded typefaces.
type fontStyle int

const (
	styleSansBold fontStyle = iota
	styleSerifItalic
)

// faceKey identifies a cached, ready-to-draw font.Face by style and size.
type faceKey struct {
	style fontStyle
	size  float64
}

// newFace builds a font.Face for the given style/size directly from the
// shared, already-parsed *sfnt.Font. Building a Face is cheap (it wraps a
// few fields, no re-parsing of the font tables), so canvas.face below
// builds one per (style, size) it's asked for and caches it for the
// lifetime of a single canvas - see canvas.go's concurrency note on why
// that cache must never be shared across canvases/goroutines.
func newFace(style fontStyle, size float64) font.Face {
	src := sansBoldFont
	if style == styleSerifItalic {
		src = serifItalicFont
	}
	face, err := opentype.NewFace(src, &opentype.FaceOptions{
		Size:    size,
		DPI:     72, // 1pt == 1px at 72 DPI, matching the mockup's CSS px sizes
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(fmt.Sprintf("sharecard: build face %v/%v: %v", style, size, err))
	}
	return face
}

// measureWidth returns the pixel width `text` would occupy at the given
// style/size, without drawing anything.
func measureWidth(face font.Face, text string) int {
	d := font.Drawer{Face: face}
	return d.MeasureString(text).Ceil()
}

// measureTracked is measureWidth plus extra letter-spacing (tracking, in
// px) added after every rune including the last - matching how CSS
// letter-spacing affects layout, close enough for centering math here.
func measureTracked(face font.Face, text string, tracking float64) int {
	width := measureWidth(face, text)
	n := len([]rune(text))
	if n == 0 {
		return 0
	}
	return width + int(tracking*float64(n))
}

// measure is measureWidth when tracking is 0, measureTracked otherwise -
// the one width function wrapText and the centering helpers in layout.go
// need, regardless of whether the caller wants letter-spacing.
func measure(face font.Face, text string, tracking float64) int {
	if tracking == 0 {
		return measureWidth(face, text)
	}
	return measureTracked(face, text, tracking)
}

// drawText draws a single line of text with its top-left corner at (x, y)
// using the given style/size/color, and returns the y coordinate
// immediately below it (based on the face's line height), so callers can
// stack lines without recomputing line heights by hand.
func (c *canvas) drawText(text string, x, y int, style fontStyle, size float64, col color.Color) int {
	face := c.face(style, size)
	m := face.Metrics()
	if text != "" {
		d := &font.Drawer{
			Dst:  c.img,
			Src:  image.NewUniform(col),
			Face: face,
			Dot:  fixed.P(x, 0),
		}
		d.Dot.Y = fixed.I(y) + m.Ascent
		d.DrawString(text)
	}
	return y + m.Height.Ceil()
}

// drawTracked behaves like drawText but adds letter-spacing (tracking, in
// px) between glyphs, for the mockup's many tracked uppercase labels.
// font.Drawer has no native letter-spacing support, so this draws rune by
// rune and manually advances the pen by (glyph advance + tracking).
func (c *canvas) drawTracked(text string, x, y int, style fontStyle, size, tracking float64, col color.Color) int {
	face := c.face(style, size)
	m := face.Metrics()
	if text != "" {
		d := &font.Drawer{
			Dst:  c.img,
			Src:  image.NewUniform(col),
			Face: face,
		}
		dot := fixed.P(x, 0)
		dot.Y = fixed.I(y) + m.Ascent
		trackFixed := fixed.Int26_6(tracking * 64)
		prev := rune(-1)
		for _, r := range text {
			if prev >= 0 {
				dot.X += face.Kern(prev, r)
			}
			d.Dot = dot
			d.DrawString(string(r))
			adv, ok := face.GlyphAdvance(r)
			if !ok {
				adv, _ = face.GlyphAdvance(' ')
			}
			dot.X += adv + trackFixed
			prev = r
		}
	}
	return y + m.Height.Ceil()
}

// drawLine draws one line of text, routing to drawText (better kerning,
// via font.Drawer.DrawString) when tracking is 0, or drawTracked when the
// caller wants letter-spacing. This is the one entry point layout.go's
// component drawers use, so they never have to choose between the two
// themselves.
func (c *canvas) drawLine(text string, x, y int, style fontStyle, size, tracking float64, col color.Color) int {
	if tracking == 0 {
		return c.drawText(text, x, y, style, size, col)
	}
	return c.drawTracked(text, x, y, style, size, tracking, col)
}

// drawCenteredLine behaves like drawLine but horizontally centers the text
// within the full canvas width.
func (c *canvas) drawCenteredLine(text string, y int, style fontStyle, size, tracking float64, col color.Color) int {
	width := measure(c.face(style, size), text, tracking)
	x := (CanvasWidth - width) / 2
	if x < 0 {
		x = 0
	}
	return c.drawLine(text, x, y, style, size, tracking, col)
}

// drawBlock wraps text to fit within areaW, draws up to maxLines of it
// (dropping any remainder rather than overflowing past the canvas edge or
// clipping mid-sentence) either left-aligned at areaX or centered within
// [areaX, areaX+areaW), and returns the y coordinate below the last drawn
// line. This is the one wrapping/multi-line primitive every higher-level
// component in layout.go builds on (hero copy, card headlines, tile
// labels, empty-state messages, ...), so alignment/wrapping behaves
// identically everywhere.
func (c *canvas) drawBlock(text string, areaX, areaW, y int, style fontStyle, size, tracking float64, col color.Color, maxLines, lineGap int, centered bool) int {
	face := c.face(style, size)
	lines := wrapText(face, text, areaW, tracking)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for _, line := range lines {
		lx := areaX
		if centered {
			w := measure(face, line, tracking)
			lx = areaX + (areaW-w)/2
			if lx < areaX {
				lx = areaX
			}
		}
		y = c.drawLine(line, lx, y, style, size, tracking, col)
		y += lineGap
	}
	return y
}

// drawCenteredBlock is drawBlock centered within the full canvas width.
func (c *canvas) drawCenteredBlock(text string, y int, style fontStyle, size, tracking float64, col color.Color, maxWidthPx, maxLines, lineGap int) int {
	areaX := (CanvasWidth - maxWidthPx) / 2
	return c.drawBlock(text, areaX, maxWidthPx, y, style, size, tracking, col, maxLines, lineGap, true)
}

// drawLeftBlock is drawBlock left-aligned at x.
func (c *canvas) drawLeftBlock(text string, x, y int, style fontStyle, size, tracking float64, col color.Color, maxWidthPx, maxLines, lineGap int) int {
	return c.drawBlock(text, x, maxWidthPx, y, style, size, tracking, col, maxLines, lineGap, false)
}

// drawCenteredInRect draws a single line of (optionally tracked) text
// centered both horizontally and vertically within r - used for the small
// "FOTO" placeholder label inside the torró-card photo box.
func (c *canvas) drawCenteredInRect(r image.Rectangle, text string, style fontStyle, size, tracking float64, col color.Color) {
	face := c.face(style, size)
	w := measure(face, text, tracking)
	h := face.Metrics().Height.Ceil()
	x := r.Min.X + (r.Dx()-w)/2
	y := r.Min.Y + (r.Dy()-h)/2
	c.drawLine(text, x, y, style, size, tracking, col)
}

// wrapText greedily wraps text into lines that fit within maxWidthPx at the
// given style/size/tracking, breaking on word boundaries. It never splits a
// single word, so a word wider than maxWidthPx will still overflow on its
// own line - acceptable for a card with short, human-authored copy.
func wrapText(face font.Face, text string, maxWidthPx int, tracking float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		candidate := current + " " + w
		if measure(face, candidate, tracking) > maxWidthPx {
			lines = append(lines, current)
			current = w
			continue
		}
		current = candidate
	}
	lines = append(lines, current)

	return lines
}

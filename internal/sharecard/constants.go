package sharecard

import "image/color"

// ---------------------------------------------------------------------
// Rendering constants
//
// Everything visual lives in this one file on purpose: canvas geometry,
// colors, spacing and font scaling. The data-assembly and PNG-encoding
// plumbing in sharecard.go should never need to change when the brand
// pass lands — only the values below should. Keep it that way: don't let
// magic numbers leak into sharecard.go or text.go.
//
// TODO(design-pass): this whole file is a placeholder. Colors mirror the
// current public/css/main.css tokens and layout is a plain stacked-band
// design chosen for legibility, not looks. Swap freely once locked brand
// mockups exist.
// ---------------------------------------------------------------------

// Canvas dimensions match the Instagram/WhatsApp Stories aspect ratio (9:16).
const (
	CanvasWidth  = 1080
	CanvasHeight = 1920
)

// Layout bands, in pixels.
const (
	headerHeight    = 300
	footerHeight    = 220
	contentPaddingX = 90
	contentTop      = headerHeight + 160
)

// Brand colors, mirrored from the --color-* custom properties in
// public/css/main.css.
var (
	colorPrimary    = color.RGBA{R: 0x4E, G: 0x00, B: 0x11, A: 0xFF} // --color-primary
	colorBackground = color.RGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF} // --color-background
	colorWhite      = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // --color-white
	colorText       = color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF} // --color-text
	colorTextLight  = color.RGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xFF} // --color-text-light
)

// Font scale factors. basicfont.Face7x13 is a fixed 7x13 bitmap glyph; we
// blow it up with nearest-neighbor scaling to get readable sizes at story
// resolution (see text.go). No TTF is embedded on purpose — sourcing and
// licensing a real display font is a design-pass decision, not this one.
const (
	scaleWordmark = 6
	scaleName     = 5
	scaleLabel    = 3
	scaleBody     = 3
	scaleFooter   = 3
)

// Vertical spacing between drawn lines/blocks.
const (
	lineGapName  = 24
	lineGapLabel = 40
	blockGap     = 70
)

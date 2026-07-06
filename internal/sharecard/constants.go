package sharecard

import "image/color"

// ---------------------------------------------------------------------
// Rendering constants
//
// Everything visual lives in this one file on purpose: canvas geometry,
// colors, spacing and type scale. The shared drawing engine in canvas.go
// and the per-variant frame builders in sharecard.go/wrapped.go/
// presskit.go should never need to change when the brand pass moves on
// to its next iteration — only the values below should.
//
// This is the "dark cacau/gold" design from
// docs/design-deliverables/Torrorendum Story Card.dc.html, replacing the
// v1 engine's plain placeholder palette entirely.
// ---------------------------------------------------------------------

// Canvas dimensions match the Instagram/WhatsApp Stories aspect ratio (9:16),
// and the mockup's own $preview size, so every pixel value lifted from the
// mockup's inline styles below applies at 1:1 scale.
const (
	CanvasWidth  = 1080
	CanvasHeight = 1920
)

// Background gradient stops, from docs/design-deliverables's
// linear-gradient(165deg, #241812 0%, #332119 45%, #4A2130 100%). 165deg is
// close enough to straight down (180deg) that a top-to-bottom vertical
// gradient is a faithful simplification of the diagonal one - not worth the
// extra complexity of a true diagonal gradient for a background wash.
var (
	colorBgTop    = color.RGBA{R: 0x24, G: 0x18, B: 0x12, A: 0xFF}
	colorBgMid    = color.RGBA{R: 0x33, G: 0x21, B: 0x19, A: 0xFF}
	colorBgBottom = color.RGBA{R: 0x4A, G: 0x21, B: 0x30, A: 0xFF}
)

// Core brand colors.
//
// These are all color.NRGBA (straight, non-premultiplied alpha), not
// color.RGBA, on purpose: color.RGBA's own RGBA() method treats its fields
// as already alpha-premultiplied (per image/color's doc comment) and just
// bit-replicates them, so a color.RGBA{R: 0xFF, ..., A: 0x29} is *invalid*
// premultiplied data (R can't exceed A) and blends wrongly - too bright,
// with a visible fringe - through both image/draw's Over operator and
// font.Drawer's glyph-mask compositing (which is itself Over underneath).
// color.NRGBA is the standard type for exactly this "straight alpha"
// case and premultiplies correctly on the way in, which is what every
// translucent fill/text color below needs. Fully-opaque colors (A: 0xFF)
// behave identically either way, but everything here is NRGBA for
// consistency and so nothing has to be re-typed the day it gains an alpha.
var (
	colorCream      = color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: 0xFF} // #FFFAF1 - headline text, card background
	colorGold       = color.NRGBA{R: 0xEF, G: 0xC2, B: 0x6E, A: 0xFF} // #EFC26E - accent/highlight
	colorCardInk    = color.NRGBA{R: 0x3A, G: 0x2A, B: 0x1C, A: 0xFF} // #3A2A1C - dark text on the cream card
	colorTagInk     = color.NRGBA{R: 0xB9, G: 0x6F, B: 0x26, A: 0xFF} // #B96F26 - card tag/category line
	colorMetaInk    = color.NRGBA{R: 0x6B, G: 0x5A, B: 0x44, A: 0xFF} // #6B5A44 - small labels on the cream card
	colorHeadToHead = color.NRGBA{R: 0x8A, G: 0x26, B: 0x38, A: 0xFF} // #8A2638 - the "en duels" stat value
	colorCardBg     = color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: 0xFF} // #FFFAF1 - result/info card fill
	colorTileBg     = color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: 0x0F} // rgba(255,250,241,.06)
	colorTileBorder = color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: 0x24} // rgba(255,250,241,.14)
	colorDotDim     = color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: 0x29} // rgba(255,250,241,.16)
	colorPhotoA     = color.NRGBA{R: 0xF0, G: 0xDF, B: 0xC2, A: 0xFF} // card photo placeholder stripes
	colorPhotoB     = color.NRGBA{R: 0xE2, G: 0xCB, B: 0x9E, A: 0xFF}
	colorPhotoInk   = color.NRGBA{R: 0x8A, G: 0x74, B: 0x58, A: 0xFF}
)

// creamAlpha returns colorCream at a given straight alpha (0-255), for the
// many translucent-cream text/line treatments the mockup uses (kicker
// label, footer tagline, hashtag, sponsor line, dot-grid caption, divider
// rule...). See the color.NRGBA doc comment above for why straight (not
// premultiplied) alpha matters here.
func creamAlpha(a uint8) color.NRGBA {
	return color.NRGBA{R: 0xFF, G: 0xFA, B: 0xF1, A: a}
}

// goldAlpha returns colorGold at a given straight alpha, for the ambient
// glow/ring and pill borders.
func goldAlpha(a uint8) color.NRGBA {
	return color.NRGBA{R: 0xEF, G: 0xC2, B: 0x6E, A: a}
}

// Layout: fixed bands, in pixels. The header is anchored to the top of the
// canvas and the footer to the bottom (footerTop); the hero/dots/tiles/
// divider/cards flow top-down from contentTop in between. Values below are
// lifted directly from the mockup's inline padding/margin-top
// declarations, except footerTop (the mockup pins its footer with
// `margin-top:auto` in a flex column, which has no direct fixed-canvas
// equivalent - see layout.go's drawFooter).
const (
	headerPaddingTop = 60
	paddingX         = 72 // header's horizontal padding
	heroPaddingX     = 92 // hero block's horizontal padding
	cardPaddingX     = 88 // divider/card/tile-grid horizontal padding
	dotsPaddingX     = 118

	contentTop = 240  // where the hero block starts, below the header row
	footerTop  = 1500 // where the footer block starts

	headerRowHeight      = 40
	headerPillPadX       = 22
	headerDotRadius      = 4
	headerGap            = 10
	wordmarkRingDiameter = 34
	wordmarkRingStroke   = 3
	wordmarkDotRadius    = 5

	heroLineGap  = 6 // extra gap between wrapped lines within one hero block
	heroGapTiny  = 8
	heroGapSmall = 16
	heroGapMed   = 28

	sectionGap = 20 // between the hero and whatever section follows it (dots/tiles)

	dotSize       = 15
	dotGap        = 10
	dotCols       = 20
	dotCaptionGap = 22

	tileGap            = 18
	tilePadXWide       = 28 // 2-col (Wrapped) tiles
	tilePadXCentered   = 16 // 3-col (press kit) tiles
	tilePadY           = 26
	tileHeightWide     = 148
	tileHeightCentered = 128
	tileValueLabelGap  = 8
	tileLabelLineGap   = 4

	dividerGapY = 40

	cardRadius      = 40
	cardBorderWidth = 2
	cardPadX        = 44
	cardPadY        = 40
	cardPhotoW      = 140
	cardPhotoH      = 196
	cardPhotoRadius = 22
	cardPhotoStripe = 12
	cardHeadlineGap = 10
	cardStatsGapTop = 28
	cardStatsPadTop = 24
	cardFootnoteGap = 24
	cardColumnGap   = 8 // between a stat column's value and its label

	footerGapA  = 22 // after the tagline line
	footerGapB  = 20 // after the QR box
	footerGapC  = 16 // after the hashtag
	sponsorPadX = 22
	sponsorPadY = 9
)

// Type scale, in px (passed straight through as opentype.FaceOptions.Size
// at 72 DPI, so 1 unit == 1px). Names mirror the mockup's font-size values.
const (
	sizeKicker    = 15.0
	sizeWordmark  = 26.0
	sizeHeroIntro = 26.0
	sizeHeroLabel = 24.0
	sizeHeroBig   = 190.0 // mockup uses 224px; trimmed slightly so 3-4
	// digit values (e.g. "842.310") keep clear of
	// the canvas edges without a dynamic-size pass
	sizeHeroTagline   = 38.0
	sizeHeroPill      = 21.0
	sizeDotCaption    = 20.0
	sizeTileValueLg   = 38.0
	sizeTileValueSm   = 32.0
	sizeTileLabel     = 16.0
	sizeDivider       = 21.0
	sizeCardHeadline  = 42.0
	sizeCardTag       = 18.0
	sizeCardStatVal   = 36.0
	sizeCardStatLbl   = 15.0
	sizeCardFootnote  = 25.0
	sizeFooterTagline = 23.0
	sizeHashtag       = 21.0
	sizeSponsor       = 14.0
)

// Letter-tracking (extra space added between glyphs), in px, for the
// mockup's many letter-spaced uppercase labels.
const (
	trackKicker     = 2.0
	trackHeroLbl    = 5.0
	trackPill       = 2.5
	trackDivider    = 3.0
	trackTileLbl    = 0.4
	trackTag        = 0.4
	trackStatLbl    = 0.4
	trackHashtag    = 1.5
	trackSponsor    = 0.4
	trackDotCaption = 0.4
)

// QR placeholder geometry (see canvas.go's drawQR): an 8x8 grid of 16px
// cells inside a 128px box, matching the mockup's fake-but-legible QR
// pattern exactly (it isn't a real scannable code there either).
const (
	qrBoxSize  = 128
	qrPad      = 14
	qrGrid     = 8
	qrCellSize = 16
	qrCellFill = 13
	qrFinder   = 48
)

// sponsorPlaceholder is the mockup's literal placeholder copy for the
// footer's dashed sponsor line - there is no real sponsor yet, and this
// package has no way to know when one is signed, so every variant that
// shows the sponsor row shows this exact text (see the mockup's own
// showSponsor default, which is unconditional across all three variants).
const sponsorPlaceholder = "amb la col·laboració de — [espai reservat]"

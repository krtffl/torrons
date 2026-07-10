// Package sharecard renders every 1080x1920 ("Stories" aspect ratio) PNG
// this app shares to Instagram/WhatsApp: the plain per-user result card
// (GET /share/card.png), the personal "Torrorèndum Wrapped" recap (GET
// /wrapped/card.png), and the aggregate press-kit one-pager (GET
// /press-kit/card.png). It knows nothing about HTTP, cookies, or the
// database — callers (internal/http's Handler) assemble a variant's Data
// value from domain data and hand it to that variant's Render function.
//
// # One shared engine, several variants
//
// All three variants share the exact same visual chrome: the dark
// cacau/gold canvas (see constants.go for the full palette, lifted from
// docs/design-deliverables/Torrorendum Story Card.dc.html), the header
// badge+wordmark, a hero stat block, an optional dot-grid or stat-tile
// visualization, a divider, zero or more rounded "info cards", and the
// footer (short link, QR, hashtag, sponsor placeholder).
//
// That chrome lives here, in canvas.go/shapes.go/text.go, as one
// draw-a-frame engine. Each variant's exported Data type (Data in
// sharecard.go, WrappedData in wrapped.go, PressKitData in presskit.go)
// only has to do one thing: translate its own domain-shaped fields into a
// frame value (see toFrame in each of those files) and hand it to
// renderFrame. Adding a future variant - the roadmap's "reveal" persona
// card is the next one planned - means writing a new Data type and a
// toFrame method; nothing in this file needs to change, because frame
// already has to accommodate arbitrary combinations of these same pieces
// (a dot-grid variant with no stat tiles, a multi-card variant with
// neither, etc).
//
// # Fonts
//
// The v1 engine (see git history) used golang.org/x/image/font/basicfont,
// a built-in 7x13 bitmap face covering printable ASCII only, plus an
// accent-stripping fallback for Catalan diacritics that fell outside it.
// This engine embeds two real outline fonts instead (see text.go and
// assets/fonts/README.md) and draws every string as-is: no folding, no
// stripping, full Catalan coverage.
package sharecard

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"time"

	"golang.org/x/image/font"
)

// ---------------------------------------------------------------------
// Render concurrency gate
// ---------------------------------------------------------------------
//
// Every card render allocates a full CanvasWidth x CanvasHeight RGBA
// (~8.3MB for the 1080x1920 cards) and pegs a CPU core for the duration of
// the paint+encode. With no bound, a burst of concurrent PNG requests can
// allocate dozens of these at once and OOM/saturate the (small) VPS. This
// package-level counting semaphore caps how many renders run at a time; the
// four PNG HTTP handlers acquire a slot before rendering and release it
// after, shedding load with HTTP 503 rather than piling up when the cap is
// hit (see TryAcquireRenderSlot).

// RenderSlotWait bounds how long a caller waits for a render slot before
// giving up and shedding load. Kept short so a burst sheds fast instead of
// queueing requests that would themselves time out; a client that gets 503
// is told (via Retry-After) to come back shortly.
const RenderSlotWait = 250 * time.Millisecond

// renderSlots is the counting semaphore: a buffered channel with one slot
// per allowed concurrent render. A send acquires a slot, a receive releases
// it. Sized to GOMAXPROCS (the renders are CPU-bound) with a floor of 2 so
// even a single-core box still serves two callers rather than fully
// serializing.
var renderSlots = make(chan struct{}, maxRenderConcurrency())

func maxRenderConcurrency() int {
	if n := runtime.GOMAXPROCS(0); n > 2 {
		return n
	}
	return 2
}

// TryAcquireRenderSlot waits up to RenderSlotWait (deriving its deadline
// from parent, so a cancelled request gives up immediately) for a render
// slot. It returns true if a slot was acquired - in which case the caller
// MUST pair it with exactly one ReleaseRenderSlot(), typically deferred -
// or false if the cap is saturated, in which case the caller should shed
// load (HTTP 503) and NOT call ReleaseRenderSlot.
func TryAcquireRenderSlot(parent context.Context) bool {
	ctx, cancel := context.WithTimeout(parent, RenderSlotWait)
	defer cancel()
	select {
	case renderSlots <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// ReleaseRenderSlot returns a slot acquired via a successful
// TryAcquireRenderSlot. Call it exactly once per true acquisition.
func ReleaseRenderSlot() {
	<-renderSlots
}

// pngEncoder is the shared PNG encoder for every card. It uses BestSpeed
// compression on purpose: png.DefaultCompression accounted for ~68% of a
// render's CPU, and share cards are ephemeral social images regenerated on
// demand, so trading a modestly larger byte size for markedly less CPU is
// the right call. png.Encoder is safe for concurrent use when only
// CompressionLevel is set (Encode never mutates it), so one package-level
// value serves all renders.
var pngEncoder = png.Encoder{CompressionLevel: png.BestSpeed}

// canvas wraps the RGBA image being built plus a font.Face cache.
//
// font.Face (see golang.org/x/image/font/opentype) is explicitly
// documented as "not safe to use concurrently" - each Face owns mutable
// scratch state (a glyph-loading buffer and rasterizer) that a concurrent
// second render would corrupt. So: a canvas, and its face cache, must
// never be shared across goroutines. newCanvas is called once per Render
// call (i.e. once per HTTP request), and canvases are never stored
// anywhere longer-lived than that - the underlying *sfnt.Font values
// (sansBoldFont/serifItalicFont in text.go) are parsed once at package
// init and ARE safe to share, since building a Face from them is cheap and
// every canvas builds its own.
type canvas struct {
	img   *image.RGBA
	faces map[faceKey]font.Face
}

func newCanvas() *canvas {
	c := &canvas{
		img:   image.NewRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight)),
		faces: make(map[faceKey]font.Face),
	}
	c.paintBackground()
	c.paintDecoration()
	return c
}

func (c *canvas) face(style fontStyle, size float64) font.Face {
	key := faceKey{style, size}
	if f, ok := c.faces[key]; ok {
		return f
	}
	f := newFace(style, size)
	c.faces[key] = f
	return f
}

func (c *canvas) encode() ([]byte, error) {
	var buf bytes.Buffer
	if err := pngEncoder.Encode(&buf, c.img); err != nil {
		return nil, fmt.Errorf("sharecard: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

// bgColorAt returns the background gradient's color at row y, so small
// decorations that sit on top of the gradient (the wordmark ring's
// "punched" center, the QR box's rounded corners) can blend seamlessly
// with whatever the gradient looks like at that height instead of using a
// flat color that would show a seam.
func bgColorAt(y int) color.NRGBA {
	t := float64(y) / float64(CanvasHeight)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	// Two-segment lerp: colorBgTop -> colorBgMid at 45%, then
	// colorBgMid -> colorBgBottom for the remaining 55%, matching the
	// mockup's `linear-gradient(165deg, #241812 0%, #332119 45%, #4A2130
	// 100%)` stops (see constants.go on why a vertical gradient is a
	// faithful simplification of the 165deg diagonal one).
	if t <= 0.45 {
		return lerpRGB(colorBgTop, colorBgMid, t/0.45)
	}
	return lerpRGB(colorBgMid, colorBgBottom, (t-0.45)/0.55)
}

func lerpRGB(a, b color.RGBA, t float64) color.NRGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return color.NRGBA{R: l(a.R, b.R), G: l(a.G, b.G), B: l(a.B, b.B), A: 0xFF}
}

func (c *canvas) paintBackground() {
	for y := 0; y < CanvasHeight; y++ {
		col := bgColorAt(y)
		rgba := color.RGBA{R: col.R, G: col.G, B: col.B, A: 0xFF}
		for x := 0; x < CanvasWidth; x++ {
			c.img.SetRGBA(x, y, rgba)
		}
	}
}

// paintDecoration paints the mockup's purely-ambient background flourishes
// (glow, dashed ring, confetti accents) - fixed positions/sizes, the same
// on every render regardless of variant or data.
func (c *canvas) paintDecoration() {
	fillGlow(c.img, CanvasWidth/2, 520, 410, 40, colorGold)
	dashedRing(c.img, CanvasWidth/2, 600, 330, 2, goldAlpha(0x48))

	fillCircle(c.img, 104, 130, 7, goldAlpha(0x59))
	fillDiamond(c.img, CanvasWidth-86, 960, 14, color.NRGBA{R: 0x8A, G: 0x26, B: 0x38, A: 0x52})
	fillCircle(c.img, 72, 1760, 6, goldAlpha(0x40))
}

// ---------------------------------------------------------------------
// frame: the fully-resolved visual content for one card render
// ---------------------------------------------------------------------

// frame is what a variant's Data type builds (see toFrame in
// sharecard.go/wrapped.go/presskit.go) and hands to renderFrame. Every
// field is optional except kicker/footer: a nil dots/tiles, an empty
// dividerLabel, or a nil-length cards slice simply isn't drawn, which is
// how the three variants end up with different amounts of content without
// needing different code paths here.
type frame struct {
	// empty, if non-nil, short-circuits the whole content flow below the
	// header: only a centered heading+sub message is drawn (still with the
	// normal header/footer around it). Used for "not enough data yet"
	// states (no votes, Wrapped not unlocked, no champion decided).
	empty *emptyMessage

	kicker string // header badge pill text, e.g. "RESUM 2026"
	hero   heroContent

	dots         *dotGrid  // nil => no dot-grid stat visualization
	tiles        *tileGrid // nil => no stat-tile grid
	dividerLabel string    // "" => no section divider drawn
	cards        []infoCard

	footer footerContent
}

type emptyMessage struct {
	heading string
	sub     string
}

// heroContent is the centered stat block directly under the header: an
// italic intro line, an optional small tracked label, the huge headline
// stat, an optional small tracked unit label, an italic tagline, and an
// optional outlined pill/badge.
type heroContent struct {
	intro      string
	labelAbove string // "" => omit
	big        string
	unitBelow  string // "" => omit
	tagline    string // "" => omit
	pill       string // "" => omit
}

// dotGrid is the percentile-style "N of 100" visualization: total dots
// arranged in a fixed-width grid, `highlight` of them drawn in gold.
type dotGrid struct {
	total     int
	highlight int
	caption   string
}

// tile is a (value, label) pair, used by tileGrid (the Wrapped-style 2/3
// column stat grid).
type tile struct {
	value string
	label string
}

type tileGrid struct {
	cols     int
	centered bool // true => 3-col press-kit style (smaller, centered text)
	tiles    []tile
}

// statColumn is a tile-like (value, label) pair for infoCard's stat
// columns, plus an explicit accent flag: unlike tileGrid (where every
// value is always gold), an infoCard's last column is sometimes a
// head-to-head score that the mockup renders in a distinct red accent
// (see the torró-result card's "EN DUELS" column) and sometimes isn't
// (e.g. Data's "TORRONS VALORATS" count) - accent says which, rather than
// layout.go guessing from column position.
type statColumn struct {
	value  string
	label  string
	accent bool
}

// infoCard is the shared rounded "cream card" component: the mockup's
// torró-result card generalizes cleanly to any headline + up to a few
// stat columns + an optional italic footnote, which is enough to express
// the torró-name-and-vote-count card (Data/PressKit) as well as Wrapped's
// contested-duel/unpopular-pick/bracket-path panels without forcing
// unrelated data through the same field names.
type infoCard struct {
	headline string
	sub      string // "" => omit (category tag line)
	photo    bool   // whether to draw the diagonal-stripe photo placeholder
	columns  []statColumn
	footnote string // "" => omit
}

type footerContent struct {
	shortLink   string
	showSponsor bool
	sponsorLine string
}

// renderFrame draws f onto a fresh canvas and returns it PNG-encoded. It
// never errors on the drawing itself; the returned error only reflects PNG
// encoding failures.
func renderFrame(f frame) ([]byte, error) {
	c := newCanvas()

	c.drawHeader(f.kicker)
	c.drawFooter(f.footer)

	if f.empty != nil {
		c.drawEmptyMessage(*f.empty)
	} else {
		y := c.drawHero(f.hero)
		if f.dots != nil {
			y += sectionGap
			y = c.drawDotGrid(*f.dots, y)
		}
		if f.tiles != nil {
			y += sectionGap
			y = c.drawTileGrid(*f.tiles, y)
		}
		if f.dividerLabel != "" {
			y = c.drawDivider(f.dividerLabel, y)
		}
		for _, card := range f.cards {
			y = c.drawInfoCard(card, y)
		}
	}

	return c.encode()
}

// Package sharecard renders a user's voting result as a 1080x1920 PNG
// ("Stories" aspect ratio) suitable for sharing to Instagram/WhatsApp.
//
// This is the v1 engine: plain data in, PNG bytes out. It knows nothing
// about HTTP, cookies, or the database — callers (see
// internal/http.Handler.shareCard) assemble a Data value from domain
// data and hand it to Render. That separation is deliberate: the roadmap
// describes a later "personality reveal" / "Wrapped" feature reusing this
// same rendering pipeline with richer data, and a separate design pass
// reworking the visuals — neither should need to touch the other.
//
// The visual treatment here is intentionally plain, using the app's
// current brand colors as placeholders (see constants.go). It is not
// meant to be the final design.
package sharecard

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

// Data is the plain input the rendering engine needs. It carries only
// what the v1 card shows: the user's top-rated torró (from their
// personal ELO leaderboard, see domain.UserEloSnapshotRepo.
// GetUserGlobalLeaderboard), that torró's rank among the torrons the
// user has rated, and their total vote count.
//
// TODO(roadmap): a future "torró personality" / archetype feature (e.g.
// "you're in the 8% who prefer bold flavors") is explicitly out of scope
// here — it needs real persona copywriting and almost certainly a
// cross-user aggregate query, not just this user's own leaderboard. This
// struct should grow a richer variant for that feature rather than
// bolting persona fields onto this v1 shape.
type Data struct {
	// HasVotes is false when the user hasn't voted yet (or has no rated
	// torrons); Render then renders a fallback "vota per generar la teva
	// targeta" card instead of a result.
	HasVotes bool

	// TotalVotes is the user's all-time vote count (domain.User.VoteCount).
	TotalVotes int

	// TopTorroName is the name of the user's highest-rated torron.
	TopTorroName string

	// TopTorroRank is the 1-based rank of TopTorroName within the user's
	// own leaderboard. By construction this is almost always 1 — it is
	// kept explicit (rather than hardcoded) so the "#N of M" framing
	// stays correct if the leaderboard ever includes ties or the caller
	// picks a torró other than the literal top one.
	TopTorroRank int

	// RatedTorronCount is how many distinct torrons the user has voted
	// on enough to have a personal rating for. Combined with
	// TopTorroRank this gives simple, honest context ("#1 of 12") in
	// place of a fabricated percentile — see package doc TODO above.
	RatedTorronCount int
}

// Render draws Data onto a CanvasWidth x CanvasHeight canvas and returns
// it PNG-encoded. It never errors on the drawing itself (there is no
// partial/invalid Data state); the returned error only reflects PNG
// encoding failures.
func Render(data Data) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))

	draw.Draw(img, img.Bounds(), image.NewUniform(colorBackground), image.Point{}, draw.Src)

	drawHeader(img)
	drawFooter(img)

	if data.HasVotes {
		drawResult(img, data)
	} else {
		drawEmptyState(img)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("sharecard: encode png: %w", err)
	}

	return buf.Bytes(), nil
}

// drawHeader paints the top band with the app wordmark.
func drawHeader(img *image.RGBA) {
	band := image.Rect(0, 0, CanvasWidth, headerHeight)
	draw.Draw(img, band, image.NewUniform(colorPrimary), image.Point{}, draw.Src)

	y := (headerHeight - textFace.Height*scaleWordmark) / 2
	drawCenteredLine(img, "TORRORENDUM", y, colorWhite, scaleWordmark)
}

// drawFooter paints the bottom band with a plain call-to-action tagline.
func drawFooter(img *image.RGBA) {
	band := image.Rect(0, CanvasHeight-footerHeight, CanvasWidth, CanvasHeight)
	draw.Draw(img, band, image.NewUniform(colorPrimary), image.Point{}, draw.Src)

	y := CanvasHeight - footerHeight + (footerHeight-textFace.Height*scaleFooter)/2
	drawCenteredLine(img, "Vota el teu preferit a torrorendum", y, colorWhite, scaleFooter)
}

// drawResult renders the "you voted, here's your result" state.
func drawResult(img *image.RGBA, data Data) {
	maxTextWidth := CanvasWidth - 2*contentPaddingX
	y := contentTop

	y = drawLine(img, "El teu torro preferit", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel

	nameLines := wrapText(data.TopTorroName, maxTextWidth, scaleName)
	const maxNameLines = 3
	if len(nameLines) > maxNameLines {
		nameLines = nameLines[:maxNameLines]
	}
	for _, line := range nameLines {
		y = drawLine(img, line, contentPaddingX, y, colorPrimary, scaleName)
		y += lineGapName
	}

	y += blockGap
	rankLine := fmt.Sprintf("Ranquing personal: #%d de %d", data.TopTorroRank, data.RatedTorronCount)
	y = drawLine(img, rankLine, contentPaddingX, y, colorText, scaleBody)

	y += blockGap
	votesLine := fmt.Sprintf("Vots totals: %d", data.TotalVotes)
	drawLine(img, votesLine, contentPaddingX, y, colorText, scaleBody)
}

// drawEmptyState renders the fallback card for users who haven't voted
// (or don't yet have a rated torron) instead of erroring.
func drawEmptyState(img *image.RGBA) {
	y := contentTop + 200

	y = drawCenteredLine(img, "Encara no has votat", y, colorText, scaleName)
	y += blockGap
	drawCenteredLine(img, "Vota per generar la teva targeta", y, colorTextLight, scaleBody)
}

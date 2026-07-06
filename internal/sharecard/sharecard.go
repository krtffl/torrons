package sharecard

import "fmt"

// Data is the plain input for the v1 per-user result card (GET
// /share/card.png): the user's top-rated torró (from their personal ELO
// leaderboard, see domain.UserEloSnapshotRepo.GetUserGlobalLeaderboard),
// that torró's rank among the torrons the user has rated, and their total
// vote count.
//
// This is the simplest of the three variants this engine renders - it
// predates the richer Wrapped/press-kit data models and isn't one of the
// mockup's documented variants (reveal/wrapped/presskit), so its hero copy
// (see toFrame) is this package's own reasonable design decision rather
// than lifted verbatim from the mockup - see the package doc in canvas.go.
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
	// own leaderboard. By construction this is almost always 1 - it is
	// kept explicit (rather than hardcoded) so the "#N of M" framing stays
	// correct if the leaderboard ever includes ties or the caller picks a
	// torró other than the literal top one.
	TopTorroRank int

	// RatedTorronCount is how many distinct torrons the user has voted on
	// enough to have a personal rating for. Combined with TopTorroRank
	// this gives simple, honest context ("#1 of 12") in place of a
	// fabricated percentile.
	//
	// TODO(roadmap): a real "torró personality"/archetype percentile (e.g.
	// "you're in the 8% who prefer bold flavors") is the roadmap's
	// "reveal" variant - it needs real persona copywriting and a
	// cross-user aggregate query, not just this user's own leaderboard,
	// and its product-owner copy review is still pending. This engine's
	// frame/dotGrid/heroContent shapes already accommodate it (see
	// canvas.go's package doc) whenever that lands.
	RatedTorronCount int
}

// Render draws data onto a CanvasWidth x CanvasHeight canvas and returns it
// PNG-encoded. It never errors on the drawing itself (there is no
// partial/invalid Data state); the returned error only reflects PNG
// encoding failures.
func Render(data Data) ([]byte, error) {
	return renderFrame(data.toFrame())
}

func (d Data) toFrame() frame {
	f := frame{
		kicker: "EL TEU RESULTAT",
		footer: footerContent{
			shortLink:   "torrorendum.cat",
			showSponsor: true,
			sponsorLine: sponsorPlaceholder,
		},
	}

	if !d.HasVotes {
		f.empty = &emptyMessage{
			heading: "Encara no has votat",
			sub:     "Vota per generar la teva targeta",
		}
		return f
	}

	f.hero = heroContent{
		intro:      fmt.Sprintf("Després de %d vots, això és el que has triat.", d.TotalVotes),
		labelAbove: "EL TEU RÀNQUING",
		big:        fmt.Sprintf("#%d", d.TopTorroRank),
		unitBelow:  fmt.Sprintf("DE %d TORRONS VOTATS", d.RatedTorronCount),
		tagline:    fmt.Sprintf("El teu preferit és %s.", d.TopTorroName),
	}
	f.dividerLabel = "EL TEU TORRÓ"
	f.cards = []infoCard{{
		headline: d.TopTorroName,
		photo:    true,
		columns: []statColumn{
			{value: fmt.Sprintf("#%d", d.TopTorroRank), label: "RÀNQUING PERSONAL"},
			{value: fmt.Sprintf("%d", d.RatedTorronCount), label: "TORRONS VALORATS"},
		},
	}}

	return f
}

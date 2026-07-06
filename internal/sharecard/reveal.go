package sharecard

import "fmt"

// RevealData is the input for the "torró personality reveal" card (GET
// /reveal/card.png, design prompt 14) - the "reveal" variant sketched out
// in docs/design-deliverables/Torrorendum Story Card.dc.html (its bespoke
// "Torrorendum Reveal Moment.dc.html" full-screen ceremony wrapper is not
// reproduced here - see reveal.html's own doc comment on that scope cut).
// It carries a user's most-voted arena persona (from domain.PersonaStats,
// itself sourced from Users.ClassVotes) plus their single most-voted-for
// torró within that arena.
//
// PersonaBadge/PersonaTagline/TopTorroTag arrive already resolved to their
// final display copy (internal/http/reveal_handler.go owns the approved
// copy table and the arenaTagForClass tag lookup) - this type's own
// toFrame only has to decide layout, not word choice, mirroring how
// WrappedData receives already-resolved torró names rather than raw IDs.
type RevealData struct {
	// HasEnoughVotes is false when the user hasn't cleared the reveal
	// unlock threshold yet (same getMinVotesForClass("5") threshold as
	// Wrapped). toFrame then builds a "not unlocked yet" empty state.
	HasEnoughVotes bool
	// VotesRemaining is meaningful only when !HasEnoughVotes.
	VotesRemaining int

	// TotalVotes is the user's all-time vote count.
	TotalVotes int

	// HasClearFavorite is false on an exact tie between two or more voting
	// arenas (domain.PersonaStats.HasClearFavorite) - the "ELS EQUILIBRATS"
	// persona. Percentile/the dot-grid visualization are only meaningful/
	// drawn when this is true - see toFrame.
	HasClearFavorite bool
	// Percentile (1-100) is meaningful only when HasClearFavorite.
	Percentile int

	// PersonaBadge is the persona's badge label (e.g. "ELS ATREVITS",
	// "ELS EQUILIBRATS").
	PersonaBadge string
	// PersonaTagline is, when HasClearFavorite, the clause that completes
	// "Ets del X% que <PersonaTagline>" (e.g. "que prefereix els sabors
	// més atrevits.", matching this variant's own default prop in the
	// mockup above) - toFrame places it directly under the "ETS DEL X%"
	// stat, never re-wrapping it itself. When !HasClearFavorite it's
	// instead a full standalone sentence (there's no percentile to
	// complete a clause about).
	PersonaTagline string

	// TopTorroName/TopTorroTag/TopTorroVotesCast/TopTorroWins/
	// TopTorroLosses describe the single torró the user voted for (won)
	// most often within their top arena (or, on a tie, across every
	// arena) - see domain.PersonaStats. TopTorroName == "" means no torró
	// card is shown at all (see toFrame) - a defensive empty state that
	// shouldn't arise in practice once HasEnoughVotes is true.
	TopTorroName      string
	TopTorroTag       string
	TopTorroVotesCast int
	TopTorroWins      int
	TopTorroLosses    int
}

// RenderReveal draws data onto a CanvasWidth x CanvasHeight canvas and
// returns it PNG-encoded. It never errors on the drawing itself; the
// returned error only reflects PNG encoding failures.
func RenderReveal(data RevealData) ([]byte, error) {
	return renderFrame(data.toFrame())
}

func (d RevealData) toFrame() frame {
	f := frame{
		kicker: "MOMENT DESTACAT",
		footer: footerContent{
			shortLink:   "torrorendum.cat",
			showSponsor: true,
			sponsorLine: sponsorPlaceholder,
		},
	}

	if !d.HasEnoughVotes {
		f.empty = &emptyMessage{
			heading: "Encara no has desbloquejat la teva revelació",
			sub:     fmt.Sprintf("Et falten %d vots", d.VotesRemaining),
		}
		return f
	}

	f.hero = heroContent{
		intro: fmt.Sprintf("Després de %d vots, el torrorèndum ha parlat de tu.", d.TotalVotes),
		pill:  d.PersonaBadge,
	}

	if d.HasClearFavorite {
		f.hero.labelAbove = "ETS DEL"
		f.hero.big = fmt.Sprintf("%d%%", d.Percentile)
		f.hero.tagline = d.PersonaTagline

		f.dots = &dotGrid{
			total:     100,
			highlight: d.Percentile,
			caption:   fmt.Sprintf("%d de cada 100 votants trien com tu", d.Percentile),
		}
	} else {
		// ELS EQUILIBRATS: there's no single arena to headline with a
		// percentile, so the hero's big-stat slot falls back to the total
		// vote count instead (the same slot/copy Wrapped's own hero uses
		// for it), and the persona sentence stands on its own rather than
		// completing "Ets del X% que..." - see
		// internal/http/reveal_handler.go's equilibratsTagline. No dot
		// grid: there's no percentile to visualize on an exact tie.
		f.hero.labelAbove = "EL TEU"
		f.hero.big = fmt.Sprintf("%d", d.TotalVotes)
		f.hero.unitBelow = "VOTS EMESOS"
		f.hero.tagline = d.PersonaTagline
	}

	if d.TopTorroName != "" {
		f.dividerLabel = "EL TEU TORRÓ"
		f.cards = []infoCard{{
			headline: d.TopTorroName,
			sub:      d.TopTorroTag,
			photo:    true,
			columns: []statColumn{
				{value: fmt.Sprintf("%d", d.TopTorroVotesCast), label: "VOTS EMESOS"},
				{value: headToHeadScore(d.TopTorroWins, d.TopTorroLosses), label: "EN DUELS", accent: true},
			},
			footnote: "Cap altre torró t'ha convençut tant.",
		}}
	}

	return f
}

// headToHeadScore formats a "wins–losses" record using the mockup's own en
// dash separator (e.g. "9–1" - see Torrorendum Story Card.dc.html's
// headToHead prop), not a plain hyphen.
func headToHeadScore(wins, losses int) string {
	return fmt.Sprintf("%d–%d", wins, losses)
}

package sharecard

import (
	"fmt"
	"time"
)

// WrappedData is the input for the "Torrorèndum Wrapped" personal recap
// card (GET /wrapped/card.png). Unlike Data (the plain v1 share card),
// this carries cross-source data: total votes, the user's most contested
// duel and most unpopular pick (both from
// domain.WrappedStatsRepo.DuelStats), and their Phase 2 knockout bracket
// participation (domain.WrappedStatsRepo.BracketPath).
type WrappedData struct {
	// HasEnoughVotes is false when the user hasn't cleared the Wrapped
	// unlock threshold yet (see getMinVotesForClass("5") in
	// internal/http/user_api.go). toFrame then builds a "not unlocked yet"
	// empty state instead of any stat.
	HasEnoughVotes bool
	// VotesRemaining is meaningful only when !HasEnoughVotes: how many
	// more votes the user needs to unlock their Wrapped.
	VotesRemaining int

	// TotalVotes is the user's all-time vote count (domain.User.VoteCount).
	TotalVotes int

	// HasContestedDuel is false if the user never voted on a pairing that
	// met the minimum-total-votes threshold (a legitimate empty state, not
	// an error - see domain.WrappedStatsRepo.DuelStats).
	HasContestedDuel    bool
	ContestedTorroAName string
	ContestedTorroBName string
	ContestedPercentA   int
	ContestedPercentB   int

	// HasUnpopularPick is false under the same conditions as
	// HasContestedDuel above (they're derived from the same dataset).
	HasUnpopularPick  bool
	UnpopularPickName string
	// UnpopularPercent is the % of the crowd that agreed with this pick -
	// low means unpopular.
	UnpopularPercent int

	// HasBracketVotes is false if the user never voted in the Phase 2
	// knockout bracket at all.
	HasBracketVotes       bool
	BracketRoundsVoted    int
	BracketMatchesDecided int
	BracketPicksCorrect   int
	// HasChampion is whether the bracket itself is fully decided yet.
	HasChampion  bool
	ChampionName string
	// MatchedChampion is meaningful only if HasChampion: did the user ever
	// pick the champion in one of their voted matches.
	MatchedChampion bool
}

// RenderWrapped draws data onto a CanvasWidth x CanvasHeight canvas and
// returns it PNG-encoded. It never errors on the drawing itself; the
// returned error only reflects PNG encoding failures.
func RenderWrapped(data WrappedData) ([]byte, error) {
	return renderFrame(data.toFrame())
}

func (d WrappedData) toFrame() frame {
	f := frame{
		// The mockup's Wrapped variant kicker is "RESUM {year}"; this
		// package has no campaign-year field to draw from (WrappedData
		// doesn't carry one), so it uses the wall-clock year at render
		// time - correct for "this year's recap" without needing a new
		// field, and harmless since this is a display label, not data used
		// in any calculation.
		kicker: fmt.Sprintf("RESUM %d", time.Now().Year()),
		footer: footerContent{
			shortLink:   "torrorendum.cat",
			showSponsor: true,
			sponsorLine: sponsorPlaceholder,
		},
	}

	if !d.HasEnoughVotes {
		f.empty = &emptyMessage{
			heading: "Encara no has desbloquejat el teu Wrapped",
			sub:     fmt.Sprintf("Et falten %d vots", d.VotesRemaining),
		}
		return f
	}

	f.hero = heroContent{
		intro:      "Un any sencer de duels. Això és el que has votat.",
		labelAbove: "EL TEU",
		big:        fmt.Sprintf("%d", d.TotalVotes),
		unitBelow:  "VOTS EMESOS",
		tagline:    d.wrappedTagline(),
		pill:       d.wrappedPill(),
	}

	f.tiles = &tileGrid{
		cols: 2,
		tiles: []tile{
			{value: fmt.Sprintf("%d", d.BracketRoundsVoted), label: "RONDES VOTADES A LA FASE FINAL"},
			{value: bracketScore(d.BracketPicksCorrect, d.BracketMatchesDecided), label: "ENCERTS A LA FASE FINAL"},
		},
	}

	if card, ok := d.featuredCard(); ok {
		f.dividerLabel = "UN MOMENT DESTACAT"
		f.cards = []infoCard{card}
	}

	return f
}

// wrappedTagline picks the hero's italic sub-line from whatever bracket
// data is actually available, honestly reflecting "not decided/voted yet"
// rather than fabricating a persona-style tagline this data model doesn't
// have (see Data.RatedTorronCount's doc comment on that TODO).
func (d WrappedData) wrappedTagline() string {
	switch {
	case d.HasChampion && d.MatchedChampion:
		return "Vas encertar el campió de la Gran Final."
	case d.HasChampion:
		return fmt.Sprintf("El campió va ser %s.", d.ChampionName)
	case d.HasBracketVotes:
		return "La Gran Final encara no s'ha decidit."
	default:
		return "Encara no has votat a la fase de knockout."
	}
}

// wrappedPill picks the hero's badge label, or "" to omit it when there's
// no bracket participation yet to badge.
func (d WrappedData) wrappedPill() string {
	if !d.HasBracketVotes {
		return ""
	}
	return fmt.Sprintf("%d RONDES VOTADES", d.BracketRoundsVoted)
}

// featuredCard picks the single richest available stat (champion result,
// then the closest duel, then the boldest pick) to feature in the card
// below the divider, in that priority order. Showing all three that could
// be present, each as its own card, would overflow a single 1080x1920
// canvas (each card needs real room for a torró name, a photo placeholder
// and a stat row) - see canvas.go's package doc on this engine choosing
// depth of content over cramming in every fact at once. ok is false when
// none of the three are available yet, in which case the caller skips the
// divider and card entirely.
func (d WrappedData) featuredCard() (infoCard, bool) {
	switch {
	case d.HasChampion:
		footnote := "El campió et va sorprendre."
		matched := "NO"
		if d.MatchedChampion {
			footnote = "Vas endevinar el campió a les teves votacions."
			matched = "SÍ"
		}
		return infoCard{
			headline: d.ChampionName,
			sub:      "CAMPIÓ DE LA GRAN FINAL",
			photo:    true,
			columns:  []statColumn{{value: matched, label: "EL VAS ENCERTAR"}},
			footnote: footnote,
		}, true

	case d.HasContestedDuel:
		return infoCard{
			headline: "El teu duel més ajustat",
			columns: []statColumn{
				{value: fmt.Sprintf("%d%%", d.ContestedPercentA), label: d.ContestedTorroAName},
				{value: fmt.Sprintf("%d%%", d.ContestedPercentB), label: d.ContestedTorroBName, accent: true},
			},
		}, true

	case d.HasUnpopularPick:
		return infoCard{
			headline: d.UnpopularPickName,
			sub:      "LA TEVA TRIA MÉS ATREVIDA",
			photo:    true,
			columns:  []statColumn{{value: fmt.Sprintf("%d%%", d.UnpopularPercent), label: "DEL PÚBLIC HI COINCIDEIX"}},
		}, true

	default:
		return infoCard{}, false
	}
}

// bracketScore formats a "correct/decided" fraction, or an em dash when no
// matches have been decided yet (avoiding a misleading "0/0").
func bracketScore(correct, decided int) string {
	if decided <= 0 {
		return "—"
	}
	return fmt.Sprintf("%d/%d", correct, decided)
}

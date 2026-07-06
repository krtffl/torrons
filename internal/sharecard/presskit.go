package sharecard

import "fmt"

// PressKitData is the input for the press-kit one-pager PNG (GET
// /press-kit/card.png): a global aggregate card (the same for every
// viewer, unlike Data/WrappedData which are per-user) summarizing the
// Phase 2 knockout bracket's result for the press page. See
// internal/http/press_handler.go.
type PressKitData struct {
	// HasChampion is false while the Gran Final hasn't been decided yet
	// (no bracket, or a bracket still in progress). toFrame then builds an
	// empty state instead of a result.
	HasChampion bool

	ChampionName  string
	ChampionVotes int
}

// RenderPressKit draws data onto a CanvasWidth x CanvasHeight canvas and
// returns it PNG-encoded. It never errors on the drawing itself; the
// returned error only reflects PNG encoding failures.
func RenderPressKit(data PressKitData) ([]byte, error) {
	return renderFrame(data.toFrame())
}

func (d PressKitData) toFrame() frame {
	f := frame{
		kicker: "RESULTATS OFICIALS",
		footer: footerContent{
			shortLink:   "torrorendum.cat",
			showSponsor: true,
			sponsorLine: sponsorPlaceholder,
		},
	}

	if !d.HasChampion {
		// Reuses the exact same Catalan wording already established on the
		// /premsa page (see press.html) so the two surfaces read
		// consistently.
		f.empty = &emptyMessage{heading: "La Gran Final encara no s'ha disputat"}
		return f
	}

	f.hero = heroContent{
		intro:     "El torrorèndum ha arribat a la Gran Final.",
		big:       fmt.Sprintf("%d", d.ChampionVotes),
		unitBelow: "VOTS GUANYADORS",
		tagline:   "El poble ha parlat.",
		pill:      fmt.Sprintf("CAMPIÓ: %s", d.ChampionName),
	}
	f.dividerLabel = "TORRÓ CAMPIÓ"
	f.cards = []infoCard{{
		headline: d.ChampionName,
		photo:    true,
		columns: []statColumn{
			{value: fmt.Sprintf("%d", d.ChampionVotes), label: "VOTS EMESOS"},
		},
	}}

	return f
}

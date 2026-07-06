package sharecard

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderRevealProducesValidCanvas(t *testing.T) {
	tests := map[string]RevealData{
		"locked - not enough votes": {
			HasEnoughVotes: false,
			VotesRemaining: 12,
		},
		"unlocked - clear favorite, full torró card": {
			HasEnoughVotes:    true,
			TotalVotes:        212,
			HasClearFavorite:  true,
			Percentile:        8,
			PersonaBadge:      "ELS ATREVITS",
			PersonaTagline:    "que prefereix els sabors més atrevits, signats per un xef.",
			TopTorroName:      "Mandarina i yuzu - Adrià Natura",
			TopTorroTag:       "ADRIÀ NATURA · EDICIÓ LIMITADA",
			TopTorroVotesCast: 16,
			TopTorroWins:      9,
			TopTorroLosses:    1,
		},
		"unlocked - Global arena persona": {
			HasEnoughVotes:    true,
			TotalVotes:        340,
			HasClearFavorite:  true,
			Percentile:        22,
			PersonaBadge:      "ELS ÀRBITRES",
			PersonaTagline:    "que no es conforma amb una categoria: vol saber qui guanya de veritat.",
			TopTorroName:      "Torró de xocolata amb ametlles senceres",
			TopTorroTag:       "GLOBAL · EL REPTE DEFINITIU",
			TopTorroVotesCast: 40,
			TopTorroWins:      28,
			TopTorroLosses:    12,
		},
		"unlocked - ELS EQUILIBRATS tie, still has a top torró": {
			HasEnoughVotes:    true,
			TotalVotes:        180,
			HasClearFavorite:  false,
			PersonaBadge:      "ELS EQUILIBRATS",
			PersonaTagline:    "Reparteix els vots per igual entre totes les categories.",
			TopTorroName:      "Torró de massapà",
			TopTorroTag:       "CLÀSSICS · L'ORIGINAL",
			TopTorroVotesCast: 5,
			TopTorroWins:      3,
			TopTorroLosses:    2,
		},
		"unlocked - no torró resolved yet (defensive empty state)": {
			HasEnoughVotes:   true,
			TotalVotes:       50,
			HasClearFavorite: true,
			Percentile:       50,
			PersonaBadge:     "ELS TRADICIONALISTES",
			PersonaTagline:   "que no necessita res més que ametlla, mel i tradició.",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := RenderReveal(data)
			if err != nil {
				t.Fatalf("RenderReveal() error = %v", err)
			}

			img, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("png.Decode() error = %v; output is not a valid PNG", err)
			}

			bounds := img.Bounds()
			if bounds.Dx() != CanvasWidth || bounds.Dy() != CanvasHeight {
				t.Fatalf("got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), CanvasWidth, CanvasHeight)
			}
		})
	}
}

func TestHeadToHeadScore(t *testing.T) {
	if got := headToHeadScore(9, 1); got != "9–1" {
		t.Errorf("headToHeadScore(9, 1) = %q, want \"9–1\"", got)
	}
	if got := headToHeadScore(0, 0); got != "0–0" {
		t.Errorf("headToHeadScore(0, 0) = %q, want \"0–0\"", got)
	}
}

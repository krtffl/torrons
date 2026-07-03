package sharecard

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderWrappedProducesValidCanvas(t *testing.T) {
	tests := map[string]WrappedData{
		"locked - not enough votes": {
			HasEnoughVotes: false,
			VotesRemaining: 37,
		},
		"unlocked - no sub-stats yet": {
			HasEnoughVotes:   true,
			TotalVotes:       50,
			HasContestedDuel: false,
			HasUnpopularPick: false,
			HasBracketVotes:  false,
		},
		"unlocked - full stats, bracket in progress": {
			HasEnoughVotes:        true,
			TotalVotes:            312,
			HasContestedDuel:      true,
			ContestedTorroAName:   "Torró d'Ametlla i Xocolata Negra 70%",
			ContestedTorroBName:   "Torró de Xocolata amb Ametlles Senceres",
			ContestedPercentA:     51,
			ContestedPercentB:     49,
			HasUnpopularPick:      true,
			UnpopularPickName:     "Torró Dur de Torrons Vicens",
			UnpopularPercent:      12,
			HasBracketVotes:       true,
			BracketRoundsVoted:    3,
			BracketMatchesDecided: 3,
			BracketPicksCorrect:   2,
			HasChampion:           false,
		},
		"unlocked - full stats, champion decided and matched": {
			HasEnoughVotes:        true,
			TotalVotes:            500,
			HasContestedDuel:      true,
			ContestedTorroAName:   "Torró A",
			ContestedTorroBName:   "Torró B",
			ContestedPercentA:     60,
			ContestedPercentB:     40,
			HasUnpopularPick:      true,
			UnpopularPickName:     "Torró Menys Popular",
			UnpopularPercent:      5,
			HasBracketVotes:       true,
			BracketRoundsVoted:    4,
			BracketMatchesDecided: 4,
			BracketPicksCorrect:   4,
			HasChampion:           true,
			ChampionName:          "Torró Campió del Torrorèndum",
			MatchedChampion:       true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := RenderWrapped(data)
			if err != nil {
				t.Fatalf("RenderWrapped() error = %v", err)
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

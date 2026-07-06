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
		"unlocked - champion decided but not matched": {
			HasEnoughVotes:        true,
			TotalVotes:            212,
			HasBracketVotes:       true,
			BracketRoundsVoted:    4,
			BracketMatchesDecided: 4,
			BracketPicksCorrect:   1,
			HasChampion:           true,
			ChampionName:          "Torró Sorpresa",
			MatchedChampion:       false,
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

func TestWrappedFeaturedCardPriority(t *testing.T) {
	tests := []struct {
		name         string
		data         WrappedData
		wantCard     bool
		wantHeadline string
	}{
		{
			name:     "nothing available yet",
			data:     WrappedData{HasEnoughVotes: true},
			wantCard: false,
		},
		{
			name: "unpopular pick only",
			data: WrappedData{
				HasEnoughVotes:    true,
				HasUnpopularPick:  true,
				UnpopularPickName: "Torró Atrevit",
			},
			wantCard:     true,
			wantHeadline: "Torró Atrevit",
		},
		{
			name: "contested duel beats unpopular pick",
			data: WrappedData{
				HasEnoughVotes:      true,
				HasUnpopularPick:    true,
				UnpopularPickName:   "Torró Atrevit",
				HasContestedDuel:    true,
				ContestedTorroAName: "A",
				ContestedTorroBName: "B",
			},
			wantCard:     true,
			wantHeadline: "El teu duel més ajustat",
		},
		{
			name: "champion beats everything",
			data: WrappedData{
				HasEnoughVotes:      true,
				HasUnpopularPick:    true,
				HasContestedDuel:    true,
				HasChampion:         true,
				ChampionName:        "Torró Campió",
				ContestedTorroAName: "A",
				ContestedTorroBName: "B",
			},
			wantCard:     true,
			wantHeadline: "Torró Campió",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, ok := tt.data.featuredCard()
			if ok != tt.wantCard {
				t.Fatalf("featuredCard() ok = %v, want %v", ok, tt.wantCard)
			}
			if ok && card.headline != tt.wantHeadline {
				t.Errorf("featuredCard() headline = %q, want %q", card.headline, tt.wantHeadline)
			}
		})
	}
}

func TestBracketScore(t *testing.T) {
	if got := bracketScore(0, 0); got != "—" {
		t.Errorf("bracketScore(0, 0) = %q, want em dash", got)
	}
	if got := bracketScore(2, 3); got != "2/3" {
		t.Errorf("bracketScore(2, 3) = %q, want \"2/3\"", got)
	}
}

func TestDotHighlightSetSpreadsEvenly(t *testing.T) {
	set := dotHighlightSet(100, 14)
	if len(set) != 14 {
		t.Fatalf("dotHighlightSet(100, 14): got %d highlighted dots, want 14", len(set))
	}

	// Edge cases: 0 total/highlight must not panic or divide by zero.
	if s := dotHighlightSet(0, 10); len(s) != 0 {
		t.Errorf("dotHighlightSet(0, 10) = %d entries, want 0", len(s))
	}
	if s := dotHighlightSet(100, 0); len(s) != 0 {
		t.Errorf("dotHighlightSet(100, 0) = %d entries, want 0", len(s))
	}
	if s := dotHighlightSet(100, 150); len(s) != 100 {
		t.Errorf("dotHighlightSet(100, 150) = %d entries, want 100 (clamped)", len(s))
	}
}

package sharecard

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderProducesValidCanvas(t *testing.T) {
	tests := map[string]Data{
		"with votes": {
			HasVotes:         true,
			TotalVotes:       137,
			TopTorroName:     "Torró d'Ametlla i Xocolata Negra 70%",
			TopTorroRank:     1,
			RatedTorronCount: 42,
		},
		"no votes yet": {
			HasVotes: false,
		},
		"short name, single vote": {
			HasVotes:         true,
			TotalVotes:       1,
			TopTorroName:     "Torró Dur",
			TopTorroRank:     1,
			RatedTorronCount: 1,
		},
		"long catalan name with diacritics": {
			HasVotes:         true,
			TotalVotes:       9999,
			TopTorroName:     "Torró de Xocolata amb Ametlles i Praliné, l'Especialitat més Atrevida",
			TopTorroRank:     3,
			RatedTorronCount: 87,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := Render(data)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
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

// TestRenderEmptyStateOmitsResult checks that the "no votes" fallback
// really takes the empty-message path (no hero/card content computed from
// zero-value fields) rather than happening to render harmlessly.
func TestRenderEmptyStateOmitsResult(t *testing.T) {
	f := Data{HasVotes: false}.toFrame()
	if f.empty == nil {
		t.Fatal("toFrame() with HasVotes=false: expected empty state, got nil")
	}
	if len(f.cards) != 0 {
		t.Errorf("toFrame() with HasVotes=false: expected no cards, got %d", len(f.cards))
	}
}

func TestDataToFrameResultCard(t *testing.T) {
	data := Data{
		HasVotes:         true,
		TotalVotes:       50,
		TopTorroName:     "Torró de Xocolata",
		TopTorroRank:     2,
		RatedTorronCount: 10,
	}
	f := data.toFrame()

	if f.empty != nil {
		t.Fatal("toFrame() with HasVotes=true: expected no empty state")
	}
	if len(f.cards) != 1 {
		t.Fatalf("toFrame(): expected exactly 1 card, got %d", len(f.cards))
	}
	if f.cards[0].headline != data.TopTorroName {
		t.Errorf("card headline = %q, want %q", f.cards[0].headline, data.TopTorroName)
	}
	if len(f.cards[0].columns) != 2 {
		t.Fatalf("card columns = %d, want 2", len(f.cards[0].columns))
	}
}

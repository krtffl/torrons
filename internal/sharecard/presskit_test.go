package sharecard

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderPressKitProducesValidCanvas(t *testing.T) {
	tests := map[string]PressKitData{
		"no champion yet": {
			HasChampion: false,
		},
		"champion decided": {
			HasChampion:   true,
			ChampionName:  "Torró d'Ametlla i Xocolata Negra 70%",
			ChampionVotes: 1284,
		},
		"champion decided, huge vote total": {
			HasChampion:   true,
			ChampionName:  "Torró de Xocolata",
			ChampionVotes: 842310,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := RenderPressKit(data)
			if err != nil {
				t.Fatalf("RenderPressKit() error = %v", err)
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

func TestPressKitEmptyState(t *testing.T) {
	f := PressKitData{HasChampion: false}.toFrame()
	if f.empty == nil {
		t.Fatal("toFrame() with HasChampion=false: expected empty state")
	}
}

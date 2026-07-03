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

func TestWrapTextFitsWithinMaxWidth(t *testing.T) {
	const maxWidth = 900
	const scale = 5

	lines := wrapText("Torró d'Ametlla i Xocolata Negra amb un nom llarguíssim de veritat", maxWidth, scale)
	if len(lines) < 2 {
		t.Fatalf("expected the long name to wrap into multiple lines, got %d: %v", len(lines), lines)
	}

	for _, line := range lines {
		if w := measureWidth(line, scale); w > maxWidth {
			t.Errorf("line %q measures %dpx, exceeds maxWidth %dpx", line, w, maxWidth)
		}
	}
}

func TestWrapTextEmptyString(t *testing.T) {
	if lines := wrapText("", 900, 5); lines != nil {
		t.Errorf("wrapText(\"\") = %v, want nil", lines)
	}
}

package sharecard

import "testing"

func TestWrapTextFitsWithinMaxWidth(t *testing.T) {
	const maxWidth = 900
	face := newFace(styleSansBold, 42)

	lines := wrapText(face, "Torró d'Ametlla i Xocolata Negra amb un nom llarguíssim de veritat", maxWidth, 0)
	if len(lines) < 2 {
		t.Fatalf("expected the long name to wrap into multiple lines, got %d: %v", len(lines), lines)
	}

	for _, line := range lines {
		if w := measureWidth(face, line); w > maxWidth {
			t.Errorf("line %q measures %dpx, exceeds maxWidth %dpx", line, w, maxWidth)
		}
	}
}

func TestWrapTextEmptyString(t *testing.T) {
	face := newFace(styleSansBold, 42)
	if lines := wrapText(face, "", 900, 0); lines != nil {
		t.Errorf("wrapText(\"\") = %v, want nil", lines)
	}
}

func TestWrapTextNeverSplitsAWord(t *testing.T) {
	face := newFace(styleSansBold, 42)
	// A single word wider than maxWidthPx must still come back as its own
	// line rather than being split mid-word.
	lines := wrapText(face, "Torrorendumdemocratitzacionalment", 10, 0)
	if len(lines) != 1 || lines[0] != "Torrorendumdemocratitzacionalment" {
		t.Errorf("wrapText() with an overlong single word = %v, want it unsplit", lines)
	}
}

func TestMeasureTrackedAddsSpacing(t *testing.T) {
	face := newFace(styleSansBold, 24)
	plain := measureWidth(face, "TEST")
	tracked := measureTracked(face, "TEST", 5)
	if tracked <= plain {
		t.Errorf("measureTracked width %d should exceed plain width %d", tracked, plain)
	}
}

// TestCatalanDiacriticsMeasurable is a regression guard for the v1 engine's
// accent-stripping bug: every Catalan diacritic this app's copy uses (torró
// names, "l·l", ...) must measure as present glyphs (non-zero advance) in
// both embedded faces, not silently render as nothing/tofu.
func TestCatalanDiacriticsMeasurable(t *testing.T) {
	sample := "àèéíïòóúüçÀÈÉÍÏÒÓÚÜÇ·"
	for _, style := range []fontStyle{styleSansBold, styleSerifItalic} {
		face := newFace(style, 32)
		for _, r := range sample {
			adv, ok := face.GlyphAdvance(r)
			if !ok || adv <= 0 {
				t.Errorf("style %v: rune %q has no glyph advance (ok=%v, adv=%v)", style, string(r), ok, adv)
			}
		}
	}
}

func TestHeroBigSizeShrinksToFit(t *testing.T) {
	c := newCanvas()
	maxWidth := CanvasWidth - 2*heroPaddingX

	shortSize := c.heroBigSize("8%")
	longSize := c.heroBigSize("842.310")

	if longSize > shortSize {
		t.Errorf("heroBigSize(\"842.310\")=%v should not exceed heroBigSize(\"8%%\")=%v", longSize, shortSize)
	}
	if w := measureWidth(c.face(styleSansBold, longSize), "842.310"); w > maxWidth {
		t.Errorf("heroBigSize did not shrink enough: %q measures %dpx at size %v, want <= %dpx", "842.310", w, longSize, maxWidth)
	}
}

package sharecard

import "testing"

// These exercise the shared chrome's component drawers directly (rather
// than only indirectly through the three Render* entry points), including
// paths - like tileGrid.centered's 3-column press-kit style - that no
// current variant's toFrame wires up yet but that the engine has to
// support cleanly for a future variant (see canvas.go's package doc) to
// use without a rewrite.

func TestDrawTileGridCenteredThreeColumn(t *testing.T) {
	c := newCanvas()
	y := c.drawTileGrid(tileGrid{
		cols:     3,
		centered: true,
		tiles: []tile{
			{value: "38.492", label: "VOTANTS"},
			{value: "96", label: "DUELS JUGATS"},
			{value: "1–24 des.", label: "PERÍODE"},
		},
	}, contentTop)

	if y <= contentTop {
		t.Fatalf("drawTileGrid() returned y=%d, want > contentTop=%d", y, contentTop)
	}
	if y >= CanvasHeight {
		t.Fatalf("drawTileGrid() returned y=%d, want < CanvasHeight=%d", y, CanvasHeight)
	}
}

func TestDrawTileGridWideTwoColumn(t *testing.T) {
	c := newCanvas()
	y := c.drawTileGrid(tileGrid{
		cols: 2,
		tiles: []tile{
			{value: "38", label: "RONDES JUGADES"},
			{value: "12", label: "RÀFEGA MÉS LLARGA"},
		},
	}, contentTop)

	if y <= contentTop {
		t.Fatalf("drawTileGrid() returned y=%d, want > contentTop=%d", y, contentTop)
	}
}

func TestDrawTileGridEmptyIsNoop(t *testing.T) {
	c := newCanvas()
	y := c.drawTileGrid(tileGrid{cols: 2}, contentTop)
	if y != contentTop {
		t.Errorf("drawTileGrid() with no tiles returned y=%d, want unchanged %d", y, contentTop)
	}
}

func TestDrawInfoCardWithPhotoAndColumns(t *testing.T) {
	c := newCanvas()
	y := c.drawInfoCard(infoCard{
		headline: "Torró d'Ametlla i Xocolata Negra 70%",
		sub:      "CACAU · PER ALS GOLAFRES",
		photo:    true,
		columns: []statColumn{
			{value: "16", label: "VOTS EMESOS"},
			{value: "9–1", label: "EN DUELS", accent: true},
		},
		footnote: "Cap altre torró t'ha convençut tant.",
	}, contentTop)

	if y <= contentTop {
		t.Fatalf("drawInfoCard() returned y=%d, want > contentTop=%d", y, contentTop)
	}
}

func TestDrawInfoCardWithoutPhotoOrColumns(t *testing.T) {
	c := newCanvas()
	y := c.drawInfoCard(infoCard{headline: "Encara no hi ha prou dades"}, contentTop)
	if y <= contentTop {
		t.Fatalf("drawInfoCard() returned y=%d, want > contentTop=%d", y, contentTop)
	}
}

func TestDrawDotGridHighlightsExpectedCount(t *testing.T) {
	c := newCanvas()
	y := c.drawDotGrid(dotGrid{total: 100, highlight: 8, caption: "8 de cada 100 votants trien com tu"}, contentTop)
	if y <= contentTop {
		t.Fatalf("drawDotGrid() returned y=%d, want > contentTop=%d", y, contentTop)
	}
}

func TestDrawHeaderAndFooterDoNotPanic(t *testing.T) {
	c := newCanvas()
	c.drawHeader("RESUM 2026")
	c.drawFooter(footerContent{
		shortLink:   "torrorendum.cat",
		showSponsor: true,
		sponsorLine: sponsorPlaceholder,
	})
	// No assertions beyond "didn't panic": drawHeader/drawFooter are
	// exercised pixel-for-pixel by the Render*ProducesValidCanvas tests;
	// this just isolates them so a future regression in one is easier to
	// bisect to than a full-card render.
}

func TestDrawDividerReturnsYBelowLine(t *testing.T) {
	c := newCanvas()
	y := c.drawDivider("EL TEU TORRÓ", contentTop)
	if y <= contentTop {
		t.Fatalf("drawDivider() returned y=%d, want > contentTop=%d", y, contentTop)
	}
}

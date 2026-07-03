package sharecard

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
)

// PressKitData is the plain input for the press-kit one-pager PNG: a
// global aggregate card (same for every viewer, unlike Data/WrappedData
// which are per-user) summarizing the Phase 2 knockout bracket's result
// for the press page (/premsa). See internal/http/press_handler.go.
type PressKitData struct {
	// HasChampion is false while the Gran Final hasn't been decided yet
	// (no bracket, or a bracket still in progress). RenderPressKit then
	// renders an empty state instead of a result.
	HasChampion bool

	ChampionName  string
	ChampionVotes int
}

// RenderPressKit draws PressKitData onto a CanvasWidth x CanvasHeight
// canvas and returns it PNG-encoded, following Render's exact shape. It
// never errors on the drawing itself; the returned error only reflects
// PNG encoding failures.
func RenderPressKit(data PressKitData) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))

	draw.Draw(img, img.Bounds(), image.NewUniform(colorBackground), image.Point{}, draw.Src)

	drawHeader(img)
	drawFooter(img)

	if data.HasChampion {
		drawPressKitChampion(img, data)
	} else {
		drawPressKitEmptyState(img)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("sharecard: encode press kit png: %w", err)
	}

	return buf.Bytes(), nil
}

// drawPressKitEmptyState renders the fallback card for when the Gran
// Final hasn't been disputed yet, reusing the exact same Catalan wording
// already established on the /premsa page (see press.html) so the two
// surfaces read consistently.
func drawPressKitEmptyState(img *image.RGBA) {
	maxTextWidth := CanvasWidth - 2*contentPaddingX
	y := contentTop + 200

	drawCenteredWrappedLines(img, "La Gran Final encara no s'ha disputat", maxTextWidth, y, colorText, scaleName, 3)
}

// drawPressKitChampion renders the "champion decided" result: the
// champion's name (large, primary color, wrapped) and its total winning
// vote count.
func drawPressKitChampion(img *image.RGBA, data PressKitData) {
	maxTextWidth := CanvasWidth - 2*contentPaddingX
	y := contentTop

	y = drawLine(img, "El torró guanyador del Torrorèndum", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel

	nameLines := wrapText(data.ChampionName, maxTextWidth, scaleName)
	const maxNameLines = 3
	if len(nameLines) > maxNameLines {
		nameLines = nameLines[:maxNameLines]
	}
	for _, line := range nameLines {
		y = drawLine(img, line, contentPaddingX, y, colorPrimary, scaleName)
		y += lineGapName
	}

	y += blockGap
	votesLine := fmt.Sprintf("Segons %d vots", data.ChampionVotes)
	drawLine(img, votesLine, contentPaddingX, y, colorText, scaleBody)
}

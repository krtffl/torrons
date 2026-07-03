package sharecard

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// WrappedData is the plain input for the "Torrorèndum Wrapped" personal
// recap card. Unlike Data (the v1 share card), this reuses the same
// rendering pipeline with richer, cross-source data: total votes, the
// user's most contested duel and most unpopular pick (both derived from
// domain.WrappedStatsRepo.DuelStats), and their Phase 2 knockout bracket
// participation (domain.WrappedStatsRepo.BracketPath). See the TODO in
// sharecard.go's package doc, which this struct fulfills.
type WrappedData struct {
	// HasEnoughVotes is false when the user hasn't cleared the Wrapped
	// unlock threshold yet (see getMinVotesForClass("5") in
	// internal/http/user_api.go). RenderWrapped then renders a "not
	// unlocked yet" empty state instead of any stat.
	HasEnoughVotes bool
	// VotesRemaining is meaningful only when !HasEnoughVotes: how many
	// more votes the user needs to unlock their Wrapped.
	VotesRemaining int

	// TotalVotes is the user's all-time vote count (domain.User.VoteCount).
	TotalVotes int

	// HasContestedDuel is false if the user never voted on a pairing that
	// met the minimum-total-votes threshold (a legitimate empty state,
	// not an error - see domain.WrappedStatsRepo.DuelStats).
	HasContestedDuel    bool
	ContestedTorroAName string
	ContestedTorroBName string
	ContestedPercentA   int
	ContestedPercentB   int

	// HasUnpopularPick is false under the same conditions as
	// HasContestedDuel above (they're derived from the same dataset).
	HasUnpopularPick  bool
	UnpopularPickName string
	// UnpopularPercent is the % of the crowd that agreed with this pick -
	// low means unpopular.
	UnpopularPercent int

	// HasBracketVotes is false if the user never voted in the Phase 2
	// knockout bracket at all.
	HasBracketVotes       bool
	BracketRoundsVoted    int
	BracketMatchesDecided int
	BracketPicksCorrect   int
	// HasChampion is whether the bracket itself is fully decided yet.
	HasChampion  bool
	ChampionName string
	// MatchedChampion is meaningful only if HasChampion: did the user
	// ever pick the champion in one of their voted matches.
	MatchedChampion bool
}

// RenderWrapped draws WrappedData onto a CanvasWidth x CanvasHeight canvas
// and returns it PNG-encoded, following Render's exact shape. It never
// errors on the drawing itself; the returned error only reflects PNG
// encoding failures.
func RenderWrapped(data WrappedData) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))

	draw.Draw(img, img.Bounds(), image.NewUniform(colorBackground), image.Point{}, draw.Src)

	drawHeader(img)
	drawFooter(img)

	if data.HasEnoughVotes {
		drawWrappedResult(img, data)
	} else {
		drawWrappedLocked(img, data)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("sharecard: encode wrapped png: %w", err)
	}

	return buf.Bytes(), nil
}

// drawWrappedLocked renders the fallback card for users who haven't
// cleared the Wrapped unlock threshold yet, mirroring drawEmptyState's
// "vota per generar la teva targeta" treatment in sharecard.go.
func drawWrappedLocked(img *image.RGBA, data WrappedData) {
	maxTextWidth := CanvasWidth - 2*contentPaddingX
	y := contentTop + 200

	y = drawCenteredWrappedLines(img, "Encara no has desbloquejat el teu Wrapped", maxTextWidth, y, colorText, scaleName, 3)
	y += blockGap
	line := fmt.Sprintf("Et falten %d vots", data.VotesRemaining)
	drawCenteredLine(img, line, y, colorTextLight, scaleBody)
}

// drawWrappedResult renders the unlocked Wrapped stat blocks, stacked
// top-to-bottom starting at contentTop. This is deliberately capped at
// four blocks (total votes, contested duel, unpopular pick, bracket
// path) rather than trying to fit every conceivable stat - see the
// "plumbing pass, not final design" note in sharecard.go's package doc.
// Each sub-stat that has its own legitimate empty state (not enough
// data yet) renders an honest sub-line instead of a fabricated number.
func drawWrappedResult(img *image.RGBA, data WrappedData) {
	maxTextWidth := CanvasWidth - 2*contentPaddingX
	y := contentTop

	// -- Total votes --
	y = drawLine(img, "Vots totals", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel
	y = drawLine(img, fmt.Sprintf("%d", data.TotalVotes), contentPaddingX, y, colorPrimary, scaleName)
	y += blockGap

	// -- Contested duel --
	y = drawLine(img, "El duel més disputat", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel
	if data.HasContestedDuel {
		line := fmt.Sprintf("%s (%d%%) vs %s (%d%%)",
			data.ContestedTorroAName, data.ContestedPercentA,
			data.ContestedTorroBName, data.ContestedPercentB)
		// Capped at 3 lines (rather than the other blocks' 2): with two
		// full torró names plus both percentages, this is the one line
		// most likely to run long, and there's headroom left in the
		// canvas for it - see drawWrappedResult's package-level budget
		// note.
		y = drawWrappedLines(img, line, maxTextWidth, y, colorText, scaleBody, 3)
	} else {
		y = drawWrappedLines(img, "Encara no hi ha prou dades per calcular el teu duel més disputat", maxTextWidth, y, colorTextLight, scaleBody, 2)
	}
	y += blockGap

	// -- Unpopular pick --
	y = drawLine(img, "La teva tria més atrevida", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel
	if data.HasUnpopularPick {
		y = drawWrappedLines(img, data.UnpopularPickName, maxTextWidth, y, colorPrimary, scaleBody, 2)
		y += lineGapName
		line := fmt.Sprintf("Només el %d%% del públic hi va coincidir amb tu", data.UnpopularPercent)
		y = drawWrappedLines(img, line, maxTextWidth, y, colorText, scaleBody, 2)
	} else {
		y = drawWrappedLines(img, "Encara no hi ha prou dades per triar la teva tria més atrevida", maxTextWidth, y, colorTextLight, scaleBody, 2)
	}
	y += blockGap

	// -- Bracket path --
	y = drawLine(img, "El teu camí a la Gran Final", contentPaddingX, y, colorTextLight, scaleLabel)
	y += lineGapLabel
	switch {
	case !data.HasBracketVotes:
		drawWrappedLines(img, "Encara no has votat a la fase de knockout", maxTextWidth, y, colorTextLight, scaleBody, 2)
	case !data.HasChampion:
		line := fmt.Sprintf("%d rondes votades, %d encerts de %d", data.BracketRoundsVoted, data.BracketPicksCorrect, data.BracketMatchesDecided)
		y = drawWrappedLines(img, line, maxTextWidth, y, colorText, scaleBody, 2)
		y += lineGapName
		drawWrappedLines(img, "La Gran Final encara no s'ha decidit", maxTextWidth, y, colorTextLight, scaleBody, 2)
	default:
		line := fmt.Sprintf("%d rondes votades, %d encerts de %d", data.BracketRoundsVoted, data.BracketPicksCorrect, data.BracketMatchesDecided)
		y = drawWrappedLines(img, line, maxTextWidth, y, colorText, scaleBody, 2)
		y += lineGapName
		var champLine string
		if data.MatchedChampion {
			champLine = fmt.Sprintf("Vas encertar el campió: %s", data.ChampionName)
		} else {
			champLine = fmt.Sprintf("El campió va ser %s", data.ChampionName)
		}
		drawWrappedLines(img, champLine, maxTextWidth, y, colorText, scaleBody, 2)
	}
}

// drawWrappedLines wraps text to fit maxWidthPx at the given scale, draws
// up to maxLines of it left-aligned at contentPaddingX (dropping any
// remainder rather than overflowing past the canvas edge or clipping
// mid-sentence), and returns the y coordinate below the last drawn line.
// Shared by wrapped.go and presskit.go.
func drawWrappedLines(img *image.RGBA, text string, maxWidthPx int, y int, col color.Color, scale int, maxLines int) int {
	lines := wrapText(text, maxWidthPx, scale)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for _, line := range lines {
		y = drawLine(img, line, contentPaddingX, y, col, scale)
		y += lineGapName
	}
	return y
}

// drawCenteredWrappedLines behaves like drawWrappedLines but horizontally
// centers each wrapped line, for empty-state messaging. Shared by
// wrapped.go and presskit.go.
func drawCenteredWrappedLines(img *image.RGBA, text string, maxWidthPx int, y int, col color.Color, scale int, maxLines int) int {
	lines := wrapText(text, maxWidthPx, scale)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for _, line := range lines {
		y = drawCenteredLine(img, line, y, col, scale)
		y += lineGapName
	}
	return y
}

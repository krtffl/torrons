package sharecard

import "image"

// This file holds the shared chrome's higher-level component drawers -
// header, footer, hero, dot-grid, tile-grid, divider, info card, QR
// placeholder - each built from the primitives in text.go/shapes.go. This
// is the "one shared engine" the variant Data types (see sharecard.go/
// wrapped.go/presskit.go) all render through via renderFrame in canvas.go.

// vcenter returns the y coordinate that vertically centers a line of the
// given lineHeight within [containerTop, containerTop+containerHeight).
func vcenter(containerTop, containerHeight, lineHeight int) int {
	return containerTop + (containerHeight-lineHeight)/2
}

// ---------------------------------------------------------------------
// Header: kicker badge (left) + wordmark (right)
// ---------------------------------------------------------------------

func (c *canvas) drawHeader(kicker string) {
	rowTop := headerPaddingTop
	rowCenterY := rowTop + headerRowHeight/2

	// Kicker badge pill.
	kickerFace := c.face(styleSansBold, sizeKicker)
	kickerW := measureTracked(kickerFace, kicker, trackKicker)
	pillW := headerPillPadX*2 + headerDotRadius*2 + headerGap + kickerW
	pillRect := image.Rect(paddingX, rowTop, paddingX+pillW, rowTop+headerRowHeight)
	dashedPill(c.img, pillRect, 2, creamAlpha(0x59))
	fillCircle(c.img, paddingX+headerPillPadX+headerDotRadius, rowCenterY, headerDotRadius, colorGold)
	textX := paddingX + headerPillPadX + headerDotRadius*2 + headerGap
	textY := vcenter(rowTop, headerRowHeight, kickerFace.Metrics().Height.Ceil())
	c.drawTracked(kicker, textX, textY, styleSansBold, sizeKicker, trackKicker, creamAlpha(0xBF))

	// Wordmark.
	const wordmarkText = "torrorèndum"
	wmFace := c.face(styleSansBold, sizeWordmark)
	wmTextW := measureWidth(wmFace, wordmarkText)
	totalW := wordmarkRingDiameter + headerGap + wmTextW
	wmX := CanvasWidth - paddingX - totalW
	ringCX := wmX + wordmarkRingDiameter/2
	fillCircle(c.img, ringCX, rowCenterY, wordmarkRingDiameter/2, colorGold)
	fillCircle(c.img, ringCX, rowCenterY, wordmarkRingDiameter/2-wordmarkRingStroke, bgColorAt(rowCenterY))
	fillCircle(c.img, ringCX, rowCenterY, wordmarkDotRadius, colorGold)
	wmTextX := ringCX + wordmarkRingDiameter/2 + headerGap
	wmTextY := vcenter(rowTop, headerRowHeight, wmFace.Metrics().Height.Ceil())
	c.drawLine(wordmarkText, wmTextX, wmTextY, styleSansBold, sizeWordmark, 0, colorCream)
}

// ---------------------------------------------------------------------
// Footer: "discover yours at {link}" + QR placeholder + hashtag + sponsor
// ---------------------------------------------------------------------

// drawFooter is bottom-of-canvas content, but unlike the mockup's flex
// column with `margin-top:auto` (which pushes the footer to the bottom of
// whatever space is left, however much content sits above it), this
// engine draws it starting at the fixed footerTop constant. A fixed
// canvas has no equivalent of "whatever space is left"; footerTop is
// chosen generously enough that the content flowed above it (hero/dots/
// tiles/divider/cards, each capped in how much they can grow) never
// collides with it in practice - see the package's render+inspect tests.
func (c *canvas) drawFooter(f footerContent) {
	y := footerTop

	const prefix = "Descobreix el teu torró a "
	taglineFace := c.face(styleSerifItalic, sizeFooterTagline)
	linkFace := c.face(styleSansBold, sizeFooterTagline)
	prefixW := measureWidth(taglineFace, prefix)
	linkW := measureWidth(linkFace, f.shortLink)
	startX := (CanvasWidth - (prefixW + linkW)) / 2
	c.drawLine(prefix, startX, y, styleSerifItalic, sizeFooterTagline, 0, creamAlpha(0x9E))
	y = c.drawLine(f.shortLink, startX+prefixW, y, styleSansBold, sizeFooterTagline, 0, colorGold)
	y += footerGapA

	qrTotal := qrBoxSize + 2*qrPad
	qrX := (CanvasWidth - qrTotal) / 2
	c.drawQR(qrX, y)
	y += qrTotal + footerGapB

	y = c.drawCenteredLine("#Torrorèndum", y, styleSansBold, sizeHashtag, trackHashtag, creamAlpha(0x80))

	if f.showSponsor && f.sponsorLine != "" {
		y += footerGapC
		c.drawSponsorPill(f.sponsorLine, y)
	}
}

func (c *canvas) drawSponsorPill(label string, y int) {
	face := c.face(styleSansBold, sizeSponsor)
	textW := measureTracked(face, label, trackSponsor)
	lineH := face.Metrics().Height.Ceil()
	pillW := textW + 2*sponsorPadX
	pillH := lineH + 2*sponsorPadY
	x0 := (CanvasWidth - pillW) / 2
	rect := image.Rect(x0, y, x0+pillW, y+pillH)
	dashedPill(c.img, rect, 2, creamAlpha(0x47))
	c.drawTracked(label, x0+sponsorPadX, y+sponsorPadY, styleSansBold, sizeSponsor, trackSponsor, creamAlpha(0x66))
}

// drawQR paints the QR-code-shaped placeholder (a deterministic pseudo-random
// module pattern plus the three finder squares) in a cream rounded box
// at top-left corner (x, y). This mirrors the mockup's own placeholder
// exactly (it isn't a real scannable QR code there either - generating one
// would need a QR encoding library this task doesn't call for).
func (c *canvas) drawQR(x, y int) {
	boxSize := qrBoxSize + 2*qrPad
	fillRoundedRect(c.img, image.Rect(x, y, x+boxSize, y+boxSize), 20, colorCream)

	gx, gy := x+qrPad, y+qrPad
	for row := 0; row < qrGrid; row++ {
		for col := 0; col < qrGrid; col++ {
			inTL := row < 3 && col < 3
			inTR := row < 3 && col > 4
			inBL := row > 4 && col < 3
			if inTL || inTR || inBL {
				continue
			}
			idx := uint32(row*qrGrid + col)
			hash := idx * 2654435761
			if hash%5 >= 2 {
				continue
			}
			cx := gx + col*qrCellSize
			cy := gy + row*qrCellSize
			fillRoundedRect(c.img, image.Rect(cx, cy, cx+qrCellFill, cy+qrCellFill), 2, colorCardInk)
		}
	}

	drawFinder := func(fx, fy int) {
		fillRoundedRect(c.img, image.Rect(fx, fy, fx+qrFinder, fy+qrFinder), 6, colorCardInk)
		pad := 8
		fillRoundedRect(c.img, image.Rect(fx+pad, fy+pad, fx+qrFinder-pad, fy+qrFinder-pad), 3, colorCream)
		inner := 16
		ix := fx + (qrFinder-inner)/2
		iy := fy + (qrFinder-inner)/2
		fillRoundedRect(c.img, image.Rect(ix, iy, ix+inner, iy+inner), 2, colorCardInk)
	}
	drawFinder(gx, gy)
	drawFinder(gx+qrGrid*qrCellSize-qrFinder, gy)
	drawFinder(gx, gy+qrGrid*qrCellSize-qrFinder)
}

// ---------------------------------------------------------------------
// Empty-state message (locked Wrapped, no votes yet, no champion yet)
// ---------------------------------------------------------------------

func (c *canvas) drawEmptyMessage(m emptyMessage) {
	maxWidth := CanvasWidth - 2*heroPaddingX
	y := contentTop + 260

	y = c.drawCenteredBlock(m.heading, y, styleSansBold, 44, 0, colorCream, maxWidth, 3, 10)
	if m.sub != "" {
		y += heroGapMed
		c.drawCenteredBlock(m.sub, y, styleSerifItalic, sizeHeroTagline, 0, creamAlpha(0xA6), maxWidth, 2, heroLineGap)
	}
}

// ---------------------------------------------------------------------
// Hero: intro / label-above / big stat / unit-below / tagline / pill
// ---------------------------------------------------------------------

func (c *canvas) drawHero(h heroContent) int {
	maxWidth := CanvasWidth - 2*heroPaddingX
	y := contentTop

	if h.intro != "" {
		y = c.drawCenteredBlock(h.intro, y, styleSerifItalic, sizeHeroIntro, 0, creamAlpha(0xA6), maxWidth, 2, heroLineGap)
		y += heroGapSmall
	}
	if h.labelAbove != "" {
		y = c.drawCenteredLine(h.labelAbove, y, styleSansBold, sizeHeroLabel, trackHeroLbl, colorGold)
		y += heroGapTiny
	}

	bigSize := c.heroBigSize(h.big)
	y = c.drawCenteredLine(h.big, y, styleSansBold, bigSize, 0, colorCream)

	if h.unitBelow != "" {
		y += 2
		y = c.drawCenteredLine(h.unitBelow, y, styleSansBold, sizeHeroLabel, trackHeroLbl, colorGold)
	}
	if h.tagline != "" {
		y += heroGapMed
		y = c.drawCenteredBlock(h.tagline, y, styleSerifItalic, sizeHeroTagline, 0, colorCream, maxWidth, 2, heroLineGap)
	}
	if h.pill != "" {
		y += heroGapMed
		y = c.drawPill(h.pill, y)
	}
	return y
}

// heroBigSize shrinks the hero's headline stat font size until it fits
// within the hero's content width, so long formatted numbers (a growing
// campaign vote total, e.g. "842.310") never overflow past the canvas
// edge instead of silently clipping.
func (c *canvas) heroBigSize(text string) float64 {
	maxWidth := CanvasWidth - 2*heroPaddingX
	size := sizeHeroBig
	for size > 60 {
		if measureWidth(c.face(styleSansBold, size), text) <= maxWidth {
			return size
		}
		size -= 10
	}
	return size
}

// drawPill draws the hero's outlined badge/persona pill and returns the y
// coordinate below it.
func (c *canvas) drawPill(label string, y int) int {
	const padX, padY = 34, 15
	face := c.face(styleSansBold, sizeHeroPill)
	textW := measureTracked(face, label, trackPill)
	lineH := face.Metrics().Height.Ceil()
	pillW := textW + 2*padX
	pillH := lineH + 2*padY
	x0 := (CanvasWidth - pillW) / 2
	rect := image.Rect(x0, y, x0+pillW, y+pillH)
	strokeRoundedRect(c.img, rect, pillH/2, 2, goldAlpha(0x99))
	c.drawTracked(label, x0+padX, y+padY, styleSansBold, sizeHeroPill, trackPill, colorGold)
	return y + pillH
}

// ---------------------------------------------------------------------
// Dot grid: the "N of 100" percentile/match visualization
// ---------------------------------------------------------------------

// dotHighlightSet ports the mockup's own dot-selection algorithm: it
// spreads `highlight` highlighted dots as evenly as possible across
// `total` dots.
func dotHighlightSet(total, highlight int) map[int]bool {
	set := make(map[int]bool, highlight)
	if total <= 0 || highlight <= 0 {
		return set
	}
	if highlight > total {
		highlight = total
	}
	spacing := total / highlight
	if spacing < 1 {
		spacing = 1
	}
	for k := 0; k < highlight; k++ {
		idx := (k*spacing + spacing/2) % total
		for set[idx] {
			idx = (idx + 1) % total
		}
		set[idx] = true
	}
	return set
}

func (c *canvas) drawDotGrid(d dotGrid, y int) int {
	if d.total <= 0 {
		return y
	}
	rows := (d.total + dotCols - 1) / dotCols
	gridWidth := dotCols*dotSize + (dotCols-1)*dotGap
	x0 := (CanvasWidth - gridWidth) / 2
	hl := dotHighlightSet(d.total, d.highlight)

	for i := 0; i < d.total; i++ {
		row, col := i/dotCols, i%dotCols
		cx := x0 + col*(dotSize+dotGap) + dotSize/2
		cy := y + row*(dotSize+dotGap) + dotSize/2
		if hl[i] {
			fillCircle(c.img, cx, cy, dotSize/2+2, goldAlpha(0x30)) // soft glow halo
			fillCircle(c.img, cx, cy, dotSize/2, colorGold)
		} else {
			fillCircle(c.img, cx, cy, dotSize/2, colorDotDim)
		}
	}

	y += rows*(dotSize+dotGap) - dotGap + dotCaptionGap
	if d.caption != "" {
		y = c.drawCenteredBlock(d.caption, y, styleSansBold, sizeDotCaption, trackDotCaption, creamAlpha(0x80), CanvasWidth-2*dotsPaddingX, 2, heroLineGap)
	}
	return y
}

// ---------------------------------------------------------------------
// Tile grid: Wrapped's 2-col stat grid / press kit's 3-col stat row
// ---------------------------------------------------------------------

func (c *canvas) drawTileGrid(t tileGrid, y int) int {
	if len(t.tiles) == 0 {
		return y
	}
	cols := t.cols
	if cols < 1 {
		cols = 1
	}
	areaW := CanvasWidth - 2*cardPaddingX
	tileW := (areaW - tileGap*(cols-1)) / cols

	padX, tileH, valSize := tilePadXWide, tileHeightWide, sizeTileValueLg
	if t.centered {
		padX, tileH, valSize = tilePadXCentered, tileHeightCentered, sizeTileValueSm
	}
	innerW := tileW - 2*padX

	rows := (len(t.tiles) + cols - 1) / cols
	for i, tl := range t.tiles {
		row, col := i/cols, i%cols
		x0 := cardPaddingX + col*(tileW+tileGap)
		y0 := y + row*(tileH+tileGap)
		rect := image.Rect(x0, y0, x0+tileW, y0+tileH)
		fillRoundedRect(c.img, rect, 22, colorTileBg)
		strokeRoundedRect(c.img, rect, 22, 1, colorTileBorder)

		ty := y0 + tilePadY
		ty = c.drawBlock(tl.value, x0+padX, innerW, ty, styleSansBold, valSize, 0, colorGold, 2, heroLineGap, t.centered)
		ty += tileValueLabelGap
		c.drawBlock(tl.label, x0+padX, innerW, ty, styleSansBold, sizeTileLabel, trackTileLbl, creamAlpha(0x8C), 2, tileLabelLineGap, t.centered)
	}
	return y + rows*(tileH+tileGap) - tileGap
}

// ---------------------------------------------------------------------
// Divider: "— LABEL —" section rule between the hero/dots/tiles and the
// info card(s).
// ---------------------------------------------------------------------

func (c *canvas) drawDivider(label string, y int) int {
	y += dividerGapY
	face := c.face(styleSansBold, sizeDivider)
	labelW := measureTracked(face, label, trackDivider)
	lineH := face.Metrics().Height.Ceil()
	const gap = 20
	lineY := y + lineH/2
	labelX0 := (CanvasWidth - labelW) / 2

	lineCol := creamAlpha(0x2E)
	for x := cardPaddingX; x < labelX0-gap; x++ {
		blendOver(c.img, x, lineY, lineCol)
	}
	for x := labelX0 + labelW + gap; x < CanvasWidth-cardPaddingX; x++ {
		blendOver(c.img, x, lineY, lineCol)
	}
	c.drawTracked(label, labelX0, y, styleSansBold, sizeDivider, trackDivider, creamAlpha(0x99))
	return y + lineH
}

// ---------------------------------------------------------------------
// Info card: the shared rounded "cream card" (torró result / duel / pick)
// ---------------------------------------------------------------------

func (c *canvas) drawInfoCard(card infoCard, y int) int {
	y += 32
	cardW := CanvasWidth - 2*cardPaddingX
	innerW := cardW - 2*cardPadX

	headlineFace := c.face(styleSansBold, sizeCardHeadline)
	headlineAreaW := innerW
	if card.photo {
		headlineAreaW = innerW - cardPhotoW - 28
	}
	headlineLines := wrapText(headlineFace, card.headline, headlineAreaW, 0)
	if len(headlineLines) > 3 {
		headlineLines = headlineLines[:3]
	}
	headlineLineH := headlineFace.Metrics().Height.Ceil() + 4
	topBlockH := len(headlineLines) * headlineLineH
	if card.sub != "" {
		topBlockH += cardHeadlineGap + int(sizeCardTag) + 4
	}
	if card.photo && cardPhotoH > topBlockH {
		topBlockH = cardPhotoH
	}

	statsH := 0
	if len(card.columns) > 0 {
		statsH = cardStatsGapTop + 1 + cardStatsPadTop + int(sizeCardStatVal) + cardColumnGap + int(sizeCardStatLbl)
	}
	footnoteH := 0
	if card.footnote != "" {
		footnoteH = cardFootnoteGap + int(sizeCardFootnote)*2 + 6
	}

	cardH := cardPadY*2 + topBlockH + statsH + footnoteH
	rect := image.Rect(cardPaddingX, y, CanvasWidth-cardPaddingX, y+cardH)
	fillRoundedRect(c.img, rect, cardRadius, colorCardBg)
	strokeRoundedRect(c.img, rect, cardRadius, cardBorderWidth, goldAlpha(0xFF))

	contentX := cardPaddingX + cardPadX
	contentTopY := y + cardPadY
	textX := contentX

	if card.photo {
		photoRect := image.Rect(contentX, contentTopY, contentX+cardPhotoW, contentTopY+cardPhotoH)
		fillDiagonalStripesRounded(c.img, photoRect, cardPhotoRadius, cardPhotoStripe, colorPhotoA, colorPhotoB)
		c.drawCenteredInRect(photoRect, "FOTO", styleSansBold, 15, trackTag, colorPhotoInk)
		textX = contentX + cardPhotoW + 28
	}

	textY := contentTopY + (topBlockH-len(headlineLines)*headlineLineH)/2
	if textY < contentTopY {
		textY = contentTopY
	}
	for _, line := range headlineLines {
		textY = c.drawLine(line, textX, textY, styleSansBold, sizeCardHeadline, 0, colorCardInk)
		textY += 4
	}
	if card.sub != "" {
		textY += cardHeadlineGap - 4
		c.drawTracked(card.sub, textX, textY, styleSansBold, sizeCardTag, trackTag, colorTagInk)
	}

	y = contentTopY + topBlockH

	if len(card.columns) > 0 {
		y += cardStatsGapTop
		// The mockup's card divider rule (`border-top:1px solid #F0DFC2`)
		// happens to reuse the exact same tan as the photo placeholder's
		// lighter stripe (colorPhotoA) - not a coincidence, the mockup
		// reuses that hex for both, so this does too rather than adding a
		// near-duplicate constant.
		for x := contentX; x < CanvasWidth-cardPaddingX-cardPadX; x++ {
			blendOver(c.img, x, y, colorPhotoA)
		}
		y += cardStatsPadTop
		colW := innerW / len(card.columns)
		for i, col := range card.columns {
			cx0 := contentX + i*colW
			valFace := c.face(styleSansBold, sizeCardStatVal)
			valW := measureWidth(valFace, col.value)
			valColor := colorCardInk
			if col.accent {
				valColor = colorHeadToHead
			}
			c.drawLine(col.value, cx0+(colW-valW)/2, y, styleSansBold, sizeCardStatVal, 0, valColor)
			lblFace := c.face(styleSansBold, sizeCardStatLbl)
			lblW := measureTracked(lblFace, col.label, trackStatLbl)
			c.drawTracked(col.label, cx0+(colW-lblW)/2, y+int(sizeCardStatVal)+cardColumnGap, styleSansBold, sizeCardStatLbl, trackStatLbl, colorMetaInk)
		}
		y += int(sizeCardStatVal) + cardColumnGap + int(sizeCardStatLbl)
	}

	if card.footnote != "" {
		y += cardFootnoteGap
		c.drawCenteredBlock(card.footnote, y, styleSerifItalic, sizeCardFootnote, 0, colorPhotoInk, innerW, 2, 6)
	}

	return contentTopY + cardH - cardPadY
}

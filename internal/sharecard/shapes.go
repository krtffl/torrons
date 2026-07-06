package sharecard

import (
	"image"
	"image/color"
	"math"
)

// This file holds hand-rolled raster primitives (rounded rects, circles,
// dashed/dotted strokes, diagonal stripes) that image/draw doesn't provide
// directly. Everything composites through blendOver, so every shape here
// takes a color.NRGBA (straight alpha) - see constants.go's doc comment on
// why color.RGBA would blend incorrectly for anything translucent.

// blendOver composites col onto img at (x, y), assuming (as is always true
// for this package: the canvas is painted with an opaque background before
// anything else is drawn) that the destination pixel is already fully
// opaque. That assumption means plain straight-alpha "src over opaque dst"
// math is all that's needed - no need to track/propagate a resulting
// alpha channel.
func blendOver(img *image.RGBA, x, y int, col color.NRGBA) {
	if x < img.Rect.Min.X || x >= img.Rect.Max.X || y < img.Rect.Min.Y || y >= img.Rect.Max.Y {
		return
	}
	if col.A == 0 {
		return
	}
	if col.A == 0xFF {
		img.SetRGBA(x, y, color.RGBA{R: col.R, G: col.G, B: col.B, A: 0xFF})
		return
	}
	dst := img.RGBAAt(x, y)
	a := uint32(col.A)
	inv := 0xFF - a
	r := (uint32(col.R)*a + uint32(dst.R)*inv) / 0xFF
	g := (uint32(col.G)*a + uint32(dst.G)*inv) / 0xFF
	b := (uint32(col.B)*a + uint32(dst.B)*inv) / 0xFF
	img.SetRGBA(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF})
}

// insideRoundedRect reports whether (x, y) falls within a rounded
// rectangle spanning [minX,maxX) x [minY,maxY) with corner radius rad.
func insideRoundedRect(x, y, minX, minY, maxX, maxY, rad int) bool {
	if x < minX || x >= maxX || y < minY || y >= maxY {
		return false
	}
	if rad <= 0 {
		return true
	}
	inCornerCol := x < minX+rad || x >= maxX-rad
	inCornerRow := y < minY+rad || y >= maxY-rad
	if !inCornerCol || !inCornerRow {
		return true // in the cross-shaped core, not near any corner
	}
	cx := minX + rad
	if x >= maxX-rad {
		cx = maxX - rad - 1
	}
	cy := minY + rad
	if y >= maxY-rad {
		cy = maxY - rad - 1
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= rad*rad
}

// fillRoundedRect fills a rounded rectangle.
func fillRoundedRect(img *image.RGBA, r image.Rectangle, radius int, col color.NRGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if insideRoundedRect(x, y, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y, radius) {
				blendOver(img, x, y, col)
			}
		}
	}
}

// strokeRoundedRect draws a rounded-rectangle outline of the given width.
func strokeRoundedRect(img *image.RGBA, r image.Rectangle, radius, width int, col color.NRGBA) {
	innerRadius := radius - width
	if innerRadius < 0 {
		innerRadius = 0
	}
	inner := image.Rect(r.Min.X+width, r.Min.Y+width, r.Max.X-width, r.Max.Y-width)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if !insideRoundedRect(x, y, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y, radius) {
				continue
			}
			if insideRoundedRect(x, y, inner.Min.X, inner.Min.Y, inner.Max.X, inner.Max.Y, innerRadius) {
				continue
			}
			blendOver(img, x, y, col)
		}
	}
}

// fillCircle fills a filled disc centered at (cx, cy).
func fillCircle(img *image.RGBA, cx, cy, radius int, col color.NRGBA) {
	if radius <= 0 {
		return
	}
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				blendOver(img, x, y, col)
			}
		}
	}
}

// fillDiamond fills a 45-degree-rotated square (a diamond) centered at
// (cx, cy), matching the mockup's `transform:rotate(45deg)` confetti square.
func fillDiamond(img *image.RGBA, cx, cy, halfDiagonal int, col color.NRGBA) {
	for y := cy - halfDiagonal; y <= cy+halfDiagonal; y++ {
		for x := cx - halfDiagonal; x <= cx+halfDiagonal; x++ {
			dx, dy := x-cx, y-cy
			if dx < 0 {
				dx = -dx
			}
			if dy < 0 {
				dy = -dy
			}
			if dx+dy <= halfDiagonal {
				blendOver(img, x, y, col)
			}
		}
	}
}

// fillGlow paints a soft radial falloff (maxAlpha at the center, 0 at
// radius) - a cheap stand-in for the mockup's
// `radial-gradient(circle, rgba(gold,.24) 0%, rgba(gold,0) 70%)` ambient
// glow.
func fillGlow(img *image.RGBA, cx, cy, radius int, maxAlpha uint8, col color.NRGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := float64(x-cx), float64(y-cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > float64(radius) {
				continue
			}
			t := 1 - dist/float64(radius)
			a := uint8(float64(maxAlpha) * t)
			if a == 0 {
				continue
			}
			blendOver(img, x, y, color.NRGBA{R: col.R, G: col.G, B: col.B, A: a})
		}
	}
}

// dashedRing stamps small dots evenly around a circle's circumference,
// approximating the mockup's `border: 1.5px dashed` ring - a true dashed
// circular *stroke* would need arc-length-accurate path math for little
// visual benefit at this dot size, so this "dotted" ring is a deliberate,
// much simpler stand-in.
func dashedRing(img *image.RGBA, cx, cy, radius, dotRadius int, col color.NRGBA) {
	circumference := 2 * math.Pi * float64(radius)
	step := float64(dotRadius*3) / circumference * 2 * math.Pi
	if step <= 0 {
		return
	}
	for a := 0.0; a < 2*math.Pi; a += step * 2 { // *2 so half the steps are gaps
		x := cx + int(float64(radius)*math.Cos(a))
		y := cy + int(float64(radius)*math.Sin(a))
		fillCircle(img, x, y, dotRadius, col)
	}
}

// dashedPill stamps small dots around a stadium (pill/fully-rounded
// rectangle) outline - the same "dotted" simplification as dashedRing,
// used for the kicker badge and sponsor-placeholder borders.
func dashedPill(img *image.RGBA, r image.Rectangle, dotRadius int, col color.NRGBA) {
	rad := r.Dy() / 2
	cx1, cx2 := r.Min.X+rad, r.Max.X-rad
	cy := r.Min.Y + rad
	step := dotRadius * 3
	if step <= 0 {
		return
	}
	for x := cx1; x <= cx2; x += step {
		fillCircle(img, x, r.Min.Y, dotRadius, col)
		fillCircle(img, x, r.Max.Y, dotRadius, col)
	}
	angleStep := float64(step) / float64(rad)
	if angleStep <= 0 {
		return
	}
	for a := math.Pi / 2; a <= 3*math.Pi/2; a += angleStep {
		x := cx1 + int(float64(rad)*math.Cos(a))
		y := cy + int(float64(rad)*math.Sin(a))
		fillCircle(img, x, y, dotRadius, col)
	}
	for a := -math.Pi / 2; a <= math.Pi/2; a += angleStep {
		x := cx2 + int(float64(rad)*math.Cos(a))
		y := cy + int(float64(rad)*math.Sin(a))
		fillCircle(img, x, y, dotRadius, col)
	}
}

// fillDiagonalStripesRounded fills a rounded rectangle with a two-tone
// 45-degree diagonal stripe pattern, matching the mockup's torró-photo
// placeholder (`repeating-linear-gradient(135deg, ...)` inside a
// border-radius box).
func fillDiagonalStripesRounded(img *image.RGBA, r image.Rectangle, radius, stripeWidth int, a, b color.NRGBA) {
	if stripeWidth <= 0 {
		stripeWidth = 1
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if !insideRoundedRect(x, y, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y, radius) {
				continue
			}
			idx := ((x - r.Min.X) + (y - r.Min.Y)) / stripeWidth
			col := a
			if idx%2 != 0 {
				col = b
			}
			blendOver(img, x, y, col)
		}
	}
}

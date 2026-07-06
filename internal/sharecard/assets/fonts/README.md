# Fonts embedded in this package

The design mockup (`docs/design-deliverables/Torrorendum Story Card.dc.html`)
specifies **Bricolage Grotesque** (bold sans, for headlines/labels/numbers)
and **Newsreader** (italic serif, for intro/tagline copy). Both are Google
Fonts, fetched from `fonts.googleapis.com` in the mockup's `<link>` tag.

This build environment has no network access, so those two families could
not be downloaded. Rather than ship the v1 bitmap font (`basicfont.Face7x13`)
plus its accent-stripping fallback for Catalan diacritics, this package
embeds a same-purpose, properly-licensed substitute pair that was already
present on the build machine as a standard OS font package:

- `LiberationSans-Bold.ttf` — stand-in for Bricolage Grotesque Bold/ExtraBold
  (headlines, the big headline stat, tracked uppercase labels, the wordmark).
- `LiberationSerif-Italic.ttf` — stand-in for Newsreader Italic (the italic
  intro/tagline/footnote lines).

Both are the **Liberation Fonts** project (Google/Red Hat), licensed under
the **SIL Open Font License 1.1** (see `LICENSE.txt`), which explicitly
permits bundling/embedding in software. Both are real outline (TrueType)
fonts with full Latin Extended coverage, so they render Catalan diacritics
(à, è, é, í, ï, ò, ó, ú, ü, ç, the "l·l" interpunct, …) correctly and
natively — no accent-stripping fallback is needed with these fonts.

Swapping in the actual Bricolage Grotesque / Newsreader `.ttf` files later
(once they can be sourced under their own license, e.g. the SIL OFL they
ship under from Google Fonts) is a two-file replacement in this directory;
nothing in `text.go` or `canvas.go` needs to change beyond the `//go:embed`
paths and the `sansBoldFont`/`serifItalicFont` variable names, since all
sizing/metrics are read from the font at runtime, not hardcoded.

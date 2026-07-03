package sharecard

import "strings"

// basicfont.Face7x13 (see text.go) only has glyphs for printable ASCII
// (U+0020-U+007E) plus the U+FFFD replacement character — see
// golang.org/x/image/font/basicfont's doc comment. Torró names and other
// real copy in this app are Catalan and routinely contain diacritics
// (à, ç, è, é, í, ï, ò, ó, ú, ü...) that fall outside that range; drawing
// them unmodified renders as a broken replacement-character box instead
// of the letter.
//
// asciiFold degrades those characters to a plain ASCII approximation so
// text stays legible with this bitmap-only v1 font. This is a stopgap
// for basicfont's limited coverage, not a general Unicode transliterator
// — it covers the accented Latin letters that actually show up in this
// app's Catalan/Spanish copy and product names. A future TTF (the design
// pass, see constants.go) would make this unnecessary.
var asciiFoldTable = map[rune]rune{
	'à': 'a', 'á': 'a', 'â': 'a', 'ä': 'a', 'À': 'A', 'Á': 'A', 'Â': 'A', 'Ä': 'A',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'ö': 'o', 'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Ö': 'O',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
	'ç': 'c', 'Ç': 'C',
	'ñ': 'n', 'Ñ': 'N',
	'·': '.', // interpunct, used in geminate "l·l"
	'’': '\'', '‘': '\'',
	'“': '"', '”': '"',
	'–': '-', '—': '-',
}

// asciiFold applies asciiFoldTable and drops any remaining rune outside
// basicfont's printable-ASCII range, rather than letting it fall through
// to a garbled replacement-character glyph.
func asciiFold(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		if r >= 0x20 && r <= 0x7e {
			b.WriteRune(r)
			continue
		}
		if folded, ok := asciiFoldTable[r]; ok {
			b.WriteRune(folded)
			continue
		}
		// Unsupported rune with no known fold: drop it rather than draw
		// basicfont's replacement-character box.
	}

	return b.String()
}

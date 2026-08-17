package http

import (
	"encoding/json"
	"html/template"
	"regexp"
	"strings"
	"testing"

	torrons "github.com/krtffl/torro"
)

// TestTorroTemplateJSONLD renders the product detail page with and without
// the optional product fields and asserts every JSON-LD block stays valid
// JSON — the Product block builds its optional lines with template
// conditionals, and a stray comma would silently invalidate it.
func TestTorroTemplateJSONLD(t *testing.T) {
	tmpls, err := template.New("").Funcs(templateFuncs).ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	full := TorroDetail{
		Id: "42", Name: `Torró "Únic" & Fi`, Image: "u.webp", Rating: 1600,
		HasDescription: true, Description: `Amb ametlla i "mel" d'Alcarria`,
		HasWeight: true, Weight: "300g",
		HasClassName: true, ClassName: "Clàssics",
	}
	minimal := TorroDetail{Id: "43", Name: "Torró Nu", Image: "n.webp", Rating: 1500}

	re := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)

	for name, detail := range map[string]TorroDetail{"full": full, "minimal": minimal} {
		t.Run(name, func(t *testing.T) {
			var sb strings.Builder
			if err := tmpls.ExecuteTemplate(&sb, "torro.html", TorroDetailContent{Torro: detail}); err != nil {
				t.Fatalf("failed to render: %v", err)
			}
			matches := re.FindAllStringSubmatch(sb.String(), -1)
			if len(matches) < 2 {
				t.Fatalf("expected at least 2 JSON-LD blocks (Product + BreadcrumbList), got %d", len(matches))
			}
			for i, m := range matches {
				var v any
				if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
					t.Errorf("JSON-LD block %d is not valid JSON: %v\n%s", i, err, m[1])
				}
			}
		})
	}
}

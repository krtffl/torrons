package http

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
)

// TestStaticContentPages exercises the evergreen SEO content pages (About/
// FAQ, IGP explainer, Agramunt-vs-Xixona comparison, torró glossary): each
// is a zero-DB static page, so this renders the real parsed template tree
// without needing a database.
func TestStaticContentPages(t *testing.T) {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	h := &Handler{
		template: tmpls,
		bpool:    bpool.NewBufferPool(8),
	}

	cases := []struct {
		name    string
		path    string
		handler func(http.ResponseWriter, *http.Request)
		want    string
	}{
		{"about", "/sobre", h.about, "Sobre Torrorèndum"},
		{"igp", "/torro-agramunt-igp", h.igpExplainer, "IGP Torró d'Agramunt"},
		{"comparativa", "/torro-agramunt-vs-xixona", h.agramuntVsXixona, "Torró d'Agramunt vs Torró de Xixona"},
		{"glossari", "/tipus-de-torrons", h.torroGlossary, "Glossari de torrons"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Errorf("expected body to contain %q", tc.want)
			}
		})
	}
}

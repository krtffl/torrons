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

// TestIndexHXFragment guards against the index page rendering full page
// chrome (DOCTYPE/head/header/topbar/footer) on an htmx-boosted request.
// Every other page template gates its chrome behind {{if not .HX}}, but
// index.html and its handler both missed it, so hx-boosted navigation to
// "/" (e.g. clicking the topbar logo, which targets #main-content) nested a
// second full page inside the existing header/topbar/footer, doubling them
// on screen. See torro.cat production bug, 2026-07-09.
func TestIndexHXFragment(t *testing.T) {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	h := &Handler{
		template: tmpls,
		bpool:    bpool.NewBufferPool(8),
	}

	t.Run("full navigation returns full page chrome", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.index(rec, req)

		body := rec.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("expected full page load to include DOCTYPE, got: %s", body)
		}
		if !strings.Contains(body, `id="topbar"`) {
			t.Errorf("expected full page load to include #topbar chrome")
		}
	})

	t.Run("htmx request returns fragment without page chrome", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.index(rec, req)

		body := rec.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("htmx fragment must not include a nested DOCTYPE/full page")
		}
		if strings.Contains(body, `id="topbar"`) {
			t.Errorf("htmx fragment must not include a nested #topbar (would duplicate the page's existing one)")
		}
		if strings.Contains(body, `id="header"`) {
			t.Errorf("htmx fragment must not include a nested #header (would duplicate the page's existing one)")
		}
	})
}

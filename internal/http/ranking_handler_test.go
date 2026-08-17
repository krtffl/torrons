package http

import (
	"encoding/json"
	"html/template"
	"regexp"
	"strings"
	"testing"
	"time"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
)

// rankingTestContent builds a representative RankingContent payload,
// including a torró name with characters that must survive JSON-LD escaping.
func rankingTestContent() RankingContent {
	return RankingContent{
		Entries: []LeaderboardEntry{
			{Rank: 1, TorronId: "1", TorronName: `Crema Cremada "L'Original"`, TorronImage: "a.webp", Rating: 1710.5, RatingPercentage: 100},
			{Rank: 2, TorronId: "2", TorronName: "Xocolata & Avellana", TorronImage: "b.webp", Rating: 1650.2, RatingPercentage: 61},
			{Rank: 3, TorronId: "3", TorronName: "Praliné", TorronImage: "c.webp", Rating: 1590.9, RatingPercentage: 10},
		},
		Categories: []RankingCategory{
			{
				Class:   &domain.Class{Id: "1", Name: "Clàssics"},
				Entries: []LeaderboardEntry{{Rank: 1, TorronId: "1", TorronName: "Crema Cremada", Rating: 1710.5}},
			},
		},
		TotalVotes:   12345,
		UpdatedAt:    "17 d'agost de 2026",
		UpdatedAtISO: "2026-08-17",
	}
}

// TestRankingTemplate renders the public ranking page both as a full page
// and as an htmx fragment, and validates that every JSON-LD block on the
// full page parses as JSON (guarding the range-generated ItemList).
func TestRankingTemplate(t *testing.T) {
	tmpls, err := template.New("").Funcs(templateFuncs).ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	t.Run("full page", func(t *testing.T) {
		var sb strings.Builder
		content := rankingTestContent()
		if err := tmpls.ExecuteTemplate(&sb, "ranquing.html", content); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		body := sb.String()

		for _, want := range []string{
			"<!DOCTYPE html>",
			"Rànquing de torrons",
			"12345",
			`rel="canonical" href="https://torro.cat/ranquing-de-torrons"`,
			`"@type": "ItemList"`,
			"/torro/1",
			"Clàssics",
			"17 d&#39;agost de 2026",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("expected body to contain %q", want)
			}
		}

		// Every JSON-LD block must be valid JSON after template execution -
		// a trailing comma from the range loop would silently invalidate the
		// structured data.
		re := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
		matches := re.FindAllStringSubmatch(body, -1)
		if len(matches) < 2 {
			t.Fatalf("expected at least 2 JSON-LD blocks, got %d", len(matches))
		}
		for i, m := range matches {
			var v any
			if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
				t.Errorf("JSON-LD block %d is not valid JSON: %v\n%s", i, err, m[1])
			}
		}
	})

	t.Run("millors-vicens full page", func(t *testing.T) {
		var sb strings.Builder
		content := rankingTestContent()
		if err := tmpls.ExecuteTemplate(&sb, "millors_vicens.html", content); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		body := sb.String()

		for _, want := range []string{
			"Els millors torrons Vicens",
			`rel="canonical" href="https://torro.cat/millors-torrons-vicens"`,
			"projecte de fans independent",
			"/torro/1",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("expected body to contain %q", want)
			}
		}

		re := regexp.MustCompile(`(?s)<script type="application/ld\+json">(.*?)</script>`)
		for i, m := range re.FindAllStringSubmatch(body, -1) {
			var v any
			if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
				t.Errorf("JSON-LD block %d is not valid JSON: %v\n%s", i, err, m[1])
			}
		}
	})

	t.Run("turron-agramunt-es full page", func(t *testing.T) {
		var sb strings.Builder
		if err := tmpls.ExecuteTemplate(&sb, "turron_agramunt_es.html", Content{}); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		body := sb.String()
		for _, want := range []string{
			`<html lang="es">`,
			`hreflang="ca" href="https://torro.cat/torro-agramunt-igp"`,
			`hreflang="x-default"`,
			"Turrón de Agramunt",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("expected body to contain %q", want)
			}
		}
	})

	t.Run("hx fragment", func(t *testing.T) {
		var sb strings.Builder
		content := rankingTestContent()
		content.HX = true
		if err := tmpls.ExecuteTemplate(&sb, "ranquing.html", content); err != nil {
			t.Fatalf("failed to render: %v", err)
		}
		if strings.Contains(sb.String(), "<!DOCTYPE html>") {
			t.Error("hx fragment must not include the full page shell")
		}
	})
}

// TestFormatCatalanDate covers the de/d' vowel elision rule.
func TestFormatCatalanDate(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2026, time.August, 17, 0, 0, 0, 0, time.UTC), "17 d'agost de 2026"},
		{time.Date(2026, time.January, 6, 0, 0, 0, 0, time.UTC), "6 de gener de 2026"},
		{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), "1 d'octubre de 2026"},
		{time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC), "25 de desembre de 2026"},
	}
	for _, tc := range cases {
		if got := formatCatalanDate(tc.date); got != tc.want {
			t.Errorf("formatCatalanDate(%v) = %q, want %q", tc.date, got, tc.want)
		}
	}
}

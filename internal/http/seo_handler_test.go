package http

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
)

// fakeTorroRepo is a minimal stand-in for domain.TorroRepo, used only by
// sitemapXML's use of List.
type fakeTorroRepo struct {
	torros []*domain.Torro
}

func (f *fakeTorroRepo) Get(ctx context.Context, id string) (*domain.Torro, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeTorroRepo) List(ctx context.Context) ([]*domain.Torro, error) {
	return f.torros, nil
}
func (f *fakeTorroRepo) ListByClass(ctx context.Context, classId string) ([]*domain.Torro, error) {
	return nil, nil
}
func (f *fakeTorroRepo) ListFiltered(ctx context.Context, classId string, filter domain.TorroFilter) ([]*domain.Torro, error) {
	return nil, nil
}
func (f *fakeTorroRepo) Update(ctx context.Context, id string, rating float64) (*domain.Torro, error) {
	return nil, nil
}
func (f *fakeTorroRepo) TopNByClass(ctx context.Context, classId string, n int) ([]*domain.Torro, error) {
	return nil, nil
}
func (f *fakeTorroRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.Torro, error) {
	return nil, nil
}
func (f *fakeTorroRepo) UpdateTx(tx *sql.Tx, ctx context.Context, id string, rating float64) (*domain.Torro, error) {
	return nil, nil
}

// fakeBracketRepo is a minimal stand-in for domain.BracketRepo, used only by
// sitemapXML's use of GetLatestByClass. The embedded nil interface satisfies
// every other method the real interface requires - safe here because
// sitemapXML never calls them, unlike the full-DB integration tests that
// exercise the rest of BracketRepo.
type fakeBracketRepo struct {
	domain.BracketRepo
	brackets map[string]*domain.Bracket // classId -> latest bracket, absent = none
}

func (f *fakeBracketRepo) GetLatestByClass(ctx context.Context, classId string) (*domain.Bracket, error) {
	return f.brackets[classId], nil
}

func TestRobotsTxt(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()

	robotsTxt(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "User-agent: *") || !strings.Contains(body, "Allow: /") {
		t.Errorf("robots.txt missing expected directives, got: %q", body)
	}
	if !strings.Contains(body, "Sitemap: https://torro.cat/sitemap.xml") {
		t.Errorf("robots.txt missing sitemap reference, got: %q", body)
	}
}

func TestSitemapXML(t *testing.T) {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	h := &Handler{
		template: tmpls,
		bpool:    bpool.NewBufferPool(8),
		torroRepo: &fakeTorroRepo{torros: []*domain.Torro{
			{Id: "torro-1", Name: "Torró de Xocolata"},
			{Id: "torro-2", Name: "Torró d'Ametlla"},
		}},
		classRepo: &fakeClassRepo{classes: []*domain.Class{
			{Id: "1", Name: "Clàssics"},
			{Id: "5", Name: "Global"},
		}},
		bracketRepo: &fakeBracketRepo{brackets: map[string]*domain.Bracket{
			// Only class "5" has a bracket - class "1" deliberately absent
			// to exercise the "no bracket yet" skip path.
			"5": {Id: "bracket-1", ClassId: "5"},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec := httptest.NewRecorder()

	h.sitemapXML(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/xml") {
		t.Errorf("expected application/xml content type, got %q", got)
	}
	for _, want := range []string{
		"<loc>https://torro.cat/</loc>",
		"<loc>https://torro.cat/torro/torro-1</loc>",
		"<loc>https://torro.cat/torro/torro-2</loc>",
		"<loc>https://torro.cat/bracket/5</loc>",
		"<loc>https://torro.cat/sobre</loc>",
		"<loc>https://torro.cat/torro-agramunt-igp</loc>",
		"<loc>https://torro.cat/torro-agramunt-vs-xixona</loc>",
		"<loc>https://torro.cat/tipus-de-torrons</loc>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("sitemap missing expected URL %q, got: %s", want, body)
		}
	}
	if strings.Contains(body, "<loc>https://torro.cat/bracket/1</loc>") {
		t.Errorf("sitemap should omit /bracket/1 - class 1 has no bracket yet, got: %s", body)
	}
}

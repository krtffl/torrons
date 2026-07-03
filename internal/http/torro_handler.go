package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// TorroDetail is the presentation model for the product detail page.
// Extended product fields were added to the "Torrons" table in migration
// 000011 and are still being backfilled, so every nullable/optional domain
// field is resolved here into a plain field plus a "Has*" flag - the
// template never has to deal with nil pointers or empty slices directly.
type TorroDetail struct {
	Id     string
	Name   string
	Image  string
	Class  string
	Rating float64

	HasRank bool
	Rank    int

	HasDescription bool
	Description    string

	HasWeight bool
	Weight    string

	HasPrice bool
	Price    float64

	HasProductUrl bool
	ProductUrl    string

	Allergens       []string
	MainIngredients []string

	IsVegan       bool
	IsGlutenFree  bool
	IsLactoseFree bool
	IsOrganic     bool
	HasDietInfo   bool

	HasIntensity   bool
	IntensityLevel int
	IntensityDots  []bool // len 5, true = filled dot
}

// TorroDetailContent holds data for the product detail page template.
type TorroDetailContent struct {
	HX    bool
	Torro TorroDetail
}

// torroDetail renders the product detail page for a single torró.
func (h *Handler) torroDetail(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - TorroDetail] Incoming request")

	id := chi.URLParam(r, "id")

	t, err := h.torroRepo.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), string(domain.NotFoundError)) {
			logger.Warn("[Handler - TorroDetail] Torró not found: %s", id)
			http.Error(w, "Torró no trobat", http.StatusNotFound)
			return
		}
		logger.Error("[Handler - TorroDetail] Couldn't get torro %s. %v", id, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rank := h.getRankInClass(r.Context(), t)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "torro.html", TorroDetailContent{
		HX:    isHX(r),
		Torro: newTorroDetail(t, rank),
	}); err != nil {
		logger.Error("[Handler - TorroDetail] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// getRankInClass computes a torró's 1-based rank within its own class based
// on its global rating. Returns 0 (unknown/not applicable) if the torró is
// discontinued, has no class, or the rank can't be determined.
func (h *Handler) getRankInClass(ctx context.Context, t *domain.Torro) int {
	if t.Discontinued || t.Class == "" {
		return 0
	}

	torrons, err := h.torroRepo.ListFiltered(ctx, t.Class, domain.TorroFilter{})
	if err != nil {
		logger.Warn("[Handler - TorroDetail] Couldn't compute rank for %s. %v", t.Id, err)
		return 0
	}

	rank := 0
	for _, candidate := range torrons {
		if candidate.Discontinued {
			continue
		}
		rank++
		if candidate.Id == t.Id {
			return rank
		}
	}

	return 0
}

// newTorroDetail builds the presentation model for the detail page from the
// domain torró, resolving nullable/optional fields defensively so nothing
// ever renders as a bare pointer or "undefined".
func newTorroDetail(t *domain.Torro, rank int) TorroDetail {
	detail := TorroDetail{
		Id:              t.Id,
		Name:            t.Name,
		Image:           t.Image,
		Class:           t.Class,
		Rating:          t.Rating,
		Allergens:       t.Allergens,
		MainIngredients: t.MainIngredients,
		IsVegan:         t.IsVegan,
		IsGlutenFree:    t.IsGlutenFree,
		IsLactoseFree:   t.IsLactoseFree,
		IsOrganic:       t.IsOrganic,
		HasDietInfo:     t.IsVegan || t.IsGlutenFree || t.IsLactoseFree || t.IsOrganic,
		IntensityDots:   make([]bool, 5),
	}

	if rank > 0 {
		detail.HasRank = true
		detail.Rank = rank
	}

	if t.Description != nil && *t.Description != "" {
		detail.HasDescription = true
		detail.Description = *t.Description
	}

	if t.Weight != nil && *t.Weight != "" {
		detail.HasWeight = true
		detail.Weight = *t.Weight
	}

	if t.Price != nil {
		detail.HasPrice = true
		detail.Price = *t.Price
	}

	if t.ProductUrl != nil && *t.ProductUrl != "" {
		detail.HasProductUrl = true
		detail.ProductUrl = *t.ProductUrl
	}

	if t.IntensityLevel != nil && *t.IntensityLevel > 0 {
		level := *t.IntensityLevel
		if level > 5 {
			level = 5
		}
		detail.HasIntensity = true
		detail.IntensityLevel = level
		for i := 0; i < level; i++ {
			detail.IntensityDots[i] = true
		}
	}

	return detail
}

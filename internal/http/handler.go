package http

import (
	"fmt"
	"html/template"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

const K = 33

type Content struct {
	Torrons  []*domain.Torro
	Pairing  *domain.Pairing
	Pairings []*domain.Pairing
	Classes  []*domain.Class
	HX       bool
}

type Handler struct {
	template    *template.Template
	bpool       *bpool.BufferPool
	pairingRepo domain.PairingRepo
	torroRepo   domain.TorroRepo
	classRepo   domain.ClassRepo
	resultRepo  domain.ResultRepo
}

func NewHandler(
	bpool *bpool.BufferPool,
	pairingRep domain.PairingRepo,
	torroRepo domain.TorroRepo,
	classRepo domain.ClassRepo,
	resultRepo domain.ResultRepo,
) *Handler {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		logger.Fatal("[Handler] - Failed to parse templates. %v", err)
	}

	return &Handler{
		template:    tmpls,
		bpool:       bpool,
		pairingRepo: pairingRep,
		torroRepo:   torroRepo,
		classRepo:   classRepo,
		resultRepo:  resultRepo,
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Index] Incoming request")

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "index.html", Content{}); err != nil {
		logger.Error("[Handler - Index] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func (h *Handler) classes(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Classes] Incoming request")

	classes, err := h.classRepo.List()
	if err != nil {
		logger.Error("[Handler - Classes] Couldn't list classes. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "classes.html", Content{
		Classes: classes,
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Classes ] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Vote] Incoming request")

	classId := chi.URLParam(r, "id")

	p, err := h.pairingRepo.GetRandom(classId)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get random pairing. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t1, err := h.torroRepo.Get(p.Torro1)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t2, err := h.torroRepo.Get(p.Torro2)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t1.Pairing = p.Id
	t2.Pairing = p.Id

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "vote.html", Content{
		Pairing: p,
		Torrons: []*domain.Torro{t1, t2},
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Classes ] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func (h *Handler) result(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Result] Incoming request")

	pairingId := chi.URLParam(r, "id")

	p, err := h.pairingRepo.Get(pairingId)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get pairing with ID %s. %v", pairingId, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t1, err := h.torroRepo.Get(p.Torro1)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t2, err := h.torroRepo.Get(p.Torro2)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	exp1 := 1.0 / (1.0 + math.Pow(10, (t2.Rating-t1.Rating)/400))
	exp2 := 1.0 / (1.0 + math.Pow(10, (t1.Rating-t2.Rating)/400))

	winnderId := r.URL.Query().Get("id")

	var new1, new2 float64
	if winnderId == t1.Id {
		new1 = t1.Rating + K*(1-exp1)
		new2 = t2.Rating + K*(0-exp2)
	} else {
		new1 = t1.Rating + K*(0-exp1)
		new2 = t2.Rating + K*(1-exp2)
	}

	_, err = h.resultRepo.Create(&domain.Result{
		Pairing: pairingId,
		Rat1Bef: t1.Rating,
		Rat2Bef: t2.Rating,
		Winner:  winnderId,
		Rat1Aft: new1,
		Rat2Aft: new2,
	})
	if err != nil {
		logger.Error("[Handler - Result] Couldn't create result. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return

	}

	_, err = h.torroRepo.Update(t1.Id, new1)
	if err != nil {
		logger.Error("[Handler - Result] Coulnd't update rating: %s. %v", t1.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return

	}

	_, err = h.torroRepo.Update(t2.Id, new2)
	if err != nil {
		logger.Error("[Handler - Result] Coulnd't update rating: %s. %v", t2.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	fmt.Print("Votació OK")
	newP, err := h.pairingRepo.GetRandom(p.Class)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get random pairing. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt1, err := h.torroRepo.Get(p.Torro1)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt2, err := h.torroRepo.Get(p.Torro2)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt1.Pairing = newP.Id
	newt2.Pairing = newP.Id

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "pairing.html", Content{
		Pairing: newP,
		Torrons: []*domain.Torro{newt1, newt2},
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Classes ] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func isHX(r *http.Request) bool {
	if r.Header.Get("HX-Request") == "true" {
		return true
	}
	return false
}

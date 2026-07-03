package http

import (
	"hash/fnv"
	"net/http"
	"time"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// adventClassRotation lists which classes are featured for the daily duel,
// rotating deterministically by day-of-year. Judgment call: rotate through
// the four regular categories (Clàssics, Novetats, Xocolata, Albert Adrià)
// rather than always featuring Global, so the daily duel showcases a
// different category across the season instead of always the same one.
var adventClassRotation = []string{"1", "2", "3", "4"}

// AdventContent holds data for the advent page template
type AdventContent struct {
	HX             bool
	CampaignActive bool
	AlreadyVoted   bool
	ClassName      string
	Torrons        []*domain.Torro
}

// advent handles the "advent daily duel" page: one featured, pre-selected
// pairing per calendar day, the same for every visitor, gated to one vote
// per user per day.
func (h *Handler) advent(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Advent] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	content := AdventContent{HX: isHX(r)}

	// No active campaign right now -- friendly boundary state instead of an
	// error. "Advent" only makes sense while a campaign is running.
	if _, err := h.campaignRepo.GetActive(r.Context()); err != nil {
		content.CampaignActive = false
		h.renderAdvent(w, content)
		return
	}
	content.CampaignActive = true

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	classId := adventClassRotation[now.YearDay()%len(adventClassRotation)]

	if classes, err := h.classRepo.List(r.Context()); err == nil {
		for _, c := range classes {
			if c.Id == classId {
				content.ClassName = c.Name
				break
			}
		}
	}

	if userId != "" {
		voted, err := h.adventVoteRepo.HasVotedToday(r.Context(), userId, dateStr)
		if err != nil {
			logger.Error("[Handler - Advent] Couldn't check today's advent vote. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if voted {
			content.AlreadyVoted = true
			h.renderAdvent(w, content)
			return
		}
	}

	seed := adventSeed(dateStr, classId)
	p, err := h.pairingRepo.GetDeterministic(r.Context(), classId, seed)
	if err != nil {
		logger.Error("[Handler - Advent] Couldn't get today's pairing. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t1, err := h.torroRepo.Get(r.Context(), p.Torro1)
	if err != nil {
		logger.Error("[Handler - Advent] Couldn't get torro. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t2, err := h.torroRepo.Get(r.Context(), p.Torro2)
	if err != nil {
		logger.Error("[Handler - Advent] Couldn't get torro. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t1.Pairing = p.Id
	t2.Pairing = p.Id
	content.Torrons = []*domain.Torro{t1, t2}

	h.renderAdvent(w, content)
}

func (h *Handler) renderAdvent(w http.ResponseWriter, content AdventContent) {
	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "advent.html", content); err != nil {
		logger.Error("[Handler - Advent] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// adventSeed deterministically derives a seed from the date and featured
// class so the same pairing is picked by every request/replica on a given
// day, without needing to persist "today's pick" anywhere.
func adventSeed(dateStr, classId string) int64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(dateStr + ":" + classId))
	return int64(hasher.Sum64())
}

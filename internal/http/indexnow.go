package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/krtffl/torro/internal/logger"
)

// indexNowEndpoint is the shared submission API; a ping here propagates to
// every IndexNow-participating engine (Bing, Yandex, Seznam, Naver...).
// Google does not support IndexNow - it is covered by the sitemap instead.
const indexNowEndpoint = "https://api.indexnow.org/indexnow"

// indexNowInterval is how often the pinger checks whether the voting data
// changed. Daily matches how fast the ranking pages meaningfully move and
// stays far inside IndexNow's rate expectations.
const indexNowInterval = 24 * time.Hour

// indexNowPages returns the vote-driven pages whose content genuinely
// changes as votes arrive - the only URLs it is honest to re-submit on data
// change. Must stay in sync with the sitemap entries stamped with
// votesLastMod (seo_handler.go); the per-category pages are derived from
// categoryPages so new ones are picked up automatically.
func indexNowPages() []string {
	pages := []string{
		siteBaseURL + "/ranquing-de-torrons",
		siteBaseURL + "/es/ranking-de-turrones",
		siteBaseURL + "/millors-torrons-vicens",
	}
	for _, p := range categoryPages {
		pages = append(pages, siteBaseURL+"/"+p.Slug)
	}
	return pages
}

// validIndexNowKey reports whether key matches the IndexNow spec's allowed
// charset (a-z, A-Z, 0-9 and dashes, 8-128 chars). Anything else must not
// be interpolated into a chi route pattern ('{', '}' and '*' panic the
// router at startup) or a query string.
func validIndexNowKey(key string) bool {
	if len(key) < 8 || len(key) > 128 {
		return false
	}
	for _, c := range key {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-':
		default:
			return false
		}
	}
	return true
}

// runIndexNowPinger loops until ctx is cancelled, submitting the vote-driven
// pages to IndexNow whenever a vote landed since the previous check. The
// first check runs shortly after boot, then every indexNowInterval.
func (h *Handler) runIndexNowPinger(ctx context.Context, key string) {
	// Small boot delay so a crash-looping process can't hammer the API.
	timer := time.NewTimer(1 * time.Minute)
	defer timer.Stop()

	// Seeded one interval back rather than at zero: a restart only
	// re-submits when a vote actually landed within the last interval
	// (content genuinely changed recently), instead of every boot
	// re-announcing months-old data as if it were fresh.
	lastSeen := time.Now().Add(-indexNowInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		// Bounded like submitIndexNow's request: with the unbounded server
		// ctx, one wedged query would silently kill this loop for the life
		// of the process.
		readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		latest, err := h.pressStatsRepo.LatestVoteTime(readCtx)
		cancel()
		if err != nil {
			logger.Warn("[IndexNow] Couldn't read latest vote time. %v", err)
		} else if latest != nil && latest.After(lastSeen) {
			pages := indexNowPages()
			if err := submitIndexNow(ctx, key, pages); err != nil {
				logger.Warn("[IndexNow] Submission failed. %v", err)
			} else {
				logger.Info("[IndexNow] Submitted %d URLs (data changed at %s)",
					len(pages), latest.Format(time.RFC3339))
				lastSeen = *latest
			}
		}

		timer.Reset(indexNowInterval)
	}
}

// submitIndexNow POSTs the URL list to the shared IndexNow endpoint.
func submitIndexNow(ctx context.Context, key string, urls []string) error {
	host := strings.TrimPrefix(siteBaseURL, "https://")
	payload, err := json.Marshal(map[string]any{
		"host":    host,
		"key":     key,
		"urlList": urls,
	})
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, indexNowEndpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	// Drain before closing so the keep-alive connection can be reused.
	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	// 200 and 202 both mean accepted per the IndexNow spec.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("indexnow returned %d", resp.StatusCode)
	}

	return nil
}

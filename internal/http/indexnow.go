package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// indexNowPages are the vote-driven pages whose content genuinely changes
// as votes arrive - the only URLs it is honest to re-submit on data change.
var indexNowPages = []string{
	siteBaseURL + "/ranquing-de-torrons",
	siteBaseURL + "/millors-torrons-vicens",
}

// runIndexNowPinger loops until ctx is cancelled, submitting the vote-driven
// pages to IndexNow whenever a vote landed since the previous check. The
// first check runs shortly after boot (so a deploy with fresh data
// propagates without waiting a day), then every indexNowInterval.
func (h *Handler) runIndexNowPinger(ctx context.Context, key string) {
	// Small boot delay so a crash-looping process can't hammer the API.
	timer := time.NewTimer(1 * time.Minute)
	defer timer.Stop()

	var lastSeen time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		latest, err := h.pressStatsRepo.LatestVoteTime(ctx)
		if err != nil {
			logger.Warn("[IndexNow] Couldn't read latest vote time. %v", err)
		} else if latest != nil && latest.After(lastSeen) {
			if err := submitIndexNow(ctx, key, indexNowPages); err != nil {
				logger.Warn("[IndexNow] Submission failed. %v", err)
			} else {
				logger.Info("[IndexNow] Submitted %d URLs (data changed at %s)",
					len(indexNowPages), latest.Format(time.RFC3339))
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
	defer resp.Body.Close()

	// 200 and 202 both mean accepted per the IndexNow spec.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("indexnow returned %d", resp.StatusCode)
	}

	return nil
}

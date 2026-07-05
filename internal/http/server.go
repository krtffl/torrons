package http

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/logger"
)

type Server struct {
	ctx        context.Context
	shutdownFn context.CancelFunc

	port    uint
	handler *Handler
}

func New(
	port uint,
	handler *Handler,
) *Server {
	ctx, shutdownFn := context.WithCancel(context.Background())

	return &Server{
		ctx:        ctx,
		shutdownFn: shutdownFn,
		port:       port,
		handler:    handler,
	}
}

func (srv *Server) Run() error {
	logger.Info("[HTTP] - Starting to listen on port %d", srv.port)
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Rate limiting middleware (100 requests per minute per IP)
	r.Use(httprate.Limit(
		100,                                     // requests
		1*time.Minute,                           // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP), // per IP address
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn("[HTTP Server] Rate limit exceeded for IP: %s", r.RemoteAddr)
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		}),
	))

	// Per-user rate limiting for vote-casting routes, layered on top of the
	// blanket per-IP limit above. Keyed by the anonymous user cookie (set
	// by UserMiddleware, registered below - by the time a request reaches
	// this route-scoped middleware every global middleware has already
	// run) rather than IP, so it throttles one identity hammering the vote
	// endpoint rather than one network address. 20/minute comfortably
	// covers fast human clicking while meaningfully slowing scripted ELO
	// manipulation - see docs/design-prompts or project notes for why a
	// permanent per-pairing block isn't used instead: GetRandom draws
	// repeatedly from the same small fixed pairing pool per class, so
	// blocking a re-vote outright would eventually lock a user out of
	// voting entirely once they'd seen every pairing once.
	voteRateLimiter := httprate.Limit(
		20,
		1*time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return GetUserIDFromContext(r.Context()), nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn("[HTTP Server] Vote rate limit exceeded for user %s", GetUserIDFromContext(r.Context()))
			http.Error(w, "You're voting too quickly. Please slow down.", http.StatusTooManyRequests)
		}),
	)

	// Security headers middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS: Force HTTPS for 1 year, including subdomains
			// Note: Only enable after deploying with HTTPS!
			// Uncomment in production with HTTPS enabled
			// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// Prevent MIME sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Enable XSS filter
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Control referrer information
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy: Restrict browser features
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// X-Permitted-Cross-Domain-Policies
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// The embeddable leaderboard widget (/embed/*) is meant to be
			// loaded cross-origin inside a third party's <iframe> - that's
			// the entire point of shipping it (backlink/SEO value).
			// X-Frame-Options: DENY blocks all framing regardless of CSP,
			// so that path gets its own, deliberately permissive framing
			// policy instead. Every other route keeps the exact previous
			// behavior unchanged.
			if strings.HasPrefix(r.URL.Path, "/embed/") {
				w.Header().Set("Content-Security-Policy",
					"default-src 'self'; "+
						"style-src 'self' 'unsafe-inline'; "+
						"img-src 'self' data:; "+
						"frame-ancestors *")
			} else {
				// Prevent clickjacking
				w.Header().Set("X-Frame-Options", "DENY")

				// Content Security Policy (configured for HTMX app)
				w.Header().Set("Content-Security-Policy",
					"default-src 'self'; "+
						"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
						"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
						"font-src 'self' https://fonts.gstatic.com; "+
						"img-src 'self' data:")
			}

			next.ServeHTTP(w, r)
		})
	})

	// User tracking middleware - identifies users via cookies
	r.Use(srv.handler.UserMiddleware)

	assets, err := fs.Sub(torrons.Public, "public")
	if err != nil {
		logger.Fatal("[HTTP Server] - Failed to run templates. %v", err)
	}

	fs := http.FileServer(http.FS(assets))

	r.Handle(
		"/public/*",
		http.StripPrefix("/public/", fs),
	)

	// ********** W E B  U I **********
	r.Route("/", func(r chi.Router) {
		r.Get("/healthcheck", handleHealthcheck)

		r.Get("/", srv.handler.index)

		r.Get("/classes", srv.handler.classes)
		r.Get("/classes/{id}/vote", srv.handler.vote)

		r.With(voteRateLimiter).Post("/pairings/{id}/vote", srv.handler.result)

		// Product detail page
		r.Get("/torro/{id}", srv.handler.torroDetail)

		// Leaderboard visualization
		r.Get("/leaderboard", srv.handler.leaderboard)

		// User statistics page
		r.Get("/stats", srv.handler.stats)

		// Voting history page
		r.Get("/history", srv.handler.history)

		// Shareable result card (PNG). Registered without the ".png"
		// suffix: the global middleware.URLFormat (registered above)
		// strips any trailing ".ext" from the routing path before chi
		// matches it, so a route registered as "/share/card.png" would
		// 404 on that exact request. The public URL clients hit is still
		// GET /share/card.png.
		r.Get("/share/card", srv.handler.shareCard)

		// Phase 2 - The knockout bracket (separate mechanic from the
		// open-voting/ELO pairings above; see internal/domain/bracket.go).
		r.Get("/bracket/{classId}", srv.handler.bracketOverview)
		r.Get("/bracket/{classId}/vote", srv.handler.bracketVote)
		r.With(voteRateLimiter).Post("/bracket/match/{matchId}/vote", srv.handler.bracketMatchVote)

		// Admin-only bracket management, gated by a shared-secret bearer
		// token. See Handler.RequireAdminToken (middleware.go) and the
		// AdminToken config value (internal/config/config.go).
		r.With(srv.handler.RequireAdminToken).Post("/bracket/{classId}/create", srv.handler.bracketCreate)
		r.With(srv.handler.RequireAdminToken).Post("/bracket/{bracketId}/advance", srv.handler.bracketAdvance)

		// Advent daily duel: one featured pairing per calendar day
		r.Get("/advent", srv.handler.advent)

		// Friend circles: shareable, invite-based leaderboards
		r.Get("/friends", srv.handler.friendsIndex)
		r.Post("/friends/create", srv.handler.friendsCreate)
		r.Get("/friends/join/{inviteCode}", srv.handler.friendsJoin)
		r.Get("/friends/{circleId}", srv.handler.friendsLeaderboard)

		// Embeddable leaderboard widget, designed to be loaded cross-origin
		// inside a third party's <iframe> (see the security-headers and
		// UserMiddleware /embed/ special-casing above/in middleware.go).
		r.Get("/embed/leaderboard", srv.handler.embedLeaderboard)

		// Press/stats data page: a permanent, screenshot-friendly page of
		// public aggregate stats, plus the embed snippet generator for the
		// widget above.
		r.Get("/premsa", srv.handler.press)

		// Personal "Torrorèndum Wrapped" campaign recap: page + PNG story
		// card, gated behind the same minimum-vote threshold as the Global
		// leaderboard. Registered without the ".png" suffix for the same
		// reason as /share/card above.
		r.Get("/wrapped", srv.handler.wrapped)
		r.Get("/wrapped/card", srv.handler.wrappedCard)

		// Press-kit aggregate one-pager PNG (the Gran Final's champion),
		// same ".png"-suffix-stripping caveat as above.
		r.Get("/press-kit/card", srv.handler.pressKitCard)
	})
	// **********        **********

	// ********** U S E R  A P I **********
	r.Route("/api/user", func(r chi.Router) {
		// Get current user's statistics
		r.Get("/stats", srv.handler.handleUserStats)

		// Get personalized leaderboard for a class
		r.Get("/leaderboard/class/{classId}", srv.handler.handleUserLeaderboard)

		// Get personalized global leaderboard
		r.Get("/leaderboard/global", srv.handler.handleUserGlobalLeaderboard)
	})
	// **********           **********

	// ********** C A M P A I G N  A P I **********
	r.Route("/api/campaign", func(r chi.Router) {
		// Get countdown to campaign end (JSON)
		r.Get("/countdown", srv.handler.handleCountdown)

		// Get countdown widget (HTML)
		r.Get("/countdown/widget", srv.handler.handleCountdownWidget)

		// Get active campaign information
		r.Get("/info", srv.handler.handleCampaignInfo)
	})

	r.Route("/api/leaderboard", func(r chi.Router) {
		// Get global leaderboard across all categories
		r.Get("/global", srv.handler.handleGlobalLeaderboard)

		// Get class-specific global leaderboard
		r.Get("/class/{classId}", srv.handler.handleClassLeaderboard)
	})
	// **********                **********

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", srv.port),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	go func() {
		<-srv.ctx.Done()
		if err := httpServer.Shutdown(srv.ctx); err != nil {
			logger.Error("[HTTP Server] - Failed to shutdown on port %d. %v", srv.port, err)
		}
		logger.Info("[HTTP Server] - Server on port %d has shutdown", srv.port)
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("[HTTP Server] - %v", err)
	}

	return nil
}

func (srv *Server) Shutdown() {
	logger.Info("[HTTP Server] - Server on port %d shutting down", srv.port)
	srv.shutdownFn()
	logger.Info("[HTTP Server] - Server on port %d shutted down", srv.port)
}

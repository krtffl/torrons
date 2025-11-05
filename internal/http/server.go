package http

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
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
		100,                       // requests
		1*time.Minute,             // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP), // per IP address
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.Warn("[HTTP Server] Rate limit exceeded for IP: %s", r.RemoteAddr)
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		}),
	))

	// Security headers middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS: Force HTTPS for 1 year, including subdomains
			// Note: Only enable after deploying with HTTPS!
			// Uncomment in production with HTTPS enabled
			// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// Prevent MIME sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// Enable XSS filter
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Control referrer information
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy: Restrict browser features
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// X-Permitted-Cross-Domain-Policies
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Content Security Policy (configured for HTMX app)
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"img-src 'self' data:")

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

		r.Post("/pairings/{id}/vote", srv.handler.result)

		// Leaderboard visualization
		r.Get("/leaderboard", srv.handler.leaderboard)
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
		// Get countdown to campaign end
		r.Get("/countdown", srv.handler.handleCountdown)

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

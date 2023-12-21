package http

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/logger"
)

type Server struct {
	ctx        context.Context
	shutdownFn context.CancelFunc

	port uint

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
	})
	// **********        **********

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.port),
		Handler: r,
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

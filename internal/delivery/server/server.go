package server

import (
	"betera-tz/internal/config"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type AppServer struct {
	Server *http.Server
}

func NewAppServer(scfg config.ServerConfig, si ServerInterface) *AppServer {
	r := chi.NewRouter()
	r.Use(middleware.RedirectSlashes)
	h := HandlerFromMux(si, r)
	return &AppServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", scfg.Host, scfg.Port),
			Handler: h,
		},
	}
}

func (as *AppServer) MustClose(ctx context.Context) {
	if err := as.Server.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("faield to close app server: %w", err))
	}
}

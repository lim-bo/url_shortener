package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/limbo/url_shortener/internal/settings"
)

type LinksManager interface {
	SaveURL(link string) (string, error)
	GetLinkByCode(code string) (string, error)
}

type Server struct {
	mx *chi.Mux
	lm LinksManager
}

func New(lm LinksManager) *Server {
	return &Server{
		mx: chi.NewMux(),
		lm: lm,
	}
}

func (s *Server) moundEndpoints() {
	cfg := settings.GetConfig()
	s.mx.Use(s.CORSMiddleware)
	s.mx.Route("/api/v"+cfg.GetString("api_version"), func(r chi.Router) {
		r.Post("/shorten", s.shorten)
		r.Route("/stats", func(r chi.Router) {
			r.Get("/{short_code}", s.clickStats)
		})
	})
	s.mx.Get("/{short_code}", s.redirect)
}

func (s *Server) Run() error {
	s.moundEndpoints()
	cfg := settings.GetConfig()
	address := cfg.GetString("api_address")
	return http.ListenAndServe(address, s.mx)
}

package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/limbo/url_shortener/internal/settings"
	"github.com/limbo/url_shortener/models"
	httpSwagger "github.com/swaggo/http-swagger"
)

type LinksManager interface {
	SaveURL(link string) (string, error)
	GetLinkByCode(code string) (string, error)
}

type CacheManager interface {
	CacheLink(shortCode string, link string) error
	GetLink(shortCode string) (string, error)
}

type StatsManager interface {
	IncreaseClicks(link, code string) error
	GetStats(code string) (*models.ClicksStat, error)
}

type Server struct {
	mx        *chi.Mux
	links     LinksManager
	cache     CacheManager
	statistic StatsManager
}

type ServerCfg struct {
	Lm LinksManager
	Cm CacheManager
	Sm StatsManager
}

func New(cfg ServerCfg) *Server {
	return &Server{
		mx:        chi.NewMux(),
		links:     cfg.Lm,
		cache:     cfg.Cm,
		statistic: cfg.Sm,
	}
}

func (s *Server) moundEndpoints() {
	cfg := settings.GetConfig()
	s.mx.Use(s.CORSMiddleware, s.RequestIDMiddleware, s.HTTPMetricsMiddleware)
	s.mx.Route("/api/v"+cfg.GetString("api_version"), func(r chi.Router) {
		r.Post("/shorten", s.shorten)
		r.Route("/stats", func(r chi.Router) {
			r.Get("/{short_code}", s.clickStats)
		})
	})
	s.mx.Get("/{short_code}", s.redirect)

	s.mx.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}

func (s *Server) Run() error {
	s.moundEndpoints()
	cfg := settings.GetConfig()
	address := cfg.GetString("api_address")
	slog.Info("server running on " + address)
	return http.ListenAndServe(address, s.mx)
}

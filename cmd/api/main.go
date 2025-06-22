// @title URL-shortener API
// @description API for url-shortener app
// @BasePath /
// @schemes http
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/limbo/url_shortener/docs"
	"github.com/limbo/url_shortener/internal/api"
	"github.com/limbo/url_shortener/internal/logger"
	"github.com/limbo/url_shortener/internal/settings"
	cache "github.com/limbo/url_shortener/internal/url_cache_manager"
	urlmanager "github.com/limbo/url_shortener/internal/url_manager"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := settings.GetConfig()

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	go func() {
		r := chi.NewMux()
		r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Fatal(http.ListenAndServe(":2112", r))
	}()

	lm := urlmanager.New(urlmanager.DBCfg{
		Address:  cfg.GetString("links_db_address"),
		User:     cfg.GetString("links_db_user"),
		Password: cfg.GetString("links_db_pass"),
		DBName:   cfg.GetString("links_db_name"),
	}, &urlmanager.CodeGenerator{})

	cm := cache.New(cache.RedisConfig{
		Address:  cfg.GetString("redis_address"),
		Password: cfg.GetString("redis_pass"),
	})

	lgger := slog.New(
		logger.NewContextHandler(slog.NewTextHandler(os.Stdout, nil)),
	)
	slog.SetDefault(lgger)

	serv := api.New(lm, cm)
	log.Fatal(serv.Run())
}

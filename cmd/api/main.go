// @title URL-shortener API
// @description API for url-shortener app
// @BasePath /
// @schemes http
package main

import (
	"log"
	"log/slog"
	"os"

	_ "github.com/limbo/url_shortener/docs"
	"github.com/limbo/url_shortener/internal/api"
	"github.com/limbo/url_shortener/internal/logger"
	"github.com/limbo/url_shortener/internal/settings"
	"github.com/limbo/url_shortener/internal/stats"
	cache "github.com/limbo/url_shortener/internal/url_cache_manager"
	urlmanager "github.com/limbo/url_shortener/internal/url_manager"
)

func main() {
	cfg := settings.GetConfig()

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

	sm := stats.New(stats.ClicksDBCfg{
		Address:  cfg.GetString("ch_address"),
		Database: cfg.GetString("ch_db"),
		Username: cfg.GetString("ch_user"),
		Password: cfg.GetString("ch_pass"),
	})

	lgger := slog.New(
		logger.NewContextHandler(slog.NewTextHandler(os.Stdout, nil)),
	)
	slog.SetDefault(lgger)

	serv := api.New(api.ServerCfg{
		Lm: lm,
		Cm: cm,
		Sm: sm,
	})
	log.Fatal(serv.Run())
}

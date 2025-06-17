package main

import (
	"log"

	"github.com/limbo/url_shortener/internal/api"
	"github.com/limbo/url_shortener/internal/settings"
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

	serv := api.New(lm)
	log.Fatal(serv.Run())
}

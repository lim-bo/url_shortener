package api

import (
	"log/slog"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/limbo/url_shortener/internal/settings"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) shorten(w http.ResponseWriter, r *http.Request) {
	var originalLink string
	request := make(map[string]interface{}, 0)
	err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.Error("error unmarshalling request body", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalLink, ok := request["link"].(string)
	if !ok {
		slog.Error("shorten request with invalid link", slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortLink, err := s.lm.SaveURL(originalLink)
	if err != nil {
		slog.Error("error while saving link", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]string{"link": settings.GetConfig().GetString("api_address") + "/" + shortLink})
	if err != nil {
		slog.Error("error while marshalling results", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("successfully provided short link", slog.String("endpoint", "/shorten"))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) clickStats(w http.ResponseWriter, r *http.Request) {

}

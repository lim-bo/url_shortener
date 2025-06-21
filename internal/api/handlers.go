package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/limbo/url_shortener/internal/settings"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		ctx := context.WithValue(r.Context(), "requestID", requestID.String())
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID.String())
		next.ServeHTTP(w, r)
	})
}

type ShortenRequest struct {
	Link string `json:"link"`
}

type ShortenResponse struct {
	Link string `json:"link"`
}

// @Router /shorten [post]
// @Summary Recieves link to be shorted and provides short one
// @Description Recieves request with json data {"link": ...}, if there such link was already saved
// @Description provides its short version, otherwise generates new one and saves data to db.
// @Param link body ShortenRequest true "Link to be shorted" example({"link": "google.com"})
// @Success 200 {object} ShortenResponse "Link saved and short version returned" example({"link": "host/12345678"})
// @Failure 400 "Invalid request body"
// @Failure 500 "Error while fetching databases"
// @Accept json
// @Produce json
func (s *Server) shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var originalLink string
	var request ShortenRequest
	err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.ErrorContext(ctx, "error unmarshalling request body", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalLink = request.Link
	shortCode, err := s.links.SaveURL(originalLink)
	if err != nil {
		slog.ErrorContext(ctx, "error while saving link", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func() {
		err := s.cache.CacheLink(shortCode, originalLink)
		if err != nil {
			slog.ErrorContext(ctx, "error caching link", slog.String("error", err.Error()))
		} else {
			slog.InfoContext(ctx, "link cached")
		}
	}()
	err = sonic.ConfigFastest.NewEncoder(w).Encode(ShortenResponse{Link: settings.GetConfig().GetString("api_address") + "/" + shortCode})
	if err != nil {
		slog.ErrorContext(ctx, "error while marshalling results", slog.String("error", err.Error()), slog.String("endpoint", "/shorten"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.InfoContext(ctx, "successfully provided short link", slog.String("endpoint", "/shorten"))
}

// @Router /{short_code} [get]
// @Summary Redirects to original link with provided short code
// @Description Takes shortcode in path value, searchs original url
// @Description linked with it and then redirects with 308 code.
// @Param short_code path string true "short code of original link" example("12345678")
// @Success 308 "Redirected"
// @Failure 404 "Invalid code or internal error"
func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.PathValue("short_code")
	if code == "" {
		slog.ErrorContext(ctx, "redirect request with invalid short code")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	originalLink, err := s.cache.GetLink(code)
	if err != nil && err != ErrNoKey {
		slog.ErrorContext(ctx, "error getting link cache", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == ErrNoKey {
		originalLink, err = s.links.GetLinkByCode(code)
		if err != nil {
			slog.ErrorContext(ctx, "getting original link error", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	http.Redirect(w, r, originalLink, http.StatusPermanentRedirect)
	slog.InfoContext(ctx, "successfull redirect")
}

func (s *Server) clickStats(w http.ResponseWriter, r *http.Request) {

}

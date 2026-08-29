package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/wxvn/grpc-mesh/internal/url/transport"
)

type Server struct {
	service    transport.URLService
	middleware *Middleware
}

func NewServer(service transport.URLService, middleware *Middleware) *Server {
	return &Server{service: service, middleware: middleware}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{shortCode}", s.redirect)
	mux.Handle("POST /urls", s.middleware.Auth(http.HandlerFunc(s.createShortURL)))
	mux.Handle("GET /urls", s.middleware.Auth(http.HandlerFunc(s.getMyURLs)))
	mux.Handle("GET /urls/{id}", s.middleware.Auth(http.HandlerFunc(s.getURLStats)))
	mux.Handle("DELETE /urls/{id}", s.middleware.Auth(http.HandlerFunc(s.deleteURL)))
	return mux
}

func (s *Server) createShortURL(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	userID, ok := getUserID(r.Context())
	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	result, err := s.service.CreateShortURL(r.Context(), userID, request.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) getMyURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r.Context())
	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	result, err := s.service.GetMyURLs(
		r.Context(),
		userID,
		limit,
		offset,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getURLStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r.Context())
	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid url id", http.StatusBadRequest)
		return
	}

	result, err := s.service.GetURLStats(r.Context(), userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) deleteURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r.Context())
	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid url id", http.StatusBadRequest)
		return
	}

	if err := s.service.DeleteURL(r.Context(), userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	originalURL, err := s.service.Redirect(r.Context(), shortCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

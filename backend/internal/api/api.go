package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jlcoulter/pageturner/internal/db/generated"
	"github.com/jlcoulter/pageturner/internal/repository"
	"github.com/jlcoulter/pageturner/internal/types"
)

type Server struct {
	BookRepo        *repository.BookRepo
	OpenLibraryRepo *repository.OpenLibraryRepo
}

func NewServer(bookRepo *repository.BookRepo, openLibraryRepo *repository.OpenLibraryRepo) *Server {
	return &Server{
		BookRepo:        bookRepo,
		OpenLibraryRepo: openLibraryRepo,
	}
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Routes().ServeHTTP(w, r)
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	// Structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r.Use(middleware.RequestID)
	r.Use(requestLogger(logger))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)

	// CORS middleware
	r.Use(s.corsMiddleware)

	// Health check
	r.Get("/health", s.HandleHealth)

	// Book endpoints
	r.Get("/books", s.HandleGetBooks)
	r.Post("/book", s.HandleBook)
	r.Post("/library/search", s.HandleSearchLibrary)

	// OpenLibrary search endpoints
	r.Post("/openlibrary/search", s.HandleOpenLibrarySearch)

	return r
}

// requestLogger creates a logger middleware using slog
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := middleware.GetReqID(r.Context())

			wr := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wr, r)

			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wr.Status(),
				"duration", time.Since(start).String(),
				"request_id", requestID,
			)
		})
	}
}

// corsMiddleware returns a configurable CORS handler
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	//allowedOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")
	allowedMethods := getEnv("CORS_METHODS", "GET, POST, OPTIONS")
	//allowedHeaders := getEnv("CORS_HEADERS", "Content-Type")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) HandleGetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := s.BookRepo.GetAllBooks(r.Context())
	if err != nil {
		slog.Error("failed to get books", "error", err)
		writeJsonError(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// OpenLibrarySearchRequest represents an OpenLibrary search request
type OpenLibrarySearchRequest struct {
	Term     string `json:"term"`
	SearchBy string `json:"searchBy"` // "author", "title", or "both" (default)
}

func (s *Server) HandleOpenLibrarySearch(w http.ResponseWriter, r *http.Request) {
	var req OpenLibrarySearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Term) == "" {
		writeJsonError(w, http.StatusBadRequest, "search term is required")
		return
	}

	searchBy := strings.ToLower(req.SearchBy)
	if searchBy == "" {
		searchBy = "both"
	}

	var result any

	switch searchBy {

	default: // "both"
		bothResults, err := s.OpenLibraryRepo.Search(r.Context(), req.Term)
		if err != nil {
			slog.Error("failed to search openlibrary", "error", err)
			writeJsonError(w, http.StatusInternalServerError, "failed to search")
			return
		}
		if bothResults == nil {
			bothResults = []generated.SearchOpenLibraryRow{}
		}
		result = bothResults
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (s *Server) HandleSearchLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	books, err := s.BookRepo.SearchBooks(r.Context(), req.SearchTerm)
	if err != nil {
		slog.Error("failed to fetch books", "error", err)
		writeJsonError(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// HandleBook handles POST /api/book
func (s *Server) HandleBook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry types.BookEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		writeJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if err := validateBookEntry(entry); err != nil {
		writeJsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := s.BookRepo.SaveBook(r.Context(), entry)
	if err != nil {
		slog.Error("failed to save book", "error", err)
		writeJsonError(w, http.StatusInternalServerError, "failed to save book")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func validateBookEntry(entry types.BookEntry) error {
	if strings.TrimSpace(entry.Book) == "" {
		return errors.New("book title is required")
	}
	if entry.Rating < 0 || entry.Rating > 10 {
		return errors.New("rating must be between 0 and 10")
	}
	return nil
}

func writeJsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

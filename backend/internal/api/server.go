package api

import (
	"database/sql"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"recipick/backend/internal/auth"
	"recipick/backend/internal/handler"
	"recipick/backend/internal/repository"
	"recipick/backend/internal/service"
)

func allowDevOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" {
		return false
	}
	if port := parsed.Port(); port != "" && port != "5173" {
		return false
	}

	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback()
}

func originAllowSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		normalized := strings.TrimSpace(origin)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

func NewServer(db *sql.DB, authMiddleware *auth.Middleware, allowedOrigins []string) http.Handler {
	repo := repository.NewRecipeRepository(db)
	importer := service.NewJSONLDImporter()
	recipeHandler := handler.NewRecipeHandler(repo, importer)
	allowedOriginSet := originAllowSet(allowedOrigins)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			if _, ok := allowedOriginSet[origin]; ok {
				return true
			}
			return allowDevOrigin(origin)
		},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Mount("/v1", recipeHandler.Router(authMiddleware.Handler))
	return r
}

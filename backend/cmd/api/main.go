package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"recipick/backend/internal/api"
	"recipick/backend/internal/auth"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	jwtIssuer := os.Getenv("AUTH_JWT_ISSUER")
	if jwtIssuer == "" {
		log.Fatal("AUTH_JWT_ISSUER is required")
	}
	jwksURL := os.Getenv("AUTH_JWT_JWKS_URL")
	if jwksURL == "" {
		log.Fatal("AUTH_JWT_JWKS_URL is required")
	}
	jwtAudience := os.Getenv("AUTH_JWT_AUDIENCE")
	corsAllowedOrigins := splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))

	authMiddleware, err := auth.NewMiddleware(auth.Config{
		Issuer:        jwtIssuer,
		Audience:      jwtAudience,
		JWKSURL:       jwksURL,
		HTTPTimeout:   5 * time.Second,
		RefreshWindow: 5 * time.Minute,
		Leeway:        60 * time.Second,
	})
	if err != nil {
		log.Fatalf("init auth middleware: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, api.NewServer(db, authMiddleware, corsAllowedOrigins)); err != nil {
		log.Fatal(err)
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

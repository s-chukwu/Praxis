package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"github.com/s-chukwu/Praxis/backend/internal/config"
	"github.com/s-chukwu/Praxis/backend/internal/httpapi"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

func main() {
	// Sentry only starts when SENTRY_DSN is set. SendDefaultPII stays off, so
	// request headers such as Authorization are not sent with error reports.
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      os.Getenv("PRAXIS_ENV"),
			Release:          os.Getenv("PRAXIS_RELEASE"),
			TracesSampleRate: 0.1,
		}); err != nil {
			log.Printf("sentry init: %v", err)
		} else {
			defer sentry.Flush(2 * time.Second)
		}
	}

	cfg := config.Load()

	// Say plainly which settings are missing, so a 503 in production can be traced
	// from the startup log. Only the names are logged, never the values.
	if cfg.SupabaseURL == "" || cfg.SupabaseServiceRoleKey == "" {
		log.Println("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY is not set: login routes will answer 503 auth_not_configured")
	}
	if len(cfg.AIServiceSharedSecret) < 32 {
		log.Println("AI_SERVICE_SHARED_SECRET is missing or shorter than 32 bytes: AI routes will answer 503 ai_not_configured")
	}

	var db *store.Store
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		opened, err := store.New(ctx, cfg.DatabaseURL)
		cancel()
		if err != nil {
			log.Fatalf("connect to database: %v", err)
		}
		defer opened.Close()
		db = opened
	} else {
		log.Println("DATABASE_URL is not set: attempt, review and plan routes will answer 503")
	}

	handler := httpapi.NewRouterWithStore(cfg, db)
	handler = sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}).Handle(handler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("praxis-backend listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
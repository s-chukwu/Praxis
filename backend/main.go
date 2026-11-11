package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/s-chukwu/Praxis/backend/internal/config"
	"github.com/s-chukwu/Praxis/backend/internal/httpapi"
	"github.com/s-chukwu/Praxis/backend/internal/store"
)

func main() {
	cfg := config.Load()

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
		log.Println("DATABASE_URL is not set: attempt routes will answer 503")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouterWithStore(cfg, db),
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

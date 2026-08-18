package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aquafarm/internal/handler"
	"aquafarm/internal/monitor"
	"aquafarm/internal/service"
	"aquafarm/internal/store"
)

func main() {
	// Use file-based SQLite, NOT :memory:
	dbPath := getEnvDefault("AQUAFARM_DB", "./aquafarm.db")
	listenAddr := getEnvDefault("AQUAFARM_ADDR", ":8585")

	repo, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	svc := service.New(repo)
	mon := monitor.New(svc)
	h := handler.New(svc, mon)

	// Start background monitoring loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mon.Run(ctx)

	router := h.Router()

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")

		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		srv.Shutdown(shutCtx)
		cancel()
	}()

	fmt.Printf("aquafarm listening on %s\n", listenAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KrishnaGrg1/hookfire/internal/api"
	"github.com/KrishnaGrg1/hookfire/internal/config"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/KrishnaGrg1/hookfire/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DB_URL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Cannot ping database:", err)
	}
	log.Println("Connected to database")

	q, err := queue.New(cfg.REDIS_URL)
	if err != nil {
		log.Fatal("Cannot connect to Redis:", err)
	}
	log.Println("Connected to Redis")

	s := store.New(pool)

	// Start workers in background
	ctx, cancel := context.WithCancel(context.Background())
	dispatcher := worker.NewDispatcher(q, s, cfg.WORKER_COUNT)
	go dispatcher.Start(ctx)

	router := api.NewRouter(s, q)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Server running on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("Server stopped")
}

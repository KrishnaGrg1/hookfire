package main

import (
	"context"
	"log"
	"net/http"

	"github.com/KrishnaGrg1/hookfire/internal/api"
	"github.com/KrishnaGrg1/hookfire/internal/config"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DB_URL)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer pool.Close()

	if err := pool.Ping((context.Background())); err != nil {
		log.Fatal("Cannot ping database", err)
	}
	log.Println("Connected to database")

	s := store.New(pool)
	router := api.NewRouter(s)

	log.Printf("Server running on port %s", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, router)
}

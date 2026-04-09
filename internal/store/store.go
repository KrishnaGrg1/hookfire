package store

import (
	"context"

	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: db.New(pool),
	}
}

func Connect(dbURL string) (*Store, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return New(pool), nil
}

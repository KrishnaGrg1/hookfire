package store

import (
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

package repository

import "github.com/jackc/pgx/v4/pgxpool"

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool}
}

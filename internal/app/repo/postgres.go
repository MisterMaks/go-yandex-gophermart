package repo

import "database/sql"

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) (*AppRepo, error) {
	return &AppRepo{db: db}, nil
}

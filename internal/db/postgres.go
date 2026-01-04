package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/leora/leora-server/internal/config"
	_ "github.com/lib/pq"
)

// NewPostgres creates a new sqlx DB pool.
func NewPostgres(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

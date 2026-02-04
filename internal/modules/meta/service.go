package meta

import (
	"context"

	"github.com/jmoiron/sqlx"
	appErrors "github.com/leora/leora-server/internal/errors"
)

type Service struct {
	db *sqlx.DB
}

type Language struct {
	Code      string `json:"code" db:"code"`
	Name      string `json:"name" db:"name"`
	IsActive  bool   `json:"isActive" db:"is_active"`
	IsDefault bool   `json:"isDefault" db:"is_default"`
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListLanguages(ctx context.Context) ([]Language, error) {
	var languages []Language
	if err := s.db.SelectContext(ctx, &languages, `
		SELECT code, name, is_active, is_default
		FROM meta_languages
		WHERE is_active = true
		ORDER BY is_default DESC, name ASC
	`); err != nil {
		return nil, appErrors.DatabaseError
	}
	return languages, nil
}

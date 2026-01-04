package search

import "context"

// Service executes search flows.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Query(ctx context.Context, term string) ([]*Result, error) {
	return s.repo.Query(ctx, term)
}

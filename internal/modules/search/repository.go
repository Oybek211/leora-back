package search

import (
	"context"
	"strings"
	"sync"
)

// Repository defines search behavior.
type Repository interface {
	Query(ctx context.Context, term string) ([]*Result, error)
}

// InMemoryRepository stores search results in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items []*Result
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: []*Result{}}
}

func (s *InMemoryRepository) Query(ctx context.Context, term string) ([]*Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if term == "" {
		return append([]*Result{}, s.items...), nil
	}
	needle := strings.ToLower(strings.TrimSpace(term))
	results := make([]*Result, 0, len(s.items))
	for _, item := range s.items {
		if item == nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Title), needle) {
			results = append(results, item)
		}
	}
	return results, nil
}

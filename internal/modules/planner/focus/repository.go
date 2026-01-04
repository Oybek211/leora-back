package focus

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository for focus sessions.
type Repository interface {
	List(ctx context.Context) ([]*Session, error)
	GetByID(ctx context.Context, id string) (*Session, error)
	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*SessionStats, error)
}

// InMemoryRepository stores focus sessions in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*Session
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[string]*Session)}
}

func (r *InMemoryRepository) List(ctx context.Context) ([]*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Session, 0, len(r.items))
	for _, session := range r.items {
		if session == nil || session.DeletedAt != nil {
			continue
		}
		results = append(results, cloneSession(session))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.items[id]
	if !ok || session == nil || session.DeletedAt != nil {
		return nil, appErrors.FocusSessionNotFound
	}
	return cloneSession(session), nil
}

func (r *InMemoryRepository) Create(ctx context.Context, session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	session.CreatedAt = now
	session.UpdatedAt = now
	r.items[session.ID] = cloneSession(session)
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.items[session.ID]
	if !ok || current == nil || current.DeletedAt != nil {
		return appErrors.FocusSessionNotFound
	}
	session.CreatedAt = current.CreatedAt
	session.UpdatedAt = utils.NowUTC()
	r.items[session.ID] = cloneSession(session)
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.items[id]
	if !ok || session == nil || session.DeletedAt != nil {
		return appErrors.FocusSessionNotFound
	}
	now := utils.NowUTC()
	session.DeletedAt = &now
	session.UpdatedAt = now
	r.items[id] = session
	return nil
}

func (r *InMemoryRepository) GetStats(ctx context.Context) (*SessionStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stats := &SessionStats{}
	for _, session := range r.items {
		if session == nil || session.DeletedAt != nil {
			continue
		}
		stats.TotalSessions++
		stats.TotalMinutes += session.ActualMinutes
		stats.TotalInterruptions += session.InterruptionsCount
		if session.Status == "completed" {
			stats.CompletedSessions++
		}
	}
	if stats.CompletedSessions > 0 {
		stats.AverageMinutes = float64(stats.TotalMinutes) / float64(stats.CompletedSessions)
	}
	return stats, nil
}

func cloneSession(session *Session) *Session {
	if session == nil {
		return nil
	}
	copy := *session
	return &copy
}

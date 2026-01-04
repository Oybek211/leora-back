package goals

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository persists goals.
type Repository interface {
	List(ctx context.Context) ([]*Goal, error)
	GetByID(ctx context.Context, id string) (*Goal, error)
	Create(ctx context.Context, goal *Goal) error
	Update(ctx context.Context, goal *Goal) error
	Delete(ctx context.Context, id string) error
	CreateCheckIn(ctx context.Context, checkIn *CheckIn) error
	GetCheckIns(ctx context.Context, goalID string) ([]*CheckIn, error)
}

// InMemoryRepository stores goals in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*Goal
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[string]*Goal)}
}

func (r *InMemoryRepository) List(ctx context.Context) ([]*Goal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Goal, 0, len(r.items))
	for _, goal := range r.items {
		if goal == nil || goal.DeletedAt != nil {
			continue
		}
		results = append(results, cloneGoal(goal))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Goal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	goal, ok := r.items[id]
	if !ok || goal == nil || goal.DeletedAt != nil {
		return nil, appErrors.GoalNotFound
	}
	return cloneGoal(goal), nil
}

func (r *InMemoryRepository) Create(ctx context.Context, goal *Goal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if goal.ID == "" {
		goal.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	goal.CreatedAt = now
	goal.UpdatedAt = now
	r.items[goal.ID] = cloneGoal(goal)
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, goal *Goal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.items[goal.ID]
	if !ok || current == nil || current.DeletedAt != nil {
		return appErrors.GoalNotFound
	}
	goal.CreatedAt = current.CreatedAt
	goal.UpdatedAt = utils.NowUTC()
	r.items[goal.ID] = cloneGoal(goal)
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	goal, ok := r.items[id]
	if !ok || goal == nil || goal.DeletedAt != nil {
		return appErrors.GoalNotFound
	}
	now := utils.NowUTC()
	goal.DeletedAt = &now
	goal.UpdatedAt = now
	r.items[id] = goal
	return nil
}

func (r *InMemoryRepository) CreateCheckIn(ctx context.Context, checkIn *CheckIn) error {
	// InMemory implementation - not used in production
	return nil
}

func (r *InMemoryRepository) GetCheckIns(ctx context.Context, goalID string) ([]*CheckIn, error) {
	// InMemory implementation - not used in production
	return []*CheckIn{}, nil
}

func cloneGoal(goal *Goal) *Goal {
	if goal == nil {
		return nil
	}
	copy := *goal
	return &copy
}

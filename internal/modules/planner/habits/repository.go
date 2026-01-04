package habits

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository for habit persistence.
type Repository interface {
	List(ctx context.Context) ([]*Habit, error)
	GetByID(ctx context.Context, id string) (*Habit, error)
	Create(ctx context.Context, habit *Habit) error
	Update(ctx context.Context, habit *Habit) error
	Delete(ctx context.Context, id string) error
	CreateCompletion(ctx context.Context, completion *HabitCompletion) error
	GetCompletions(ctx context.Context, habitID string) ([]*HabitCompletion, error)
	GetStats(ctx context.Context, habitID string) (*HabitStats, error)
}

// InMemoryRepository stores habits in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*Habit
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[string]*Habit)}
}

func (r *InMemoryRepository) List(ctx context.Context) ([]*Habit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Habit, 0, len(r.items))
	for _, habit := range r.items {
		if habit == nil || habit.DeletedAt != nil {
			continue
		}
		results = append(results, cloneHabit(habit))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Habit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	habit, ok := r.items[id]
	if !ok || habit == nil || habit.DeletedAt != nil {
		return nil, appErrors.HabitNotFound
	}
	return cloneHabit(habit), nil
}

func (r *InMemoryRepository) Create(ctx context.Context, habit *Habit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if habit.ID == "" {
		habit.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	habit.CreatedAt = now
	habit.UpdatedAt = now
	r.items[habit.ID] = cloneHabit(habit)
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, habit *Habit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.items[habit.ID]
	if !ok || current == nil || current.DeletedAt != nil {
		return appErrors.HabitNotFound
	}
	habit.CreatedAt = current.CreatedAt
	habit.UpdatedAt = utils.NowUTC()
	r.items[habit.ID] = cloneHabit(habit)
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	habit, ok := r.items[id]
	if !ok || habit == nil || habit.DeletedAt != nil {
		return appErrors.HabitNotFound
	}
	now := utils.NowUTC()
	habit.DeletedAt = &now
	habit.UpdatedAt = now
	r.items[id] = habit
	return nil
}

func (r *InMemoryRepository) CreateCompletion(ctx context.Context, completion *HabitCompletion) error {
	// InMemory implementation - not used in production
	return nil
}

func (r *InMemoryRepository) GetCompletions(ctx context.Context, habitID string) ([]*HabitCompletion, error) {
	// InMemory implementation - not used in production
	return []*HabitCompletion{}, nil
}

func (r *InMemoryRepository) GetStats(ctx context.Context, habitID string) (*HabitStats, error) {
	// InMemory implementation - not used in production
	return &HabitStats{}, nil
}

func cloneHabit(habit *Habit) *Habit {
	if habit == nil {
		return nil
	}
	copy := *habit
	return &copy
}

package widgets

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository persists widget definitions.
type Repository interface {
	List(ctx context.Context) ([]*Widget, error)
	GetByID(ctx context.Context, id string) (*Widget, error)
	Create(ctx context.Context, widget *Widget) error
	Update(ctx context.Context, widget *Widget) error
	Delete(ctx context.Context, id string) error
}

// InMemoryRepository stores widgets in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*Widget
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[string]*Widget)}
}

func (r *InMemoryRepository) List(ctx context.Context) ([]*Widget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Widget, 0, len(r.items))
	for _, widget := range r.items {
		if widget == nil || widget.DeletedAt != "" {
			continue
		}
		results = append(results, cloneWidget(widget))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Widget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	widget, ok := r.items[id]
	if !ok || widget == nil || widget.DeletedAt != "" {
		return nil, appErrors.WidgetNotFound
	}
	return cloneWidget(widget), nil
}

func (r *InMemoryRepository) Create(ctx context.Context, widget *Widget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if widget.ID == "" {
		widget.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	widget.CreatedAt = now
	widget.UpdatedAt = now
	r.items[widget.ID] = cloneWidget(widget)
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, widget *Widget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.items[widget.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.WidgetNotFound
	}
	widget.CreatedAt = current.CreatedAt
	widget.UpdatedAt = utils.NowUTC()
	r.items[widget.ID] = cloneWidget(widget)
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	widget, ok := r.items[id]
	if !ok || widget == nil || widget.DeletedAt != "" {
		return appErrors.WidgetNotFound
	}
	widget.DeletedAt = utils.NowUTC()
	widget.UpdatedAt = widget.DeletedAt
	r.items[id] = widget
	return nil
}

func cloneWidget(widget *Widget) *Widget {
	if widget == nil {
		return nil
	}
	copy := *widget
	return &copy
}

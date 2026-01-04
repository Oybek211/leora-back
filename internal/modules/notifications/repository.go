package notifications

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository persists notifications.
type Repository interface {
	List(ctx context.Context) ([]*Notification, error)
	GetByID(ctx context.Context, id string) (*Notification, error)
	Create(ctx context.Context, notification *Notification) error
	Update(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, id string) error
}

// InMemoryRepository stores notifications in memory.
type InMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*Notification
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{items: make(map[string]*Notification)}
}

func (r *InMemoryRepository) List(ctx context.Context) ([]*Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Notification, 0, len(r.items))
	for _, notification := range r.items {
		if notification == nil || notification.DeletedAt != "" {
			continue
		}
		results = append(results, cloneNotification(notification))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, id string) (*Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	notification, ok := r.items[id]
	if !ok || notification == nil || notification.DeletedAt != "" {
		return nil, appErrors.NotificationNotFound
	}
	return cloneNotification(notification), nil
}

func (r *InMemoryRepository) Create(ctx context.Context, notification *Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if notification.ID == "" {
		notification.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	notification.CreatedAt = now
	notification.UpdatedAt = now
	r.items[notification.ID] = cloneNotification(notification)
	return nil
}

func (r *InMemoryRepository) Update(ctx context.Context, notification *Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.items[notification.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.NotificationNotFound
	}
	notification.CreatedAt = current.CreatedAt
	notification.UpdatedAt = utils.NowUTC()
	r.items[notification.ID] = cloneNotification(notification)
	return nil
}

func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	notification, ok := r.items[id]
	if !ok || notification == nil || notification.DeletedAt != "" {
		return appErrors.NotificationNotFound
	}
	notification.DeletedAt = utils.NowUTC()
	notification.UpdatedAt = notification.DeletedAt
	r.items[id] = notification
	return nil
}

func cloneNotification(notification *Notification) *Notification {
	if notification == nil {
		return nil
	}
	copy := *notification
	return &copy
}

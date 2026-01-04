package premium

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/leora/leora-server/internal/common/utils"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// Repository for subscription state.
type Repository interface {
	ListSubscriptions(ctx context.Context) ([]*Subscription, error)
	GetSubscriptionByID(ctx context.Context, id string) (*Subscription, error)
	CreateSubscription(ctx context.Context, subscription *Subscription) error
	UpdateSubscription(ctx context.Context, subscription *Subscription) error
	DeleteSubscription(ctx context.Context, id string) error

	ListPlans(ctx context.Context) ([]*Plan, error)
	GetPlanByID(ctx context.Context, id string) (*Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) error
	UpdatePlan(ctx context.Context, plan *Plan) error
	DeletePlan(ctx context.Context, id string) error
}

// InMemoryRepository stores subscriptions and plans in memory.
type InMemoryRepository struct {
	mu            sync.RWMutex
	subscriptions map[string]*Subscription
	plans         map[string]*Plan
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		subscriptions: make(map[string]*Subscription),
		plans:         make(map[string]*Plan),
	}
}

func (r *InMemoryRepository) ListSubscriptions(ctx context.Context) ([]*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Subscription, 0, len(r.subscriptions))
	for _, sub := range r.subscriptions {
		if sub == nil || sub.DeletedAt != "" {
			continue
		}
		results = append(results, cloneSubscription(sub))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetSubscriptionByID(ctx context.Context, id string) (*Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.subscriptions[id]
	if !ok || sub == nil || sub.DeletedAt != "" {
		return nil, appErrors.SubscriptionNotFound
	}
	return cloneSubscription(sub), nil
}

func (r *InMemoryRepository) CreateSubscription(ctx context.Context, subscription *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if subscription.ID == "" {
		subscription.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now
	r.subscriptions[subscription.ID] = cloneSubscription(subscription)
	return nil
}

func (r *InMemoryRepository) UpdateSubscription(ctx context.Context, subscription *Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.subscriptions[subscription.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.SubscriptionNotFound
	}
	subscription.CreatedAt = current.CreatedAt
	subscription.UpdatedAt = utils.NowUTC()
	r.subscriptions[subscription.ID] = cloneSubscription(subscription)
	return nil
}

func (r *InMemoryRepository) DeleteSubscription(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub, ok := r.subscriptions[id]
	if !ok || sub == nil || sub.DeletedAt != "" {
		return appErrors.SubscriptionNotFound
	}
	sub.DeletedAt = utils.NowUTC()
	sub.UpdatedAt = sub.DeletedAt
	r.subscriptions[id] = sub
	return nil
}

func (r *InMemoryRepository) ListPlans(ctx context.Context) ([]*Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]*Plan, 0, len(r.plans))
	for _, plan := range r.plans {
		if plan == nil || plan.DeletedAt != "" {
			continue
		}
		results = append(results, clonePlan(plan))
	}
	sort.Slice(results, func(i, j int) bool {
		return utils.ParseRFC3339(results[i].CreatedAt).After(utils.ParseRFC3339(results[j].CreatedAt))
	})
	return results, nil
}

func (r *InMemoryRepository) GetPlanByID(ctx context.Context, id string) (*Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plan, ok := r.plans[id]
	if !ok || plan == nil || plan.DeletedAt != "" {
		return nil, appErrors.PlanNotFound
	}
	return clonePlan(plan), nil
}

func (r *InMemoryRepository) CreatePlan(ctx context.Context, plan *Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if plan.ID == "" {
		plan.ID = uuid.NewString()
	}
	now := utils.NowUTC()
	plan.CreatedAt = now
	plan.UpdatedAt = now
	r.plans[plan.ID] = clonePlan(plan)
	return nil
}

func (r *InMemoryRepository) UpdatePlan(ctx context.Context, plan *Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.plans[plan.ID]
	if !ok || current == nil || current.DeletedAt != "" {
		return appErrors.PlanNotFound
	}
	plan.CreatedAt = current.CreatedAt
	plan.UpdatedAt = utils.NowUTC()
	r.plans[plan.ID] = clonePlan(plan)
	return nil
}

func (r *InMemoryRepository) DeletePlan(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	plan, ok := r.plans[id]
	if !ok || plan == nil || plan.DeletedAt != "" {
		return appErrors.PlanNotFound
	}
	plan.DeletedAt = utils.NowUTC()
	plan.UpdatedAt = plan.DeletedAt
	r.plans[id] = plan
	return nil
}

func cloneSubscription(subscription *Subscription) *Subscription {
	if subscription == nil {
		return nil
	}
	copy := *subscription
	return &copy
}

func clonePlan(plan *Plan) *Plan {
	if plan == nil {
		return nil
	}
	copy := *plan
	return &copy
}

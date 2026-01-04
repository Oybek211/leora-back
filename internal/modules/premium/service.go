package premium

import "context"

// Service encapsulates premium flows.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListSubscriptions(ctx context.Context) ([]*Subscription, error) {
	return s.repo.ListSubscriptions(ctx)
}

func (s *Service) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	return s.repo.GetSubscriptionByID(ctx, id)
}

func (s *Service) CreateSubscription(ctx context.Context, subscription *Subscription) (*Subscription, error) {
	if err := s.repo.CreateSubscription(ctx, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (s *Service) UpdateSubscription(ctx context.Context, id string, subscription *Subscription) (*Subscription, error) {
	subscription.ID = id
	if err := s.repo.UpdateSubscription(ctx, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (s *Service) PatchSubscription(ctx context.Context, id string, fields map[string]interface{}) (*Subscription, error) {
	current, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["tier"].(string); ok {
		current.Tier = v
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if v, ok := fields["cancelAtPeriodEnd"].(bool); ok {
		current.CancelAtPeriodEnd = v
	}
	if err := s.repo.UpdateSubscription(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteSubscription(ctx context.Context, id string) error {
	return s.repo.DeleteSubscription(ctx, id)
}

func (s *Service) ListPlans(ctx context.Context) ([]*Plan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *Service) GetPlan(ctx context.Context, id string) (*Plan, error) {
	return s.repo.GetPlanByID(ctx, id)
}

func (s *Service) CreatePlan(ctx context.Context, plan *Plan) (*Plan, error) {
	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) UpdatePlan(ctx context.Context, id string, plan *Plan) (*Plan, error) {
	plan.ID = id
	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) PatchPlan(ctx context.Context, id string, fields map[string]interface{}) (*Plan, error) {
	current, err := s.repo.GetPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["name"].(string); ok {
		current.Name = v
	}
	if v, ok := fields["interval"].(string); ok {
		current.Interval = v
	}
	if v, ok := fields["prices"].(map[string]interface{}); ok {
		converted := make(map[string]float64)
		for key, raw := range v {
			if val, ok := raw.(float64); ok {
				converted[key] = val
			}
		}
		if len(converted) > 0 {
			current.Prices = converted
		}
	}
	if err := s.repo.UpdatePlan(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeletePlan(ctx context.Context, id string) error {
	return s.repo.DeletePlan(ctx, id)
}

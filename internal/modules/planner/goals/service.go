package goals

import "context"

// Service handles goal logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Goal, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Goal, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, goal *Goal) (*Goal, error) {
	if err := s.repo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *Service) Update(ctx context.Context, id string, goal *Goal) (*Goal, error) {
	goal.ID = id
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Goal, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["goalType"].(string); ok {
		current.GoalType = v
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if v, ok := fields["progressPercent"].(float64); ok {
		current.ProgressPercent = v
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) CreateCheckIn(ctx context.Context, checkIn *CheckIn) (*CheckIn, error) {
	if err := s.repo.CreateCheckIn(ctx, checkIn); err != nil {
		return nil, err
	}
	return checkIn, nil
}

func (s *Service) GetCheckIns(ctx context.Context, goalID string) ([]*CheckIn, error) {
	return s.repo.GetCheckIns(ctx, goalID)
}

func (s *Service) Complete(ctx context.Context, id string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	goal.Status = "completed"
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *Service) Reactivate(ctx context.Context, id string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	goal.Status = "active"
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

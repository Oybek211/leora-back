package habits

import "context"

// Service orchestrates habit operations.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Habit, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Habit, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, habit *Habit) (*Habit, error) {
	if err := s.repo.Create(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *Service) Update(ctx context.Context, id string, habit *Habit) (*Habit, error) {
	habit.ID = id
	if err := s.repo.Update(ctx, habit); err != nil {
		return nil, err
	}
	return habit, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Habit, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["habitType"].(string); ok {
		current.HabitType = v
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) CreateCompletion(ctx context.Context, completion *HabitCompletion) (*HabitCompletion, error) {
	if err := s.repo.CreateCompletion(ctx, completion); err != nil {
		return nil, err
	}
	return completion, nil
}

func (s *Service) GetCompletions(ctx context.Context, habitID string) ([]*HabitCompletion, error) {
	return s.repo.GetCompletions(ctx, habitID)
}

func (s *Service) GetStats(ctx context.Context, habitID string) (*HabitStats, error) {
	return s.repo.GetStats(ctx, habitID)
}

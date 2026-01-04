package tasks

import "context"

// Service orchestrates planner tasks.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Task, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, payload *Task) (*Task, error) {
	if err := s.repo.Create(ctx, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *Service) Update(ctx context.Context, id string, payload *Task) (*Task, error) {
	payload.ID = id
	if err := s.repo.Update(ctx, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Task, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if v, ok := fields["priority"].(string); ok {
		current.Priority = v
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Complete(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Status = "completed"
	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Reopen(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Status = "in_progress"
	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) UpdateChecklistItem(ctx context.Context, taskID, itemID string, completed bool) error {
	return s.repo.UpdateChecklistItem(ctx, taskID, itemID, completed)
}

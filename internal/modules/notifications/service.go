package notifications

import "context"

// Service handles notification flows.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Notification, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, payload *Notification) (*Notification, error) {
	if err := s.repo.Create(ctx, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *Service) Update(ctx context.Context, id string, payload *Notification) (*Notification, error) {
	payload.ID = id
	if err := s.repo.Update(ctx, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Notification, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["message"].(string); ok {
		current.Message = v
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

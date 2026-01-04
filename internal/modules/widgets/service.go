package widgets

import "context"

// Service handles widget concerns.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Widget, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Widget, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, widget *Widget) (*Widget, error) {
	if err := s.repo.Create(ctx, widget); err != nil {
		return nil, err
	}
	return widget, nil
}

func (s *Service) Update(ctx context.Context, id string, widget *Widget) (*Widget, error) {
	widget.ID = id
	if err := s.repo.Update(ctx, widget); err != nil {
		return nil, err
	}
	return widget, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Widget, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["config"].(map[string]interface{}); ok {
		current.Config = v
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

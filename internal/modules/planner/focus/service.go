package focus

import "context"

// Service for focus session orchestration.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Session, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Session, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, session *Session) (*Session, error) {
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) Update(ctx context.Context, id string, session *Session) (*Session, error) {
	session.ID = id
	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) Patch(ctx context.Context, id string, fields map[string]interface{}) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := fields["taskId"].(string); ok {
		current.TaskID = &v
	}
	if v, ok := fields["goalId"].(string); ok {
		current.GoalID = &v
	}
	if v, ok := fields["plannedMinutes"].(float64); ok {
		current.PlannedMinutes = int(v)
	}
	if v, ok := fields["actualMinutes"].(float64); ok {
		current.ActualMinutes = int(v)
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if v, ok := fields["notes"].(string); ok {
		current.Notes = &v
	}
	if v, ok := fields["interruptionsCount"].(float64); ok {
		current.InterruptionsCount = int(v)
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetStats(ctx context.Context) (*SessionStats, error) {
	return s.repo.GetStats(ctx)
}

// Pause pauses an in-progress session
func (s *Service) Pause(ctx context.Context, id string) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Status = "paused"
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

// Resume resumes a paused session
func (s *Service) Resume(ctx context.Context, id string) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Status = "in_progress"
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

// Complete completes a session with actual minutes and notes
func (s *Service) Complete(ctx context.Context, id string, actualMinutes int, notes *string) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Status = "completed"
	current.ActualMinutes = actualMinutes
	current.Notes = notes
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

// Cancel cancels a session
func (s *Service) Cancel(ctx context.Context, id string) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Status = "canceled"
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

// Interrupt adds an interruption to a session
func (s *Service) Interrupt(ctx context.Context, id string) (*Session, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.InterruptionsCount++
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

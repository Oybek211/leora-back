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

// BulkDelete soft deletes multiple habits
func (s *Service) BulkDelete(ctx context.Context, ids []string) (int64, error) {
	return s.repo.BulkDelete(ctx, ids)
}

// ToggleCompletion toggles the completion status for a habit on a specific date
func (s *Service) ToggleCompletion(ctx context.Context, habitID, dateKey string) (*HabitCompletion, error) {
	return s.repo.ToggleCompletion(ctx, habitID, dateKey)
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

// FinanceEvaluationResult represents the result of evaluating a habit against finance data
type FinanceEvaluationResult struct {
	HabitID          string  `json:"habitId"`
	DateKey          string  `json:"dateKey"`
	TransactionCount int     `json:"transactionCount"`
	TotalAmount      float64 `json:"totalAmount"`
	HasFinanceRule   bool    `json:"hasFinanceRule"`
	RuleSatisfied    bool    `json:"ruleSatisfied"`
	AutoCompleted    bool    `json:"autoCompleted"`
	Message          string  `json:"message"`
}

// EvaluateFinance evaluates a habit against finance transactions for the given date
func (s *Service) EvaluateFinance(ctx context.Context, habitID string, dateKey string) (*FinanceEvaluationResult, error) {
	// Get habit to check finance rule
	habit, err := s.repo.GetByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	result := &FinanceEvaluationResult{
		HabitID:        habitID,
		DateKey:        dateKey,
		HasFinanceRule: habit.FinanceRule != nil,
	}

	// Get transactions for this habit on the given date
	count, totalAmount, err := s.repo.GetTransactionCountForHabit(ctx, habitID, dateKey)
	if err != nil {
		return nil, err
	}

	result.TransactionCount = count
	result.TotalAmount = totalAmount

	// If no finance rule, just report the transaction data
	if habit.FinanceRule == nil {
		result.Message = "No finance rule configured for this habit"
		return result, nil
	}

	// Evaluate the finance rule
	rule := habit.FinanceRule
	ruleSatisfied := false

	switch rule.Type {
	case "spend_in_categories":
		// Check if minimum amount was spent in specified categories
		if rule.MinAmount != nil && totalAmount >= *rule.MinAmount {
			ruleSatisfied = true
		} else if rule.MinAmount == nil && count > 0 {
			ruleSatisfied = true
		}
	case "daily_spend_under":
		// Check if daily spending is under the limit (for "quit" habits)
		if rule.Amount != nil && totalAmount <= *rule.Amount {
			ruleSatisfied = true
		}
	case "no_spend_in_categories":
		// Check if no spending occurred in specified categories
		if count == 0 {
			ruleSatisfied = true
		}
	case "has_any_transactions":
		// Any transaction satisfies the rule
		if count > 0 {
			ruleSatisfied = true
		}
	default:
		// Unknown rule type - check if any transactions exist
		if count > 0 {
			ruleSatisfied = true
		}
	}

	result.RuleSatisfied = ruleSatisfied

	// Auto-complete if rule is satisfied
	if ruleSatisfied {
		value := totalAmount
		completion := &HabitCompletion{
			HabitID: habitID,
			DateKey: dateKey,
			Status:  "done",
			Value:   &value,
		}
		if err := s.repo.CreateCompletion(ctx, completion); err == nil {
			result.AutoCompleted = true
			result.Message = "Habit auto-completed based on finance activity"
		} else {
			result.Message = "Finance rule satisfied"
		}
	} else {
		result.Message = "Finance rule not satisfied"
	}

	return result, nil
}

// EvaluateAllFinance evaluates finance rules for all habits for the given date.
func (s *Service) EvaluateAllFinance(ctx context.Context, dateKey string) ([]*FinanceEvaluationResult, error) {
	habits, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]*FinanceEvaluationResult, 0, len(habits))
	for _, habit := range habits {
		if habit == nil || habit.ID == "" {
			continue
		}
		result, err := s.EvaluateFinance(ctx, habit.ID, dateKey)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

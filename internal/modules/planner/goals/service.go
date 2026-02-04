package goals

import "context"

// Service handles goal logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// calculateProgress computes progressPercent based on currentValue, targetValue, and direction
// Returns a value between 0 and 1 (clamped)
// Direction: "increase" means current should go UP to reach target (e.g., savings goal)
// Direction: "decrease" means current should go DOWN to reach target (e.g., weight loss)
func calculateProgress(currentValue float64, targetValue *float64, initialValue *float64, direction string) float64 {
	if targetValue == nil {
		return 0
	}

	target := *targetValue

	// If initialValue is set, use it for relative progress
	initial := float64(0)
	if initialValue != nil {
		initial = *initialValue
	}

	// Handle decrease direction (e.g., weight loss: 100kg -> 80kg)
	if direction == "decrease" {
		// For decrease: initial=100, target=80, current=100 -> 0%
		// For decrease: initial=100, target=80, current=90 -> 50%
		// For decrease: initial=100, target=80, current=80 -> 100%
		if initial == target {
			if currentValue <= target {
				return 1.0
			}
			return 0
		}
		// Progress = how much we've decreased / total decrease needed
		progress := (initial - currentValue) / (initial - target)

		// Clamp between 0 and 1
		if progress < 0 {
			return 0
		}
		if progress > 1 {
			return 1
		}
		return progress
	}

	// Default: increase direction (e.g., savings: 0 -> 1000)
	// Handle case where target == initial
	if target == initial {
		if currentValue >= target {
			return 1.0
		}
		return 0
	}

	// For increase: target must be > 0 for meaningful progress
	if target <= 0 {
		return 0
	}

	progress := (currentValue - initial) / (target - initial)

	// Clamp between 0 and 1
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

func (s *Service) List(ctx context.Context) ([]*Goal, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Goal, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, goal *Goal) (*Goal, error) {
	// Calculate progress on create (include direction for decrease goals like weight loss)
	goal.ProgressPercent = calculateProgress(goal.CurrentValue, goal.TargetValue, goal.InitialValue, goal.Direction)

	if err := s.repo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *Service) Update(ctx context.Context, id string, goal *Goal) (*Goal, error) {
	goal.ID = id
	// Recalculate progress on full update (include direction for decrease goals)
	goal.ProgressPercent = calculateProgress(goal.CurrentValue, goal.TargetValue, goal.InitialValue, goal.Direction)

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

	needsProgressRecalc := false

	if v, ok := fields["title"].(string); ok {
		current.Title = v
	}
	if v, ok := fields["goalType"].(string); ok {
		current.GoalType = v
	}
	if v, ok := fields["status"].(string); ok {
		current.Status = v
	}
	if v, ok := fields["currentValue"].(float64); ok {
		current.CurrentValue = v
		needsProgressRecalc = true
	}
	if v, ok := fields["targetValue"].(float64); ok {
		current.TargetValue = &v
		needsProgressRecalc = true
	}
	if v, ok := fields["initialValue"].(float64); ok {
		current.InitialValue = &v
		needsProgressRecalc = true
	}
	// Allow explicit progressPercent override, but only if not recalculating
	if v, ok := fields["progressPercent"].(float64); ok && !needsProgressRecalc {
		current.ProgressPercent = v
	}

	// Recalculate progress if value fields changed (include direction for decrease goals)
	if needsProgressRecalc {
		current.ProgressPercent = calculateProgress(current.CurrentValue, current.TargetValue, current.InitialValue, current.Direction)
	}

	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// BulkDelete soft deletes multiple goals and unlinks associated finance items
func (s *Service) BulkDelete(ctx context.Context, ids []string) (int64, []string, []string, error) {
	return s.repo.BulkDelete(ctx, ids)
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

func (s *Service) GetStats(ctx context.Context, goalID string) (*GoalStats, error) {
	return s.repo.GetStats(ctx, goalID)
}

func (s *Service) GetTasks(ctx context.Context, goalID string) ([]*TaskSummary, error) {
	return s.repo.ListTasksByGoal(ctx, goalID)
}

func (s *Service) GetHabits(ctx context.Context, goalID string) ([]*HabitSummary, error) {
	return s.repo.ListHabitsByGoal(ctx, goalID)
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

// LinkBudget links a budget to a goal (bidirectional)
func (s *Service) LinkBudget(ctx context.Context, goalID, budgetID string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	goal.LinkedBudgetID = &budgetID
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	// Update budget's linked_goal_id via repository
	if err := s.repo.UpdateBudgetGoalLink(ctx, budgetID, goalID); err != nil {
		// Log error but don't fail - goal link was successful
	}
	return s.repo.GetByID(ctx, goalID)
}

// UnlinkBudget removes budget link from a goal
func (s *Service) UnlinkBudget(ctx context.Context, goalID string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	oldBudgetID := goal.LinkedBudgetID
	goal.LinkedBudgetID = nil
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	// Clear budget's linked_goal_id
	if oldBudgetID != nil {
		_ = s.repo.UpdateBudgetGoalLink(ctx, *oldBudgetID, "")
	}
	return s.repo.GetByID(ctx, goalID)
}

// LinkDebt links a debt to a goal (bidirectional)
func (s *Service) LinkDebt(ctx context.Context, goalID, debtID string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	goal.LinkedDebtID = &debtID
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	// Update debt's linked_goal_id via repository
	if err := s.repo.UpdateDebtGoalLink(ctx, debtID, goalID); err != nil {
		// Log error but don't fail - goal link was successful
	}
	return s.repo.GetByID(ctx, goalID)
}

// UnlinkDebt removes debt link from a goal
func (s *Service) UnlinkDebt(ctx context.Context, goalID string) (*Goal, error) {
	goal, err := s.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	oldDebtID := goal.LinkedDebtID
	goal.LinkedDebtID = nil
	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	// Clear debt's linked_goal_id
	if oldDebtID != nil {
		_ = s.repo.UpdateDebtGoalLink(ctx, *oldDebtID, "")
	}
	return s.repo.GetByID(ctx, goalID)
}

// FinanceProgress represents finance-based progress for a goal
type FinanceProgress struct {
	GoalID              string   `json:"goalId"`
	LinkedBudgetID      *string  `json:"linkedBudgetId,omitempty"`
	LinkedDebtID        *string  `json:"linkedDebtId,omitempty"`
	BudgetSpentAmount   *float64 `json:"budgetSpentAmount,omitempty"`
	BudgetLimitAmount   *float64 `json:"budgetLimitAmount,omitempty"`
	BudgetPercentUsed   *float64 `json:"budgetPercentUsed,omitempty"`
	DebtPrincipalAmount *float64 `json:"debtPrincipalAmount,omitempty"`
	DebtPaidAmount      *float64 `json:"debtPaidAmount,omitempty"`
	DebtPercentPaid     *float64 `json:"debtPercentPaid,omitempty"`
	FinanceProgressPct  float64  `json:"financeProgressPercent"`
}

// GetFinanceProgress calculates finance-based progress for a goal
func (s *Service) GetFinanceProgress(ctx context.Context, goalID string) (*FinanceProgress, error) {
	goal, err := s.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}

	progress := &FinanceProgress{
		GoalID:         goalID,
		LinkedBudgetID: goal.LinkedBudgetID,
		LinkedDebtID:   goal.LinkedDebtID,
	}

	// Get budget progress if linked
	if goal.LinkedBudgetID != nil {
		budgetProgress, err := s.repo.GetBudgetProgress(ctx, *goal.LinkedBudgetID)
		if err == nil && budgetProgress != nil {
			progress.BudgetSpentAmount = &budgetProgress.SpentAmount
			progress.BudgetLimitAmount = &budgetProgress.LimitAmount
			if budgetProgress.LimitAmount > 0 {
				pct := (budgetProgress.SpentAmount / budgetProgress.LimitAmount) * 100
				progress.BudgetPercentUsed = &pct
				progress.FinanceProgressPct = pct
			}
		}
	}

	// Get debt progress if linked
	if goal.LinkedDebtID != nil {
		debtProgress, err := s.repo.GetDebtProgress(ctx, *goal.LinkedDebtID)
		if err == nil && debtProgress != nil {
			progress.DebtPrincipalAmount = &debtProgress.PrincipalAmount
			progress.DebtPaidAmount = &debtProgress.PaidAmount
			if debtProgress.PrincipalAmount > 0 {
				pct := (debtProgress.PaidAmount / debtProgress.PrincipalAmount) * 100
				progress.DebtPercentPaid = &pct
				progress.FinanceProgressPct = pct
			}
		}
	}

	return progress, nil
}

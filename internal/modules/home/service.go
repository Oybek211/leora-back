package home

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// --- Response types ---

type HomeTask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Time      string `json:"time"`
	Completed bool   `json:"completed"`
	Priority  string `json:"priority"`
	Context   string `json:"context"`
}

type HomeGoal struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Progress float64 `json:"progress"`
	Current  float64 `json:"current"`
	Target   float64 `json:"target"`
	Unit     string  `json:"unit"`
	Category string  `json:"category"`
}

type HabitSummary struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
}

type TaskSummary struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
}

type WeeklyStats struct {
	TasksCompleted int     `json:"tasksCompleted"`
	TotalTasks     int     `json:"totalTasks"`
	FocusHours     float64 `json:"focusHours"`
	Streak         int     `json:"streak"`
}

type FocusSessionItem struct {
	ID        string `json:"id"`
	Task      string `json:"task"`
	Duration  int    `json:"duration"`
	Completed bool   `json:"completed"`
}

type FocusSessionsSummary struct {
	Completed          int  `json:"completed"`
	TotalMinutes       int  `json:"totalMinutes"`
	NextSessionMinutes *int `json:"nextSessionMinutes"`
}

type TrendPoint struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type FocusSessionsData struct {
	Sessions []FocusSessionItem   `json:"sessions"`
	Summary  FocusSessionsSummary `json:"summary"`
	Trend    []TrendPoint         `json:"trend"`
}

type TransactionItem struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

type CategoryAmount struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type SpendingSummary struct {
	Categories []CategoryAmount `json:"categories"`
	Total      float64          `json:"total"`
}

type BudgetProgress struct {
	Label       string  `json:"label"`
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	PercentUsed float64 `json:"percentUsed"`
}

type CashFlowDay struct {
	Label   string  `json:"label"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type CashFlowTimeline struct {
	Days []CashFlowDay `json:"days"`
}

type HabitItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Streak    int    `json:"streak"`
	Completed bool   `json:"completed"`
	HabitType string `json:"habitType"`
}

type WidgetPayloads struct {
	WeeklyStats     WeeklyStats       `json:"weeklyStats"`
	FocusSessions   FocusSessionsData `json:"focusSessions"`
	Transactions    []TransactionItem `json:"transactions"`
	SpendingSummary SpendingSummary   `json:"spendingSummary"`
	BudgetList      []BudgetProgress  `json:"budgetList"`
	CashFlow        CashFlowTimeline  `json:"cashFlow"`
	Habits          []HabitItem       `json:"habits"`
}

type HomeResponse struct {
	Tasks         []HomeTask     `json:"tasks"`
	Goals         []HomeGoal     `json:"goals"`
	Progress      ProgressData   `json:"progress"`
	HabitsSummary HabitSummary   `json:"habitsSummary"`
	TasksSummary  TaskSummary    `json:"tasksSummary"`
	BudgetScore   float64        `json:"budgetScore"`
	Widgets       WidgetPayloads `json:"widgets"`
}

type ProgressData struct {
	Tasks  float64 `json:"tasks"`
	Budget float64 `json:"budget"`
	Focus  float64 `json:"focus"`
}

func (s *Service) Summary(ctx context.Context, date string) (*HomeResponse, error) {
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	userID, _ := ctx.Value("user_id").(string)

	response := &HomeResponse{}

	// --- Tasks for date ---
	s.loadTasks(ctx, response, userID, date)

	// --- Active goals (top 5) ---
	s.loadGoals(ctx, response, userID)

	// --- Task counts + progress ---
	var tasksDue, tasksCompleted int
	_ = s.db.GetContext(ctx, &tasksDue, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2 AND deleted_at IS NULL`, userID, date)
	_ = s.db.GetContext(ctx, &tasksCompleted, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2 AND status = 'completed' AND deleted_at IS NULL`, userID, date)
	if tasksDue > 0 {
		response.Progress.Tasks = math.Round((float64(tasksCompleted) / float64(tasksDue)) * 100)
	}
	response.TasksSummary = TaskSummary{Total: tasksDue, Completed: tasksCompleted}

	// --- Budget score + progress ---
	budgetScore := s.computeBudgetScore(ctx, userID)
	response.BudgetScore = budgetScore
	response.Progress.Budget = budgetScore

	// --- Focus progress ---
	var focusTotal, focusCompleted int
	_ = s.db.GetContext(ctx, &focusTotal, `SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND started_at::date = $2 AND deleted_at IS NULL`, userID, date)
	_ = s.db.GetContext(ctx, &focusCompleted, `SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND started_at::date = $2 AND status = 'completed' AND deleted_at IS NULL`, userID, date)
	if focusTotal > 0 {
		response.Progress.Focus = math.Round((float64(focusCompleted) / float64(focusTotal)) * 100)
	}

	// --- Habits summary ---
	var habitsTotal int
	_ = s.db.GetContext(ctx, &habitsTotal, `SELECT COUNT(*) FROM habits WHERE user_id = $1 AND deleted_at IS NULL AND show_status = 'active'`, userID)
	var habitsCompleted int
	_ = s.db.GetContext(ctx, &habitsCompleted, `SELECT COUNT(DISTINCT habit_id) FROM habit_completions WHERE date_key = $1 AND status = 'done'`, date)
	response.HabitsSummary = HabitSummary{Total: habitsTotal, Completed: habitsCompleted}

	// --- Widget data ---
	parsedDate, _ := time.Parse("2006-01-02", date)
	response.Widgets.WeeklyStats = s.buildWeeklyStats(ctx, userID, parsedDate)
	response.Widgets.FocusSessions = s.buildFocusSessions(ctx, userID, parsedDate)
	response.Widgets.Transactions = s.buildRecentTransactions(ctx, userID, date)
	response.Widgets.SpendingSummary = s.buildSpendingSummary(ctx, userID, date)
	response.Widgets.BudgetList = s.buildBudgetList(ctx, userID)
	response.Widgets.CashFlow = s.buildCashFlowTimeline(ctx, userID, parsedDate)
	response.Widgets.Habits = s.buildHabits(ctx, userID, parsedDate, date)

	return response, nil
}

func (s *Service) loadTasks(ctx context.Context, resp *HomeResponse, userID, date string) {
	rows, err := s.db.QueryxContext(ctx,
		`SELECT id, title, time_of_day, status, priority, context
		 FROM tasks WHERE user_id = $1 AND due_date = $2 AND deleted_at IS NULL
		 ORDER BY created_at DESC`, userID, date)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var task HomeTask
		var status string
		var timeOfDay, taskCtx *string
		if err := rows.Scan(&task.ID, &task.Title, &timeOfDay, &status, &task.Priority, &taskCtx); err == nil {
			if timeOfDay != nil {
				task.Time = *timeOfDay
			}
			task.Completed = status == "completed"
			if taskCtx != nil {
				task.Context = *taskCtx
			}
			resp.Tasks = append(resp.Tasks, task)
		}
	}
}

func (s *Service) loadGoals(ctx context.Context, resp *HomeResponse, userID string) {
	rows, err := s.db.QueryxContext(ctx,
		`SELECT id, title, progress_percent, current_value, target_value, unit, goal_type
		 FROM goals WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT 5`, userID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goal HomeGoal
		var unit *string
		if err := rows.Scan(&goal.ID, &goal.Title, &goal.Progress, &goal.Current, &goal.Target, &unit, &goal.Category); err == nil {
			if unit != nil {
				goal.Unit = *unit
			}
			resp.Goals = append(resp.Goals, goal)
		}
	}
}

func (s *Service) computeBudgetScore(ctx context.Context, userID string) float64 {
	type budgetRow struct {
		LimitAmount float64 `db:"limit_amount"`
		Spent       float64 `db:"spent"`
	}
	var rows []budgetRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT b.limit_amount,
		       COALESCE(SUM(ABS(t.converted_amount_to_base)), 0) as spent
		FROM budgets b
		LEFT JOIN transactions t ON (t.budget_id = b.id OR t.category_id = ANY(b.category_ids))
			AND t.type = 'expense' AND t.deleted_at IS NULL
			AND (b.period_type = 'none' OR (t.date >= b.start_date AND t.date <= b.end_date))
		WHERE b.user_id = $1 AND b.deleted_at IS NULL AND b.is_archived = false AND b.limit_amount > 0
		GROUP BY b.id, b.limit_amount
	`, userID)
	if err != nil || len(rows) == 0 {
		return 0
	}

	var totalUsage float64
	for _, r := range rows {
		usage := math.Min(r.Spent/r.LimitAmount, 1)
		totalUsage += usage
	}
	avg := totalUsage / float64(len(rows))
	return math.Round(math.Max(0, 100-avg*100))
}

func (s *Service) buildWeeklyStats(ctx context.Context, userID string, date time.Time) WeeklyStats {
	start := date.AddDate(0, 0, -6).Format("2006-01-02")
	end := date.Format("2006-01-02")

	var total, completed int
	_ = s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date BETWEEN $2 AND $3 AND deleted_at IS NULL`, userID, start, end)
	_ = s.db.GetContext(ctx, &completed, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date BETWEEN $2 AND $3 AND status = 'completed' AND deleted_at IS NULL`, userID, start, end)

	var focusMinutes float64
	_ = s.db.GetContext(ctx, &focusMinutes, `SELECT COALESCE(SUM(COALESCE(actual_minutes, planned_minutes, 0)), 0) FROM focus_sessions WHERE user_id = $1 AND started_at::date BETWEEN $2 AND $3 AND deleted_at IS NULL`, userID, start, end)

	var topStreak int
	_ = s.db.GetContext(ctx, &topStreak, `SELECT COALESCE(MAX(streak_current), 0) FROM habits WHERE user_id = $1 AND deleted_at IS NULL`, userID)

	hours := math.Round(focusMinutes/60*10) / 10
	return WeeklyStats{
		TasksCompleted: completed,
		TotalTasks:     total,
		FocusHours:     hours,
		Streak:         topStreak,
	}
}

func (s *Service) buildFocusSessions(ctx context.Context, userID string, date time.Time) FocusSessionsData {
	dateStr := date.Format("2006-01-02")
	result := FocusSessionsData{
		Sessions: []FocusSessionItem{},
		Trend:    []TrendPoint{},
	}

	// Daily sessions
	rows, err := s.db.QueryxContext(ctx, `
		SELECT fs.id, COALESCE(t.title, 'Focus session') as task_title,
		       COALESCE(fs.actual_minutes, fs.planned_minutes, 25) as duration,
		       fs.status
		FROM focus_sessions fs
		LEFT JOIN tasks t ON t.id = fs.task_id
		WHERE fs.user_id = $1 AND fs.started_at::date = $2 AND fs.deleted_at IS NULL
		ORDER BY fs.started_at DESC LIMIT 4
	`, userID, dateStr)
	if err == nil {
		defer rows.Close()
		var totalMinutes, completedCount int
		var nextSession *int
		for rows.Next() {
			var item FocusSessionItem
			var status string
			if err := rows.Scan(&item.ID, &item.Task, &item.Duration, &status); err == nil {
				item.Completed = status == "completed"
				if item.Completed {
					completedCount++
				} else if nextSession == nil {
					d := item.Duration
					nextSession = &d
				}
				totalMinutes += item.Duration
				result.Sessions = append(result.Sessions, item)
			}
		}
		result.Summary = FocusSessionsSummary{
			Completed:          completedCount,
			TotalMinutes:       totalMinutes,
			NextSessionMinutes: nextSession,
		}
	}

	// 5-day trend
	for i := -4; i <= 0; i++ {
		d := date.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")
		var minutes int
		_ = s.db.GetContext(ctx, &minutes, `SELECT COALESCE(SUM(COALESCE(actual_minutes, planned_minutes, 0)), 0) FROM focus_sessions WHERE user_id = $1 AND started_at::date = $2 AND deleted_at IS NULL`, userID, dStr)
		label := d.Format("Mon")
		result.Trend = append(result.Trend, TrendPoint{Label: label, Value: min(100, minutes)})
	}

	return result
}

func (s *Service) buildRecentTransactions(ctx context.Context, userID, date string) []TransactionItem {
	var items []TransactionItem
	rows, err := s.db.QueryxContext(ctx, `
		SELECT id, type, amount, currency, COALESCE(category_id, 'General') as category, date
		FROM transactions
		WHERE user_id = $1 AND date = $2 AND deleted_at IS NULL
		ORDER BY date DESC LIMIT 5
	`, userID, date)
	if err != nil {
		return items
	}
	defer rows.Close()
	for rows.Next() {
		var item TransactionItem
		if err := rows.Scan(&item.ID, &item.Type, &item.Amount, &item.Currency, &item.Category, &item.Date); err == nil {
			items = append(items, item)
		}
	}
	return items
}

func (s *Service) buildSpendingSummary(ctx context.Context, userID, date string) SpendingSummary {
	summary := SpendingSummary{Categories: []CategoryAmount{}}
	rows, err := s.db.QueryxContext(ctx, `
		SELECT COALESCE(category_id, 'Other') as category, SUM(ABS(converted_amount_to_base)) as total
		FROM transactions
		WHERE user_id = $1 AND date = $2 AND type = 'expense' AND deleted_at IS NULL
		GROUP BY category_id ORDER BY total DESC LIMIT 3
	`, userID, date)
	if err != nil {
		return summary
	}
	defer rows.Close()
	for rows.Next() {
		var cat CategoryAmount
		if err := rows.Scan(&cat.Label, &cat.Amount); err == nil {
			cat.Amount = math.Round(cat.Amount)
			summary.Total += cat.Amount
			summary.Categories = append(summary.Categories, cat)
		}
	}
	return summary
}

func (s *Service) buildBudgetList(ctx context.Context, userID string) []BudgetProgress {
	var items []BudgetProgress
	type row struct {
		Name        string  `db:"name"`
		LimitAmount float64 `db:"limit_amount"`
		Spent       float64 `db:"spent"`
	}
	var rows []row
	err := s.db.SelectContext(ctx, &rows, `
		SELECT b.name,
		       b.limit_amount,
		       COALESCE(SUM(ABS(t.converted_amount_to_base)), 0) as spent
		FROM budgets b
		LEFT JOIN transactions t ON (t.budget_id = b.id OR t.category_id = ANY(b.category_ids))
			AND t.type = 'expense' AND t.deleted_at IS NULL
			AND (b.period_type = 'none' OR (t.date >= b.start_date AND t.date <= b.end_date))
		WHERE b.user_id = $1 AND b.deleted_at IS NULL AND b.is_archived = false
		GROUP BY b.id, b.name, b.limit_amount
		ORDER BY CASE WHEN b.limit_amount > 0 THEN COALESCE(SUM(ABS(t.converted_amount_to_base)), 0) / b.limit_amount ELSE 0 END DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return items
	}
	for _, r := range rows {
		pct := 0.0
		if r.LimitAmount > 0 {
			pct = math.Min(r.Spent/r.LimitAmount, 1)
		}
		items = append(items, BudgetProgress{
			Label:       r.Name,
			Used:        math.Round(r.Spent),
			Total:       math.Round(r.LimitAmount),
			PercentUsed: pct,
		})
	}
	return items
}

func (s *Service) buildCashFlowTimeline(ctx context.Context, userID string, date time.Time) CashFlowTimeline {
	cf := CashFlowTimeline{Days: []CashFlowDay{}}
	for i := -4; i <= 0; i++ {
		d := date.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")
		var income, expense float64
		_ = s.db.GetContext(ctx, &income, `SELECT COALESCE(SUM(converted_amount_to_base), 0) FROM transactions WHERE user_id = $1 AND date = $2 AND type = 'income' AND deleted_at IS NULL`, userID, dStr)
		_ = s.db.GetContext(ctx, &expense, `SELECT COALESCE(SUM(ABS(converted_amount_to_base)), 0) FROM transactions WHERE user_id = $1 AND date = $2 AND type = 'expense' AND deleted_at IS NULL`, userID, dStr)
		cf.Days = append(cf.Days, CashFlowDay{
			Label:   d.Format("Mon"),
			Income:  math.Round(income),
			Expense: math.Round(expense),
		})
	}
	return cf
}

func (s *Service) buildHabits(ctx context.Context, userID string, date time.Time, dateStr string) []HabitItem {
	var items []HabitItem
	weekday := int(date.Weekday())

	rows, err := s.db.QueryxContext(ctx, fmt.Sprintf(`
		SELECT h.id, h.title, COALESCE(h.streak_current, 0) as streak, h.habit_type,
		       CASE WHEN hc.status = 'done' THEN true ELSE false END as completed
		FROM habits h
		LEFT JOIN habit_completions hc ON hc.habit_id = h.id AND hc.date_key = $2 AND hc.status = 'done'
		WHERE h.user_id = $1 AND h.deleted_at IS NULL AND h.show_status = 'active'
		AND (
			h.frequency = 'daily'
			OR (h.frequency = 'weekly')
			OR (h.days_of_week IS NOT NULL AND $3 = ANY(h.days_of_week))
		)
		ORDER BY h.created_at DESC LIMIT 3
	`), userID, dateStr, weekday)
	if err != nil {
		return items
	}
	defer rows.Close()
	for rows.Next() {
		var item HabitItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Streak, &item.HabitType, &item.Completed); err == nil {
			items = append(items, item)
		}
	}
	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package dashboard

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

type SummaryResponse struct {
	Date     string `json:"date"`
	Progress struct {
		Tasks  float64 `json:"tasks"`
		Budget float64 `json:"budget"`
		Focus  float64 `json:"focus"`
	} `json:"progress"`
	Counts struct {
		TasksDue     int `json:"tasksDue"`
		HabitsDue    int `json:"habitsDue"`
		GoalsActive  int `json:"goalsActive"`
		Transactions int `json:"transactions"`
	} `json:"counts"`
	Finance struct {
		Income   float64 `json:"income"`
		Expense  float64 `json:"expense"`
		Net      float64 `json:"net"`
		Currency string  `json:"currency"`
	} `json:"finance"`
}

type CalendarEntry struct {
	Progress struct {
		Tasks  float64 `json:"tasks"`
		Budget float64 `json:"budget"`
		Focus  float64 `json:"focus"`
	} `json:"progress"`
	Events struct {
		Tasks   int `json:"tasks"`
		Habits  int `json:"habits"`
		Goals   int `json:"goals"`
		Finance int `json:"finance"`
	} `json:"events"`
}

func (s *Service) Summary(ctx context.Context, date string) (*SummaryResponse, error) {
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	userID, _ := ctx.Value("user_id").(string)

	response := &SummaryResponse{Date: date}

	var tasksDue, tasksCompleted int
	_ = s.db.GetContext(ctx, &tasksDue, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2 AND deleted_at IS NULL`, userID, date)
	_ = s.db.GetContext(ctx, &tasksCompleted, `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2 AND status = 'completed' AND deleted_at IS NULL`, userID, date)
	response.Counts.TasksDue = tasksDue
	if tasksDue > 0 {
		response.Progress.Tasks = (float64(tasksCompleted) / float64(tasksDue)) * 100
	}

	var habitsDue int
	_ = s.db.GetContext(ctx, &habitsDue, `SELECT COUNT(*) FROM habits WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL`, userID)
	response.Counts.HabitsDue = habitsDue

	var goalsActive int
	_ = s.db.GetContext(ctx, &goalsActive, `SELECT COUNT(*) FROM goals WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL`, userID)
	response.Counts.GoalsActive = goalsActive

	var transactionsCount int
	_ = s.db.GetContext(ctx, &transactionsCount, `SELECT COUNT(*) FROM transactions WHERE user_id = $1 AND date = $2 AND deleted_at IS NULL`, userID, date)
	response.Counts.Transactions = transactionsCount

	var income float64
	_ = s.db.GetContext(ctx, &income, `SELECT COALESCE(SUM(converted_amount_to_base), 0) FROM transactions WHERE user_id = $1 AND date = $2 AND type = 'income' AND deleted_at IS NULL`, userID, date)
	var expense float64
	_ = s.db.GetContext(ctx, &expense, `SELECT COALESCE(SUM(converted_amount_to_base), 0) FROM transactions WHERE user_id = $1 AND date = $2 AND type = 'expense' AND deleted_at IS NULL`, userID, date)
	response.Finance.Income = income
	response.Finance.Expense = expense
	response.Finance.Net = income - expense
	response.Finance.Currency = "UZS"

	var focusTotal, focusCompleted int
	_ = s.db.GetContext(ctx, &focusTotal, `SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND started_at::date = $2 AND deleted_at IS NULL`, userID, date)
	_ = s.db.GetContext(ctx, &focusCompleted, `SELECT COUNT(*) FROM focus_sessions WHERE user_id = $1 AND started_at::date = $2 AND status = 'completed' AND deleted_at IS NULL`, userID, date)
	if focusTotal > 0 {
		response.Progress.Focus = (float64(focusCompleted) / float64(focusTotal)) * 100
	}

	var budgetLimit float64
	_ = s.db.GetContext(ctx, &budgetLimit, `SELECT COALESCE(SUM(limit_amount), 0) FROM budgets WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if budgetLimit > 0 {
		response.Progress.Budget = (expense / budgetLimit) * 100
	}

	return response, nil
}

func (s *Service) Calendar(ctx context.Context, fromDate, toDate string) (map[string]CalendarEntry, error) {
	entries := make(map[string]CalendarEntry)
	if fromDate == "" || toDate == "" {
		return entries, nil
	}
	start, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return entries, nil
	}
	end, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return entries, nil
	}

	// Initialize all days in range
	for day := start; !day.After(end); day = day.Add(24 * time.Hour) {
		key := day.Format("2006-01-02")
		entries[key] = CalendarEntry{}
	}

	userID, _ := ctx.Value("user_id").(string)

	// Batch query: task events per day
	type dayCount struct {
		Day       string `db:"day"`
		Total     int    `db:"total"`
		Completed int    `db:"completed"`
	}
	var taskCounts []dayCount
	_ = s.db.SelectContext(ctx, &taskCounts, `
		SELECT due_date::text as day,
		       COUNT(*) as total,
		       SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks
		WHERE user_id = $1 AND due_date BETWEEN $2 AND $3 AND deleted_at IS NULL
		GROUP BY due_date
	`, userID, fromDate, toDate)
	for _, tc := range taskCounts {
		if e, ok := entries[tc.Day]; ok {
			e.Events.Tasks = tc.Total
			if tc.Total > 0 {
				e.Progress.Tasks = (float64(tc.Completed) / float64(tc.Total)) * 100
			}
			entries[tc.Day] = e
		}
	}

	// Batch query: habit completions per day
	type dateCount struct {
		Day   string `db:"day"`
		Count int    `db:"count"`
	}
	var habitCounts []dateCount
	_ = s.db.SelectContext(ctx, &habitCounts, `
		SELECT date_key as day, COUNT(DISTINCT habit_id) as count
		FROM habit_completions
		WHERE date_key BETWEEN $1 AND $2 AND status = 'done'
		GROUP BY date_key
	`, fromDate, toDate)
	for _, hc := range habitCounts {
		if e, ok := entries[hc.Day]; ok {
			e.Events.Habits = hc.Count
			entries[hc.Day] = e
		}
	}

	// Batch query: transaction events per day
	var txnCounts []dateCount
	_ = s.db.SelectContext(ctx, &txnCounts, `
		SELECT date::text as day, COUNT(*) as count
		FROM transactions
		WHERE user_id = $1 AND date BETWEEN $2 AND $3 AND deleted_at IS NULL
		GROUP BY date
	`, userID, fromDate, toDate)
	for _, tc := range txnCounts {
		if e, ok := entries[tc.Day]; ok {
			e.Events.Finance = tc.Count
			entries[tc.Day] = e
		}
	}

	// Batch query: focus session events per day
	var focusCounts []dayCount
	_ = s.db.SelectContext(ctx, &focusCounts, `
		SELECT started_at::date::text as day,
		       COUNT(*) as total,
		       SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM focus_sessions
		WHERE user_id = $1 AND started_at::date BETWEEN $2 AND $3 AND deleted_at IS NULL
		GROUP BY started_at::date
	`, userID, fromDate, toDate)
	for _, fc := range focusCounts {
		if e, ok := entries[fc.Day]; ok {
			if fc.Total > 0 {
				e.Progress.Focus = (float64(fc.Completed) / float64(fc.Total)) * 100
			}
			entries[fc.Day] = e
		}
	}

	return entries, nil
}

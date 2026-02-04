package reports

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

type FinanceSummary struct {
	Currency    string  `json:"currency"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	Net         float64 `json:"net"`
	SavingsRate float64 `json:"savingsRate"`
	Period      struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"period"`
}

type CategoryBreakdown struct {
	Total      float64         `json:"total"`
	Categories []CategoryEntry `json:"categories"`
}

type CategoryEntry struct {
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Share        float64 `json:"share"`
}

type CashflowBucket struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

type CashflowReport struct {
	Granularity string           `json:"granularity"`
	Series      []CashflowBucket `json:"series"`
}

type DebtDueSoon struct {
	DebtID           string  `json:"debtId"`
	CounterpartyName string  `json:"counterpartyName"`
	DueDate          string  `json:"dueDate"`
	RemainingAmount  float64 `json:"remainingAmount"`
}

type DebtReport struct {
	ActiveCount int           `json:"activeCount"`
	DueSoon     []DebtDueSoon `json:"dueSoon"`
}

type ProductivityReport struct {
	TasksCompleted int `json:"tasksCompleted"`
	FocusMinutes   int `json:"focusMinutes"`
	HabitStreaks   int `json:"habitStreaks"`
}

func (s *Service) FinanceSummary(ctx context.Context, fromDate, toDate string) (*FinanceSummary, error) {
	fromDate, toDate = normalizeRange(fromDate, toDate)
	summary := &FinanceSummary{Currency: "UZS"}
	summary.Period.From = fromDate
	summary.Period.To = toDate

	_ = s.db.GetContext(ctx, &summary.Income, `SELECT COALESCE(SUM(converted_amount_to_base), 0) FROM transactions WHERE date BETWEEN $1 AND $2 AND type = 'income' AND deleted_at IS NULL`, fromDate, toDate)
	_ = s.db.GetContext(ctx, &summary.Expense, `SELECT COALESCE(SUM(converted_amount_to_base), 0) FROM transactions WHERE date BETWEEN $1 AND $2 AND type = 'expense' AND deleted_at IS NULL`, fromDate, toDate)
	summary.Net = summary.Income - summary.Expense
	if summary.Income > 0 {
		summary.SavingsRate = (summary.Net / summary.Income) * 100
	}
	return summary, nil
}

func (s *Service) FinanceCategories(ctx context.Context, fromDate, toDate string) (*CategoryBreakdown, error) {
	fromDate, toDate = normalizeRange(fromDate, toDate)
	breakdown := &CategoryBreakdown{}
	type row struct {
		CategoryID string  `db:"category_id"`
		Amount     float64 `db:"amount"`
	}
	rows := []row{}
	query := `
		SELECT COALESCE(category_id, 'uncategorized') as category_id, COALESCE(SUM(converted_amount_to_base), 0) as amount
		FROM transactions
		WHERE date BETWEEN $1 AND $2 AND type = 'expense' AND deleted_at IS NULL
		GROUP BY category_id
		ORDER BY amount DESC
	`
	_ = s.db.SelectContext(ctx, &rows, query, fromDate, toDate)
	for _, entry := range rows {
		breakdown.Total += entry.Amount
		breakdown.Categories = append(breakdown.Categories, CategoryEntry{
			CategoryID:   entry.CategoryID,
			CategoryName: entry.CategoryID,
			Amount:       entry.Amount,
		})
	}
	for i := range breakdown.Categories {
		if breakdown.Total > 0 {
			breakdown.Categories[i].Share = (breakdown.Categories[i].Amount / breakdown.Total) * 100
		}
	}
	return breakdown, nil
}

func (s *Service) Cashflow(ctx context.Context, fromDate, toDate, granularity string) (*CashflowReport, error) {
	fromDate, toDate = normalizeRange(fromDate, toDate)
	if granularity == "" {
		granularity = "day"
	}
	report := &CashflowReport{Granularity: granularity}
	var rows []CashflowBucket
	query := `
		SELECT date::text as date,
			COALESCE(SUM(CASE WHEN type = 'income' THEN converted_amount_to_base ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN converted_amount_to_base ELSE 0 END), 0) as expense,
			COALESCE(SUM(CASE WHEN type = 'income' THEN converted_amount_to_base ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN type = 'expense' THEN converted_amount_to_base ELSE 0 END), 0) as net
		FROM transactions
		WHERE date BETWEEN $1 AND $2 AND deleted_at IS NULL
		GROUP BY date
		ORDER BY date
	`
	_ = s.db.SelectContext(ctx, &rows, query, fromDate, toDate)
	report.Series = rows
	return report, nil
}

func (s *Service) DebtReport(ctx context.Context) (*DebtReport, error) {
	report := &DebtReport{}
	_ = s.db.GetContext(ctx, &report.ActiveCount, `SELECT COUNT(*) FROM debts WHERE status = 'active' AND deleted_at IS NULL`)
	var dueSoon []DebtDueSoon
	query := `
		SELECT id as debt_id, counterparty_name, due_date::text as due_date, remaining_amount
		FROM debts
		WHERE status = 'active' AND due_date IS NOT NULL AND due_date <= (NOW()::date + INTERVAL '7 days') AND deleted_at IS NULL
		ORDER BY due_date ASC
	`
	_ = s.db.SelectContext(ctx, &dueSoon, query)
	report.DueSoon = dueSoon
	return report, nil
}

func (s *Service) Productivity(ctx context.Context, fromDate, toDate string) (*ProductivityReport, error) {
	fromDate, toDate = normalizeRange(fromDate, toDate)
	report := &ProductivityReport{}
	_ = s.db.GetContext(ctx, &report.TasksCompleted, `SELECT COUNT(*) FROM tasks WHERE status = 'completed' AND updated_at::date BETWEEN $1 AND $2 AND deleted_at IS NULL`, fromDate, toDate)
	_ = s.db.GetContext(ctx, &report.FocusMinutes, `SELECT COALESCE(SUM(actual_minutes), 0) FROM focus_sessions WHERE started_at::date BETWEEN $1 AND $2 AND deleted_at IS NULL`, fromDate, toDate)
	_ = s.db.GetContext(ctx, &report.HabitStreaks, `SELECT COALESCE(SUM(streak_current), 0) FROM habits WHERE deleted_at IS NULL`)
	return report, nil
}

func (s *Service) InsightsContext(ctx context.Context, fromDate, toDate string) (map[string]interface{}, error) {
	fromDate, toDate = normalizeRange(fromDate, toDate)
	context := map[string]interface{}{
		"period": map[string]string{"from": fromDate, "to": toDate},
	}
	return context, nil
}

func normalizeRange(fromDate, toDate string) (string, string) {
	if fromDate == "" || toDate == "" {
		end := time.Now().UTC().Format("2006-01-02")
		start := time.Now().AddDate(0, 0, -30).UTC().Format("2006-01-02")
		return start, end
	}
	return fromDate, toDate
}

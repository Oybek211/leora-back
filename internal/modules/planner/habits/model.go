package habits

// Habit represents a recurring behavior to track
type Habit struct {
	ID                 string        `json:"id"`
	Title              string        `json:"title"`
	Description        *string       `json:"description,omitempty"`
	IconID             *string       `json:"iconId,omitempty"`
	HabitType          string        `json:"habitType"`
	Status             string        `json:"status"`
	ShowStatus         string        `json:"showStatus"`
	GoalID             *string       `json:"goalId,omitempty"`
	Frequency          string        `json:"frequency"`
	DaysOfWeek         []int         `json:"daysOfWeek,omitempty"`
	TimesPerWeek       *int          `json:"timesPerWeek,omitempty"`
	TimeOfDay          *string       `json:"timeOfDay,omitempty"`
	CompletionMode     string        `json:"completionMode"`
	TargetPerDay       *float64      `json:"targetPerDay,omitempty"`
	Unit               *string       `json:"unit,omitempty"`
	CountingType       string        `json:"countingType"`
	Difficulty         string        `json:"difficulty"`
	Priority           string        `json:"priority"`
	ChallengeLengthDays *int         `json:"challengeLengthDays,omitempty"`
	ReminderEnabled    bool          `json:"reminderEnabled"`
	ReminderTime       *string       `json:"reminderTime,omitempty"`
	StreakCurrent      int           `json:"streakCurrent"`
	StreakBest         int           `json:"streakBest"`
	CompletionRate30d  float64       `json:"completionRate30d"`
	FinanceRule        *FinanceRule  `json:"financeRule,omitempty"`
	LinkedGoalIDs      []string      `json:"linkedGoalIds,omitempty"`
	CreatedAt          string        `json:"createdAt,omitempty"`
	UpdatedAt          string        `json:"updatedAt,omitempty"`
	DeletedAt          *string       `json:"deletedAt,omitempty"`
}

// FinanceRule defines automatic completion rules for finance habits
type FinanceRule struct {
	Type        string   `json:"type"`                  // no_spend_in_categories, spend_in_categories, has_any_transactions, daily_spend_under
	CategoryIDs []string `json:"categoryIds,omitempty"` // For category-based rules
	AccountIDs  []string `json:"accountIds,omitempty"`  // For account-based rules
	MinAmount   *float64 `json:"minAmount,omitempty"`   // For spend_in_categories
	Amount      *float64 `json:"amount,omitempty"`      // For daily_spend_under
	Currency    *string  `json:"currency,omitempty"`    // Currency for amounts
}

// HabitCompletion represents a single completion entry for a habit
type HabitCompletion struct {
	ID        string   `json:"id" db:"id"`
	HabitID   string   `json:"habitId" db:"habit_id"`
	DateKey   string   `json:"dateKey" db:"date_key"`
	Status    string   `json:"status" db:"status"`
	Value     *float64 `json:"value,omitempty" db:"value"`
	CreatedAt string   `json:"createdAt,omitempty" db:"created_at"`
}

// HabitStats represents streak and statistics for a habit
type HabitStats struct {
	StreakCurrent     int     `json:"streakCurrent"`
	StreakBest        int     `json:"streakBest"`
	CompletionRate30d float64 `json:"completionRate30d"`
	TotalCompletions  int     `json:"totalCompletions"`
	TotalMisses       int     `json:"totalMisses"`
}

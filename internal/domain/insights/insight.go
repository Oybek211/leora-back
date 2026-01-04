package insights

// InsightSummary captures aggregated insights for planner and finance.
type InsightSummary struct {
    TasksCompleted int     `json:"tasksCompleted"`
    HabitsStreak   int     `json:"habitsStreak"`
    FocusMinutes   int     `json:"focusMinutes"`
    SavingsGrowth  float64 `json:"savingsGrowth"`
}

// InsightDetail represents a time-series insight.
type InsightDetail struct {
    Period string  `json:"period"`
    Value  float64 `json:"value"`
}

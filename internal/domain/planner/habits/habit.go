package habits

import "github.com/leora/leora-server/internal/domain/common"

// HabitType enumerates different habit categories.
type HabitType string

const (
    HabitTypeHealth       HabitType = "health"
    HabitTypeFinance      HabitType = "finance"
    HabitTypeProductivity HabitType = "productivity"
    HabitTypeEducation    HabitType = "education"
    HabitTypePersonal     HabitType = "personal"
    HabitTypeCustom       HabitType = "custom"
)

// HabitFrequency defines repetition schemes.
type HabitFrequency string

const (
    HabitFrequencyDaily   HabitFrequency = "daily"
    HabitFrequencyWeekly  HabitFrequency = "weekly"
    HabitFrequencyCustom  HabitFrequency = "custom"
)

// CompletionMode indicates how a habit is marked done.
type CompletionMode string

const (
    CompletionModeBoolean CompletionMode = "boolean"
    CompletionModeNumeric CompletionMode = "numeric"
)

// CountingType defines the habit intention.
type CountingType string

const (
    CountingTypeCreate CountingType = "create"
    CountingTypeQuit   CountingType = "quit"
)

// Difficulty levels.
type Difficulty string

const (
    DifficultyEasy   Difficulty = "easy"
    DifficultyMedium Difficulty = "medium"
    DifficultyHard   Difficulty = "hard"
)

// Habit represents a recurring action the user tracks.
type Habit struct {
    common.BaseEntity
    Title               string         `json:"title"`
    Description         string         `json:"description,omitempty"`
    IconID              string         `json:"iconId,omitempty"`
    Type                HabitType      `json:"habitType"`
    Status              string         `json:"status"`
    Frequency           HabitFrequency `json:"frequency"`
    DaysOfWeek          []int          `json:"daysOfWeek,omitempty"`
    TimesPerWeek        int            `json:"timesPerWeek,omitempty"`
    TimeOfDay           string         `json:"timeOfDay,omitempty"`
    CompletionMode      CompletionMode `json:"completionMode"`
    TargetPerDay        float64        `json:"targetPerDay,omitempty"`
    Unit                string         `json:"unit,omitempty"`
    CountingType        CountingType   `json:"countingType"`
    Difficulty          Difficulty     `json:"difficulty"`
    Priority            string         `json:"priority"`
    ReminderEnabled     bool           `json:"reminderEnabled"`
    ReminderTime        string         `json:"reminderTime,omitempty"`
    StreakCurrent       int            `json:"streakCurrent"`
    StreakBest          int            `json:"streakBest"`
    CompletionRate30d   float64        `json:"completionRate30d"`
}

// HabitCompletionEntry logs daily completions.
type HabitCompletionEntry struct {
    ID        string `json:"id"`
    HabitID   string `json:"habitId"`
    DateKey   string `json:"dateKey"`
    Status    string `json:"status"`
    Value     float64 `json:"value,omitempty"`
    CreatedAt string `json:"createdAt"`
}

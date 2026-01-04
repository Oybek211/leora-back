package errors

import "net/http"

type Error struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func StatusFromType(errType string) int {
	switch errType {
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

var (
	UserNotFound      = &Error{Code: -2000, Type: "NOT_FOUND", Message: "User not found"}
	UserAlreadyExists = &Error{Code: -2001, Type: "CONFLICT", Message: "User already exists"}
	InvalidUserData   = &Error{Code: -2002, Type: "BAD_REQUEST", Message: "Invalid user data"}
)

var (
	TaskNotFound         = &Error{Code: -3000, Type: "NOT_FOUND", Message: "Task not found"}
	GoalNotFound         = &Error{Code: -3001, Type: "NOT_FOUND", Message: "Goal not found"}
	HabitNotFound        = &Error{Code: -3002, Type: "NOT_FOUND", Message: "Habit not found"}
	FocusSessionNotFound = &Error{Code: -3003, Type: "NOT_FOUND", Message: "Focus session not found"}
	InvalidPlannerData   = &Error{Code: -3004, Type: "BAD_REQUEST", Message: "Invalid planner data"}
)

var (
	InvalidCredentials = &Error{Code: -4000, Type: "UNAUTHORIZED", Message: "Invalid credentials"}
	InvalidToken       = &Error{Code: -4001, Type: "UNAUTHORIZED", Message: "Invalid token"}
	TokenExpired       = &Error{Code: -4002, Type: "UNAUTHORIZED", Message: "Token expired"}
	PermissionDenied   = &Error{Code: -4003, Type: "FORBIDDEN", Message: "Permission denied"}
)

var (
	AccountNotFound     = &Error{Code: -5000, Type: "NOT_FOUND", Message: "Account not found"}
	TransactionNotFound = &Error{Code: -5001, Type: "NOT_FOUND", Message: "Transaction not found"}
	BudgetNotFound      = &Error{Code: -5002, Type: "NOT_FOUND", Message: "Budget not found"}
	DebtNotFound        = &Error{Code: -5003, Type: "NOT_FOUND", Message: "Debt not found"}
	InvalidFinanceData  = &Error{Code: -5004, Type: "BAD_REQUEST", Message: "Invalid finance data"}
)

var (
	NotificationNotFound   = &Error{Code: -6000, Type: "NOT_FOUND", Message: "Notification not found"}
	InvalidNotificationData = &Error{Code: -6001, Type: "BAD_REQUEST", Message: "Invalid notification data"}
)

var (
	InternalServerError = &Error{Code: -9000, Type: "INTERNAL", Message: "Internal server error"}
	DatabaseError       = &Error{Code: -9001, Type: "INTERNAL", Message: "Database error"}
	RedisUnavailable    = &Error{Code: -9002, Type: "INTERNAL", Message: "Redis unavailable"}
	WidgetNotFound      = &Error{Code: -9003, Type: "NOT_FOUND", Message: "Widget not found"}
	InvalidWidgetData   = &Error{Code: -9004, Type: "BAD_REQUEST", Message: "Invalid widget data"}
	SubscriptionNotFound = &Error{Code: -9005, Type: "NOT_FOUND", Message: "Subscription not found"}
	PlanNotFound        = &Error{Code: -9006, Type: "NOT_FOUND", Message: "Plan not found"}
	InvalidSubscriptionData = &Error{Code: -9007, Type: "BAD_REQUEST", Message: "Invalid subscription data"}
	SearchError         = &Error{Code: -9008, Type: "INTERNAL", Message: "Search error"}
)

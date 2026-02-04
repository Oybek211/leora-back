package errors

import "net/http"

type Error struct {
	Code    int                    `json:"code"`
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Slug    string                 `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func StatusFromType(errType string) int {
	switch errType {
	case "VALIDATION":
		return http.StatusBadRequest
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "ACCESS_TOKEN_EXPIRED":
		return http.StatusUnauthorized
	case "REFRESH_TOKEN_EXPIRED":
		return http.StatusUnauthorized
	case "INVALID_REFRESH_TOKEN":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "ACCOUNT_HAS_TRANSACTIONS":
		return http.StatusConflict
	case "TXN_IMMUTABLE":
		return http.StatusConflict
	case "RATE_LIMITED":
		return http.StatusTooManyRequests
	case "ACCOUNT_REQUIRED":
		return http.StatusBadRequest
	case "INSUFFICIENT_FUNDS":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func WithDetails(err *Error, details map[string]interface{}) *Error {
	if err == nil {
		return nil
	}
	copyErr := *err
	copyErr.Details = details
	return &copyErr
}

var (
	UserNotFound      = &Error{Code: -2000, Type: "NOT_FOUND", Message: "User not found"}
	UserAlreadyExists = &Error{Code: -2001, Type: "CONFLICT", Message: "User already exists"}
	InvalidUserData   = &Error{Code: -2002, Type: "VALIDATION", Message: "Invalid user data"}
)

var (
	TaskNotFound         = &Error{Code: -3000, Type: "NOT_FOUND", Message: "Task not found"}
	GoalNotFound         = &Error{Code: -3001, Type: "NOT_FOUND", Message: "Goal not found"}
	HabitNotFound        = &Error{Code: -3002, Type: "NOT_FOUND", Message: "Habit not found"}
	FocusSessionNotFound = &Error{Code: -3003, Type: "NOT_FOUND", Message: "Focus session not found"}
	InvalidPlannerData   = &Error{Code: -3004, Type: "VALIDATION", Message: "Invalid planner data"}
)

var (
	InvalidCredentials = &Error{Code: -4000, Type: "UNAUTHORIZED", Message: "Invalid credentials"}
	InvalidToken       = &Error{Code: -4001, Type: "UNAUTHORIZED", Message: "Invalid token"}
	TokenExpired       = &Error{Code: -4002, Type: "ACCESS_TOKEN_EXPIRED", Message: "Access token expired"}
	PermissionDenied   = &Error{Code: -4003, Type: "FORBIDDEN", Message: "Permission denied"}
	RefreshTokenExpired = &Error{Code: -4004, Type: "REFRESH_TOKEN_EXPIRED", Message: "Refresh token expired"}
	InvalidRefreshToken = &Error{Code: -4005, Type: "INVALID_REFRESH_TOKEN", Message: "Invalid refresh token"}
	InvalidGoogleToken  = &Error{Code: -4006, Type: "UNAUTHORIZED", Message: "Invalid Google ID token"}
	InvalidAppleToken   = &Error{Code: -4007, Type: "UNAUTHORIZED", Message: "Invalid Apple ID token"}
)

var (
	AccountNotFound      = &Error{Code: -5000, Type: "NOT_FOUND", Message: "Account not found", Slug: "FIN_ACCOUNT_NOT_FOUND"}
	TransactionNotFound  = &Error{Code: -5001, Type: "NOT_FOUND", Message: "Transaction not found", Slug: "FIN_TRANSACTION_NOT_FOUND"}
	BudgetNotFound       = &Error{Code: -5002, Type: "NOT_FOUND", Message: "Budget not found", Slug: "FIN_BUDGET_NOT_FOUND"}
	DebtNotFound         = &Error{Code: -5003, Type: "NOT_FOUND", Message: "Debt not found", Slug: "FIN_DEBT_NOT_FOUND"}
	InvalidFinanceData   = &Error{Code: -5004, Type: "VALIDATION", Message: "Invalid finance data", Slug: "FIN_INVALID_INPUT"}
	CounterpartyNotFound = &Error{Code: -5005, Type: "NOT_FOUND", Message: "Counterparty not found"}
	DebtPaymentNotFound  = &Error{Code: -5006, Type: "NOT_FOUND", Message: "Debt payment not found"}
	FXRateNotFound       = &Error{Code: -5007, Type: "NOT_FOUND", Message: "FX rate not found", Slug: "FX_RATE_NOT_FOUND"}
	InvalidAmount        = &Error{Code: -5026, Type: "VALIDATION", Message: "Invalid amount", Slug: "FIN_INVALID_AMOUNT"}
	InvalidCurrency      = &Error{Code: -5027, Type: "VALIDATION", Message: "Invalid currency", Slug: "FIN_INVALID_CURRENCY"}
	CategoryNotFound     = &Error{Code: -5028, Type: "NOT_FOUND", Message: "Category not found", Slug: "FIN_CATEGORY_NOT_FOUND"}

	// Debt counterparty validation errors
	CounterpartyRequired      = &Error{Code: -5010, Type: "VALIDATION", Message: "Counterparty is required for debt"}
	CounterpartyNameTooShort  = &Error{Code: -5011, Type: "VALIDATION", Message: "Counterparty name must be at least 2 characters"}
	InvalidDebtDirection      = &Error{Code: -5012, Type: "VALIDATION", Message: "Direction must be 'i_owe' or 'they_owe_me'"}
	CounterpartyHasDebts      = &Error{Code: -5013, Type: "CONFLICT", Message: "Cannot delete counterparty that has linked debts"}
	CounterpartyForbidden     = &Error{Code: -5014, Type: "FORBIDDEN", Message: "Counterparty does not belong to this user"}
	InvalidDueDateRange       = &Error{Code: -5015, Type: "VALIDATION", Message: "Due date must be on or after start date"}
	InvalidDebtAmount         = &Error{Code: -5016, Type: "VALIDATION", Message: "Principal amount must be greater than 0"}
	DebtStartDateRequired     = &Error{Code: -5017, Type: "VALIDATION", Message: "startDate is required"}
	InvalidDebtStartDate      = &Error{Code: -5018, Type: "VALIDATION", Message: "startDate must be YYYY-MM-DD"}
	PrincipalCurrencyRequired = &Error{Code: -5019, Type: "VALIDATION", Message: "principalCurrency is required"}
	InvalidDebtDueDate        = &Error{Code: -5020, Type: "VALIDATION", Message: "dueDate must be YYYY-MM-DD"}
	InvalidTransactionDate    = &Error{Code: -5021, Type: "VALIDATION", Message: "date must be YYYY-MM-DD"}
	AccountRequired           = &Error{Code: -5022, Type: "ACCOUNT_REQUIRED", Message: "Please create an account first", Slug: "FIN_ACCOUNT_REQUIRED"}
	AccountHasTransactions    = &Error{Code: -5023, Type: "ACCOUNT_HAS_TRANSACTIONS", Message: "Account has transactions and cannot be deleted", Slug: "FIN_ACCOUNT_HAS_TRANSACTIONS"}
	InsufficientFunds         = &Error{Code: -5024, Type: "INSUFFICIENT_FUNDS", Message: "Insufficient funds", Slug: "FIN_INSUFFICIENT_FUNDS"}
	TransactionImmutable      = &Error{Code: -5025, Type: "TXN_IMMUTABLE", Message: "Transaction cannot be modified", Slug: "FIN_TXN_IMMUTABLE"}
)

var (
	NotificationNotFound    = &Error{Code: -6000, Type: "NOT_FOUND", Message: "Notification not found"}
	InvalidNotificationData = &Error{Code: -6001, Type: "VALIDATION", Message: "Invalid notification data"}
)

var (
	InternalServerError     = &Error{Code: -9000, Type: "INTERNAL", Message: "Internal server error"}
	DatabaseError           = &Error{Code: -9001, Type: "INTERNAL", Message: "Database error"}
	RedisUnavailable        = &Error{Code: -9002, Type: "INTERNAL", Message: "Redis unavailable"}
	WidgetNotFound          = &Error{Code: -9003, Type: "NOT_FOUND", Message: "Widget not found"}
	InvalidWidgetData       = &Error{Code: -9004, Type: "VALIDATION", Message: "Invalid widget data"}
	SubscriptionNotFound    = &Error{Code: -9005, Type: "NOT_FOUND", Message: "Subscription not found"}
	PlanNotFound            = &Error{Code: -9006, Type: "NOT_FOUND", Message: "Plan not found"}
	InvalidSubscriptionData = &Error{Code: -9007, Type: "VALIDATION", Message: "Invalid subscription data"}
	SearchError             = &Error{Code: -9008, Type: "INTERNAL", Message: "Search error"}
)

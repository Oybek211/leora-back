package finance

const (
	TransactionTypeIncome                  = "income"
	TransactionTypeExpense                 = "expense"
	TransactionTypeTransfer                = "transfer"
	TransactionTypeTransferIn              = "transfer_in"
	TransactionTypeTransferOut             = "transfer_out"
	TransactionTypeSystemOpening           = "system_opening"
	TransactionTypeSystemAdjustment        = "system_adjustment"
	TransactionTypeSystemArchive           = "system_archive"
	TransactionTypeDebtCreate              = "debt_create"
	TransactionTypeDebtPayment             = "debt_payment"
	TransactionTypeDebtAdjustment          = "debt_adjustment"
	TransactionTypeAccountCreateFunding    = "account_create_funding"
	TransactionTypeAccountDeleteWithdrawal = "account_delete_withdrawal"
	TransactionTypeBudgetAddValue          = "budget_add_value"
	TransactionTypeDebtAddValue            = "debt_add_value"
	TransactionTypeDebtFullPayment         = "debt_full_payment"
)

const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

// Account represents a financial account.
type Account struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	Name           string  `json:"name"`
	AccountType    string  `json:"accountType"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initialBalance"`
	CurrentBalance float64 `json:"currentBalance"`
	LinkedGoalID   *string `json:"linkedGoalId,omitempty"`
	CustomTypeID   *string `json:"customTypeId,omitempty"`
	IsMain         bool    `json:"isMain"`
	IsArchived     bool    `json:"isArchived"`
	ShowStatus     string  `json:"showStatus"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	UpdatedAt      string  `json:"updatedAt,omitempty"`
	DeletedAt      string  `json:"-"`
}

// Transaction models a ledger entry.
type Transaction struct {
	ID                    string                 `json:"id"`
	UserID                string                 `json:"userId"`
	Type                  string                 `json:"type"`
	Status                string                 `json:"status"`
	AccountID             *string                `json:"accountId,omitempty"`
	FromAccountID         *string                `json:"fromAccountId,omitempty"`
	ToAccountID           *string                `json:"toAccountId,omitempty"`
	ReferenceType         *string                `json:"referenceType,omitempty"`
	ReferenceID           *string                `json:"referenceId,omitempty"`
	Amount                float64                `json:"amount"`
	Currency              string                 `json:"currency"`
	BaseCurrency          string                 `json:"baseCurrency"`
	RateUsedToBase        float64                `json:"rateUsedToBase"`
	ConvertedAmountToBase float64                `json:"convertedAmountToBase"`
	ToAmount              float64                `json:"toAmount"`
	ToCurrency            *string                `json:"toCurrency,omitempty"`
	EffectiveRateFromTo   float64                `json:"effectiveRateFromTo"`
	FeeAmount             float64                `json:"feeAmount"`
	FeeCategoryID         *string                `json:"feeCategoryId,omitempty"`
	CategoryID            *string                `json:"categoryId,omitempty"`
	SubcategoryID         *string                `json:"subcategoryId,omitempty"`
	Name                  *string                `json:"name,omitempty"`
	Description           *string                `json:"description,omitempty"`
	Date                  string                 `json:"date"`
	Time                  *string                `json:"time,omitempty"`
	GoalID                *string                `json:"goalId,omitempty"`
	BudgetID              *string                `json:"budgetId,omitempty"`
	DebtID                *string                `json:"debtId,omitempty"`
	HabitID               *string                `json:"habitId,omitempty"`
	CounterpartyID        *string                `json:"counterpartyId,omitempty"`
	RecurringID           *string                `json:"recurringId,omitempty"`
	Attachments           []string               `json:"attachments"`
	Tags                  []string               `json:"tags"`
	IsBalanceAdjustment   bool                   `json:"isBalanceAdjustment"`
	SkipBudgetMatching    bool                   `json:"skipBudgetMatching"`
	ShowStatus            string                 `json:"showStatus"`
	CreatedAt             string                 `json:"createdAt,omitempty"`
	OccurredAt            string                 `json:"occurredAt,omitempty"`
	UpdatedAt             string                 `json:"updatedAt,omitempty"`
	DeletedAt             string                 `json:"-"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`

	RelatedBudgetID  *string `json:"relatedBudgetId,omitempty"`
	RelatedDebtID    *string `json:"relatedDebtId,omitempty"`
	GoalName         *string `json:"goalName,omitempty"`
	GoalType         *string `json:"goalType,omitempty"`
	PlannedAmount    float64 `json:"plannedAmount"`
	PaidAmount       float64 `json:"paidAmount"`
	OriginalCurrency *string `json:"originalCurrency,omitempty"`
	OriginalAmount   float64 `json:"originalAmount"`
	ConversionRate   float64 `json:"conversionRate"`
}

// Budget tracks spending goals.
type Budget struct {
	ID                string   `json:"id"`
	UserID            string   `json:"userId"`
	Name              string   `json:"name"`
	BudgetType        string   `json:"budgetType"`
	CategoryIDs       []string `json:"categoryIds"`
	LinkedGoalID      *string  `json:"linkedGoalId,omitempty"`
	AccountID         *string  `json:"accountId,omitempty"`
	TransactionType   *string  `json:"transactionType,omitempty"`
	Currency          string   `json:"currency"`
	LimitAmount       float64  `json:"limitAmount"`
	PeriodType        string   `json:"periodType"`
	StartDate         *string  `json:"startDate,omitempty"`
	EndDate           *string  `json:"endDate,omitempty"`
	SpentAmount       float64  `json:"spentAmount"`
	RemainingAmount   float64  `json:"remainingAmount"`
	PercentUsed       float64  `json:"percentUsed"`
	IsOverspent       bool     `json:"isOverspent"`
	RolloverMode      string   `json:"rolloverMode"`
	NotifyOnExceed    bool     `json:"notifyOnExceed"`
	ContributionTotal float64  `json:"contributionTotal"`
	CurrentBalance    float64  `json:"currentBalance"`
	IsArchived        bool     `json:"isArchived"`
	ShowStatus        string   `json:"showStatus"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	UpdatedAt         string   `json:"updatedAt,omitempty"`
	DeletedAt         string   `json:"-"`
}

// Debt represents an owed balance.
type Debt struct {
	ID                           string  `json:"id"`
	UserID                       string  `json:"userId"`
	Direction                    string  `json:"direction"`
	CounterpartyID               *string `json:"counterpartyId,omitempty"`
	CounterpartyName             string  `json:"counterpartyName"`
	Description                  *string `json:"description,omitempty"`
	PrincipalAmount              float64 `json:"principalAmount"`
	PrincipalCurrency            string  `json:"principalCurrency"`
	PrincipalOriginalAmount      float64 `json:"principalOriginalAmount"`
	PrincipalOriginalCurrency    *string `json:"principalOriginalCurrency,omitempty"`
	BaseCurrency                 string  `json:"baseCurrency"`
	RateOnStart                  float64 `json:"rateOnStart"`
	PrincipalBaseValue           float64 `json:"principalBaseValue"`
	RepaymentCurrency            *string `json:"repaymentCurrency,omitempty"`
	RepaymentAmount              float64 `json:"repaymentAmount"`
	RepaymentRateOnStart         float64 `json:"repaymentRateOnStart"`
	IsFixedRepaymentAmount       bool    `json:"isFixedRepaymentAmount"`
	StartDate                    string  `json:"startDate"`
	DueDate                      *string `json:"dueDate,omitempty"`
	InterestMode                 *string `json:"interestMode,omitempty"`
	InterestRateAnnual           float64 `json:"interestRateAnnual"`
	ScheduleHint                 *string `json:"scheduleHint,omitempty"`
	LinkedGoalID                 *string `json:"linkedGoalId,omitempty"`
	LinkedBudgetID               *string `json:"linkedBudgetId,omitempty"`
	FundingAccountID             *string `json:"fundingAccountId,omitempty"`
	FundingTransactionID         *string `json:"fundingTransactionId,omitempty"`
	LentFromAccountID            *string `json:"lentFromAccountId,omitempty"`
	ReturnToAccountID            *string `json:"returnToAccountId,omitempty"`
	ReceivedToAccountID          *string `json:"receivedToAccountId,omitempty"`
	PayFromAccountID             *string `json:"payFromAccountId,omitempty"`
	CustomRateUsed               float64 `json:"customRateUsed"`
	ExchangeRateCurrent          float64 `json:"exchangeRateCurrent"`
	ReminderEnabled              bool    `json:"reminderEnabled"`
	ReminderTime                 *string `json:"reminderTime,omitempty"`
	Status                       string  `json:"status"`
	SettledAt                    *string `json:"settledAt,omitempty"`
	FinalRateUsed                float64 `json:"finalRateUsed"`
	FinalProfitLoss              float64 `json:"finalProfitLoss"`
	FinalProfitLossCurrency      *string `json:"finalProfitLossCurrency,omitempty"`
	TotalPaidInRepaymentCurrency float64 `json:"totalPaidInRepaymentCurrency"`
	RemainingAmount              float64 `json:"remainingAmount"`
	TotalPaid                    float64 `json:"totalPaid"`
	PercentPaid                  float64 `json:"percentPaid"`
	ShowStatus                   string  `json:"showStatus"`
	CreatedAt                    string  `json:"createdAt,omitempty"`
	UpdatedAt                    string  `json:"updatedAt,omitempty"`
	DeletedAt                    string  `json:"-"`
}

// DebtResponse wraps Debt with embedded counterparty for API responses.
type DebtResponse struct {
	*Debt
	Counterparty *CounterpartyEmbed `json:"counterparty,omitempty"`
}

// CounterpartyEmbed is a lightweight counterparty object for embedding in responses.
type CounterpartyEmbed struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// CreateDebtRequest supports multiple ways to specify counterparty.
type CreateDebtRequest struct {
	Debt
	// Option B: Inline counterparty creation
	InlineCounterparty *InlineCounterparty `json:"counterparty,omitempty"`
}

// InlineCounterparty is used when creating a debt with a new counterparty inline.
type InlineCounterparty struct {
	DisplayName string  `json:"displayName"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// DebtPayment represents a payment against a debt.
type DebtPayment struct {
	ID                    string  `json:"id"`
	DebtID                string  `json:"debtId"`
	Amount                float64 `json:"amount"`
	Currency              string  `json:"currency"`
	BaseCurrency          string  `json:"baseCurrency"`
	RateUsedToBase        float64 `json:"rateUsedToBase"`
	ConvertedAmountToBase float64 `json:"convertedAmountToBase"`
	RateUsedToDebt        float64 `json:"rateUsedToDebt"`
	ConvertedAmountToDebt float64 `json:"convertedAmountToDebt"`
	PaymentDate           string  `json:"paymentDate"`
	AccountID             *string `json:"accountId,omitempty"`
	Note                  *string `json:"note,omitempty"`
	RelatedTransactionID  *string `json:"relatedTransactionId,omitempty"`
	AppliedRate           float64 `json:"appliedRate"`
	CreatedAt             string  `json:"createdAt,omitempty"`
	UpdatedAt             string  `json:"updatedAt,omitempty"`
	DeletedAt             string  `json:"-"`
	TransactionType       string  `json:"-"`
}

// Counterparty represents a person or organization.
type Counterparty struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	DisplayName    string  `json:"displayName"`
	PhoneNumber    *string `json:"phoneNumber,omitempty"`
	Comment        *string `json:"comment,omitempty"`
	SearchKeywords *string `json:"searchKeywords,omitempty"`
	ShowStatus     string  `json:"showStatus"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	UpdatedAt      string  `json:"updatedAt,omitempty"`
	DeletedAt      string  `json:"-"`
}

// FXRate represents a stored exchange rate.
type FXRate struct {
	ID            string  `json:"id"`
	Date          string  `json:"date"`
	FromCurrency  string  `json:"fromCurrency"`
	ToCurrency    string  `json:"toCurrency"`
	Rate          float64 `json:"rate"`
	RateMid       float64 `json:"rateMid"`
	RateBid       float64 `json:"rateBid"`
	RateAsk       float64 `json:"rateAsk"`
	Nominal       float64 `json:"nominal"`
	SpreadPercent float64 `json:"spreadPercent"`
	Source        string  `json:"source"`
	CreatedAt     string  `json:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
}

// BudgetSpendingItem captures spending by category.
type BudgetSpendingItem struct {
	CategoryID   string  `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

// CurrencyBalance summarizes totals per currency.
type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

// SupportedCurrency describes display metadata for a currency.
type SupportedCurrency struct {
	Code     string `json:"code"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
}

// FinanceCategory defines admin-driven categories.
type FinanceCategory struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	NameI18n  map[string]string `json:"nameI18n"`
	IconName  string            `json:"iconName"`
	Color     *string           `json:"color,omitempty"`
	IsDefault bool              `json:"isDefault"`
	SortOrder int               `json:"sortOrder"`
	IsActive  bool              `json:"isActive"`
	CreatedAt string            `json:"createdAt,omitempty"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

// QuickExpenseCategory stores user-selected quick categories.
type QuickExpenseCategory struct {
	Tag  string `json:"tag"`
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

// FinanceSummary holds aggregated balances and totals.
type FinanceSummaryTotals struct {
	Balance float64 `json:"balance"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

type FinanceSummaryPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type FinanceSummaryAccount struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Balance  float64 `json:"balance"`
	BalanceBase float64 `json:"balanceBase"`
	Currency string  `json:"currency"`
}

type FinanceSummaryCategory struct {
	CategoryID string  `json:"categoryId"`
	Amount     float64 `json:"amount"`
}

type FinanceSummaryChanges struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type FinanceSummaryProgress struct {
	Used       float64 `json:"used"`
	Percentage float64 `json:"percentage"`
	Limit      float64 `json:"limit"`
}

type FinanceSummaryTransaction struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Date          string  `json:"date"`
	Description   string  `json:"description,omitempty"`
	CategoryID    *string `json:"categoryId,omitempty"`
	AccountID     *string `json:"accountId,omitempty"`
	FromAccountID *string `json:"fromAccountId,omitempty"`
	ToAccountID   *string `json:"toAccountId,omitempty"`
}

type FinanceSummaryEvent struct {
	ID          string `json:"id"`
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        string `json:"time"`
}

type FinanceSummary struct {
	Period        FinanceSummaryPeriod     `json:"period"`
	BaseCurrency  string                   `json:"baseCurrency"`
	Totals        FinanceSummaryTotals     `json:"totals"`
	Changes       FinanceSummaryChanges    `json:"changes"`
	ByCurrency    []CurrencyBalance        `json:"byCurrency"`
	Accounts      []FinanceSummaryAccount  `json:"accounts"`
	TopCategories []FinanceSummaryCategory `json:"topCategories"`
	Progress      FinanceSummaryProgress   `json:"progress"`
	RecentTransactions []FinanceSummaryTransaction `json:"recentTransactions"`
	Events        []FinanceSummaryEvent    `json:"events"`
}

// FinanceBootstrap is returned when initializing finance flows.
type FinanceBootstrap struct {
	HasAccounts bool           `json:"hasAccounts"`
	Accounts    []*Account     `json:"accounts"`
	Summary     FinanceSummary `json:"summary"`
}

// BalanceHistoryPoint captures balance over time.
type BalanceHistoryPoint struct {
	Date    string  `json:"date"`
	Balance float64 `json:"balance"`
}

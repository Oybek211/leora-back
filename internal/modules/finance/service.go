package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	appErrors "github.com/leora/leora-server/internal/errors"
)

// TransactionFilter captures filtering options for list endpoints.
type TransactionFilter struct {
	AccountID  string
	Type       string
	CategoryID string
	DateFrom   string
	DateTo     string
	GoalID     string
	BudgetID   string
	DebtID     string
}

// BudgetFilter captures budget list filters.
type BudgetFilter struct {
	PeriodType   string
	IsArchived   *bool
	LinkedGoalID string
}

// DebtFilter captures debt list filters.
type DebtFilter struct {
	Direction    string
	Status       string
	LinkedGoalID string
}

// CounterpartyFilter captures search options.
type CounterpartyFilter struct {
	Search string
}

// Service orchestrates finance use cases.
type Service struct {
	repo  Repository
	cache *redis.Client
}

const financeSummaryCacheTTL = 45 * time.Second

func NewService(repo Repository, cache *redis.Client) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Accounts(ctx context.Context) ([]*Account, error) {
	accounts, err := s.repo.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		account.ShowStatus = normalizeShowStatus(account.ShowStatus)
		account.IsArchived = account.ShowStatus == "archived"
	}
	return accounts, nil
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	account.ShowStatus = normalizeShowStatus(account.ShowStatus)
	account.IsArchived = account.ShowStatus == "archived"
	return account, nil
}

func (s *Service) CreateAccount(ctx context.Context, account *Account) (*Account, *Transaction, error) {
	normalizeAccount(account)
	account.CurrentBalance = account.InitialBalance
	openingTxn, err := s.repo.CreateAccount(ctx, account)
	if err != nil {
		return nil, nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	return account, openingTxn, nil
}

func (s *Service) UpdateAccount(ctx context.Context, id string, account *Account) (*Account, error) {
	account.ID = id
	normalizeAccount(account)
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	updated, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	updated.ShowStatus = normalizeShowStatus(updated.ShowStatus)
	updated.IsArchived = updated.ShowStatus == "archived"
	return updated, nil
}

func (s *Service) PatchAccount(ctx context.Context, id string, fields map[string]interface{}) (*Account, error) {
	current, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyAccountPatch(current, fields)
	normalizeAccount(current)
	if err := s.repo.UpdateAccount(ctx, current); err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	updated, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	updated.ShowStatus = normalizeShowStatus(updated.ShowStatus)
	updated.IsArchived = updated.ShowStatus == "archived"
	return updated, nil
}

func (s *Service) DeleteAccount(ctx context.Context, id string) (*Transaction, error) {
	txn, err := s.repo.DeleteAccount(ctx, id)
	if err != nil {
		log.Printf("[Service.DeleteAccount] Error for id=%s: %v", id, err)
	}
	s.invalidateFinanceSummaryCache(ctx)
	return txn, err
}

func (s *Service) Transactions(ctx context.Context, filter TransactionFilter) ([]*Transaction, error) {
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	transactions = filterTransactions(transactions, filter)
	for _, txn := range transactions {
		normalizeTransaction(txn)
	}
	return transactions, nil
}

func (s *Service) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	txn, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	normalizeTransaction(txn)
	return txn, nil
}

func (s *Service) CreateTransaction(ctx context.Context, txn *Transaction) (*Transaction, error) {
	normalizeTransaction(txn)
	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	return txn, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, id string, txn *Transaction) (*Transaction, error) {
	return nil, appErrors.TransactionImmutable
}

func (s *Service) PatchTransaction(ctx context.Context, id string, fields map[string]interface{}) (*Transaction, error) {
	return nil, appErrors.TransactionImmutable
}

func (s *Service) DeleteTransaction(ctx context.Context, id string) error {
	return appErrors.TransactionImmutable
}

func (s *Service) FinanceSummary(ctx context.Context, dateFrom, dateTo, baseCurrency string, accountIDs []string) (*FinanceSummary, error) {
	if cached := s.getFinanceSummaryCache(ctx, dateFrom, dateTo, baseCurrency, accountIDs); cached != nil {
		return cached, nil
	}
	accounts, err := s.repo.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	baseCurrency = normalizeSummaryBaseCurrency(baseCurrency, accounts)
	rateDate := resolveSummaryRateDate(dateFrom, dateTo)

	accountCurrencyMap := make(map[string]string, len(accounts))
	for _, account := range accounts {
		if account != nil {
			accountCurrencyMap[account.ID] = account.Currency
		}
	}

	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	filtered := filterTransactions(transactions, TransactionFilter{DateFrom: dateFrom, DateTo: dateTo})

	accountFilter := make(map[string]bool)
	if len(accountIDs) > 0 {
		for _, id := range accountIDs {
			if id != "" {
				accountFilter[id] = true
			}
		}
	}
	if len(accountFilter) > 0 {
		accounts = filterAccountsByIDs(accounts, accountFilter)
		filtered = filterTransactionsByAccount(filtered, accountFilter)
	}

	totalBalance := 0.0
	byCurrency := make(map[string]float64)
	accountsSummary := make([]FinanceSummaryAccount, 0, len(accounts))
	for _, account := range accounts {
		if account.Currency != "" {
			byCurrency[account.Currency] += account.CurrentBalance
		}
		baseBalance := convertToSummaryBase(s, ctx, account.CurrentBalance, account.Currency, baseCurrency, rateDate)
		totalBalance += baseBalance
		accountsSummary = append(accountsSummary, FinanceSummaryAccount{
			ID:          account.ID,
			Name:        account.Name,
			Balance:     account.CurrentBalance,
			BalanceBase: baseBalance,
			Currency:    account.Currency,
		})
	}

	totalIncome := 0.0
	totalExpense := 0.0
	categoryTotals := map[string]float64{}
	for _, txn := range filtered {
		txnCurrency := resolveTransactionCurrency(txn, accountCurrencyMap, baseCurrency)
		txnDate := resolveTransactionDate(txn, rateDate)
		baseAmount := convertToSummaryBase(s, ctx, txn.Amount, txnCurrency, baseCurrency, txnDate)
		impact := transactionImpactForSummaryAmount(txn, baseAmount)
		if impact > 0 {
			totalIncome += impact
		}
		if impact < 0 {
			totalExpense += -impact
		}
		if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Type == TransactionTypeExpense {
			categoryTotals[*txn.CategoryID] += baseAmount
		}
	}

	currencyTotals := make([]CurrencyBalance, 0, len(byCurrency))
	for currency, total := range byCurrency {
		currencyTotals = append(currencyTotals, CurrencyBalance{
			Currency: currency,
			Balance:  total,
		})
	}
	sort.Slice(currencyTotals, func(i, j int) bool {
		return currencyTotals[i].Currency < currencyTotals[j].Currency
	})

	topCategories := make([]FinanceSummaryCategory, 0, len(categoryTotals))
	for categoryID, amount := range categoryTotals {
		topCategories = append(topCategories, FinanceSummaryCategory{
			CategoryID: categoryID,
			Amount:     amount,
		})
	}
	sort.Slice(topCategories, func(i, j int) bool {
		return topCategories[i].Amount > topCategories[j].Amount
	})

	changes := FinanceSummaryChanges{}
	if prevFrom, prevTo, ok := previousSummaryPeriod(dateFrom, dateTo); ok {
		prevFiltered := filterTransactions(transactions, TransactionFilter{DateFrom: prevFrom, DateTo: prevTo})
		if len(accountFilter) > 0 {
			prevFiltered = filterTransactionsByAccount(prevFiltered, accountFilter)
		}
		prevIncome := 0.0
		prevExpense := 0.0
		for _, txn := range prevFiltered {
			txnCurrency := resolveTransactionCurrency(txn, accountCurrencyMap, baseCurrency)
			txnDate := resolveTransactionDate(txn, rateDate)
			baseAmount := convertToSummaryBase(s, ctx, txn.Amount, txnCurrency, baseCurrency, txnDate)
			impact := transactionImpactForSummaryAmount(txn, baseAmount)
			if impact > 0 {
				prevIncome += impact
			}
			if impact < 0 {
				prevExpense += -impact
			}
		}
		changes.Income = percentChange(totalIncome, prevIncome)
		changes.Expense = percentChange(totalExpense, prevExpense)
	}

	progress := buildSummaryProgress(s, ctx, accounts, dateFrom, dateTo, baseCurrency, rateDate, accountFilter, totalExpense)
	recentTransactions := buildSummaryRecentTransactions(s, ctx, filtered, accountCurrencyMap, baseCurrency, rateDate)
	events := buildSummaryEvents(s, ctx, accountFilter, dateFrom, dateTo, baseCurrency, rateDate)

	summary := &FinanceSummary{
		Period: FinanceSummaryPeriod{
			From: dateFrom,
			To:   dateTo,
		},
		BaseCurrency: baseCurrency,
		Totals: FinanceSummaryTotals{
			Balance: totalBalance,
			Income:  totalIncome,
			Expense: totalExpense,
			Net:     totalIncome - totalExpense,
		},
		Changes:       changes,
		ByCurrency:    currencyTotals,
		Accounts:      accountsSummary,
		TopCategories: topCategories,
		Progress:      progress,
		RecentTransactions: recentTransactions,
		Events:        events,
	}
	s.setFinanceSummaryCache(ctx, dateFrom, dateTo, baseCurrency, accountIDs, summary)
	return summary, nil
}

func (s *Service) FinanceBootstrap(ctx context.Context, dateFrom, dateTo, baseCurrency string, accountIDs []string) (*FinanceBootstrap, error) {
	accounts, err := s.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	summary, err := s.FinanceSummary(ctx, dateFrom, dateTo, baseCurrency, accountIDs)
	if err != nil {
		return nil, err
	}
	return &FinanceBootstrap{
		HasAccounts: len(accounts) > 0,
		Accounts:    accounts,
		Summary:     *summary,
	}, nil
}

func normalizeSummaryBaseCurrency(baseCurrency string, accounts []*Account) string {
	trimmed := strings.TrimSpace(baseCurrency)
	if trimmed != "" {
		return strings.ToUpper(trimmed)
	}
	for _, account := range accounts {
		if account != nil && strings.TrimSpace(account.Currency) != "" {
			return strings.ToUpper(account.Currency)
		}
	}
	return "USD"
}

func resolveSummaryRateDate(dateFrom, dateTo string) string {
	if strings.TrimSpace(dateTo) != "" {
		return dateTo
	}
	if strings.TrimSpace(dateFrom) != "" {
		return dateFrom
	}
	return time.Now().UTC().Format("2006-01-02")
}

func convertToSummaryBase(s *Service, ctx context.Context, amount float64, fromCurrency, baseCurrency, dateValue string) float64 {
	if strings.TrimSpace(fromCurrency) == "" || strings.EqualFold(fromCurrency, baseCurrency) {
		return amount
	}
	rate, err := s.resolveFXRate(ctx, fromCurrency, baseCurrency, dateValue)
	if err != nil || rate == 0 {
		return amount
	}
	return amount * rate
}

func resolveTransactionCurrency(txn *Transaction, accountCurrencyMap map[string]string, fallbackCurrency string) string {
	if txn == nil {
		return fallbackCurrency
	}
	if strings.TrimSpace(txn.Currency) != "" {
		return txn.Currency
	}
	if txn.AccountID != nil {
		if currency := accountCurrencyMap[*txn.AccountID]; strings.TrimSpace(currency) != "" {
			return currency
		}
	}
	if txn.FromAccountID != nil {
		if currency := accountCurrencyMap[*txn.FromAccountID]; strings.TrimSpace(currency) != "" {
			return currency
		}
	}
	if txn.ToAccountID != nil {
		if currency := accountCurrencyMap[*txn.ToAccountID]; strings.TrimSpace(currency) != "" {
			return currency
		}
	}
	return fallbackCurrency
}

func resolveTransactionDate(txn *Transaction, fallback string) string {
	if txn == nil {
		return fallback
	}
	if strings.TrimSpace(txn.Date) != "" {
		return txn.Date
	}
	return fallback
}

func transactionImpactForSummaryAmount(txn *Transaction, amount float64) float64 {
	if txn == nil {
		return 0
	}
	switch txn.Type {
	case TransactionTypeIncome:
		return amount
	case TransactionTypeExpense:
		return -amount
	case TransactionTypeSystemAdjustment, TransactionTypeDebtCreate, TransactionTypeDebtPayment, TransactionTypeDebtAdjustment, TransactionTypeDebtFullPayment, TransactionTypeBudgetAddValue, TransactionTypeDebtAddValue:
		return amount
	case TransactionTypeAccountCreateFunding:
		return amount
	case TransactionTypeAccountDeleteWithdrawal:
		return -amount
	default:
		return 0
	}
}

func previousSummaryPeriod(dateFrom, dateTo string) (string, string, bool) {
	if strings.TrimSpace(dateFrom) == "" || strings.TrimSpace(dateTo) == "" {
		return "", "", false
	}
	start, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return "", "", false
	}
	end, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return "", "", false
	}
	prevStart := start.AddDate(0, -1, 0)
	prevEnd := end.AddDate(0, -1, 0)
	return prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02"), true
}

func percentChange(current, prev float64) float64 {
	if prev == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	value := ((current - prev) / prev) * 100
	return math.Round(value*10) / 10
}

func filterAccountsByIDs(accounts []*Account, accountFilter map[string]bool) []*Account {
	filteredAccounts := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		if accountFilter[account.ID] {
			filteredAccounts = append(filteredAccounts, account)
		}
	}
	return filteredAccounts
}

func filterTransactionsByAccount(transactions []*Transaction, accountFilter map[string]bool) []*Transaction {
	filtered := make([]*Transaction, 0, len(transactions))
	for _, txn := range transactions {
		if txn.AccountID != nil && accountFilter[*txn.AccountID] {
			filtered = append(filtered, txn)
			continue
		}
		if txn.FromAccountID != nil && accountFilter[*txn.FromAccountID] {
			filtered = append(filtered, txn)
			continue
		}
		if txn.ToAccountID != nil && accountFilter[*txn.ToAccountID] {
			filtered = append(filtered, txn)
			continue
		}
	}
	return filtered
}

func transactionSortTime(txn *Transaction) time.Time {
	if txn == nil {
		return time.Time{}
	}
	if strings.TrimSpace(txn.Date) != "" {
		if parsed, err := time.Parse("2006-01-02", txn.Date); err == nil {
			return parsed
		}
	}
	if strings.TrimSpace(txn.CreatedAt) != "" {
		if parsed, err := time.Parse(time.RFC3339, txn.CreatedAt); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func buildSummaryRecentTransactions(
	s *Service,
	ctx context.Context,
	filtered []*Transaction,
	accountCurrencyMap map[string]string,
	baseCurrency string,
	rateDate string,
) []FinanceSummaryTransaction {
	if len(filtered) == 0 {
		return nil
	}
	sort.Slice(filtered, func(i, j int) bool {
		return transactionSortTime(filtered[i]).After(transactionSortTime(filtered[j]))
	})
	recent := make([]FinanceSummaryTransaction, 0, 5)
	for _, txn := range filtered {
		if txn == nil {
			continue
		}
		if txn.Type == TransactionTypeTransfer || txn.Type == TransactionTypeTransferIn || txn.Type == TransactionTypeTransferOut {
			continue
		}
		txnCurrency := resolveTransactionCurrency(txn, accountCurrencyMap, baseCurrency)
		txnDate := resolveTransactionDate(txn, rateDate)
		baseAmount := convertToSummaryBase(s, ctx, txn.Amount, txnCurrency, baseCurrency, txnDate)
		description := ""
		if txn.Description != nil && strings.TrimSpace(*txn.Description) != "" {
			description = *txn.Description
		} else if txn.Name != nil {
			description = *txn.Name
		}
		recent = append(recent, FinanceSummaryTransaction{
			ID:            txn.ID,
			Type:          txn.Type,
			Amount:        baseAmount,
			Currency:      baseCurrency,
			Date:          txn.Date,
			Description:   description,
			CategoryID:    txn.CategoryID,
			AccountID:     txn.AccountID,
			FromAccountID: txn.FromAccountID,
			ToAccountID:   txn.ToAccountID,
		})
		if len(recent) >= 5 {
			break
		}
	}
	return recent
}

func buildSummaryProgress(
	s *Service,
	ctx context.Context,
	accounts []*Account,
	dateFrom, dateTo, baseCurrency, rateDate string,
	accountFilter map[string]bool,
	fallbackExpense float64,
) FinanceSummaryProgress {
	budgets, err := s.repo.ListBudgets(ctx)
	if err != nil {
		return FinanceSummaryProgress{Used: fallbackExpense}
	}
	totalLimit := 0.0
	totalSpent := 0.0
	for _, budget := range budgets {
		if budget == nil {
			continue
		}
		if budget.AccountID != nil && len(accountFilter) > 0 && !accountFilter[*budget.AccountID] {
			continue
		}
		if !budgetMatchesRange(budget, dateFrom, dateTo) {
			continue
		}
		limitBase := convertToSummaryBase(s, ctx, budget.LimitAmount, budget.Currency, baseCurrency, rateDate)
		spentBase := convertToSummaryBase(s, ctx, budget.SpentAmount, budget.Currency, baseCurrency, rateDate)
		totalLimit += limitBase
		totalSpent += spentBase
	}
	used := totalSpent
	if used == 0 {
		used = fallbackExpense
	}
	percentage := 0.0
	if totalLimit > 0 {
		percentage = math.Round((used/totalLimit)*100)
		if percentage > 125 {
			percentage = 125
		}
	}
	return FinanceSummaryProgress{
		Used:       used,
		Percentage: percentage,
		Limit:      totalLimit,
	}
}

func budgetMatchesRange(budget *Budget, dateFrom, dateTo string) bool {
	if budget == nil {
		return false
	}
	if strings.TrimSpace(dateFrom) == "" && strings.TrimSpace(dateTo) == "" {
		return true
	}
	rangeStart, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return true
	}
	rangeEnd, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return true
	}
	budgetStart := rangeStart
	if budget.StartDate != nil && strings.TrimSpace(*budget.StartDate) != "" {
		if parsed, err := time.Parse("2006-01-02", *budget.StartDate); err == nil {
			budgetStart = parsed
		}
	}
	budgetEnd := rangeEnd
	if budget.EndDate != nil && strings.TrimSpace(*budget.EndDate) != "" {
		if parsed, err := time.Parse("2006-01-02", *budget.EndDate); err == nil {
			budgetEnd = parsed
		}
	}
	if budgetStart.After(rangeEnd) {
		return false
	}
	if budgetEnd.Before(rangeStart) {
		return false
	}
	return true
}

func buildSummaryEvents(
	s *Service,
	ctx context.Context,
	accountFilter map[string]bool,
	dateFrom, dateTo, baseCurrency, rateDate string,
) []FinanceSummaryEvent {
	events := make([]FinanceSummaryEvent, 0, 3)
	debts, err := s.repo.ListDebts(ctx)
	if err == nil && len(debts) > 0 {
		filteredDebts := make([]*Debt, 0, len(debts))
		for _, debt := range debts {
			if debt == nil {
				continue
			}
			if debt.Status == "paid" {
				continue
			}
			if debt.FundingAccountID != nil && len(accountFilter) > 0 && !accountFilter[*debt.FundingAccountID] {
				continue
			}
			filteredDebts = append(filteredDebts, debt)
		}
		sort.Slice(filteredDebts, func(i, j int) bool {
			return debtSortTime(filteredDebts[i]).Before(debtSortTime(filteredDebts[j]))
		})
		for _, debt := range filteredDebts {
			if debt == nil {
				continue
			}
			amountBase := convertToSummaryBase(s, ctx, debt.PrincipalAmount, debt.PrincipalCurrency, baseCurrency, rateDate)
			amountLabel := formatAmountForCurrency(amountBase, baseCurrency)
			description := amountLabel
			if debt.Description != nil && strings.TrimSpace(*debt.Description) != "" {
				description = fmt.Sprintf("%s • %s", amountLabel, strings.TrimSpace(*debt.Description))
			}
			title := ""
			icon := "clock"
			if debt.Direction == "they_owe_me" {
				title = fmt.Sprintf("%s owes you", debt.CounterpartyName)
				icon = "wallet"
			} else {
				title = fmt.Sprintf("You owe %s", debt.CounterpartyName)
				icon = "alert"
			}
			timeLabel := describeDueDate(debt.DueDate)
			events = append(events, FinanceSummaryEvent{
				ID:          debt.ID,
				Icon:        icon,
				Title:       title,
				Description: description,
				Time:        timeLabel,
			})
			if len(events) >= 3 {
				return events
			}
		}
	}

	if len(events) > 0 {
		return events
	}

	budgets, err := s.repo.ListBudgets(ctx)
	if err != nil {
		return events
	}
	for _, budget := range budgets {
		if budget == nil {
			continue
		}
		if budget.AccountID != nil && len(accountFilter) > 0 && !accountFilter[*budget.AccountID] {
			continue
		}
		if !budgetMatchesRange(budget, dateFrom, dateTo) {
			continue
		}
		if budget.LimitAmount <= 0 {
			continue
		}
		limitBase := convertToSummaryBase(s, ctx, budget.LimitAmount, budget.Currency, baseCurrency, rateDate)
		spentBase := convertToSummaryBase(s, ctx, budget.SpentAmount, budget.Currency, baseCurrency, rateDate)
		if spentBase <= limitBase {
			continue
		}
		spentLabel := formatAmountForCurrency(spentBase, baseCurrency)
		limitLabel := formatAmountForCurrency(limitBase, baseCurrency)
		events = append(events, FinanceSummaryEvent{
			ID:          budget.ID,
			Icon:        "alert",
			Title:       budget.Name,
			Description: fmt.Sprintf("%s / %s", spentLabel, limitLabel),
			Time:        "Limit exceeded",
		})
		if len(events) >= 3 {
			break
		}
	}
	return events
}

func debtSortTime(debt *Debt) time.Time {
	if debt == nil {
		return time.Time{}
	}
	if debt.DueDate != nil && strings.TrimSpace(*debt.DueDate) != "" {
		if parsed, err := time.Parse("2006-01-02", *debt.DueDate); err == nil {
			return parsed
		}
	}
	if strings.TrimSpace(debt.StartDate) != "" {
		if parsed, err := time.Parse("2006-01-02", debt.StartDate); err == nil {
			return parsed
		}
	}
	return time.Now().UTC()
}

func describeDueDate(dueDate *string) string {
	if dueDate == nil || strings.TrimSpace(*dueDate) == "" {
		return "No period"
	}
	target, err := time.Parse("2006-01-02", *dueDate)
	if err != nil {
		return "No period"
	}
	now := time.Now()
	diffDays := int(math.Ceil(target.Sub(now).Hours() / 24))
	if diffDays == 0 {
		return "Due today"
	}
	if diffDays > 0 {
		if diffDays == 1 {
			return "In 1 day"
		}
		return fmt.Sprintf("In %d days", diffDays)
	}
	overdue := int(math.Abs(float64(diffDays)))
	if overdue == 1 {
		return "1 day overdue"
	}
	return fmt.Sprintf("%d days overdue", overdue)
}

func formatAmountForCurrency(amount float64, currency string) string {
	rounded := roundAmountForCurrency(amount, currency)
	decimals := 2
	if strings.EqualFold(currency, "UZS") {
		decimals = 0
	}
	format := "%." + strconv.Itoa(decimals) + "f %s"
	return fmt.Sprintf(format, rounded, strings.ToUpper(currency))
}

func (s *Service) Categories(ctx context.Context, categoryType string, activeOnly bool) ([]*FinanceCategory, error) {
	all, err := s.repo.ListCategories(ctx, "", false)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		for _, category := range defaultFinanceCategories() {
			_ = s.repo.CreateCategory(ctx, category)
		}
	}
	return s.repo.ListCategories(ctx, categoryType, activeOnly)
}

func (s *Service) QuickExpenseCategories(ctx context.Context, categoryType string) ([]*QuickExpenseCategory, error) {
	return s.repo.ListQuickExpenseCategories(ctx, categoryType)
}

func (s *Service) UpdateQuickExpenseCategories(ctx context.Context, categoryType string, categories []*QuickExpenseCategory) ([]*QuickExpenseCategory, error) {
	categoryType = strings.TrimSpace(categoryType)
	if categoryType == "" {
		return nil, appErrors.InvalidFinanceData
	}
	if categoryType != "income" && categoryType != "expense" {
		return nil, appErrors.InvalidFinanceData
	}
	for _, entry := range categories {
		if entry == nil || strings.TrimSpace(entry.Tag) == "" {
			return nil, appErrors.InvalidFinanceData
		}
		entry.Type = categoryType
	}
	if err := s.repo.ReplaceQuickExpenseCategories(ctx, categoryType, categories); err != nil {
		return nil, err
	}
	return s.repo.ListQuickExpenseCategories(ctx, categoryType)
}

func (s *Service) CreateCategory(ctx context.Context, category *FinanceCategory) (*FinanceCategory, error) {
	if category.Type != "income" && category.Type != "expense" {
		return nil, appErrors.InvalidFinanceData
	}
	if category.IconName == "" || len(category.NameI18n) == 0 {
		return nil, appErrors.InvalidFinanceData
	}
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id string, category *FinanceCategory) (*FinanceCategory, error) {
	if category.Type != "income" && category.Type != "expense" {
		return nil, appErrors.InvalidFinanceData
	}
	if category.IconName == "" || len(category.NameI18n) == 0 {
		return nil, appErrors.InvalidFinanceData
	}
	category.ID = id
	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *Service) Budgets(ctx context.Context, filter BudgetFilter) ([]*Budget, error) {
	budgets, err := s.repo.ListBudgets(ctx)
	if err != nil {
		return nil, err
	}
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	budgets = filterBudgets(budgets, filter)
	for _, budget := range budgets {
		normalizeBudget(budget)
		applyBudgetRollups(budget, transactions)
	}
	return budgets, nil
}

func (s *Service) GetBudget(ctx context.Context, id string) (*Budget, error) {
	budget, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	normalizeBudget(budget)
	applyBudgetRollups(budget, transactions)
	return budget, nil
}

func (s *Service) CreateBudget(ctx context.Context, budget *Budget) (*Budget, error) {
	normalizeBudget(budget)
	if err := s.repo.CreateBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *Service) UpdateBudget(ctx context.Context, id string, budget *Budget) (*Budget, error) {
	budget.ID = id
	normalizeBudget(budget)
	if err := s.repo.UpdateBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *Service) PatchBudget(ctx context.Context, id string, fields map[string]interface{}) (*Budget, error) {
	current, err := s.repo.GetBudgetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyBudgetPatch(current, fields)
	normalizeBudget(current)
	if err := s.repo.UpdateBudget(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

type BudgetAddValueInput struct {
	AccountID      string
	Amount         float64
	AmountCurrency string
	Note           *string
	Date           *string
	CategoryID     *string
}

type BudgetAddValueResult struct {
	Budget      *Budget
	Transaction *Transaction
}

func (s *Service) AddBudgetValue(ctx context.Context, budgetID string, input BudgetAddValueInput) (*BudgetAddValueResult, error) {
	if strings.TrimSpace(budgetID) == "" {
		log.Printf("[Service.AddBudgetValue] Missing budgetID")
		return nil, appErrors.InvalidFinanceData
	}
	if input.Amount <= 0 {
		log.Printf("[Service.AddBudgetValue] Invalid amount=%.2f for budget=%s", input.Amount, budgetID)
		return nil, appErrors.InvalidAmount
	}
	if strings.TrimSpace(input.AccountID) == "" {
		log.Printf("[Service.AddBudgetValue] Missing accountID for budget=%s", budgetID)
		return nil, appErrors.InvalidFinanceData
	}

	budget, err := s.repo.GetBudgetByID(ctx, budgetID)
	if err != nil {
		log.Printf("[Service.AddBudgetValue] Failed to get budget=%s: %v", budgetID, err)
		return nil, err
	}
	normalizeBudget(budget)

	if strings.TrimSpace(input.AmountCurrency) == "" {
		return nil, appErrors.InvalidCurrency
	}
	if !strings.EqualFold(input.AmountCurrency, budget.Currency) {
		log.Printf("[Service.AddBudgetValue] Currency mismatch: input=%s, budget=%s", input.AmountCurrency, budget.Currency)
		return nil, appErrors.InvalidCurrency
	}

	account, err := s.repo.GetAccountByID(ctx, input.AccountID)
	if err != nil {
		log.Printf("[Service.AddBudgetValue] Failed to get account=%s: %v", input.AccountID, err)
		return nil, err
	}
	normalizeAccount(account)

	dateValue := normalizeDateInput("")
	if input.Date != nil && strings.TrimSpace(*input.Date) != "" {
		dateValue = normalizeDateInput(*input.Date)
	}
	if _, err := time.Parse("2006-01-02", dateValue); err != nil {
		return nil, appErrors.WithDetails(appErrors.InvalidTransactionDate, map[string]interface{}{"field": "date"})
	}

	accountCurrency := account.Currency
	budgetCurrency := budget.Currency

	budgetToAccountRate := 1.0
	debitAmount := input.Amount
	if !strings.EqualFold(accountCurrency, budgetCurrency) {
		rate, err := s.resolveFXRate(ctx, budgetCurrency, accountCurrency, dateValue)
		if err != nil {
			return nil, err
		}
		budgetToAccountRate = rate
		if budgetToAccountRate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
		debitAmount = input.Amount * budgetToAccountRate
	}

	debitAmount = roundAmountForCurrencyUp(debitAmount, accountCurrency)
	if debitAmount <= 0 {
		return nil, appErrors.InvalidAmount
	}
	if account.CurrentBalance < debitAmount {
		return nil, appErrors.WithDetails(appErrors.InsufficientFunds, map[string]interface{}{
			"required":  debitAmount,
			"available": account.CurrentBalance,
			"currency":  accountCurrency,
		})
	}

	txnType := "expense"
	if budget.TransactionType != nil && strings.TrimSpace(*budget.TransactionType) != "" {
		txnType = strings.ToLower(strings.TrimSpace(*budget.TransactionType))
	}

	var categoryID *string
	if input.CategoryID != nil && strings.TrimSpace(*input.CategoryID) != "" {
		categoryID = input.CategoryID
	} else if len(budget.CategoryIDs) > 0 {
		first := budget.CategoryIDs[0]
		categoryID = &first
	}

	amountDelta := debitAmount
	if txnType == "expense" {
		amountDelta = -debitAmount
	}

	txn := &Transaction{
		Type:                  TransactionTypeBudgetAddValue,
		AccountID:             &account.ID,
		Amount:                amountDelta,
		Currency:              accountCurrency,
		BaseCurrency:          budgetCurrency,
		RateUsedToBase:        1,
		ConvertedAmountToBase: input.Amount,
		CategoryID:            categoryID,
		BudgetID:              &budget.ID,
		Description:           input.Note,
		Date:                  dateValue,
		OriginalCurrency:      &budgetCurrency,
		OriginalAmount:        input.Amount,
		ConversionRate:        budgetToAccountRate,
	}

	createdTxn, err := s.CreateTransaction(ctx, txn)
	if err != nil {
		return nil, err
	}

	updatedBudget, err := s.GetBudget(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	return &BudgetAddValueResult{
		Budget:      updatedBudget,
		Transaction: createdTxn,
	}, nil
}

type DebtValueInput struct {
	AccountID      string
	Amount         float64
	AmountCurrency string
	Note           *string
	Date           *string
	AppliedRate    float64 // Exchange rate sent by client (for audit trail); 0 = not provided
}

type DebtValueResult struct {
	Debt        *Debt
	Payment     *DebtPayment
	Transaction *Transaction
}

func (s *Service) RepayDebt(ctx context.Context, debtID string, input DebtValueInput) (*DebtValueResult, error) {
	if strings.TrimSpace(debtID) == "" {
		log.Printf("[Service.RepayDebt] Missing debtID")
		return nil, appErrors.InvalidFinanceData
	}
	if input.Amount <= 0 {
		log.Printf("[Service.RepayDebt] Invalid amount=%.2f for debt=%s", input.Amount, debtID)
		return nil, appErrors.InvalidAmount
	}
	if strings.TrimSpace(input.AccountID) == "" {
		log.Printf("[Service.RepayDebt] Missing accountID for debt=%s", debtID)
		return nil, appErrors.InvalidFinanceData
	}

	debt, err := s.GetDebt(ctx, debtID)
	if err != nil {
		log.Printf("[Service.RepayDebt] Failed to get debt=%s: %v", debtID, err)
		return nil, err
	}

	if strings.TrimSpace(input.AmountCurrency) == "" {
		log.Printf("[Service.RepayDebt] Missing currency for debt=%s", debtID)
		return nil, appErrors.InvalidCurrency
	}

	account, err := s.repo.GetAccountByID(ctx, input.AccountID)
	if err != nil {
		log.Printf("[Service.RepayDebt] Failed to get account=%s: %v", input.AccountID, err)
		return nil, err
	}
	normalizeAccount(account)

	dateValue := normalizeDateInput("")
	if input.Date != nil && strings.TrimSpace(*input.Date) != "" {
		dateValue = normalizeDateInput(*input.Date)
	}
	if _, err := time.Parse("2006-01-02", dateValue); err != nil {
		return nil, appErrors.WithDetails(appErrors.InvalidTransactionDate, map[string]interface{}{"field": "date"})
	}

	debtCurrency := debt.PrincipalCurrency
	accountCurrency := account.Currency

	log.Printf("[Service.RepayDebt] === START === debtID=%s, inputAmount=%.2f, inputCurrency=%s, debtCurrency=%s, accountCurrency=%s, debt.ExchangeRateCurrent=%.6f, debt.RepaymentRateOnStart=%.6f, repaymentCurrency=%v",
		debtID, input.Amount, input.AmountCurrency, debtCurrency, accountCurrency, debt.ExchangeRateCurrent, debt.RepaymentRateOnStart, debt.RepaymentCurrency)

	// Payment amount in debt currency - this is what will be deducted from debt
	paymentInDebtCurrency := input.Amount
	// Debit amount - this is what will be deducted from account (in account currency)
	debitAmount := input.Amount

	// Check if input currency is valid (must be either debt currency or account currency)
	inputIsDebtCurrency := strings.EqualFold(input.AmountCurrency, debtCurrency)
	inputIsAccountCurrency := strings.EqualFold(input.AmountCurrency, accountCurrency)

	if !inputIsDebtCurrency && !inputIsAccountCurrency {
		log.Printf("[Service.RepayDebt] Currency mismatch: input=%s, debt=%s, account=%s", input.AmountCurrency, debtCurrency, accountCurrency)
		return nil, appErrors.InvalidCurrency
	}

	// Resolve the debt's stored exchange rate (principal → repayment).
	// ExchangeRateCurrent is the mutable single source of truth; falls back to RepaymentRateOnStart.
	debtStoredRate := debt.ExchangeRateCurrent
	if debtStoredRate <= 0 {
		debtStoredRate = debt.RepaymentRateOnStart
	}
	repaymentCurrency := ""
	if debt.RepaymentCurrency != nil {
		repaymentCurrency = *debt.RepaymentCurrency
	}

	// Convert amounts based on input currency
	if inputIsAccountCurrency && !strings.EqualFold(accountCurrency, debtCurrency) {
		// Input is in account currency - convert to debt currency for debt reduction
		var rate float64
		// Use debt's stored rate when account currency matches repayment currency
		if debtStoredRate > 0 && repaymentCurrency != "" && strings.EqualFold(accountCurrency, repaymentCurrency) {
			// storedRate = principal→repayment, we need repayment→principal = 1/storedRate
			rate = 1.0 / debtStoredRate
			log.Printf("[Service.RepayDebt] Using debt stored rate (inverse): %.6f for %s->%s", rate, accountCurrency, debtCurrency)
		} else {
			var err error
			rate, err = s.resolveFXRate(ctx, accountCurrency, debtCurrency, dateValue)
			if err != nil {
				return nil, err
			}
		}
		if rate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
		paymentInDebtCurrency = input.Amount * rate
		debitAmount = input.Amount // Account will be debited in its own currency
		log.Printf("[Service.RepayDebt] Converting from account currency: %.2f %s -> %.2f %s (rate=%.6f)",
			input.Amount, accountCurrency, paymentInDebtCurrency, debtCurrency, rate)
	} else if inputIsDebtCurrency && !strings.EqualFold(debtCurrency, accountCurrency) {
		// Input is in debt currency - convert to account currency for account debit
		var rate float64
		// Use debt's stored rate when account currency matches repayment currency
		if debtStoredRate > 0 && repaymentCurrency != "" && strings.EqualFold(accountCurrency, repaymentCurrency) {
			// storedRate = principal→repayment = debtCurrency→accountCurrency
			rate = debtStoredRate
			log.Printf("[Service.RepayDebt] Using debt stored rate: %.6f for %s->%s", rate, debtCurrency, accountCurrency)
		} else {
			var err error
			rate, err = s.resolveFXRate(ctx, debtCurrency, accountCurrency, dateValue)
			if err != nil {
				return nil, err
			}
		}
		if rate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
		debitAmount = input.Amount * rate
		paymentInDebtCurrency = input.Amount // Debt will be reduced by input amount
		log.Printf("[Service.RepayDebt] Converting from debt currency: %.2f %s -> %.2f %s (rate=%.6f)",
			input.Amount, debtCurrency, debitAmount, accountCurrency, rate)
	}
	// If currencies are same, both amounts remain as input.Amount

	log.Printf("[Service.RepayDebt] BEFORE round: debitAmount=%.6f, paymentInDebtCurrency=%.6f", debitAmount, paymentInDebtCurrency)
	debitAmount = roundAmountForCurrencyUp(debitAmount, accountCurrency)
	log.Printf("[Service.RepayDebt] AFTER round: debitAmount=%.6f", debitAmount)
	if debitAmount <= 0 {
		return nil, appErrors.InvalidAmount
	}
	if debt.Direction == "i_owe" && account.CurrentBalance < debitAmount {
		return nil, appErrors.WithDetails(appErrors.InsufficientFunds, map[string]interface{}{
			"required":  debitAmount,
			"available": account.CurrentBalance,
			"currency":  accountCurrency,
		})
	}

	accountToDebtRate := 1.0
	if !strings.EqualFold(accountCurrency, debtCurrency) {
		// Use debt's stored rate when account currency matches repayment currency
		if debtStoredRate > 0 && repaymentCurrency != "" && strings.EqualFold(accountCurrency, repaymentCurrency) {
			// storedRate = principal→repayment, we need repayment→principal = 1/storedRate
			accountToDebtRate = 1.0 / debtStoredRate
			log.Printf("[Service.RepayDebt] accountToDebtRate from debt stored rate (inverse): %.6f", accountToDebtRate)
		} else {
			rate, err := s.resolveFXRate(ctx, accountCurrency, debtCurrency, dateValue)
			if err != nil {
				return nil, err
			}
			accountToDebtRate = rate
		}
		if accountToDebtRate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
	}

	debtToBaseRate := 1.0
	if debt.BaseCurrency != "" && !strings.EqualFold(debt.BaseCurrency, debtCurrency) {
		rate, err := s.resolveFXRate(ctx, debtCurrency, debt.BaseCurrency, dateValue)
		if err != nil {
			return nil, err
		}
		debtToBaseRate = rate
		if debtToBaseRate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
	}

	remaining := debt.RemainingAmount
	if remaining <= 0 {
		remaining = debt.PrincipalAmount
	}
	// Round payment in debt currency for comparison
	paymentInDebtCurrency = roundAmountForCurrencyUp(paymentInDebtCurrency, debtCurrency)
	isFullPayment := remaining > 0 && paymentInDebtCurrency >= remaining-0.01

	// Determine the applied rate for audit trail:
	// Use client-provided rate if available, otherwise use debt's stored rate.
	appliedRate := input.AppliedRate
	if appliedRate <= 0 {
		appliedRate = debtStoredRate
	}

	payment := &DebtPayment{
		DebtID:                debt.ID,
		Amount:                debitAmount,
		Currency:              accountCurrency,
		BaseCurrency:          debt.BaseCurrency,
		RateUsedToBase:        debtToBaseRate,
		ConvertedAmountToBase: paymentInDebtCurrency * debtToBaseRate,
		RateUsedToDebt:        accountToDebtRate,
		ConvertedAmountToDebt: paymentInDebtCurrency,
		PaymentDate:           dateValue,
		AccountID:             &account.ID,
		Note:                  input.Note,
		AppliedRate:           appliedRate,
		TransactionType: func() string {
			if isFullPayment {
				return TransactionTypeDebtFullPayment
			} else {
				return TransactionTypeDebtPayment
			}
		}(),
	}

	createdPayment, err := s.CreateDebtPayment(ctx, debt, payment)
	if err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	var createdTxn *Transaction
	if createdPayment.RelatedTransactionID != nil && strings.TrimSpace(*createdPayment.RelatedTransactionID) != "" {
		createdTxn, _ = s.repo.GetTransactionByID(ctx, *createdPayment.RelatedTransactionID)
		if createdTxn != nil {
			normalizeTransaction(createdTxn)
		}
	}
	updatedDebt, err := s.GetDebt(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if updatedDebt.RemainingAmount <= 0.01 && updatedDebt.Status != "paid" {
		updatedDebt.Status = "paid"
		now := time.Now().UTC().Format(time.RFC3339)
		updatedDebt.SettledAt = &now
		normalizeDebt(updatedDebt)
		if err := s.repo.UpdateDebt(ctx, updatedDebt); err != nil {
			return nil, err
		}
		updatedDebt, _ = s.GetDebt(ctx, debtID)
	}

	return &DebtValueResult{
		Debt:        updatedDebt,
		Payment:     createdPayment,
		Transaction: createdTxn,
	}, nil
}

func (s *Service) AddDebtValue(ctx context.Context, debtID string, input DebtValueInput) (*DebtValueResult, error) {
	if strings.TrimSpace(debtID) == "" {
		return nil, appErrors.InvalidFinanceData
	}
	if input.Amount <= 0 {
		return nil, appErrors.InvalidAmount
	}
	if strings.TrimSpace(input.AccountID) == "" {
		return nil, appErrors.InvalidFinanceData
	}

	debt, err := s.repo.GetDebtByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	normalizeDebt(debt)

	if strings.TrimSpace(input.AmountCurrency) == "" {
		return nil, appErrors.InvalidCurrency
	}
	if !strings.EqualFold(input.AmountCurrency, debt.PrincipalCurrency) {
		return nil, appErrors.InvalidCurrency
	}

	account, err := s.repo.GetAccountByID(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	normalizeAccount(account)

	dateValue := normalizeDateInput("")
	if input.Date != nil && strings.TrimSpace(*input.Date) != "" {
		dateValue = normalizeDateInput(*input.Date)
	}
	if _, err := time.Parse("2006-01-02", dateValue); err != nil {
		return nil, appErrors.WithDetails(appErrors.InvalidTransactionDate, map[string]interface{}{"field": "date"})
	}

	debtCurrency := debt.PrincipalCurrency
	accountCurrency := account.Currency

	debtToAccountRate := 1.0
	debitAmount := input.Amount
	if !strings.EqualFold(debtCurrency, accountCurrency) {
		rate, err := s.resolveFXRate(ctx, debtCurrency, accountCurrency, dateValue)
		if err != nil {
			return nil, err
		}
		debtToAccountRate = rate
		if debtToAccountRate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
		debitAmount = input.Amount * debtToAccountRate
	}

	debitAmount = roundAmountForCurrencyUp(debitAmount, accountCurrency)
	if debitAmount <= 0 {
		return nil, appErrors.InvalidAmount
	}

	isExpense := debt.Direction == "they_owe_me"
	if isExpense && account.CurrentBalance < debitAmount {
		return nil, appErrors.WithDetails(appErrors.InsufficientFunds, map[string]interface{}{
			"required":  debitAmount,
			"available": account.CurrentBalance,
			"currency":  accountCurrency,
		})
	}

	accountToDebtRate := 1.0
	if !strings.EqualFold(accountCurrency, debtCurrency) {
		rate, err := s.resolveFXRate(ctx, accountCurrency, debtCurrency, dateValue)
		if err != nil {
			return nil, err
		}
		accountToDebtRate = rate
		if accountToDebtRate <= 0 {
			return nil, appErrors.FXRateNotFound
		}
	}

	transactionDelta := debitAmount
	if isExpense {
		transactionDelta = -debitAmount
	}

	linkedDebtID := debt.ID
	txn := &Transaction{
		Type:           TransactionTypeDebtAddValue,
		AccountID:      &account.ID,
		Amount:         transactionDelta,
		Currency:       accountCurrency,
		BaseCurrency:   debtCurrency,
		RateUsedToBase: accountToDebtRate,
		ConvertedAmountToBase: input.Amount * func() float64 {
			if transactionDelta < 0 {
				return -1
			} else {
				return 1
			}
		}(),
		DebtID:           &linkedDebtID,
		RelatedDebtID:    &linkedDebtID,
		Description:      input.Note,
		Date:             dateValue,
		OriginalCurrency: &debtCurrency,
		OriginalAmount:   input.Amount,
		ConversionRate:   debtToAccountRate,
	}

	createdTxn, err := s.CreateTransaction(ctx, txn)
	if err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)

	debt.PrincipalAmount += input.Amount
	debt.PrincipalBaseValue = debt.PrincipalAmount * debt.RateOnStart
	if err := s.repo.UpdateDebt(ctx, debt); err != nil {
		return nil, err
	}

	updatedDebt, err := s.GetDebt(ctx, debtID)
	if err != nil {
		return nil, err
	}

	return &DebtValueResult{
		Debt:        updatedDebt,
		Transaction: createdTxn,
	}, nil
}

func (s *Service) DeleteBudget(ctx context.Context, id string) error {
	return s.repo.DeleteBudget(ctx, id)
}

func (s *Service) Debts(ctx context.Context, filter DebtFilter) ([]*Debt, error) {
	debts, err := s.repo.ListDebts(ctx)
	if err != nil {
		return nil, err
	}
	debts = filterDebts(debts, filter)
	for _, debt := range debts {
		normalizeDebt(debt)
		payments, err := s.repo.ListDebtPayments(ctx, debt.ID)
		if err != nil {
			return nil, err
		}
		applyDebtRollups(debt, payments)
	}
	return debts, nil
}

func (s *Service) GetDebt(ctx context.Context, id string) (*Debt, error) {
	debt, err := s.repo.GetDebtByID(ctx, id)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.ListDebtPayments(ctx, id)
	if err != nil {
		return nil, err
	}
	normalizeDebt(debt)
	applyDebtRollups(debt, payments)
	return debt, nil
}

func (s *Service) AccountTransactions(ctx context.Context, accountID string) ([]*Transaction, error) {
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	transactions = filterTransactions(transactions, TransactionFilter{AccountID: accountID})
	for _, txn := range transactions {
		normalizeTransaction(txn)
	}
	return transactions, nil
}

func (s *Service) AccountBalanceHistory(ctx context.Context, accountID string) ([]BalanceHistoryPoint, error) {
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	transactions = filterTransactions(transactions, TransactionFilter{AccountID: accountID})
	normalizeAccount(account)
	return buildBalanceHistory(account, transactions), nil
}

func (s *Service) BudgetTransactions(ctx context.Context, budgetID string) ([]*Transaction, error) {
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	transactions = filterTransactions(transactions, TransactionFilter{BudgetID: budgetID})
	for _, txn := range transactions {
		normalizeTransaction(txn)
	}
	return transactions, nil
}

func (s *Service) BudgetSpending(ctx context.Context, budgetID string) ([]BudgetSpendingItem, error) {
	budget, err := s.repo.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	normalizeBudget(budget)
	total := 0.0
	byCategory := map[string]*BudgetSpendingItem{}
	for _, txn := range transactions {
		if txn.BudgetID == nil || *txn.BudgetID != budget.ID || txn.Type != "expense" {
			continue
		}
		if !budgetWithinPeriod(budget, txn.Date) {
			continue
		}
		if txn.CategoryID == nil {
			continue
		}
		item, ok := byCategory[*txn.CategoryID]
		if !ok {
			item = &BudgetSpendingItem{CategoryID: *txn.CategoryID, CategoryName: *txn.CategoryID}
			byCategory[*txn.CategoryID] = item
		}
		item.Amount += txn.Amount
		total += txn.Amount
	}
	items := make([]BudgetSpendingItem, 0, len(byCategory))
	for _, item := range byCategory {
		if total > 0 {
			item.Percentage = (item.Amount / total) * 100
		}
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Amount > items[j].Amount
	})
	return items, nil
}

func (s *Service) RecalculateBudget(ctx context.Context, budgetID string) (*Budget, error) {
	budget, err := s.repo.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	transactions, err := s.repo.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	normalizeBudget(budget)
	applyBudgetRollups(budget, transactions)
	if err := s.repo.UpdateBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

func (s *Service) SettleDebt(ctx context.Context, debtID string) (*Debt, error) {
	debt, err := s.GetDebt(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if debt.RemainingAmount > 0.01 {
		return nil, appErrors.InvalidFinanceData
	}
	debt.Status = "paid"
	now := time.Now().UTC().Format(time.RFC3339)
	debt.SettledAt = &now
	normalizeDebt(debt)
	if err := s.repo.UpdateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

func (s *Service) ExtendDebt(ctx context.Context, debtID, dueDate string) (*Debt, error) {
	debt, err := s.repo.GetDebtByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	debt.DueDate = &dueDate
	normalizeDebt(debt)
	if err := s.repo.UpdateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

func (s *Service) CreateDebt(ctx context.Context, debt *Debt) (*Debt, error) {
	if err := s.ensureCounterpartyExists(ctx, debt.CounterpartyID); err != nil {
		return nil, err
	}
	normalizeDebt(debt)
	if err := s.repo.CreateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

// CreateDebtWithCounterparty creates a debt and returns it with embedded counterparty.
func (s *Service) CreateDebtWithCounterparty(ctx context.Context, debt *Debt) (*DebtResponse, error) {
	if err := s.ensureCounterpartyExists(ctx, debt.CounterpartyID); err != nil {
		return nil, err
	}
	normalizeDebt(debt)
	if err := s.repo.CreateDebt(ctx, debt); err != nil {
		return nil, err
	}

	// Fetch counterparty if linked
	var counterpartyEmbed *CounterpartyEmbed
	if debt.CounterpartyID != nil && *debt.CounterpartyID != "" {
		counterparty, err := s.repo.GetCounterpartyByID(ctx, *debt.CounterpartyID)
		if err == nil && counterparty != nil {
			counterpartyEmbed = &CounterpartyEmbed{
				ID:          counterparty.ID,
				DisplayName: counterparty.DisplayName,
				PhoneNumber: counterparty.PhoneNumber,
				Comment:     counterparty.Comment,
			}
		}
	}

	// Apply rollups
	payments, _ := s.repo.ListDebtPayments(ctx, debt.ID)
	applyDebtRollups(debt, payments)

	return &DebtResponse{
		Debt:         debt,
		Counterparty: counterpartyEmbed,
	}, nil
}

// GetDebtWithCounterparty gets a debt by ID and embeds the counterparty.
func (s *Service) GetDebtWithCounterparty(ctx context.Context, id string) (*DebtResponse, error) {
	debt, err := s.repo.GetDebtByID(ctx, id)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.ListDebtPayments(ctx, id)
	if err != nil {
		return nil, err
	}
	normalizeDebt(debt)
	applyDebtRollups(debt, payments)

	// Fetch counterparty if linked
	var counterpartyEmbed *CounterpartyEmbed
	if debt.CounterpartyID != nil && *debt.CounterpartyID != "" {
		counterparty, err := s.repo.GetCounterpartyByID(ctx, *debt.CounterpartyID)
		if err == nil && counterparty != nil {
			counterpartyEmbed = &CounterpartyEmbed{
				ID:          counterparty.ID,
				DisplayName: counterparty.DisplayName,
				PhoneNumber: counterparty.PhoneNumber,
				Comment:     counterparty.Comment,
			}
		}
	}

	return &DebtResponse{
		Debt:         debt,
		Counterparty: counterpartyEmbed,
	}, nil
}

// DebtsWithCounterparties gets all debts with embedded counterparties.
func (s *Service) DebtsWithCounterparties(ctx context.Context, filter DebtFilter) ([]*DebtResponse, error) {
	debts, err := s.repo.ListDebts(ctx)
	if err != nil {
		return nil, err
	}
	debts = filterDebts(debts, filter)

	// Fetch all counterparties once
	counterpartiesMap := make(map[string]*Counterparty)
	counterparties, _ := s.repo.ListCounterparties(ctx)
	for _, cp := range counterparties {
		counterpartiesMap[cp.ID] = cp
	}

	results := make([]*DebtResponse, 0, len(debts))
	for _, debt := range debts {
		normalizeDebt(debt)
		payments, _ := s.repo.ListDebtPayments(ctx, debt.ID)
		applyDebtRollups(debt, payments)

		var counterpartyEmbed *CounterpartyEmbed
		if debt.CounterpartyID != nil && *debt.CounterpartyID != "" {
			if cp, ok := counterpartiesMap[*debt.CounterpartyID]; ok {
				counterpartyEmbed = &CounterpartyEmbed{
					ID:          cp.ID,
					DisplayName: cp.DisplayName,
					PhoneNumber: cp.PhoneNumber,
					Comment:     cp.Comment,
				}
			}
		}

		results = append(results, &DebtResponse{
			Debt:         debt,
			Counterparty: counterpartyEmbed,
		})
	}

	return results, nil
}

func (s *Service) UpdateDebt(ctx context.Context, id string, debt *Debt) (*Debt, error) {
	debt.ID = id
	normalizeDebt(debt)
	if err := s.ensureCounterpartyExists(ctx, debt.CounterpartyID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateDebt(ctx, debt); err != nil {
		return nil, err
	}
	return debt, nil
}

func (s *Service) PatchDebt(ctx context.Context, id string, fields map[string]interface{}) (*Debt, error) {
	current, err := s.repo.GetDebtByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyDebtPatch(current, fields)
	normalizeDebt(current)
	if err := s.ensureCounterpartyExists(ctx, current.CounterpartyID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateDebt(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteDebt(ctx context.Context, id string) error {
	return s.repo.DeleteDebt(ctx, id)
}

func (s *Service) DebtPayments(ctx context.Context, debtID string) ([]*DebtPayment, error) {
	payments, err := s.repo.ListDebtPayments(ctx, debtID)
	if err != nil {
		return nil, err
	}
	for _, payment := range payments {
		normalizeDebtPayment(payment, nil)
	}
	return payments, nil
}

func (s *Service) CreateDebtPayment(ctx context.Context, debt *Debt, payment *DebtPayment) (*DebtPayment, error) {
	normalizeDebtPayment(payment, debt)
	if err := s.repo.CreateDebtPayment(ctx, payment); err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	return payment, nil
}

func (s *Service) ensureCounterpartyExists(ctx context.Context, counterpartyID *string) error {
	if counterpartyID == nil || strings.TrimSpace(*counterpartyID) == "" {
		return nil
	}
	if _, err := s.repo.GetCounterpartyByID(ctx, *counterpartyID); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateDebtPayment(ctx context.Context, debt *Debt, payment *DebtPayment) (*DebtPayment, error) {
	normalizeDebtPayment(payment, debt)
	if err := s.repo.UpdateDebtPayment(ctx, payment); err != nil {
		return nil, err
	}
	s.invalidateFinanceSummaryCache(ctx)
	return payment, nil
}

func (s *Service) DeleteDebtPayment(ctx context.Context, debtID, paymentID string) error {
	if err := s.repo.DeleteDebtPayment(ctx, debtID, paymentID); err != nil {
		return err
	}
	s.invalidateFinanceSummaryCache(ctx)
	return nil
}

func (s *Service) summaryCacheKey(userID, dateFrom, dateTo, baseCurrency string, accountIDs []string) string {
	parts := make([]string, 0, len(accountIDs))
	for _, id := range accountIDs {
		if strings.TrimSpace(id) != "" {
			parts = append(parts, id)
		}
	}
	sort.Strings(parts)
	accountKey := strings.Join(parts, ",")
	return strings.Join([]string{
		"finance:summary",
		userID,
		strings.TrimSpace(dateFrom),
		strings.TrimSpace(dateTo),
		strings.TrimSpace(baseCurrency),
		accountKey,
	}, ":")
}

func (s *Service) getFinanceSummaryCache(ctx context.Context, dateFrom, dateTo, baseCurrency string, accountIDs []string) *FinanceSummary {
	if s.cache == nil {
		return nil
	}
	userID, ok := ctx.Value("user_id").(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil
	}
	key := s.summaryCacheKey(userID, dateFrom, dateTo, baseCurrency, accountIDs)
	raw, err := s.cache.Get(ctx, key).Result()
	if err != nil || raw == "" {
		return nil
	}
	var summary FinanceSummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return nil
	}
	return &summary
}

func (s *Service) setFinanceSummaryCache(ctx context.Context, dateFrom, dateTo, baseCurrency string, accountIDs []string, summary *FinanceSummary) {
	if s.cache == nil || summary == nil {
		return
	}
	userID, ok := ctx.Value("user_id").(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return
	}
	key := s.summaryCacheKey(userID, dateFrom, dateTo, baseCurrency, accountIDs)
	payload, err := json.Marshal(summary)
	if err != nil {
		return
	}
	_ = s.cache.Set(ctx, key, payload, financeSummaryCacheTTL).Err()
}

func (s *Service) invalidateFinanceSummaryCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	userID, ok := ctx.Value("user_id").(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return
	}
	pattern := "finance:summary:" + userID + ":*"
	var cursor uint64
	for {
		keys, nextCursor, err := s.cache.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_ = s.cache.Del(ctx, keys...).Err()
		}
		cursor = nextCursor
		if cursor == 0 {
			return
		}
	}
}

func (s *Service) Counterparties(ctx context.Context, filter CounterpartyFilter) ([]*Counterparty, error) {
	items, err := s.repo.ListCounterparties(ctx)
	if err != nil {
		return nil, err
	}
	items = filterCounterparties(items, filter)
	for _, item := range items {
		item.ShowStatus = normalizeShowStatus(item.ShowStatus)
	}
	return items, nil
}

func (s *Service) GetCounterparty(ctx context.Context, id string) (*Counterparty, error) {
	item, err := s.repo.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	item.ShowStatus = normalizeShowStatus(item.ShowStatus)
	return item, nil
}

func (s *Service) CreateCounterparty(ctx context.Context, counterparty *Counterparty) (*Counterparty, error) {
	normalizeCounterparty(counterparty)
	if err := s.repo.CreateCounterparty(ctx, counterparty); err != nil {
		return nil, err
	}
	return counterparty, nil
}

func (s *Service) UpdateCounterparty(ctx context.Context, id string, counterparty *Counterparty) (*Counterparty, error) {
	counterparty.ID = id
	normalizeCounterparty(counterparty)
	if err := s.repo.UpdateCounterparty(ctx, counterparty); err != nil {
		return nil, err
	}
	return counterparty, nil
}

func (s *Service) PatchCounterparty(ctx context.Context, id string, fields map[string]interface{}) (*Counterparty, error) {
	current, err := s.repo.GetCounterpartyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyCounterpartyPatch(current, fields)
	normalizeCounterparty(current)
	if err := s.repo.UpdateCounterparty(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *Service) DeleteCounterparty(ctx context.Context, id string) error {
	return s.repo.DeleteCounterparty(ctx, id)
}

func (s *Service) FXRates(ctx context.Context) ([]*FXRate, error) {
	return s.repo.ListFXRates(ctx)
}

func (s *Service) CreateFXRate(ctx context.Context, rate *FXRate) (*FXRate, error) {
	normalizeFXRate(rate)
	if err := s.repo.CreateFXRate(ctx, rate); err != nil {
		return nil, err
	}
	return rate, nil
}

func (s *Service) GetFXRate(ctx context.Context, id string) (*FXRate, error) {
	return s.repo.GetFXRateByID(ctx, id)
}

func normalizeShowStatus(value string) string {
	switch value {
	case "archived", "deleted":
		return value
	default:
		return "active"
	}
}

func normalizeAccount(account *Account) {
	account.ShowStatus = normalizeShowStatus(account.ShowStatus)
	account.IsArchived = account.ShowStatus == "archived"
}

func normalizeTransaction(txn *Transaction) {
	if txn.Type == "" {
		txn.Type = "expense"
	}
	if strings.TrimSpace(txn.Date) != "" {
		datePart := strings.Split(strings.TrimSpace(txn.Date), "T")[0]
		datePart = strings.Split(datePart, " ")[0]
		if datePart != "" {
			txn.Date = datePart
		}
	}
	if txn.BaseCurrency == "" {
		txn.BaseCurrency = txn.Currency
	}
	if txn.RateUsedToBase == 0 {
		txn.RateUsedToBase = 1
	}
	if txn.ConversionRate == 0 {
		txn.ConversionRate = 1
	}
	if txn.EffectiveRateFromTo == 0 {
		txn.EffectiveRateFromTo = 1
	}
	if txn.ConvertedAmountToBase == 0 {
		txn.ConvertedAmountToBase = txn.Amount * txn.RateUsedToBase
	}
	if txn.OriginalAmount == 0 {
		txn.OriginalAmount = txn.Amount
	}
	if len(txn.Attachments) == 0 {
		txn.Attachments = []string{}
	}
	if len(txn.Tags) == 0 {
		txn.Tags = []string{}
	}
	if txn.ShowStatus == "" {
		txn.ShowStatus = "active"
	}
	if txn.Status == "" {
		txn.Status = TransactionStatusCompleted
	}
	if strings.TrimSpace(txn.OccurredAt) == "" {
		if strings.TrimSpace(txn.Date) != "" {
			txn.OccurredAt = txn.Date + "T00:00:00Z"
		} else {
			txn.OccurredAt = time.Now().UTC().Format(time.RFC3339)
		}
	}
	if txn.Metadata == nil {
		txn.Metadata = map[string]interface{}{}
	}
	if txn.RelatedBudgetID == nil && txn.BudgetID != nil {
		txn.RelatedBudgetID = txn.BudgetID
	}
	if txn.RelatedDebtID == nil && txn.DebtID != nil {
		txn.RelatedDebtID = txn.DebtID
	}
}

func normalizeBudget(budget *Budget) {
	if budget.BudgetType == "" {
		budget.BudgetType = "category"
	}
	if budget.PeriodType == "" {
		budget.PeriodType = "none"
	}
	if budget.RolloverMode == "" {
		budget.RolloverMode = "none"
	}
	budget.ShowStatus = normalizeShowStatus(budget.ShowStatus)
	budget.IsArchived = budget.ShowStatus == "archived"
	if budget.CategoryIDs == nil {
		budget.CategoryIDs = []string{}
	}
}

func normalizeDebt(debt *Debt) {
	if debt.Direction == "" {
		debt.Direction = "i_owe"
	}
	if debt.BaseCurrency == "" {
		debt.BaseCurrency = debt.PrincipalCurrency
	}
	if debt.RateOnStart == 0 {
		debt.RateOnStart = 1
	}
	if debt.PrincipalBaseValue == 0 {
		debt.PrincipalBaseValue = debt.PrincipalAmount * debt.RateOnStart
	}
	if debt.RepaymentRateOnStart == 0 {
		debt.RepaymentRateOnStart = 1
	}
	if debt.Status == "" {
		debt.Status = "active"
	}
	debt.ShowStatus = normalizeShowStatus(debt.ShowStatus)
}

func normalizeDebtPayment(payment *DebtPayment, debt *Debt) {
	if debt != nil && payment.BaseCurrency == "" {
		payment.BaseCurrency = debt.BaseCurrency
	}
	if payment.RateUsedToBase == 0 {
		payment.RateUsedToBase = 1
	}
	if payment.RateUsedToDebt == 0 {
		payment.RateUsedToDebt = 1
	}
	if payment.ConvertedAmountToBase == 0 {
		payment.ConvertedAmountToBase = payment.Amount * payment.RateUsedToBase
	}
	if payment.ConvertedAmountToDebt == 0 {
		payment.ConvertedAmountToDebt = payment.Amount * payment.RateUsedToDebt
	}
}

func normalizeCounterparty(counterparty *Counterparty) {
	counterparty.ShowStatus = normalizeShowStatus(counterparty.ShowStatus)
}

func normalizeFXRate(rate *FXRate) {
	if rate.Nominal == 0 {
		rate.Nominal = 1
	}
	if rate.Source == "" {
		rate.Source = "manual"
	}
}

func normalizeDateInput(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Now().UTC().Format("2006-01-02")
	}
	datePart := strings.Split(trimmed, "T")[0]
	datePart = strings.Split(datePart, " ")[0]
	if datePart == "" {
		return time.Now().UTC().Format("2006-01-02")
	}
	return datePart
}

func roundAmountForCurrency(amount float64, currency string) float64 {
	decimals := 2
	if strings.EqualFold(currency, "UZS") {
		decimals = 0
	}
	factor := math.Pow(10, float64(decimals))
	return math.Round(amount*factor) / factor
}

func roundAmountForCurrencyUp(amount float64, currency string) float64 {
	decimals := 2
	if strings.EqualFold(currency, "UZS") {
		decimals = 0
	}
	factor := math.Pow(10, float64(decimals))
	return math.Ceil(amount*factor) / factor
}

func fxRateValue(rate *FXRate) float64 {
	if rate == nil {
		return 0
	}
	value := rate.Rate
	if value == 0 {
		value = rate.RateMid
	}
	if value == 0 {
		value = rate.RateAsk
	}
	if value == 0 {
		value = rate.RateBid
	}
	if value == 0 {
		return 0
	}
	nominal := rate.Nominal
	if nominal == 0 {
		nominal = 1
	}
	return value / nominal
}

func (s *Service) resolveFXRate(ctx context.Context, fromCurrency, toCurrency, dateValue string) (float64, error) {
	if strings.EqualFold(fromCurrency, toCurrency) {
		return 1, nil
	}
	rates, err := s.repo.ListFXRates(ctx)
	if err != nil {
		return 0, err
	}
	targetDate, err := time.Parse("2006-01-02", dateValue)
	if err != nil {
		targetDate = time.Now().UTC()
	}

	selectRate := func(from, to string) *FXRate {
		var best *FXRate
		var bestDate time.Time
		for _, rate := range rates {
			if !strings.EqualFold(rate.FromCurrency, from) || !strings.EqualFold(rate.ToCurrency, to) {
				continue
			}
			if strings.TrimSpace(rate.Date) == "" {
				continue
			}
			rateDate, err := time.Parse("2006-01-02", rate.Date)
			if err != nil {
				continue
			}
			if rateDate.After(targetDate) {
				continue
			}
			if best == nil || rateDate.After(bestDate) {
				best = rate
				bestDate = rateDate
			}
		}
		if best == nil {
			for _, rate := range rates {
				if strings.EqualFold(rate.FromCurrency, from) && strings.EqualFold(rate.ToCurrency, to) {
					best = rate
					break
				}
			}
		}
		return best
	}

	direct := selectRate(fromCurrency, toCurrency)
	if direct != nil {
		value := fxRateValue(direct)
		if value > 0 {
			return value, nil
		}
	}

	inverse := selectRate(toCurrency, fromCurrency)
	if inverse != nil {
		value := fxRateValue(inverse)
		if value > 0 {
			return 1 / value, nil
		}
	}

	fromFallback, okFrom := defaultFXRates[strings.ToUpper(fromCurrency)]
	toFallback, okTo := defaultFXRates[strings.ToUpper(toCurrency)]
	if okFrom && okTo && toFallback > 0 {
		return fromFallback / toFallback, nil
	}

	return 0, appErrors.FXRateNotFound
}

var defaultFXRates = map[string]float64{
	"UZS":  1,
	"USD":  12450,
	"EUR":  13600,
	"GBP":  15800,
	"TRY":  375,
	"SAR":  3300,
	"AED":  3380,
	"USDT": 12450,
	"RUB":  140,
}

func defaultFinanceCategories() []*FinanceCategory {
	defaults := []struct {
		Type      string
		NameI18n  map[string]string
		IconName  string
		Color     *string
		SortOrder int
	}{
		{Type: "expense", NameI18n: map[string]string{"en": "Food & Dining", "ru": "Еда и рестораны", "uz": "Ovqat va restoranlar"}, IconName: "Utensils", SortOrder: 1},
		{Type: "expense", NameI18n: map[string]string{"en": "Transportation", "ru": "Транспорт", "uz": "Transport"}, IconName: "Car", SortOrder: 2},
		{Type: "expense", NameI18n: map[string]string{"en": "Shopping", "ru": "Покупки", "uz": "Xaridlar"}, IconName: "ShoppingCart", SortOrder: 3},
		{Type: "expense", NameI18n: map[string]string{"en": "Bills & Utilities", "ru": "Счета и коммунальные", "uz": "Hisob-kitoblar"}, IconName: "Receipt", SortOrder: 4},
		{Type: "expense", NameI18n: map[string]string{"en": "Housing", "ru": "Жильё", "uz": "Uy-joy"}, IconName: "Home", SortOrder: 5},
		{Type: "expense", NameI18n: map[string]string{"en": "Health", "ru": "Здоровье", "uz": "Sog'liq"}, IconName: "HeartPulse", SortOrder: 6},
		{Type: "expense", NameI18n: map[string]string{"en": "Entertainment", "ru": "Развлечения", "uz": "Ko'ngilochar"}, IconName: "Film", SortOrder: 7},
		{Type: "expense", NameI18n: map[string]string{"en": "Debt Payment", "ru": "Погашение долга", "uz": "Qarz to'lovi"}, IconName: "ArrowDownCircle", SortOrder: 8},
		{Type: "income", NameI18n: map[string]string{"en": "Salary", "ru": "Зарплата", "uz": "Maosh"}, IconName: "Briefcase", SortOrder: 1},
		{Type: "income", NameI18n: map[string]string{"en": "Gifts", "ru": "Подарки", "uz": "Sovg'alar"}, IconName: "Gift", SortOrder: 2},
		{Type: "income", NameI18n: map[string]string{"en": "Interest", "ru": "Проценты", "uz": "Foiz"}, IconName: "Percent", SortOrder: 3},
	}

	results := make([]*FinanceCategory, 0, len(defaults))
	for _, item := range defaults {
		results = append(results, &FinanceCategory{
			Type:      item.Type,
			NameI18n:  item.NameI18n,
			IconName:  item.IconName,
			Color:     item.Color,
			IsDefault: true,
			SortOrder: item.SortOrder,
			IsActive:  true,
		})
	}
	return results
}

func groupTransactionsByAccount(transactions []*Transaction) map[string][]*Transaction {
	byAccount := make(map[string][]*Transaction)
	for _, txn := range transactions {
		if txn.AccountID != nil {
			byAccount[*txn.AccountID] = append(byAccount[*txn.AccountID], txn)
		}
		if txn.FromAccountID != nil {
			byAccount[*txn.FromAccountID] = append(byAccount[*txn.FromAccountID], txn)
		}
		if txn.ToAccountID != nil {
			byAccount[*txn.ToAccountID] = append(byAccount[*txn.ToAccountID], txn)
		}
	}
	return byAccount
}

func transactionDeltaForAccount(accountID string, txn *Transaction) float64 {
	if txn == nil || accountID == "" {
		return 0
	}
	switch txn.Type {
	case TransactionTypeIncome:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return txn.Amount
		}
	case TransactionTypeExpense:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return -txn.Amount
		}
	case TransactionTypeAccountDeleteWithdrawal:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return -txn.Amount
		}
	case TransactionTypeBudgetAddValue, TransactionTypeDebtAddValue:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return txn.Amount
		}
	case TransactionTypeTransfer:
		if txn.FromAccountID != nil && *txn.FromAccountID == accountID {
			return -txn.Amount
		}
		if txn.ToAccountID != nil && *txn.ToAccountID == accountID {
			amount := txn.ToAmount
			if amount == 0 {
				amount = txn.Amount
			}
			return amount
		}
	case TransactionTypeTransferOut:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return -txn.Amount
		}
		if txn.FromAccountID != nil && *txn.FromAccountID == accountID {
			return -txn.Amount
		}
	case TransactionTypeTransferIn:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return txn.Amount
		}
		if txn.ToAccountID != nil && *txn.ToAccountID == accountID {
			return txn.Amount
		}
	case TransactionTypeSystemAdjustment, TransactionTypeDebtCreate, TransactionTypeDebtPayment, TransactionTypeDebtAdjustment, TransactionTypeDebtFullPayment:
		if txn.AccountID != nil && *txn.AccountID == accountID {
			return txn.Amount
		}
	case TransactionTypeSystemOpening:
		return 0
	}
	return 0
}

func transactionImpactForSummary(txn *Transaction) float64 {
	if txn == nil {
		return 0
	}
	switch txn.Type {
	case TransactionTypeIncome:
		return txn.Amount
	case TransactionTypeExpense:
		return -txn.Amount
	case TransactionTypeSystemAdjustment, TransactionTypeDebtCreate, TransactionTypeDebtPayment, TransactionTypeDebtAdjustment, TransactionTypeDebtFullPayment, TransactionTypeBudgetAddValue, TransactionTypeDebtAddValue:
		return txn.Amount
	case TransactionTypeAccountCreateFunding:
		return txn.Amount
	case TransactionTypeAccountDeleteWithdrawal:
		return -txn.Amount
	default:
		return 0
	}
}

func computeAccountBalance(account *Account, transactions []*Transaction) float64 {
	balance := account.InitialBalance
	for _, txn := range transactions {
		balance += transactionDeltaForAccount(account.ID, txn)
	}
	return balance
}

func buildBalanceHistory(account *Account, transactions []*Transaction) []BalanceHistoryPoint {
	points := make([]BalanceHistoryPoint, 0)
	balance := account.InitialBalance
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date < transactions[j].Date
	})
	lastDate := ""
	for _, txn := range transactions {
		if txn.Date == "" {
			continue
		}
		balance += transactionDeltaForAccount(account.ID, txn)
		if txn.Date != lastDate {
			points = append(points, BalanceHistoryPoint{Date: txn.Date, Balance: balance})
			lastDate = txn.Date
		}
	}
	return points
}

func filterTransactions(transactions []*Transaction, filter TransactionFilter) []*Transaction {
	filtered := make([]*Transaction, 0, len(transactions))
	for _, txn := range transactions {
		if filter.AccountID != "" {
			matchesAccount := txn.AccountID != nil && *txn.AccountID == filter.AccountID
			if txn.FromAccountID != nil && *txn.FromAccountID == filter.AccountID {
				matchesAccount = true
			}
			if txn.ToAccountID != nil && *txn.ToAccountID == filter.AccountID {
				matchesAccount = true
			}
			if !matchesAccount {
				continue
			}
		}
		if filter.Type != "" {
			if filter.Type == TransactionTypeTransfer {
				if txn.Type != TransactionTypeTransfer && txn.Type != TransactionTypeTransferIn && txn.Type != TransactionTypeTransferOut {
					continue
				}
			} else if txn.Type != filter.Type {
				continue
			}
		}
		if filter.CategoryID != "" {
			if txn.CategoryID == nil || *txn.CategoryID != filter.CategoryID {
				continue
			}
		}
		if filter.GoalID != "" {
			if txn.GoalID == nil || *txn.GoalID != filter.GoalID {
				continue
			}
		}
		if filter.BudgetID != "" {
			if txn.BudgetID == nil || *txn.BudgetID != filter.BudgetID {
				continue
			}
		}
		if filter.DebtID != "" {
			if txn.DebtID == nil || *txn.DebtID != filter.DebtID {
				continue
			}
		}
		if filter.DateFrom != "" || filter.DateTo != "" {
			if !dateInRange(txn.Date, filter.DateFrom, filter.DateTo) {
				continue
			}
		}
		filtered = append(filtered, txn)
	}
	return filtered
}

func filterBudgets(budgets []*Budget, filter BudgetFilter) []*Budget {
	filtered := make([]*Budget, 0, len(budgets))
	for _, budget := range budgets {
		if filter.PeriodType != "" && budget.PeriodType != filter.PeriodType {
			continue
		}
		if filter.LinkedGoalID != "" {
			if budget.LinkedGoalID == nil || *budget.LinkedGoalID != filter.LinkedGoalID {
				continue
			}
		}
		if filter.IsArchived != nil {
			if budget.IsArchived != *filter.IsArchived {
				continue
			}
		}
		filtered = append(filtered, budget)
	}
	return filtered
}

func filterDebts(debts []*Debt, filter DebtFilter) []*Debt {
	filtered := make([]*Debt, 0, len(debts))
	for _, debt := range debts {
		if filter.Direction != "" && debt.Direction != filter.Direction {
			continue
		}
		if filter.Status != "" && debt.Status != filter.Status {
			continue
		}
		if filter.LinkedGoalID != "" {
			if debt.LinkedGoalID == nil || *debt.LinkedGoalID != filter.LinkedGoalID {
				continue
			}
		}
		filtered = append(filtered, debt)
	}
	return filtered
}

func filterCounterparties(items []*Counterparty, filter CounterpartyFilter) []*Counterparty {
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	if search == "" {
		return items
	}
	filtered := make([]*Counterparty, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.DisplayName), search) {
			filtered = append(filtered, item)
			continue
		}
		if item.PhoneNumber != nil && strings.Contains(strings.ToLower(*item.PhoneNumber), search) {
			filtered = append(filtered, item)
			continue
		}
		if item.SearchKeywords != nil && strings.Contains(strings.ToLower(*item.SearchKeywords), search) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func applyBudgetRollups(budget *Budget, transactions []*Transaction) {
	spent := 0.0
	trackType := "expense"
	if budget.TransactionType != nil && strings.ToLower(*budget.TransactionType) == "income" {
		trackType = "income"
	}
	for _, txn := range transactions {
		if txn.BudgetID == nil || *txn.BudgetID != budget.ID {
			continue
		}
		if txn.Type != trackType && txn.Type != TransactionTypeBudgetAddValue {
			continue
		}
		if !budgetWithinPeriod(budget, txn.Date) {
			continue
		}
		amount := txn.Amount
		if txn.OriginalCurrency != nil && strings.EqualFold(*txn.OriginalCurrency, budget.Currency) {
			amount = txn.OriginalAmount
		} else if strings.EqualFold(txn.Currency, budget.Currency) {
			amount = txn.Amount
		} else if strings.EqualFold(txn.BaseCurrency, budget.Currency) && txn.ConvertedAmountToBase > 0 {
			amount = txn.ConvertedAmountToBase
		}
		spent += amount
	}
	budget.SpentAmount = spent
	budget.RemainingAmount = budget.LimitAmount - spent
	if budget.LimitAmount > 0 {
		budget.PercentUsed = (spent / budget.LimitAmount) * 100
	} else {
		budget.PercentUsed = 0
	}
	budget.IsOverspent = spent > budget.LimitAmount && budget.LimitAmount > 0
}

func applyDebtRollups(debt *Debt, payments []*DebtPayment) {
	totalPaid := 0.0
	for _, payment := range payments {
		if payment.DebtID != debt.ID {
			continue
		}
		totalPaid += payment.ConvertedAmountToDebt
	}
	debt.TotalPaid = totalPaid
	debt.RemainingAmount = debt.PrincipalAmount - totalPaid
	if debt.PrincipalAmount > 0 {
		debt.PercentPaid = (totalPaid / debt.PrincipalAmount) * 100
	}
}

func applyAccountPatch(account *Account, fields map[string]interface{}) {
	if v, ok := fields["name"].(string); ok {
		account.Name = v
	}
	if v, ok := fields["currency"].(string); ok {
		account.Currency = v
	}
	if v, ok := fields["accountType"].(string); ok {
		account.AccountType = v
	}
	if v, ok := fields["initialBalance"].(float64); ok {
		account.InitialBalance = v
	}
	if v, ok := fields["currentBalance"].(float64); ok {
		account.CurrentBalance = v
	}
	if v, ok := fields["linkedGoalId"].(string); ok {
		account.LinkedGoalID = &v
	}
	if v, ok := fields["customTypeId"].(string); ok {
		account.CustomTypeID = &v
	}
	if v, ok := fields["isMain"].(bool); ok {
		account.IsMain = v
	}
	if v, ok := fields["showStatus"].(string); ok {
		account.ShowStatus = v
	}
}

func applyTransactionPatch(txn *Transaction, fields map[string]interface{}) {
	if v, ok := fields["type"].(string); ok {
		txn.Type = v
	}
	if v, ok := fields["accountId"].(string); ok {
		txn.AccountID = &v
	}
	if v, ok := fields["fromAccountId"].(string); ok {
		txn.FromAccountID = &v
	}
	if v, ok := fields["toAccountId"].(string); ok {
		txn.ToAccountID = &v
	}
	if v, ok := fields["amount"].(float64); ok {
		txn.Amount = v
	}
	if v, ok := fields["currency"].(string); ok {
		txn.Currency = v
	}
	if v, ok := fields["categoryId"].(string); ok {
		txn.CategoryID = &v
	}
	if v, ok := fields["subcategoryId"].(string); ok {
		txn.SubcategoryID = &v
	}
	if v, ok := fields["description"].(string); ok {
		txn.Description = &v
	}
	if v, ok := fields["date"].(string); ok {
		txn.Date = v
	}
	if v, ok := fields["budgetId"].(string); ok {
		txn.BudgetID = &v
	}
	if v, ok := fields["debtId"].(string); ok {
		txn.DebtID = &v
	}
	if v, ok := fields["goalId"].(string); ok {
		txn.GoalID = &v
	}
	if v, ok := fields["showStatus"].(string); ok {
		txn.ShowStatus = v
	}
}

func applyBudgetPatch(budget *Budget, fields map[string]interface{}) {
	if v, ok := fields["name"].(string); ok {
		budget.Name = v
	}
	if v, ok := fields["budgetType"].(string); ok {
		budget.BudgetType = v
	}
	if v, ok := fields["categoryIds"].([]interface{}); ok {
		ids := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				ids = append(ids, s)
			}
		}
		budget.CategoryIDs = ids
	}
	if v, ok := fields["linkedGoalId"].(string); ok {
		budget.LinkedGoalID = &v
	}
	if v, ok := fields["accountId"].(string); ok {
		budget.AccountID = &v
	}
	if v, ok := fields["transactionType"].(string); ok {
		budget.TransactionType = &v
	}
	if v, ok := fields["currency"].(string); ok {
		budget.Currency = v
	}
	if v, ok := fields["limitAmount"].(float64); ok {
		budget.LimitAmount = v
	}
	if v, ok := fields["periodType"].(string); ok {
		budget.PeriodType = v
	}
	if v, ok := fields["startDate"].(string); ok {
		budget.StartDate = &v
	}
	if v, ok := fields["endDate"].(string); ok {
		budget.EndDate = &v
	}
	if v, ok := fields["showStatus"].(string); ok {
		budget.ShowStatus = v
	}
}

func applyDebtPatch(debt *Debt, fields map[string]interface{}) {
	if v, ok := fields["direction"].(string); ok {
		debt.Direction = v
	}
	if v, ok := fields["counterpartyId"].(string); ok {
		debt.CounterpartyID = &v
	}
	if v, ok := fields["counterpartyName"].(string); ok {
		debt.CounterpartyName = v
	}
	if v, ok := fields["principalAmount"].(float64); ok {
		debt.PrincipalAmount = v
	}
	if v, ok := fields["principalCurrency"].(string); ok {
		debt.PrincipalCurrency = v
	}
	if v, ok := fields["status"].(string); ok {
		debt.Status = v
	}
	if v, ok := fields["showStatus"].(string); ok {
		debt.ShowStatus = v
	}
	if v, ok := fields["exchangeRateCurrent"].(float64); ok {
		debt.ExchangeRateCurrent = v
	}
	if v, ok := fields["repaymentRateOnStart"].(float64); ok {
		debt.RepaymentRateOnStart = v
	}
	if v, ok := fields["repaymentCurrency"].(string); ok {
		debt.RepaymentCurrency = &v
	}
	if v, ok := fields["repaymentAmount"].(float64); ok {
		debt.RepaymentAmount = v
	}
	if v, ok := fields["description"].(string); ok {
		debt.Description = &v
	}
	if v, ok := fields["dueDate"].(string); ok {
		debt.DueDate = &v
	}
	if v, ok := fields["startDate"].(string); ok {
		debt.StartDate = v
	}
	if v, ok := fields["reminderEnabled"].(bool); ok {
		debt.ReminderEnabled = v
	}
	if v, ok := fields["fundingAccountId"].(string); ok {
		debt.FundingAccountID = &v
	}
	if v, ok := fields["lentFromAccountId"].(string); ok {
		debt.LentFromAccountID = &v
	}
	if v, ok := fields["returnToAccountId"].(string); ok {
		debt.ReturnToAccountID = &v
	}
	if v, ok := fields["receivedToAccountId"].(string); ok {
		debt.ReceivedToAccountID = &v
	}
	if v, ok := fields["payFromAccountId"].(string); ok {
		debt.PayFromAccountID = &v
	}
}

func applyCounterpartyPatch(counterparty *Counterparty, fields map[string]interface{}) {
	if v, ok := fields["displayName"].(string); ok {
		counterparty.DisplayName = v
	}
	if v, ok := fields["phoneNumber"].(string); ok {
		counterparty.PhoneNumber = &v
	}
	if v, ok := fields["comment"].(string); ok {
		counterparty.Comment = &v
	}
	if v, ok := fields["searchKeywords"].(string); ok {
		counterparty.SearchKeywords = &v
	}
	if v, ok := fields["showStatus"].(string); ok {
		counterparty.ShowStatus = v
	}
}

func budgetWithinPeriod(budget *Budget, dateValue string) bool {
	if budget.PeriodType == "none" || dateValue == "" {
		return true
	}
	if budget.StartDate != nil && budget.EndDate != nil {
		return dateInRange(dateValue, *budget.StartDate, *budget.EndDate)
	}
	return true
}

func dateInRange(dateValue, from, to string) bool {
	if dateValue == "" {
		return false
	}
	date, err := time.Parse("2006-01-02", dateValue)
	if err != nil {
		return false
	}
	if from != "" {
		start, err := time.Parse("2006-01-02", from)
		if err == nil && date.Before(start) {
			return false
		}
	}
	if to != "" {
		end, err := time.Parse("2006-01-02", to)
		if err == nil && date.After(end) {
			return false
		}
	}
	return true
}

# Transaction Data Model

```json
{
  "id": "uuid",
  "userId": "uuid",
  "type": "income|expense|transfer",
  "accountId": "uuid|null",
  "fromAccountId": "uuid|null",
  "toAccountId": "uuid|null",
  "amount": 0,
  "currency": "string",
  "baseCurrency": "string",
  "rateUsedToBase": 1,
  "convertedAmountToBase": 0,
  "toAmount": 0,
  "toCurrency": "string|null",
  "effectiveRateFromTo": 1,
  "feeAmount": 0,
  "feeCategoryId": "string|null",
  "categoryId": "string|null",
  "subcategoryId": "string|null",
  "name": "string|null",
  "description": "string|null",
  "date": "YYYY-MM-DD",
  "time": "HH:mm|null",
  "goalId": "uuid|null",
  "budgetId": "uuid|null",
  "debtId": "uuid|null",
  "habitId": "uuid|null",
  "counterpartyId": "uuid|null",
  "recurringId": "uuid|null",
  "attachments": ["string"],
  "tags": ["string"],
  "isBalanceAdjustment": false,
  "skipBudgetMatching": false,
  "showStatus": "active|archived|deleted",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

### Extended Fields Required by UI
The frontend uses these optional fields when available:
```json
{
  "relatedBudgetId": "uuid|null",
  "relatedDebtId": "uuid|null",
  "goalName": "string|null",
  "goalType": "string|null",
  "plannedAmount": 0,
  "paidAmount": 0,
  "originalCurrency": "string|null",
  "originalAmount": 0,
  "conversionRate": 1
}
```

Notes:
- All numeric fields must be returned as numbers (default `0`).
- `relatedDebtId` is used to display debt-linked transactions in list/detail views.
- `originalAmount` and `conversionRate` are required for debt payment details.


# Debt Data Model

```json
{
  "id": "uuid",
  "userId": "uuid",
  "direction": "i_owe|they_owe_me",
  "counterpartyId": "uuid|null",
  "counterpartyName": "string",
  "description": "string|null",
  "principalAmount": 0,
  "principalCurrency": "string",
  "principalOriginalAmount": 0,
  "principalOriginalCurrency": "string|null",
  "baseCurrency": "string",
  "rateOnStart": 1,
  "principalBaseValue": 0,
  "repaymentCurrency": "string|null",
  "repaymentAmount": 0,
  "repaymentRateOnStart": 1,
  "isFixedRepaymentAmount": false,
  "startDate": "YYYY-MM-DD",
  "dueDate": "YYYY-MM-DD|null",
  "interestMode": "simple|compound|null",
  "interestRateAnnual": 0,
  "scheduleHint": "string|null",
  "linkedGoalId": "uuid|null",
  "linkedBudgetId": "uuid|null",
  "fundingAccountId": "uuid|null",
  "fundingTransactionId": "uuid|null",
  "lentFromAccountId": "uuid|null",
  "returnToAccountId": "uuid|null",
  "receivedToAccountId": "uuid|null",
  "payFromAccountId": "uuid|null",
  "customRateUsed": 0,
  "reminderEnabled": false,
  "reminderTime": "string|null",
  "status": "active|paid|overdue|canceled",
  "settledAt": "ISO8601|null",
  "finalRateUsed": 0,
  "finalProfitLoss": 0,
  "finalProfitLossCurrency": "string|null",
  "totalPaidInRepaymentCurrency": 0,
  "remainingAmount": 0,
  "totalPaid": 0,
  "percentPaid": 0,
  "showStatus": "active|archived|deleted",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```

Notes:
- `remainingAmount`, `totalPaid`, `percentPaid` are calculated server-side.


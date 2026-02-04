# Examples

## GET /goals/:id/finance-progress
```json
{
  "success": true,
  "data": {
    "goalId": "uuid",
    "linkedBudgetId": "uuid",
    "linkedDebtId": null,
    "financialProgressPercent": 45,
    "budgetSpentAmount": 450000,
    "budgetLimitAmount": 1000000,
    "debtTotalPaid": 0,
    "debtPrincipalAmount": 0
  },
  "error": null,
  "meta": null
}
```


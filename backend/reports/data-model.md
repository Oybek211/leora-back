# Reports Data Models

## Finance Summary
```json
{
  "currency": "string",
  "income": 0,
  "expense": 0,
  "net": 0,
  "savingsRate": 0,
  "period": { "from": "YYYY-MM-DD", "to": "YYYY-MM-DD" }
}
```

## Category Breakdown
```json
{
  "total": 0,
  "categories": [
    { "categoryId": "string", "categoryName": "string", "amount": 0, "share": 0 }
  ]
}
```

## Cashflow Buckets
```json
{
  "granularity": "day|week|month",
  "series": [
    { "date": "YYYY-MM-DD", "income": 0, "expense": 0, "net": 0 }
  ]
}
```

## Debt Report
```json
{
  "activeCount": 0,
  "dueSoon": [
    { "debtId": "uuid", "counterpartyName": "string", "dueDate": "YYYY-MM-DD", "remainingAmount": 0 }
  ]
}
```


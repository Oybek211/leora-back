# Debt Payment Data Model

```json
{
  "id": "uuid",
  "debtId": "uuid",
  "amount": 0,
  "currency": "string",
  "baseCurrency": "string",
  "rateUsedToBase": 1,
  "convertedAmountToBase": 0,
  "rateUsedToDebt": 1,
  "convertedAmountToDebt": 0,
  "paymentDate": "YYYY-MM-DD",
  "accountId": "uuid|null",
  "note": "string|null",
  "relatedTransactionId": "uuid|null",
  "createdAt": "ISO8601",
  "updatedAt": "ISO8601"
}
```


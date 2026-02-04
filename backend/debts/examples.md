# Examples

## POST /debts/:id/payments
Request:
```json
{
  "amount": 100,
  "currency": "USD",
  "paymentDate": "2024-01-20",
  "accountId": "uuid",
  "note": "Partial payment",
  "createTransaction": true
}
```
Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "debtId": "uuid",
    "amount": 100,
    "currency": "USD",
    "baseCurrency": "UZS",
    "rateUsedToBase": 12500,
    "convertedAmountToBase": 1250000,
    "rateUsedToDebt": 1,
    "convertedAmountToDebt": 100,
    "paymentDate": "2024-01-20",
    "accountId": "uuid",
    "note": "Partial payment",
    "relatedTransactionId": "uuid",
    "createdAt": "2024-01-20T12:00:00Z",
    "updatedAt": "2024-01-20T12:00:00Z"
  },
  "error": null,
  "meta": null
}
```


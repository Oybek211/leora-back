# Examples

## GET /debts/:id/payments
Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "debtId": "uuid",
      "amount": 50,
      "currency": "USD",
      "baseCurrency": "UZS",
      "rateUsedToBase": 12500,
      "convertedAmountToBase": 625000,
      "rateUsedToDebt": 1,
      "convertedAmountToDebt": 50,
      "paymentDate": "2024-01-10",
      "accountId": "uuid",
      "note": null,
      "relatedTransactionId": "uuid",
      "createdAt": "2024-01-10T09:00:00Z",
      "updatedAt": "2024-01-10T09:00:00Z"
    }
  ],
  "error": null,
  "meta": null
}
```


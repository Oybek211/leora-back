# Examples

## GET /budgets/:id/spending
Response:
```json
{
  "success": true,
  "data": [
    { "categoryId": "food", "categoryName": "Food", "amount": 350000, "percentage": 42 },
    { "categoryId": "transport", "categoryName": "Transport", "amount": 120000, "percentage": 15 }
  ],
  "error": null,
  "meta": null
}
```


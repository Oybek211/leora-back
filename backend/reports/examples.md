# Examples

## GET /reports/finance/categories?from=2024-01-01&to=2024-01-31
```json
{
  "success": true,
  "data": {
    "total": 820000,
    "categories": [
      { "categoryId": "food", "categoryName": "Food", "amount": 350000, "share": 42 },
      { "categoryId": "transport", "categoryName": "Transport", "amount": 120000, "share": 15 }
    ]
  },
  "error": null,
  "meta": null
}
```


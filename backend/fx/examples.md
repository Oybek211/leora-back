# Examples

## GET /fx/rates?from=USD&to=UZS&date=2024-01-15
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "date": "2024-01-15",
    "fromCurrency": "USD",
    "toCurrency": "UZS",
    "rate": 12500,
    "rateMid": 12500,
    "rateBid": 12450,
    "rateAsk": 12550,
    "nominal": 1,
    "spreadPercent": 0.8,
    "source": "cbu"
  },
  "error": null,
  "meta": null
}
```


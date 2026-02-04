# Examples

## POST /habits/:id/evaluate-finance
Request:
```json
{ "dateKey": "2024-01-15" }
```
Response:
```json
{ "success": true, "data": { "status": "done", "evaluatedBy": "finance_rule" }, "error": null, "meta": null }
```


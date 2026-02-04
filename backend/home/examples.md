# Examples

## GET /home
```json
{
  "success": true,
  "data": {
    "tasks": [
      { "id": "uuid", "title": "Morning run", "time": "07:00", "completed": false, "priority": "medium", "context": "Health" }
    ],
    "goals": [
      { "id": "uuid", "title": "Save 5k", "progress": 32, "current": 1600, "target": 5000, "unit": "USD", "category": "financial" }
    ],
    "progress": { "tasks": 60, "budget": 45, "focus": 30 }
  },
  "error": null,
  "meta": null
}
```


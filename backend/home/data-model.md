# Home Data Model

```json
{
  "tasks": [
    { "id": "uuid", "title": "string", "time": "string", "completed": false, "priority": "low|medium|high|critical", "context": "string" }
  ],
  "goals": [
    { "id": "uuid", "title": "string", "progress": 0, "current": 0, "target": 0, "unit": "string", "category": "financial|personal|professional|health" }
  ],
  "progress": { "tasks": 0, "budget": 0, "focus": 0 }
}
```


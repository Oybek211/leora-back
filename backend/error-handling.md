# Error Handling

## Envelope
All responses conform to:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 400,
    "message": "Validation failed",
    "type": "VALIDATION"
  },
  "meta": null
}
```

## Common Error Types
- `VALIDATION`: request schema invalid.
- `UNAUTHORIZED`: token missing or invalid.
- `FORBIDDEN`: access denied.
- `NOT_FOUND`: entity not found.
- `CONFLICT`: duplicate or invalid state transition.
- `RATE_LIMITED`: throttling.
- `INTERNAL`: unexpected error.

## Validation Rules
- Numeric fields always validated and defaulted to `0` when nullable.
- Date fields must be ISO 8601.
- IDs are UUIDv4.


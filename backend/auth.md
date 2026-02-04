# Authentication

## Summary
JWT-based auth with access + refresh tokens. Access token required for all protected endpoints.

## Endpoints
- POST `/auth/register`
- POST `/auth/login`
- GET `/auth/me`
- POST `/auth/forgot-password`
- POST `/auth/reset-password`
- POST `/auth/refresh`
- POST `/auth/logout`

## Token Rules
- Access token in `Authorization: Bearer <token>`.
- Refresh token stored and rotated on `/auth/refresh`.
- Logout invalidates refresh token and clears sessions.

## Request/Response Contracts
All responses use `ApiResponse<T>` envelope.

### POST /auth/register
Request:
```json
{
  "email": "user@leora.app",
  "fullName": "Jane Doe",
  "password": "strongpass",
  "confirmPassword": "strongpass",
  "region": "UZ",
  "currency": "UZS"
}
```
Response `data`:
```json
{
  "user": {
    "id": "uuid",
    "email": "user@leora.app",
    "fullName": "Jane Doe",
    "region": "UZ",
    "primaryCurrency": "UZS",
    "role": "user",
    "status": "active",
    "permissions": [],
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "accessToken": "...",
  "refreshToken": "...",
  "expiresIn": 3600
}
```

### POST /auth/login
Request:
```json
{
  "emailOrUsername": "user@leora.app",
  "password": "strongpass",
  "rememberMe": true
}
```
Response is same shape as register.

### GET /auth/me
Response `data`:
```json
{ "user": { "id": "uuid", "email": "...", "fullName": "...", "region": "...", "primaryCurrency": "...", "role": "user", "status": "active", "permissions": [], "createdAt": "...", "updatedAt": "..." } }
```


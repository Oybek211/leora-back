# Debt API Documentation

## Base URL
```
/api/v1/debts
```

## Authentication
All endpoints require Bearer token in Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

---

## Endpoints

### 1. Create Debt
**POST** `/api/v1/debts`

Creates a new debt record.

#### Request Body (Required Fields)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `direction` | string | ✅ **YES** | Must be exactly `"i_owe"` or `"they_owe_me"` |
| `counterpartyName` | string | ✅ **YES** | Name of the person/entity (min 2 characters) |
| `principalAmount` | number | ✅ **YES** | Amount of debt (must be > 0) |
| `principalCurrency` | string | ✅ **YES** | Currency code (e.g., `"UZS"`, `"USD"`) |
| `startDate` | string | ✅ **YES** | Start date in `YYYY-MM-DD` format |

#### Request Body (Optional Fields)

| Field | Type | Description |
|-------|------|-------------|
| `counterpartyId` | string | UUID of existing counterparty |
| `description` | string | Description of the debt |
| `baseCurrency` | string | Base currency for conversion (default: same as principalCurrency) |
| `dueDate` | string | Due date in `YYYY-MM-DD` format (must be after startDate) |
| `interestMode` | string | Interest calculation mode |
| `interestRateAnnual` | number | Annual interest rate (%) |
| `fundingAccountId` | string | Account ID where money was received/sent |
| `reminderEnabled` | boolean | Enable payment reminders |
| `reminderTime` | string | Reminder time |

#### Example Request - Minimal (I owe someone)
```json
{
  "direction": "i_owe",
  "counterpartyName": "Zafar",
  "principalAmount": 100000,
  "principalCurrency": "UZS",
  "startDate": "2026-01-10"
}
```

#### Example Request - Full (Someone owes me)
```json
{
  "direction": "they_owe_me",
  "counterpartyId": "a8095bf5-8564-48a5-9226-ca5b4c70a2cb",
  "counterpartyName": "Ahmad",
  "description": "Loan for car repair",
  "principalAmount": 5000000,
  "principalCurrency": "UZS",
  "baseCurrency": "UZS",
  "startDate": "2026-01-10",
  "dueDate": "2026-06-10",
  "interestRateAnnual": 0,
  "fundingAccountId": "95d8a355-8668-4f4e-92f5-fc22d5992450",
  "reminderEnabled": true
}
```

#### Example Request - With Inline Counterparty Creation
```json
{
  "direction": "i_owe",
  "principalAmount": 100000,
  "principalCurrency": "UZS",
  "startDate": "2026-01-10",
  "counterparty": {
    "displayName": "New Person",
    "phoneNumber": "+998901234567",
    "comment": "Friend from work"
  }
}
```

#### Success Response (200)
```json
{
  "success": true,
  "data": {
    "id": "5e1f069b-8943-4fde-9885-bc764cee0238",
    "userId": "83b90ae7-97ed-446f-9bfe-1cd0466da539",
    "direction": "i_owe",
    "counterpartyId": "a8095bf5-8564-48a5-9226-ca5b4c70a2cb",
    "counterpartyName": "Zafar",
    "principalAmount": 100000,
    "principalCurrency": "UZS",
    "baseCurrency": "UZS",
    "startDate": "2026-01-10",
    "status": "active",
    "remainingAmount": 100000,
    "totalPaid": 0,
    "percentPaid": 0,
    "createdAt": "2026-01-10T14:33:38Z",
    "updatedAt": "2026-01-10T14:33:38Z",
    "counterparty": {
      "id": "a8095bf5-8564-48a5-9226-ca5b4c70a2cb",
      "displayName": "Zafar"
    }
  },
  "error": null,
  "meta": null
}
```

#### Error Responses

**Invalid Direction (400)**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 400,
    "message": "Direction must be 'i_owe' or 'they_owe_me'",
    "type": "VALIDATION"
  }
}
```

**Missing Counterparty (400)**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 400,
    "message": "Counterparty is required",
    "type": "VALIDATION"
  }
}
```

**Invalid Amount (400)**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 400,
    "message": "Principal amount must be greater than 0",
    "type": "VALIDATION"
  }
}
```

**Counterparty Not Found (404)**
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": 404,
    "message": "Counterparty not found",
    "type": "NOT_FOUND"
  }
}
```

---

### 2. List Debts
**GET** `/api/v1/debts`

Returns all debts for the authenticated user.

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `direction` | string | Filter by `i_owe` or `they_owe_me` |
| `status` | string | Filter by `active`, `paid`, `overdue`, `canceled` |
| `linkedGoalId` | string | Filter by linked goal ID |
| `page` | number | Page number (default: 1) |
| `limit` | number | Items per page (default: 20) |

#### Example Request
```bash
GET /api/v1/debts?direction=i_owe&status=active
```

#### Success Response (200)
```json
{
  "success": true,
  "data": [
    {
      "id": "5e1f069b-8943-4fde-9885-bc764cee0238",
      "direction": "i_owe",
      "counterpartyName": "Zafar",
      "principalAmount": 100000,
      "principalCurrency": "UZS",
      "startDate": "2026-01-10",
      "status": "active",
      "remainingAmount": 100000,
      "totalPaid": 0,
      "percentPaid": 0,
      "counterparty": {
        "id": "a8095bf5-8564-48a5-9226-ca5b4c70a2cb",
        "displayName": "Zafar"
      }
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

---

### 3. Get Debt by ID
**GET** `/api/v1/debts/:id`

Returns a single debt with embedded counterparty.

#### Example Request
```bash
GET /api/v1/debts/5e1f069b-8943-4fde-9885-bc764cee0238
```

---

### 4. Update Debt
**PUT** `/api/v1/debts/:id`

Full update of a debt record.

#### Example Request
```json
{
  "direction": "i_owe",
  "counterpartyName": "Zafar Updated",
  "principalAmount": 150000,
  "principalCurrency": "UZS",
  "startDate": "2026-01-10",
  "dueDate": "2026-03-10"
}
```

---

### 5. Patch Debt
**PATCH** `/api/v1/debts/:id`

Partial update of a debt record.

#### Example Request
```json
{
  "status": "paid"
}
```

---

### 6. Delete Debt
**DELETE** `/api/v1/debts/:id`

Soft deletes a debt record.

---

### 7. Settle Debt
**POST** `/api/v1/debts/:id/settle`

Marks a debt as fully paid.

#### Success Response
```json
{
  "success": true,
  "data": {
    "id": "5e1f069b-8943-4fde-9885-bc764cee0238",
    "status": "paid",
    "settledAt": "2026-01-10T15:00:00Z"
  }
}
```

---

### 8. Extend Debt Due Date
**POST** `/api/v1/debts/:id/extend`

Extends the due date of a debt.

#### Request Body
```json
{
  "dueDate": "2026-12-31"
}
```

---

## Debt Payments

### 9. List Payments
**GET** `/api/v1/debts/:id/payments`

Returns all payments for a debt.

---

### 10. Create Payment
**POST** `/api/v1/debts/:id/payments`

Records a payment against a debt.

#### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `amount` | number | ✅ YES | Payment amount |
| `currency` | string | ✅ YES | Payment currency |
| `paymentDate` | string | ✅ YES | Date in `YYYY-MM-DD` format |
| `accountId` | string | No | Account used for payment |
| `note` | string | No | Payment note |
| `createTransaction` | boolean | No | Auto-create transaction record |

#### Example Request
```json
{
  "amount": 50000,
  "currency": "UZS",
  "paymentDate": "2026-01-15",
  "accountId": "95d8a355-8668-4f4e-92f5-fc22d5992450",
  "note": "First payment",
  "createTransaction": true
}
```

---

### 11. Update Payment
**PATCH** `/api/v1/debts/:id/payments/:paymentId`

Updates a payment record.

---

### 12. Delete Payment
**DELETE** `/api/v1/debts/:id/payments/:paymentId`

Deletes a payment record.

---

## Direction Values

| Value | Description |
|-------|-------------|
| `i_owe` | I owe money to someone (I borrowed) |
| `they_owe_me` | Someone owes me money (I lent) |

---

## Status Values

| Value | Description |
|-------|-------------|
| `active` | Debt is active and has remaining balance |
| `paid` | Debt is fully paid |
| `overdue` | Debt is past due date |
| `canceled` | Debt was canceled |

---

## cURL Examples

### Create Debt (I owe)
```bash
curl -X POST http://localhost:9090/api/v1/debts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "i_owe",
    "counterpartyName": "Zafar",
    "principalAmount": 100000,
    "principalCurrency": "UZS",
    "startDate": "2026-01-10"
  }'
```

### Create Debt (They owe me)
```bash
curl -X POST http://localhost:9090/api/v1/debts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "direction": "they_owe_me",
    "counterpartyName": "Ahmad",
    "principalAmount": 500000,
    "principalCurrency": "UZS",
    "startDate": "2026-01-10",
    "dueDate": "2026-03-10"
  }'
```

### List All Debts
```bash
curl http://localhost:9090/api/v1/debts \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Make a Payment
```bash
curl -X POST http://localhost:9090/api/v1/debts/DEBT_ID/payments \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "currency": "UZS",
    "paymentDate": "2026-01-15"
  }'
```

---

## TypeScript/JavaScript Example

```typescript
// Create Debt API call
const createDebt = async (token: string) => {
  const response = await fetch('http://localhost:9090/api/v1/debts', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      direction: 'i_owe',           // REQUIRED: 'i_owe' or 'they_owe_me'
      counterpartyName: 'Zafar',    // REQUIRED: min 2 chars
      principalAmount: 100000,       // REQUIRED: > 0
      principalCurrency: 'UZS',      // REQUIRED
      startDate: '2026-01-10',       // REQUIRED: YYYY-MM-DD format
      // Optional fields:
      description: 'Loan for phone',
      dueDate: '2026-06-10',
      fundingAccountId: 'account-uuid',
    }),
  });

  return response.json();
};
```

---

## Common Mistakes

### ❌ Wrong
```json
{
  "Direction": "i_owe"  // Wrong: capital D
}
```

### ❌ Wrong
```json
{
  "direction": "I_OWE"  // Wrong: uppercase
}
```

### ❌ Wrong
```json
{
  "direction": "owe"  // Wrong: invalid value
}
```

### ✅ Correct
```json
{
  "direction": "i_owe"  // Correct: lowercase, exact match
}
```

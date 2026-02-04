# Debt API - Postman Examples

## Base URL
```
{{baseUrl}}/api/v1
```

## Headers (all requests)
```
Authorization: Bearer {{accessToken}}
Content-Type: application/json
```

---

## Counterparties

### 1. List Counterparties
```http
GET {{baseUrl}}/api/v1/counterparties?search=Ali
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "userId": "user-uuid",
      "displayName": "Ali Valiyev",
      "phoneNumber": "+998901234567",
      "comment": "Work colleague",
      "searchKeywords": null,
      "showStatus": "active",
      "createdAt": "2026-01-10T10:00:00Z",
      "updatedAt": "2026-01-10T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

### 2. Create Counterparty
```http
POST {{baseUrl}}/api/v1/counterparties
```

**Request Body:**
```json
{
  "displayName": "Bobur Karimov",
  "phoneNumber": "+998909876543",
  "comment": "Friend from university"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "userId": "user-uuid",
    "displayName": "Bobur Karimov",
    "phoneNumber": "+998909876543",
    "comment": "Friend from university",
    "searchKeywords": null,
    "showStatus": "active",
    "createdAt": "2026-01-10T12:00:00Z",
    "updatedAt": "2026-01-10T12:00:00Z"
  }
}
```

### 3. Get Single Counterparty
```http
GET {{baseUrl}}/api/v1/counterparties/550e8400-e29b-41d4-a716-446655440001
```

### 4. Update Counterparty
```http
PATCH {{baseUrl}}/api/v1/counterparties/550e8400-e29b-41d4-a716-446655440001
```

**Request Body:**
```json
{
  "displayName": "Ali Valiyev Jr",
  "phoneNumber": "+998901234568"
}
```

### 5. Delete Counterparty
```http
DELETE {{baseUrl}}/api/v1/counterparties/550e8400-e29b-41d4-a716-446655440001
```

**Success Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "deleted"
  }
}
```

**Error Response (has linked debts):**
```json
{
  "success": false,
  "error": {
    "code": -5013,
    "type": "CONFLICT",
    "message": "Cannot delete counterparty that has linked debts"
  }
}
```

### 6. Get Counterparty's Debts
```http
GET {{baseUrl}}/api/v1/counterparties/550e8400-e29b-41d4-a716-446655440001/debts
```

---

## Debts

### 1. List Debts
```http
GET {{baseUrl}}/api/v1/debts?direction=i_owe&status=active
```

**Query Parameters:**
- `direction`: `i_owe` | `they_owe_me`
- `status`: `active` | `paid` | `overdue` | `canceled`
- `linkedGoalId`: UUID of linked goal
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "debt-uuid-1",
      "userId": "user-uuid",
      "direction": "i_owe",
      "counterpartyId": "550e8400-e29b-41d4-a716-446655440001",
      "counterpartyName": "Ali Valiyev",
      "counterparty": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "displayName": "Ali Valiyev",
        "phoneNumber": "+998901234567",
        "comment": "Work colleague"
      },
      "principalAmount": 500000,
      "principalCurrency": "UZS",
      "baseCurrency": "UZS",
      "rateOnStart": 1,
      "principalBaseValue": 500000,
      "startDate": "2026-01-10",
      "dueDate": "2026-02-10",
      "status": "active",
      "remainingAmount": 500000,
      "totalPaid": 0,
      "percentPaid": 0,
      "showStatus": "active",
      "createdAt": "2026-01-10T10:00:00Z",
      "updatedAt": "2026-01-10T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

### 2. Create Debt - Option A: Reference Existing Counterparty
```http
POST {{baseUrl}}/api/v1/debts
```

**Request Body:**
```json
{
  "direction": "i_owe",
  "counterpartyId": "550e8400-e29b-41d4-a716-446655440001",
  "counterpartyName": "Ali Valiyev",
  "principalAmount": 1000000,
  "principalCurrency": "UZS",
  "baseCurrency": "UZS",
  "startDate": "2026-01-10",
  "dueDate": "2026-03-10",
  "description": "Borrowed for car repairs"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-debt-uuid",
    "userId": "user-uuid",
    "direction": "i_owe",
    "counterpartyId": "550e8400-e29b-41d4-a716-446655440001",
    "counterpartyName": "Ali Valiyev",
    "counterparty": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "displayName": "Ali Valiyev",
      "phoneNumber": "+998901234567",
      "comment": "Work colleague"
    },
    "description": "Borrowed for car repairs",
    "principalAmount": 1000000,
    "principalCurrency": "UZS",
    "baseCurrency": "UZS",
    "rateOnStart": 1,
    "principalBaseValue": 1000000,
    "startDate": "2026-01-10",
    "dueDate": "2026-03-10",
    "status": "active",
    "remainingAmount": 1000000,
    "totalPaid": 0,
    "percentPaid": 0,
    "showStatus": "active",
    "createdAt": "2026-01-10T14:00:00Z",
    "updatedAt": "2026-01-10T14:00:00Z"
  }
}
```

### 3. Create Debt - Option B: Inline Counterparty Creation
```http
POST {{baseUrl}}/api/v1/debts
```

**Request Body:**
```json
{
  "direction": "they_owe_me",
  "counterparty": {
    "displayName": "Sardor Rahimov",
    "phoneNumber": "+998901112233",
    "comment": "Neighbor"
  },
  "principalAmount": 200,
  "principalCurrency": "USD",
  "baseCurrency": "UZS",
  "rateOnStart": 12500,
  "startDate": "2026-01-10",
  "dueDate": "2026-02-01",
  "description": "Lent for emergency"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-debt-uuid-2",
    "userId": "user-uuid",
    "direction": "they_owe_me",
    "counterpartyId": "new-counterparty-uuid",
    "counterpartyName": "Sardor Rahimov",
    "counterparty": {
      "id": "new-counterparty-uuid",
      "displayName": "Sardor Rahimov",
      "phoneNumber": "+998901112233",
      "comment": "Neighbor"
    },
    "description": "Lent for emergency",
    "principalAmount": 200,
    "principalCurrency": "USD",
    "baseCurrency": "UZS",
    "rateOnStart": 12500,
    "principalBaseValue": 2500000,
    "startDate": "2026-01-10",
    "dueDate": "2026-02-01",
    "status": "active",
    "remainingAmount": 200,
    "totalPaid": 0,
    "percentPaid": 0,
    "showStatus": "active",
    "createdAt": "2026-01-10T14:30:00Z",
    "updatedAt": "2026-01-10T14:30:00Z"
  }
}
```

### 4. Create Debt - Option C: Quick with Name Only
```http
POST {{baseUrl}}/api/v1/debts
```

**Request Body:**
```json
{
  "direction": "i_owe",
  "counterpartyName": "John Doe",
  "principalAmount": 50,
  "principalCurrency": "USD",
  "startDate": "2026-01-10"
}
```

### 5. Get Single Debt
```http
GET {{baseUrl}}/api/v1/debts/debt-uuid-1
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "debt-uuid-1",
    "userId": "user-uuid",
    "direction": "i_owe",
    "counterpartyId": "550e8400-e29b-41d4-a716-446655440001",
    "counterpartyName": "Ali Valiyev",
    "counterparty": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "displayName": "Ali Valiyev",
      "phoneNumber": "+998901234567",
      "comment": "Work colleague"
    },
    "principalAmount": 500000,
    "principalCurrency": "UZS",
    "baseCurrency": "UZS",
    "rateOnStart": 1,
    "principalBaseValue": 500000,
    "startDate": "2026-01-10",
    "dueDate": "2026-02-10",
    "status": "active",
    "remainingAmount": 300000,
    "totalPaid": 200000,
    "percentPaid": 40,
    "showStatus": "active",
    "createdAt": "2026-01-10T10:00:00Z",
    "updatedAt": "2026-01-10T15:00:00Z"
  }
}
```

### 6. Update Debt
```http
PUT {{baseUrl}}/api/v1/debts/debt-uuid-1
```

**Request Body:**
```json
{
  "direction": "i_owe",
  "counterpartyId": "550e8400-e29b-41d4-a716-446655440001",
  "counterpartyName": "Ali Valiyev",
  "principalAmount": 500000,
  "principalCurrency": "UZS",
  "startDate": "2026-01-10",
  "dueDate": "2026-03-10",
  "description": "Extended due date by 1 month"
}
```

### 7. Patch Debt (Partial Update)
```http
PATCH {{baseUrl}}/api/v1/debts/debt-uuid-1
```

**Request Body:**
```json
{
  "dueDate": "2026-04-10",
  "reminderEnabled": true
}
```

### 8. Settle Debt
```http
POST {{baseUrl}}/api/v1/debts/debt-uuid-1/settle
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "debt-uuid-1",
    "status": "paid",
    "settledAt": "2026-01-15T16:00:00Z",
    "remainingAmount": 0,
    "totalPaid": 500000,
    "percentPaid": 100
  }
}
```

### 9. Extend Debt Due Date
```http
POST {{baseUrl}}/api/v1/debts/debt-uuid-1/extend
```

**Request Body:**
```json
{
  "dueDate": "2026-04-10"
}
```

### 10. Delete Debt
```http
DELETE {{baseUrl}}/api/v1/debts/debt-uuid-1
```

---

## Debt Payments

### 1. List Debt Payments
```http
GET {{baseUrl}}/api/v1/debts/debt-uuid-1/payments
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "payment-uuid-1",
      "debtId": "debt-uuid-1",
      "amount": 200000,
      "currency": "UZS",
      "baseCurrency": "UZS",
      "rateUsedToBase": 1,
      "convertedAmountToBase": 200000,
      "rateUsedToDebt": 1,
      "convertedAmountToDebt": 200000,
      "paymentDate": "2026-01-15",
      "accountId": "account-uuid",
      "note": "First payment",
      "relatedTransactionId": "transaction-uuid",
      "createdAt": "2026-01-15T10:00:00Z",
      "updatedAt": "2026-01-15T10:00:00Z"
    }
  ]
}
```

### 2. Create Debt Payment
```http
POST {{baseUrl}}/api/v1/debts/debt-uuid-1/payments
```

**Request Body:**
```json
{
  "amount": 100000,
  "currency": "UZS",
  "paymentDate": "2026-01-20",
  "accountId": "account-uuid",
  "note": "Second payment",
  "createTransaction": true
}
```

### 3. Update Debt Payment
```http
PATCH {{baseUrl}}/api/v1/debts/debt-uuid-1/payments/payment-uuid-1
```

**Request Body:**
```json
{
  "amount": 250000,
  "note": "Updated payment amount"
}
```

### 4. Delete Debt Payment
```http
DELETE {{baseUrl}}/api/v1/debts/debt-uuid-1/payments/payment-uuid-1
```

---

## Error Responses

### Validation Errors

**Missing Counterparty:**
```json
{
  "success": false,
  "error": {
    "code": -5010,
    "type": "VALIDATION",
    "message": "Counterparty is required for debt"
  }
}
```

**Counterparty Name Too Short:**
```json
{
  "success": false,
  "error": {
    "code": -5011,
    "type": "VALIDATION",
    "message": "Counterparty name must be at least 2 characters"
  }
}
```

**Invalid Direction:**
```json
{
  "success": false,
  "error": {
    "code": -5012,
    "type": "VALIDATION",
    "message": "Direction must be 'i_owe' or 'they_owe_me'"
  }
}
```

**Invalid Amount:**
```json
{
  "success": false,
  "error": {
    "code": -5016,
    "type": "VALIDATION",
    "message": "Principal amount must be greater than 0"
  }
}
```

**Invalid Due Date:**
```json
{
  "success": false,
  "error": {
    "code": -5015,
    "type": "VALIDATION",
    "message": "Due date must be on or after start date"
  }
}
```

### Not Found Errors

**Debt Not Found:**
```json
{
  "success": false,
  "error": {
    "code": -5003,
    "type": "NOT_FOUND",
    "message": "Debt not found"
  }
}
```

**Counterparty Not Found:**
```json
{
  "success": false,
  "error": {
    "code": -5005,
    "type": "NOT_FOUND",
    "message": "Counterparty not found"
  }
}
```

### Conflict Errors

**Counterparty Has Debts:**
```json
{
  "success": false,
  "error": {
    "code": -5013,
    "type": "CONFLICT",
    "message": "Cannot delete counterparty that has linked debts"
  }
}
```

---

## Postman Collection Variables

```json
{
  "baseUrl": "http://localhost:8080",
  "accessToken": "your-jwt-token-here"
}
```

## Postman Pre-request Script (Auto-refresh token)

```javascript
// Add to collection pre-request script
const tokenExpiry = pm.collectionVariables.get("tokenExpiry");
if (!tokenExpiry || Date.now() > parseInt(tokenExpiry)) {
    // Token expired, refresh it
    pm.sendRequest({
        url: pm.collectionVariables.get("baseUrl") + "/api/v1/auth/refresh",
        method: "POST",
        header: {
            "Content-Type": "application/json"
        },
        body: {
            mode: "raw",
            raw: JSON.stringify({
                refreshToken: pm.collectionVariables.get("refreshToken")
            })
        }
    }, function (err, res) {
        if (!err && res.code === 200) {
            const data = res.json().data;
            pm.collectionVariables.set("accessToken", data.accessToken);
            pm.collectionVariables.set("tokenExpiry", Date.now() + (data.expiresIn * 1000));
        }
    });
}
```

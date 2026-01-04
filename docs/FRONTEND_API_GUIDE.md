# LEORA Frontend API Guide

Bu hujjat frontend dasturchilar uchun backend API bilan ishlash bo'yicha to'liq qo'llanma.

**Base URL:** `http://localhost:9090/api/v1`

**Autentifikatsiya:** Bearer Token (Authorization header)

---

## 1. AUTHENTICATION

### 1.1 Register (Ro'yxatdan o'tish)

```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "fullName": "John Doe",
  "password": "SecurePass123!",
  "confirmPassword": "SecurePass123!",
  "region": "uzbekistan",
  "currency": "UZS"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "fullName": "John Doe",
      "region": "uzbekistan",
      "primaryCurrency": "UZS",
      "role": "user",
      "status": "active",
      "permissions": ["planner:read", "planner:write", "finance:read", "notifications:read", "widgets:read"],
      "createdAt": "2026-01-04T07:39:57Z",
      "updatedAt": "2026-01-04T07:39:57Z"
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "f7207647-fe64-4773-a4d8-216c85a4e6b2",
    "expiresIn": 604800
  }
}
```

**Region values:** `uzbekistan`, `united-states`, `eurozone`, `united-kingdom`, `turkey`, `saudi-arabia`, `united-arab-emirates`, `russia`, `other`

**Currency values:** `UZS`, `USD`, `EUR`, `GBP`, `TRY`, `SAR`, `AED`, `USDT`, `RUB`

---

### 1.2 Login (Kirish)

```http
POST /auth/login
Content-Type: application/json

{
  "emailOrUsername": "user@example.com",
  "password": "SecurePass123!",
  "rememberMe": true
}
```

**Response:** Register bilan bir xil format

---

### 1.3 Get Current User (Joriy foydalanuvchi)

```http
GET /auth/me
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "fullName": "John Doe",
      ...
    }
  }
}
```

---

### 1.4 Forgot Password (Parolni unutdim)

```http
POST /auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "OTP sent to email"
  }
}
```

**Eslatma:** OTP emailga yuboriladi. Hozircha development rejimda OTP console ga chiqadi.

---

### 1.5 Reset Password (Parolni tiklash)

```http
POST /auth/reset-password
Content-Type: application/json

{
  "email": "user@example.com",
  "otp": "123456",
  "newPassword": "NewSecurePass123!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "password reset"
  }
}
```

---

### 1.6 Refresh Token (Tokenni yangilash)

```http
POST /auth/refresh
Content-Type: application/json

{
  "refreshToken": "f7207647-fe64-4773-a4d8-216c85a4e6b2"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "accessToken": "new-access-token...",
    "refreshToken": "new-refresh-token",
    "expiresIn": 604800
  }
}
```

---

### 1.7 Logout (Chiqish)

```http
POST /auth/logout
Authorization: Bearer {accessToken}
```

---

## 2. TASKS (Vazifalar)

### 2.1 List Tasks

```http
GET /tasks
GET /tasks?page=1&limit=20
GET /tasks?status=inbox&priority=high
Authorization: Bearer {accessToken}
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Sahifa raqami (default: 1) |
| limit | int | Har sahifada (default: 20) |
| status | string | inbox, planned, in_progress, completed, canceled |
| priority | string | low, medium, high |
| showStatus | string | active, archived, deleted |
| goalId | uuid | Goal bo'yicha filter |

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Task title",
      "status": "inbox",
      "showStatus": "active",
      "priority": "medium",
      "goalId": null,
      "habitId": null,
      "financeLink": null,
      "progressValue": null,
      "progressUnit": null,
      "dueDate": "2026-01-15",
      "startDate": null,
      "timeOfDay": null,
      "estimatedMinutes": 60,
      "energyLevel": null,
      "context": null,
      "notes": "Some notes",
      "lastFocusSessionId": null,
      "focusTotalMinutes": 0,
      "checklist": [],
      "dependencies": [],
      "createdAt": "2026-01-04T07:47:30Z",
      "updatedAt": "2026-01-04T07:47:30Z"
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

---

### 2.2 Create Task

```http
POST /tasks
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "New Task",
  "status": "inbox",
  "priority": "medium",
  "dueDate": "2026-01-15",
  "estimatedMinutes": 60,
  "notes": "Task description",
  "goalId": "uuid (optional)",
  "habitId": "uuid (optional)",
  "checklist": [
    { "title": "Step 1", "completed": false },
    { "title": "Step 2", "completed": false }
  ]
}
```

**Status values:** `inbox`, `planned`, `in_progress`, `completed`, `canceled`, `moved`, `overdue`

**Priority values:** `low`, `medium`, `high`

**Finance Link values:** `record_expenses`, `pay_debt`, `review_budget`, `transfer_money`, `none`

---

### 2.3 Get Task by ID

```http
GET /tasks/{id}
Authorization: Bearer {accessToken}
```

---

### 2.4 Update Task (Full)

```http
PUT /tasks/{id}
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Updated Task",
  "status": "in_progress",
  "priority": "high",
  ...
}
```

---

### 2.5 Update Task (Partial)

```http
PATCH /tasks/{id}
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "status": "in_progress",
  "priority": "high"
}
```

---

### 2.6 Complete Task

```http
POST /tasks/{id}/complete
Authorization: Bearer {accessToken}
```

**Response:** Status `completed` ga o'zgaradi

---

### 2.7 Reopen Task

```http
POST /tasks/{id}/reopen
Authorization: Bearer {accessToken}
```

**Response:** Status `inbox` ga qaytadi

---

### 2.8 Update Checklist Item

```http
PATCH /tasks/{taskId}/checklist/{itemId}
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "completed": true
}
```

---

### 2.9 Delete Task

```http
DELETE /tasks/{id}
Authorization: Bearer {accessToken}
```

---

## 3. GOALS (Maqsadlar)

### 3.1 List Goals

```http
GET /goals
GET /goals?page=1&limit=20
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Learn Go Programming",
      "description": "Master Go for backend",
      "goalType": "education",
      "status": "active",
      "showStatus": "active",
      "metricType": "count",
      "direction": "increase",
      "unit": "hours",
      "initialValue": 0,
      "targetValue": 100,
      "progressTargetValue": null,
      "currentValue": 25,
      "financeMode": null,
      "currency": null,
      "linkedBudgetId": null,
      "linkedDebtId": null,
      "startDate": "2026-01-01",
      "targetDate": "2026-06-30",
      "completedDate": null,
      "progressPercent": 25,
      "milestones": [],
      "stats": null,
      "createdAt": "2026-01-04T07:47:30Z",
      "updatedAt": "2026-01-04T07:47:30Z"
    }
  ],
  "meta": {...}
}
```

---

### 3.2 Create Goal

```http
POST /goals
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Learn Go Programming",
  "description": "Master Go for backend development",
  "goalType": "education",
  "status": "active",
  "metricType": "count",
  "direction": "increase",
  "unit": "hours",
  "initialValue": 0,
  "targetValue": 100,
  "currentValue": 0,
  "startDate": "2026-01-01",
  "targetDate": "2026-06-30",
  "milestones": [
    { "title": "Basic syntax", "targetPercent": 25 },
    { "title": "Concurrency", "targetPercent": 50 },
    { "title": "Web frameworks", "targetPercent": 75 },
    { "title": "Production ready", "targetPercent": 100 }
  ]
}
```

**Goal Type values:** `financial`, `health`, `education`, `productivity`, `personal`

**Status values:** `active`, `paused`, `completed`, `archived`

**Metric Type values:** `none`, `amount`, `weight`, `count`, `duration`, `custom`

**Direction values:** `increase`, `decrease`, `neutral`

**Finance Mode values:** `save`, `spend`, `debt_close`

---

### 3.3 Create Financial Goal

```http
POST /goals
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Save for Car",
  "goalType": "financial",
  "status": "active",
  "metricType": "amount",
  "direction": "increase",
  "financeMode": "save",
  "currency": "USD",
  "initialValue": 0,
  "targetValue": 15000,
  "currentValue": 2500,
  "targetDate": "2026-12-31"
}
```

---

### 3.4 Update Goal Progress

```http
PATCH /goals/{id}
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "currentValue": 50,
  "progressPercent": 50
}
```

---

### 3.5 Get/Delete Goal

```http
GET /goals/{id}
DELETE /goals/{id}
Authorization: Bearer {accessToken}
```

---

## 4. HABITS (Odatlar)

### 4.1 List Habits

```http
GET /habits
GET /habits?page=1&limit=20
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Morning Exercise",
      "description": "30 min workout",
      "iconId": null,
      "habitType": "health",
      "status": "active",
      "showStatus": "active",
      "goalId": null,
      "frequency": "daily",
      "daysOfWeek": [],
      "timesPerWeek": null,
      "timeOfDay": "06:00",
      "completionMode": "boolean",
      "targetPerDay": null,
      "unit": null,
      "countingType": "create",
      "difficulty": "medium",
      "priority": "high",
      "challengeLengthDays": null,
      "reminderEnabled": true,
      "reminderTime": "05:45",
      "streakCurrent": 5,
      "streakBest": 10,
      "completionRate30d": 85.5,
      "financeRule": null,
      "linkedGoalIds": [],
      "createdAt": "2026-01-04T07:48:12Z",
      "updatedAt": "2026-01-04T07:48:12Z"
    }
  ],
  "meta": {...}
}
```

---

### 4.2 Create Habit (Boolean/Daily)

```http
POST /habits
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Wake up early",
  "description": "Wake up at 6:00 AM",
  "habitType": "health",
  "status": "active",
  "frequency": "daily",
  "completionMode": "boolean",
  "countingType": "create",
  "difficulty": "medium",
  "priority": "high",
  "timeOfDay": "06:00",
  "reminderEnabled": true,
  "reminderTime": "05:45"
}
```

**Habit Type values:** `health`, `finance`, `productivity`, `education`, `personal`, `custom`

**Status values:** `active`, `paused`, `archived`

**Frequency values:** `daily`, `weekly`, `custom`

**Completion Mode values:** `boolean`, `numeric`

**Counting Type values:** `create`, `quit`

**Difficulty values:** `easy`, `medium`, `hard`

**Priority values:** `low`, `medium`, `high`

---

### 4.3 Create Habit (Numeric/Weekly)

```http
POST /habits
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Running",
  "description": "Run 5km 3 times a week",
  "habitType": "health",
  "status": "active",
  "frequency": "weekly",
  "daysOfWeek": [1, 3, 5],
  "timesPerWeek": 3,
  "completionMode": "numeric",
  "targetPerDay": 5,
  "unit": "km",
  "countingType": "create",
  "difficulty": "hard",
  "priority": "high"
}
```

**daysOfWeek:** 0=Sunday, 1=Monday, ..., 6=Saturday

---

### 4.4 Create Finance Habit (No Spend)

```http
POST /habits
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "No fast food",
  "habitType": "finance",
  "status": "active",
  "frequency": "daily",
  "completionMode": "boolean",
  "countingType": "quit",
  "difficulty": "hard",
  "financeRule": {
    "type": "no_spend_in_categories",
    "categoryIds": ["fast-food", "entertainment"]
  }
}
```

**Finance Rule Types:**
- `no_spend_in_categories` - Bu kategoriyalarda xarajat qilmaslik
- `spend_in_categories` - Bu kategoriyalarda xarajat qilish
- `has_any_transactions` - Har qanday tranzaksiya bo'lishi
- `daily_spend_under` - Kunlik xarajat limitdan kam

---

### 4.5 Create Finance Habit (Daily Limit)

```http
POST /habits
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Daily spending limit",
  "habitType": "finance",
  "status": "active",
  "frequency": "daily",
  "completionMode": "boolean",
  "countingType": "quit",
  "financeRule": {
    "type": "daily_spend_under",
    "amount": 100000,
    "currency": "UZS"
  }
}
```

---

### 4.6 Complete Habit

```http
POST /habits/{id}/complete
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "dateKey": "2026-01-04",
  "status": "done"
}
```

**Status values:** `done`, `miss`

**Numeric habit uchun:**
```json
{
  "dateKey": "2026-01-04",
  "status": "done",
  "value": 5.5
}
```

---

### 4.7 Get Habit History

```http
GET /habits/{id}/history
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "habitId": "uuid",
      "dateKey": "2026-01-04",
      "status": "done",
      "value": null,
      "createdAt": "2026-01-04T07:54:20Z"
    },
    {
      "id": "uuid",
      "habitId": "uuid",
      "dateKey": "2026-01-03",
      "status": "miss",
      "value": null,
      "createdAt": "2026-01-03T23:59:59Z"
    }
  ]
}
```

---

### 4.8 Get Habit Stats

```http
GET /habits/{id}/stats
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "streakCurrent": 5,
    "streakBest": 10,
    "completionRate30d": 85.5,
    "totalCompletions": 25,
    "totalMisses": 5
  }
}
```

---

### 4.9 Get/Update/Delete Habit

```http
GET /habits/{id}
PATCH /habits/{id}
DELETE /habits/{id}
Authorization: Bearer {accessToken}
```

---

## 5. FOCUS SESSIONS (Fokus Sessiyalari)

### 5.1 List Focus Sessions

```http
GET /focus-sessions
GET /focus-sessions?page=1&limit=20
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "taskId": null,
      "goalId": null,
      "plannedMinutes": 25,
      "actualMinutes": 20,
      "status": "completed",
      "startedAt": "2026-01-04T07:48:12Z",
      "endedAt": "2026-01-04T08:13:12Z",
      "interruptionsCount": 1,
      "notes": "Productive session",
      "createdAt": "2026-01-04T07:48:12Z",
      "updatedAt": "2026-01-04T08:13:12Z"
    }
  ],
  "meta": {...}
}
```

---

### 5.2 Start Focus Session

```http
POST /focus-sessions/start
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "plannedMinutes": 25
}
```

**Task bilan:**
```json
{
  "taskId": "uuid",
  "plannedMinutes": 25
}
```

**Goal bilan:**
```json
{
  "goalId": "uuid",
  "plannedMinutes": 50
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "plannedMinutes": 25,
    "actualMinutes": 0,
    "status": "in_progress",
    "startedAt": "2026-01-04T07:48:12Z",
    "interruptionsCount": 0,
    "createdAt": "2026-01-04T07:48:12Z",
    "updatedAt": "2026-01-04T07:48:12Z"
  }
}
```

---

### 5.3 Pause Session

```http
PATCH /focus-sessions/{id}/pause
Authorization: Bearer {accessToken}
```

**Response:** Status `paused` ga o'zgaradi

---

### 5.4 Resume Session

```http
PATCH /focus-sessions/{id}/resume
Authorization: Bearer {accessToken}
```

**Response:** Status `in_progress` ga qaytadi

---

### 5.5 Add Interruption

```http
POST /focus-sessions/{id}/interrupt
Authorization: Bearer {accessToken}
```

**Response:** `interruptionsCount` 1 ga oshadi

---

### 5.6 Complete Session

```http
POST /focus-sessions/{id}/complete
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "actualMinutes": 25,
  "notes": "Productive session!"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "plannedMinutes": 25,
    "actualMinutes": 25,
    "status": "completed",
    "startedAt": "2026-01-04T07:48:12Z",
    "endedAt": "2026-01-04T08:13:12Z",
    "interruptionsCount": 1,
    "notes": "Productive session!",
    ...
  }
}
```

---

### 5.7 Cancel Session

```http
POST /focus-sessions/{id}/cancel
Authorization: Bearer {accessToken}
```

**Response:** Status `canceled` ga o'zgaradi

---

### 5.8 Get Focus Stats

```http
GET /focus-sessions/stats
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "totalSessions": 10,
    "totalMinutes": 200,
    "completedSessions": 8,
    "averageMinutes": 25,
    "totalInterruptions": 5
  }
}
```

---

### 5.9 Get/Delete Focus Session

```http
GET /focus-sessions/{id}
DELETE /focus-sessions/{id}
Authorization: Bearer {accessToken}
```

---

## 6. FINANCE - ACCOUNTS

### 6.1 List Accounts

```http
GET /accounts
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Cash Wallet",
      "currency": "UZS",
      "accountType": "cash",
      "createdAt": "2026-01-04T07:45:29Z",
      "updatedAt": "2026-01-04T07:45:29Z"
    }
  ],
  "meta": {...}
}
```

---

### 6.2 Create Account

```http
POST /accounts
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "name": "Cash Wallet",
  "currency": "UZS",
  "type": "cash"
}
```

**Account Type values:** `cash`, `card`, `savings`, `investment`, `credit`, `debt`, `other`

---

### 6.3 Get/Update/Delete Account

```http
GET /accounts/{id}
PUT /accounts/{id}
PATCH /accounts/{id}
DELETE /accounts/{id}
Authorization: Bearer {accessToken}
```

---

## 7. FINANCE - TRANSACTIONS

### 7.1 List Transactions

```http
GET /transactions
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "accountId": "uuid",
      "amount": 50000,
      "currency": "UZS",
      "category": "food",
      "createdAt": "2026-01-04T07:45:29Z",
      "updatedAt": "2026-01-04T07:45:29Z"
    }
  ],
  "meta": {...}
}
```

---

### 7.2 Create Transaction

```http
POST /transactions
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "accountId": "uuid",
  "amount": 50000,
  "currency": "UZS",
  "category": "food"
}
```

---

### 7.3 Get/Update/Delete Transaction

```http
GET /transactions/{id}
PUT /transactions/{id}
PATCH /transactions/{id}
DELETE /transactions/{id}
Authorization: Bearer {accessToken}
```

---

## 8. FINANCE - BUDGETS

### 8.1 List Budgets

```http
GET /budgets
Authorization: Bearer {accessToken}
```

---

### 8.2 Create Budget

```http
POST /budgets
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "name": "Food Budget",
  "currency": "UZS",
  "limit": 1000000
}
```

---

### 8.3 Get/Update/Delete Budget

```http
GET /budgets/{id}
PATCH /budgets/{id}
DELETE /budgets/{id}
Authorization: Bearer {accessToken}
```

---

## 9. FINANCE - DEBTS

### 9.1 List Debts

```http
GET /debts
Authorization: Bearer {accessToken}
```

---

### 9.2 Create Debt

```http
POST /debts
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "name": "Car Loan",
  "balance": 5000000
}
```

---

### 9.3 Get/Update/Delete Debt

```http
GET /debts/{id}
PATCH /debts/{id}
DELETE /debts/{id}
Authorization: Bearer {accessToken}
```

---

## 10. NOTIFICATIONS

### 10.1 List Notifications

```http
GET /notifications
Authorization: Bearer {accessToken}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "Reminder",
      "message": "Don't forget your task!",
      "createdAt": "2026-01-04T07:45:29Z",
      "updatedAt": "2026-01-04T07:45:29Z"
    }
  ],
  "meta": {...}
}
```

---

### 10.2 Create Notification

```http
POST /notifications
Authorization: Bearer {accessToken}
Content-Type: application/json

{
  "title": "Reminder",
  "message": "Don't forget your task!"
}
```

---

### 10.3 Get/Delete Notification

```http
GET /notifications/{id}
DELETE /notifications/{id}
Authorization: Bearer {accessToken}
```

---

## ERROR HANDLING

### Error Response Format

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": -2000,
    "message": "User not found",
    "type": "NOT_FOUND"
  }
}
```

### Error Codes

| Code | Type | Description |
|------|------|-------------|
| -1000 | INTERNAL | Internal server error |
| -2000 | NOT_FOUND | User not found |
| -2001 | CONFLICT | User already exists |
| -2002 | BAD_REQUEST | Invalid user data |
| -2003 | UNAUTHORIZED | Invalid credentials |
| -2004 | UNAUTHORIZED | Invalid token |
| -3000 | NOT_FOUND | Task not found |
| -3001 | NOT_FOUND | Goal not found |
| -3002 | NOT_FOUND | Habit not found |
| -3003 | NOT_FOUND | Focus session not found |
| -3004 | BAD_REQUEST | Invalid planner data |
| -5000 | NOT_FOUND | Account not found |
| -5001 | NOT_FOUND | Transaction not found |
| -5002 | NOT_FOUND | Budget not found |
| -5003 | NOT_FOUND | Debt not found |
| -6000 | NOT_FOUND | Notification not found |
| -9001 | INTERNAL | Database error |

---

## PAGINATION

Barcha list endpointlari pagination qo'llab-quvvatlaydi:

**Request:**
```http
GET /tasks?page=2&limit=10
```

**Response Meta:**
```json
{
  "meta": {
    "page": 2,
    "limit": 10,
    "total": 45,
    "totalPages": 5
  }
}
```

---

## TYPESCRIPT TYPES

```typescript
// Auth
interface User {
  id: string;
  email: string;
  fullName: string;
  region: string;
  primaryCurrency: string;
  role: 'user' | 'admin' | 'premium';
  status: 'active' | 'suspended' | 'deleted';
  permissions: string[];
  createdAt: string;
  updatedAt: string;
}

interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

// Task
interface Task {
  id: string;
  title: string;
  status: 'inbox' | 'planned' | 'in_progress' | 'completed' | 'canceled' | 'moved' | 'overdue';
  showStatus: 'active' | 'archived' | 'deleted';
  priority: 'low' | 'medium' | 'high';
  goalId?: string;
  habitId?: string;
  financeLink?: 'record_expenses' | 'pay_debt' | 'review_budget' | 'transfer_money' | 'none';
  progressValue?: number;
  progressUnit?: string;
  dueDate?: string;
  startDate?: string;
  timeOfDay?: string;
  estimatedMinutes?: number;
  energyLevel?: number;
  context?: string;
  notes?: string;
  lastFocusSessionId?: string;
  focusTotalMinutes: number;
  checklist: ChecklistItem[];
  dependencies: TaskDependency[];
  createdAt: string;
  updatedAt: string;
}

interface ChecklistItem {
  id: string;
  taskId: string;
  title: string;
  completed: boolean;
  order: number;
}

// Goal
interface Goal {
  id: string;
  title: string;
  description?: string;
  goalType: 'financial' | 'health' | 'education' | 'productivity' | 'personal';
  status: 'active' | 'paused' | 'completed' | 'archived';
  showStatus: 'active' | 'archived' | 'deleted';
  metricType: 'none' | 'amount' | 'weight' | 'count' | 'duration' | 'custom';
  direction: 'increase' | 'decrease' | 'neutral';
  unit?: string;
  initialValue?: number;
  targetValue?: number;
  progressTargetValue?: number;
  currentValue: number;
  financeMode?: 'save' | 'spend' | 'debt_close';
  currency?: string;
  linkedBudgetId?: string;
  linkedDebtId?: string;
  startDate?: string;
  targetDate?: string;
  completedDate?: string;
  progressPercent: number;
  milestones: Milestone[];
  stats?: GoalStats;
  createdAt: string;
  updatedAt: string;
}

// Habit
interface Habit {
  id: string;
  title: string;
  description?: string;
  iconId?: string;
  habitType: 'health' | 'finance' | 'productivity' | 'education' | 'personal' | 'custom';
  status: 'active' | 'paused' | 'archived';
  showStatus: 'active' | 'archived' | 'deleted';
  goalId?: string;
  frequency: 'daily' | 'weekly' | 'custom';
  daysOfWeek: number[];
  timesPerWeek?: number;
  timeOfDay?: string;
  completionMode: 'boolean' | 'numeric';
  targetPerDay?: number;
  unit?: string;
  countingType: 'create' | 'quit';
  difficulty: 'easy' | 'medium' | 'hard';
  priority: 'low' | 'medium' | 'high';
  challengeLengthDays?: number;
  reminderEnabled: boolean;
  reminderTime?: string;
  streakCurrent: number;
  streakBest: number;
  completionRate30d: number;
  financeRule?: FinanceRule;
  linkedGoalIds: string[];
  createdAt: string;
  updatedAt: string;
}

interface FinanceRule {
  type: 'no_spend_in_categories' | 'spend_in_categories' | 'has_any_transactions' | 'daily_spend_under';
  categoryIds?: string[];
  accountIds?: string[];
  minAmount?: number;
  amount?: number;
  currency?: string;
}

interface HabitCompletion {
  id: string;
  habitId: string;
  dateKey: string;
  status: 'done' | 'miss';
  value?: number;
  createdAt: string;
}

interface HabitStats {
  streakCurrent: number;
  streakBest: number;
  completionRate30d: number;
  totalCompletions: number;
  totalMisses: number;
}

// Focus Session
interface FocusSession {
  id: string;
  taskId?: string;
  goalId?: string;
  plannedMinutes: number;
  actualMinutes: number;
  status: 'in_progress' | 'completed' | 'canceled' | 'paused';
  startedAt?: string;
  endedAt?: string;
  interruptionsCount: number;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

interface FocusStats {
  totalSessions: number;
  totalMinutes: number;
  completedSessions: number;
  averageMinutes: number;
  totalInterruptions: number;
}

// Finance
interface Account {
  id: string;
  name: string;
  currency: string;
  accountType: 'cash' | 'card' | 'savings' | 'investment' | 'credit' | 'debt' | 'other';
  createdAt: string;
  updatedAt: string;
}

interface Transaction {
  id: string;
  accountId: string;
  amount: number;
  currency: string;
  category: string;
  createdAt: string;
  updatedAt: string;
}

interface Budget {
  id: string;
  name: string;
  currency: string;
  limit: number;
  createdAt: string;
  updatedAt: string;
}

interface Debt {
  id: string;
  name: string;
  balance: number;
  createdAt: string;
  updatedAt: string;
}

// API Response
interface ApiResponse<T> {
  success: boolean;
  data: T | null;
  error: ApiError | null;
  meta: PaginationMeta | null;
}

interface ApiError {
  code: number;
  message: string;
  type: string;
}

interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}
```

---

## REACT/REACT NATIVE EXAMPLE

```typescript
// api.ts
const BASE_URL = 'http://localhost:9090/api/v1';

class ApiClient {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
  }

  clearToken() {
    this.token = null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    return response.json();
  }

  // Auth
  async login(email: string, password: string) {
    const response = await this.request<{
      user: User;
      accessToken: string;
      refreshToken: string;
      expiresIn: number;
    }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ emailOrUsername: email, password, rememberMe: true }),
    });

    if (response.success && response.data) {
      this.setToken(response.data.accessToken);
    }

    return response;
  }

  // Tasks
  async getTasks(page = 1, limit = 20) {
    return this.request<Task[]>(`/tasks?page=${page}&limit=${limit}`);
  }

  async createTask(task: Partial<Task>) {
    return this.request<Task>('/tasks', {
      method: 'POST',
      body: JSON.stringify(task),
    });
  }

  async completeTask(id: string) {
    return this.request<Task>(`/tasks/${id}/complete`, { method: 'POST' });
  }

  // Habits
  async getHabits() {
    return this.request<Habit[]>('/habits');
  }

  async completeHabit(id: string, dateKey: string, status: 'done' | 'miss', value?: number) {
    return this.request<HabitCompletion>(`/habits/${id}/complete`, {
      method: 'POST',
      body: JSON.stringify({ dateKey, status, value }),
    });
  }

  async getHabitStats(id: string) {
    return this.request<HabitStats>(`/habits/${id}/stats`);
  }

  // Focus Sessions
  async startFocusSession(plannedMinutes: number, taskId?: string, goalId?: string) {
    return this.request<FocusSession>('/focus-sessions/start', {
      method: 'POST',
      body: JSON.stringify({ plannedMinutes, taskId, goalId }),
    });
  }

  async pauseFocusSession(id: string) {
    return this.request<FocusSession>(`/focus-sessions/${id}/pause`, { method: 'PATCH' });
  }

  async resumeFocusSession(id: string) {
    return this.request<FocusSession>(`/focus-sessions/${id}/resume`, { method: 'PATCH' });
  }

  async completeFocusSession(id: string, actualMinutes: number, notes?: string) {
    return this.request<FocusSession>(`/focus-sessions/${id}/complete`, {
      method: 'POST',
      body: JSON.stringify({ actualMinutes, notes }),
    });
  }

  async getFocusStats() {
    return this.request<FocusStats>('/focus-sessions/stats');
  }
}

export const api = new ApiClient();
```

---

## POSTMAN COLLECTION

To'liq ishlaydigan Postman collection: `docs/leora_postman_collection_v4.json`

Import qilish:
1. Postman ochish
2. Import -> Upload Files
3. `leora_postman_collection_v4.json` tanlash
4. Import

Foydalanish:
1. "Register" yoki "Login" so'rovini yuborish
2. Token avtomatik saqlanadi
3. Boshqa so'rovlar avtomatik token bilan yuboriladi

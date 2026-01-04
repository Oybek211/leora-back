# LEORA BACKEND API - TO'LIQ DOKUMENTATSIYA

> **Maqsad**: Leora mobile app uchun multi-tenant backend API. Bu hujjat barcha entitylar, fieldlar, relationshiplar va API endpointlarini o'z ichiga oladi.

---

# MUNDARIJA

1. [Umumiy Ma'lumot](#1-umumiy-malumot)
2. [Authentication & Users](#2-authentication--users)
3. [Planner Module](#3-planner-module)
   - 3.1 Tasks
   - 3.2 Goals
   - 3.3 Habits
   - 3.4 Focus Sessions
4. [Finance Module](#4-finance-module)
   - 4.1 Accounts
   - 4.2 Transactions
   - 4.3 Budgets
   - 4.4 Debts
   - 4.5 FX Rates
   - 4.6 Counterparties
5. [Insights Module](#5-insights-module)
6. [Sync & Realtime](#6-sync--realtime)
7. [Entity Relationships Diagrammasi](#7-entity-relationships)

---

# 1. UMUMIY MA'LUMOT

## 1.1 Texnologiyalar (Tavsiya)
- **Backend Framework**: Node.js (NestJS) yoki Python (FastAPI)
- **Database**: PostgreSQL (asosiy) + Redis (caching)
- **Real-time**: WebSocket (Socket.io)
- **Authentication**: JWT + Refresh Token
- **File Storage**: S3 compatible (attachments uchun)

## 1.2 Base URL
```
Production: https://api.leora.app/v1
Staging: https://staging-api.leora.app/v1
```

## 1.3 Umumiy Response Format
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "totalPages": 5
  },
  "error": null
}
```

## 1.4 Error Response Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      { "field": "email", "message": "Invalid email format" }
    ]
  }
}
```

## 1.5 Umumiy Fieldlar (Barcha Entitylarda)
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `userId` | UUID | Owner user ID (FK → users.id) |
| `showStatus` | enum | `active`, `archived`, `deleted` |
| `syncStatus` | enum | `local`, `synced`, `pending`, `conflict` |
| `idempotencyKey` | string? | Duplicate prevention key |
| `createdAt` | timestamp | Yaratilgan vaqt |
| `updatedAt` | timestamp | Oxirgi yangilangan vaqt |

---

# 2. AUTHENTICATION & USERS

## 2.1 User Entity

### Fieldlar
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `email` | string | ✅ | Unique email |
| `fullName` | string | ✅ | To'liq ism |
| `username` | string | ❌ | Unique username |
| `phoneNumber` | string | ❌ | Telefon raqami |
| `passwordHash` | string | ✅ | Bcrypt hashed password |
| `bio` | string | ❌ | Bio/description |
| `birthday` | date | ❌ | Tug'ilgan kun |
| `profileImage` | string | ❌ | Profile rasm URL |
| `region` | enum | ✅ | `uzbekistan`, `russia`, `turkey`, `saudi_arabia`, `uae`, `other` |
| `primaryCurrency` | string | ✅ | Asosiy valyuta (UZS, USD, RUB, etc.) |
| `visibility` | enum | ❌ | `public`, `friends`, `private` |
| `preferences` | JSON | ❌ | User preferences |
| `isEmailVerified` | boolean | ✅ | Email tasdiqlangan |
| `isPhoneVerified` | boolean | ✅ | Telefon tasdiqlangan |
| `lastLoginAt` | timestamp | ❌ | Oxirgi login |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### User Preferences (JSON)
```json
{
  "showLevel": true,
  "showAchievements": true,
  "showStatistics": true,
  "language": "uz",
  "theme": "system",
  "notifications": {
    "push": true,
    "email": true,
    "reminders": true
  }
}
```

## 2.2 Auth Endpoints

### POST /auth/register
User ro'yxatdan o'tkazish
```json
// Request
{
  "email": "user@example.com",
  "fullName": "Alisher Karimov",
  "password": "SecurePass123!",
  "confirmPassword": "SecurePass123!",
  "region": "uzbekistan",
  "currency": "UZS"
}

// Response
{
  "success": true,
  "data": {
    "user": { ... },
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresIn": 3600
  }
}
```

### POST /auth/login
Login
```json
// Request
{
  "emailOrUsername": "user@example.com",
  "password": "SecurePass123!",
  "rememberMe": true
}

// Response
{
  "success": true,
  "data": {
    "user": { ... },
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresIn": 3600
  }
}
```

### POST /auth/refresh
Token yangilash
```json
// Request
{
  "refreshToken": "eyJhbGc..."
}

// Response
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresIn": 3600
  }
}
```

### POST /auth/forgot-password
Parolni tiklash
```json
// Request
{
  "email": "user@example.com"
}

// Response
{
  "success": true,
  "data": {
    "message": "OTP sent to email"
  }
}
```

### POST /auth/reset-password
Yangi parol o'rnatish
```json
// Request
{
  "email": "user@example.com",
  "otp": "123456",
  "newPassword": "NewSecurePass123!"
}
```

### POST /auth/logout
Logout (token invalidate)

### GET /auth/me
Joriy user ma'lumotlari

### PATCH /users/me
User profilni yangilash

---

# 3. PLANNER MODULE

## 3.1 TASKS

### Task Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `title` | string | ✅ | Task nomi (max 500) |
| `status` | enum | ✅ | Task holati |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `priority` | enum | ✅ | `low`, `medium`, `high` |
| `goalId` | UUID | ❌ | FK → goals.id |
| `habitId` | UUID | ❌ | FK → habits.id |
| `financeLink` | enum | ❌ | Finance bog'lanish turi |
| `progressValue` | decimal | ❌ | Progress qiymati |
| `progressUnit` | string | ❌ | Progress birligi |
| `dueDate` | date | ❌ | Tugash sanasi |
| `startDate` | date | ❌ | Boshlanish sanasi |
| `timeOfDay` | string | ❌ | Vaqt (HH:mm format) |
| `estimatedMinutes` | integer | ❌ | Taxminiy vaqt (daqiqa) |
| `energyLevel` | integer | ❌ | 1-3 (past-o'rta-yuqori) |
| `context` | string | ❌ | Kontekst (uy, ish, etc.) |
| `notes` | text | ❌ | Qo'shimcha izohlar |
| `lastFocusSessionId` | UUID | ❌ | FK → focus_sessions.id |
| `focusTotalMinutes` | integer | ✅ | Jami focus vaqti (default: 0) |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Task Status Values
```
inbox        - Yangi, rejalashtirilmagan
planned      - Rejalashtirilgan
in_progress  - Bajarilmoqda
completed    - Tugallangan
canceled     - Bekor qilingan
moved        - Boshqa kunga ko'chirilgan
overdue      - Muddati o'tgan
```

### Task Finance Link Values
```
record_expenses  - Xarajat yozish
pay_debt         - Qarz to'lash
review_budget    - Byudjetni ko'rish
transfer_money   - Pul o'tkazish
none             - Bog'lanmagan
```

### Task Checklist Item (Embedded/Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `taskId` | UUID | ✅ | FK → tasks.id |
| `title` | string | ✅ | Subtask nomi |
| `completed` | boolean | ✅ | Bajarilganmi (default: false) |
| `order` | integer | ✅ | Tartib raqami |

### Task Dependency (Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `taskId` | UUID | ✅ | FK → tasks.id (dependent task) |
| `dependsOnTaskId` | UUID | ✅ | FK → tasks.id (prerequisite) |
| `status` | enum | ✅ | `pending`, `met` |

### Tasks API Endpoints

#### GET /tasks
Tasklar ro'yxati (paginated, filtered)

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `page` | integer | Sahifa (default: 1) |
| `limit` | integer | Limit (default: 20, max: 100) |
| `status` | string | Filter by status |
| `showStatus` | string | `active`, `archived` |
| `priority` | string | Filter by priority |
| `goalId` | UUID | Filter by goal |
| `habitId` | UUID | Filter by habit |
| `dueDate` | date | Filter by due date |
| `dueDateFrom` | date | Due date dan |
| `dueDateTo` | date | Due date gacha |
| `search` | string | Title bo'yicha qidirish |
| `sortBy` | string | `createdAt`, `dueDate`, `priority` |
| `sortOrder` | string | `asc`, `desc` |

#### GET /tasks/:id
Bitta task

#### POST /tasks
Yangi task yaratish
```json
{
  "title": "Kitob o'qish",
  "priority": "medium",
  "dueDate": "2024-12-25",
  "timeOfDay": "09:00",
  "estimatedMinutes": 60,
  "energyLevel": 2,
  "goalId": "uuid-goal-id",
  "context": "home",
  "checklist": [
    { "title": "1-bob", "completed": false },
    { "title": "2-bob", "completed": false }
  ]
}
```

#### PATCH /tasks/:id
Task yangilash

#### DELETE /tasks/:id
Task o'chirish (soft delete → showStatus = 'deleted')

#### POST /tasks/:id/complete
Taskni tugallash

#### POST /tasks/:id/reopen
Taskni qayta ochish

#### PATCH /tasks/:id/checklist/:itemId
Checklist item yangilash

---

## 3.2 GOALS

### Goal Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `title` | string | ✅ | Goal nomi |
| `description` | text | ❌ | Tavsif |
| `goalType` | enum | ✅ | Goal turi |
| `status` | enum | ✅ | `active`, `paused`, `completed`, `archived` |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `metricType` | enum | ✅ | O'lchov turi |
| `direction` | enum | ✅ | `increase`, `decrease`, `neutral` |
| `unit` | string | ❌ | O'lchov birligi |
| `initialValue` | decimal | ❌ | Boshlang'ich qiymat |
| `targetValue` | decimal | ❌ | Maqsad qiymat |
| `progressTargetValue` | decimal | ❌ | Progress target |
| `currentValue` | decimal | ✅ | Joriy qiymat (default: 0) |
| `financeMode` | enum | ❌ | `save`, `spend`, `debt_close` |
| `currency` | string | ❌ | Valyuta (financial goals uchun) |
| `linkedBudgetId` | UUID | ❌ | FK → budgets.id |
| `linkedDebtId` | UUID | ❌ | FK → debts.id |
| `startDate` | date | ❌ | Boshlanish sanasi |
| `targetDate` | date | ❌ | Maqsad sanasi |
| `completedDate` | date | ❌ | Tugallangan sana |
| `progressPercent` | decimal | ✅ | Progress % (0-100) |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Goal Type Values
```
financial    - Moliyaviy maqsad
health       - Sog'liq maqsadi
education    - Ta'lim maqsadi
productivity - Samaradorlik maqsadi
personal     - Shaxsiy maqsad
```

### Metric Type Values
```
none     - O'lchovsiz
amount   - Pul miqdori
weight   - Vazn
count    - Son
duration - Vaqt davomiyligi
custom   - Maxsus
```

### Goal Stats (Embedded JSON yoki alohida table)
| Field | Type | Description |
|-------|------|-------------|
| `financialProgressPercent` | decimal | Moliyaviy progress % |
| `habitsProgressPercent` | decimal | Habitlar progress % |
| `tasksProgressPercent` | decimal | Tasklar progress % |
| `focusMinutesLast30` | integer | Oxirgi 30 kundagi focus daqiqalar |

### Goal Milestone (Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `goalId` | UUID | ✅ | FK → goals.id |
| `title` | string | ✅ | Milestone nomi |
| `description` | text | ❌ | Tavsif |
| `targetPercent` | decimal | ✅ | Maqsad % (0-100) |
| `dueDate` | date | ❌ | Tugash sanasi |
| `completedAt` | timestamp | ❌ | Bajarilgan vaqt |
| `order` | integer | ✅ | Tartib |

### Goal CheckIn (Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `goalId` | UUID | ✅ | FK → goals.id |
| `value` | decimal | ✅ | Kiritilgan qiymat |
| `note` | text | ❌ | Izoh |
| `sourceType` | enum | ✅ | `manual`, `task`, `habit`, `finance` |
| `sourceId` | UUID | ❌ | Source entity ID |
| `dateKey` | string | ❌ | YYYY-MM-DD format |
| `createdAt` | timestamp | ✅ | |

### Goal Linked Arrays (Many-to-Many)
- `goal_habits` - goals ↔ habits
- `goal_tasks` - goals ↔ tasks
- `goal_finance_contributions` - transactions that contribute to goal

### Goals API Endpoints

#### GET /goals
#### GET /goals/:id
#### POST /goals
```json
{
  "title": "1000$ jamg'arish",
  "description": "Yil oxirigacha",
  "goalType": "financial",
  "metricType": "amount",
  "direction": "increase",
  "unit": "USD",
  "initialValue": 0,
  "targetValue": 1000,
  "financeMode": "save",
  "currency": "USD",
  "targetDate": "2024-12-31",
  "milestones": [
    { "title": "250$ yig'ish", "targetPercent": 25 },
    { "title": "500$ yig'ish", "targetPercent": 50 },
    { "title": "750$ yig'ish", "targetPercent": 75 }
  ]
}
```

#### PATCH /goals/:id
#### DELETE /goals/:id
#### POST /goals/:id/check-in
Progress kiritish
```json
{
  "value": 150,
  "note": "Ish haqi tushdi",
  "sourceType": "manual"
}
```

#### POST /goals/:id/complete
#### POST /goals/:id/reactivate

---

## 3.3 HABITS

### Habit Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `title` | string | ✅ | Habit nomi |
| `description` | text | ❌ | Tavsif |
| `iconId` | string | ❌ | Icon identifikatori |
| `habitType` | enum | ✅ | Habit turi |
| `status` | enum | ✅ | `active`, `paused`, `archived` |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `goalId` | UUID | ❌ | Asosiy goal (FK → goals.id) |
| `frequency` | enum | ✅ | `daily`, `weekly`, `custom` |
| `daysOfWeek` | int[] | ❌ | Hafta kunlari [0-6] |
| `timesPerWeek` | integer | ❌ | Haftada necha marta |
| `timeOfDay` | string | ❌ | Vaqt (HH:mm) |
| `completionMode` | enum | ✅ | `boolean`, `numeric` |
| `targetPerDay` | decimal | ❌ | Kunlik maqsad (numeric mode) |
| `unit` | string | ❌ | Birlik (numeric mode) |
| `countingType` | enum | ✅ | `create` (yaratish), `quit` (tashlab ketish) |
| `difficulty` | enum | ✅ | `easy`, `medium`, `hard` |
| `priority` | enum | ✅ | `low`, `medium`, `high` |
| `challengeLengthDays` | integer | ❌ | Challenge davomiyligi |
| `reminderEnabled` | boolean | ✅ | Eslatma yoqilgan (default: false) |
| `reminderTime` | string | ❌ | Eslatma vaqti (HH:mm) |
| `streakCurrent` | integer | ✅ | Joriy streak (default: 0) |
| `streakBest` | integer | ✅ | Eng yaxshi streak (default: 0) |
| `completionRate30d` | decimal | ✅ | 30 kunlik completion % (default: 0) |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Habit Type Values
```
health       - Sog'liq
finance      - Moliya
productivity - Samaradorlik
education    - Ta'lim
personal     - Shaxsiy
custom       - Boshqa
```

### Habit Finance Rule (Embedded JSON yoki alohida)
Finance habitlar uchun avtomatik completion qoidalari:
```json
// Variant 1: Kategoriyalarda xarajat qilmaslik
{
  "type": "no_spend_in_categories",
  "categoryIds": ["fast-food", "entertainment"]
}

// Variant 2: Kategoriyalarda xarajat qilish
{
  "type": "spend_in_categories",
  "categoryIds": ["education", "gym"],
  "minAmount": 50000,
  "currency": "UZS"
}

// Variant 3: Har qanday transaction bo'lsa
{
  "type": "has_any_transactions",
  "accountIds": ["savings-account-id"]
}

// Variant 4: Kunlik xarajat limitdan past
{
  "type": "daily_spend_under",
  "amount": 100000,
  "currency": "UZS"
}
```

### Habit Completion Entry (Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `habitId` | UUID | ✅ | FK → habits.id |
| `dateKey` | string | ✅ | YYYY-MM-DD |
| `status` | enum | ✅ | `done`, `miss` |
| `value` | decimal | ❌ | Qiymat (numeric mode) |
| `createdAt` | timestamp | ✅ | |

### Habit Linked Goals (Many-to-Many)
`habit_goals` table:
| Field | Type |
|-------|------|
| `habitId` | UUID |
| `goalId` | UUID |

### Habits API Endpoints

#### GET /habits
#### GET /habits/:id
#### POST /habits
```json
{
  "title": "Ertalab yugurish",
  "habitType": "health",
  "frequency": "daily",
  "daysOfWeek": [1, 2, 3, 4, 5],
  "timeOfDay": "06:00",
  "completionMode": "numeric",
  "targetPerDay": 3,
  "unit": "km",
  "countingType": "create",
  "difficulty": "medium",
  "reminderEnabled": true,
  "reminderTime": "05:45",
  "linkedGoalIds": ["goal-uuid"]
}
```

#### PATCH /habits/:id
#### DELETE /habits/:id

#### POST /habits/:id/complete
Bugungi completionni yozish
```json
{
  "dateKey": "2024-12-17",
  "status": "done",
  "value": 3.5
}
```

#### GET /habits/:id/history
Completion tarixi

#### GET /habits/:id/stats
Streak va statistika

---

## 3.4 FOCUS SESSIONS

### FocusSession Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `taskId` | UUID | ❌ | FK → tasks.id |
| `goalId` | UUID | ❌ | FK → goals.id |
| `plannedMinutes` | integer | ✅ | Rejalashtirilgan daqiqalar |
| `actualMinutes` | integer | ✅ | Haqiqiy daqiqalar (default: 0) |
| `status` | enum | ✅ | Session holati |
| `startedAt` | timestamp | ✅ | Boshlangan vaqt |
| `endedAt` | timestamp | ❌ | Tugagan vaqt |
| `interruptionsCount` | integer | ✅ | Uzilishlar soni (default: 0) |
| `notes` | text | ❌ | Izohlar |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### FocusSession Status Values
```
in_progress - Davom etmoqda
completed   - Tugallangan
canceled    - Bekor qilingan
paused      - To'xtatilgan
```

### FocusSessions API Endpoints

#### GET /focus-sessions
#### GET /focus-sessions/:id
#### POST /focus-sessions/start
```json
{
  "taskId": "task-uuid",
  "plannedMinutes": 25
}
```

#### PATCH /focus-sessions/:id/pause
#### PATCH /focus-sessions/:id/resume
#### POST /focus-sessions/:id/complete
```json
{
  "actualMinutes": 23,
  "notes": "Yaxshi session bo'ldi"
}
```

#### POST /focus-sessions/:id/cancel
#### POST /focus-sessions/:id/interrupt
Interrupt qo'shish

#### GET /focus-sessions/stats
Focus statistikasi
```json
// Response
{
  "today": {
    "totalMinutes": 120,
    "sessionsCount": 5,
    "completedCount": 4
  },
  "thisWeek": {
    "totalMinutes": 840,
    "sessionsCount": 35
  },
  "thisMonth": {
    "totalMinutes": 3200,
    "avgPerDay": 106.7
  }
}
```

---

# 4. FINANCE MODULE

## 4.1 ACCOUNTS

### Account Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `name` | string | ✅ | Hisob nomi |
| `accountType` | enum | ✅ | Hisob turi |
| `currency` | string | ✅ | Valyuta kodi (USD, UZS, etc.) |
| `initialBalance` | decimal | ✅ | Boshlang'ich balans (default: 0) |
| `currentBalance` | decimal | ✅ | Joriy balans (default: 0) |
| `linkedGoalId` | UUID | ❌ | FK → goals.id |
| `customTypeId` | string | ❌ | Custom account type |
| `icon` | string | ❌ | Icon nomi |
| `color` | string | ❌ | Rang (hex) |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Account Type Values
```
cash       - Naqd pul
card       - Bank kartasi
savings    - Jamg'arma
investment - Investitsiya
credit     - Kredit kartasi
debt       - Qarz hisobi
other      - Boshqa
```

### Accounts API Endpoints

#### GET /accounts
#### GET /accounts/:id
#### POST /accounts
```json
{
  "name": "Asosiy karta",
  "accountType": "card",
  "currency": "UZS",
  "initialBalance": 500000,
  "icon": "credit-card",
  "color": "#4CAF50"
}
```

#### PATCH /accounts/:id
#### DELETE /accounts/:id

#### PATCH /accounts/:id/adjust-balance
Balansni qo'lda to'g'rilash
```json
{
  "newBalance": 750000,
  "reason": "Kassani tekshirgandan keyin"
}
```

#### GET /accounts/summary
Umumiy statistika
```json
// Response
{
  "totalBalance": {
    "UZS": 15000000,
    "USD": 500
  },
  "totalBalanceInBaseCurrency": 21250000,
  "baseCurrency": "UZS",
  "accountsCount": 5,
  "byType": {
    "card": 2,
    "cash": 1,
    "savings": 2
  }
}
```

---

## 4.2 TRANSACTIONS

### Transaction Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `type` | enum | ✅ | `income`, `expense`, `transfer` |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `accountId` | UUID | ❌ | FK → accounts.id (income/expense uchun) |
| `fromAccountId` | UUID | ❌ | FK → accounts.id (transfer uchun) |
| `toAccountId` | UUID | ❌ | FK → accounts.id (transfer uchun) |
| `amount` | decimal | ✅ | Summa |
| `currency` | string | ✅ | Transaction valyutasi |
| `baseCurrency` | string | ✅ | User base currency |
| `rateUsedToBase` | decimal | ✅ | Base valyutaga kurs |
| `convertedAmountToBase` | decimal | ✅ | Base valyutadagi summa |
| `toAmount` | decimal | ❌ | Transfer: maqsad summa |
| `toCurrency` | string | ❌ | Transfer: maqsad valyuta |
| `effectiveRateFromTo` | decimal | ❌ | Transfer: valyutalar orasidagi kurs |
| `feeAmount` | decimal | ❌ | Komissiya |
| `feeCategoryId` | string | ❌ | Komissiya kategoriyasi |
| `categoryId` | string | ✅ | Kategoriya ID |
| `subcategoryId` | string | ❌ | Subkategoriya ID |
| `name` | string | ❌ | Transaction nomi |
| `description` | text | ❌ | Tavsif |
| `date` | date | ✅ | Sana |
| `time` | string | ❌ | Vaqt (HH:mm) |
| `goalId` | UUID | ❌ | FK → goals.id |
| `budgetId` | UUID | ❌ | FK → budgets.id |
| `debtId` | UUID | ❌ | FK → debts.id |
| `habitId` | UUID | ❌ | FK → habits.id |
| `counterpartyId` | UUID | ❌ | FK → counterparties.id |
| `goalName` | string | ❌ | Linked goal nomi (denormalized) |
| `goalType` | string | ❌ | Linked goal turi (denormalized) |
| `relatedBudgetId` | UUID | ❌ | Auto-matched budget |
| `relatedDebtId` | UUID | ❌ | Related debt |
| `plannedAmount` | decimal | ❌ | Rejalashtirilgan summa |
| `paidAmount` | decimal | ❌ | To'langan summa |
| `originalCurrency` | string | ❌ | Original valyuta (debt payment) |
| `originalAmount` | decimal | ❌ | Original summa (debt payment) |
| `conversionRate` | decimal | ❌ | Konvertatsiya kursi |
| `isBalanceAdjustment` | boolean | ✅ | Balans to'g'rilashmi (default: false) |
| `recurringId` | string | ❌ | Recurring transaction ID |
| `attachments` | string[] | ❌ | Biriktirilgan fayllar URL |
| `tags` | string[] | ❌ | Teglar |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |
| `deletedAt` | timestamp | ❌ | Soft delete vaqti |

### Transaction Split (Related Table)
Bir transaction bir nechta kategoriyaga bo'linishi uchun:
| Field | Type | Required |
|-------|------|----------|
| `id` | UUID | ✅ |
| `transactionId` | UUID | ✅ |
| `categoryId` | string | ✅ |
| `amount` | decimal | ✅ |

### Transactions API Endpoints

#### GET /transactions
**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `page` | integer | |
| `limit` | integer | |
| `type` | string | `income`, `expense`, `transfer` |
| `accountId` | UUID | |
| `categoryId` | string | |
| `dateFrom` | date | |
| `dateTo` | date | |
| `minAmount` | decimal | |
| `maxAmount` | decimal | |
| `goalId` | UUID | |
| `budgetId` | UUID | |
| `debtId` | UUID | |
| `search` | string | name, description bo'yicha |
| `tags` | string | Comma-separated tags |
| `sortBy` | string | `date`, `amount`, `createdAt` |
| `sortOrder` | string | `asc`, `desc` |

#### GET /transactions/:id
#### POST /transactions
**Income:**
```json
{
  "type": "income",
  "accountId": "account-uuid",
  "amount": 5000000,
  "currency": "UZS",
  "categoryId": "salary",
  "name": "Oylik maosh",
  "date": "2024-12-15",
  "goalId": "goal-uuid"
}
```

**Expense:**
```json
{
  "type": "expense",
  "accountId": "account-uuid",
  "amount": 150000,
  "currency": "UZS",
  "categoryId": "food",
  "subcategoryId": "groceries",
  "name": "Oziq-ovqat",
  "date": "2024-12-17",
  "budgetId": "budget-uuid"
}
```

**Transfer:**
```json
{
  "type": "transfer",
  "fromAccountId": "account-uuid-1",
  "toAccountId": "account-uuid-2",
  "amount": 100,
  "currency": "USD",
  "toAmount": 1250000,
  "toCurrency": "UZS",
  "effectiveRateFromTo": 12500,
  "date": "2024-12-17"
}
```

#### PATCH /transactions/:id
#### DELETE /transactions/:id

#### GET /transactions/summary
```json
// Query: ?dateFrom=2024-12-01&dateTo=2024-12-31
// Response
{
  "period": {
    "from": "2024-12-01",
    "to": "2024-12-31"
  },
  "totals": {
    "income": 8000000,
    "expense": 3500000,
    "net": 4500000,
    "transfersIn": 1000000,
    "transfersOut": 500000
  },
  "currency": "UZS",
  "transactionsCount": 45,
  "byCategory": [
    { "categoryId": "food", "amount": 1200000, "percent": 34.3 },
    { "categoryId": "transport", "amount": 800000, "percent": 22.9 }
  ]
}
```

#### GET /transactions/analytics
Kengaytirilgan analitika

---

## 4.3 BUDGETS

### Budget Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `name` | string | ✅ | Byudjet nomi |
| `budgetType` | enum | ✅ | `category`, `project` |
| `categoryIds` | string[] | ❌ | Kategoriyalar ro'yxati |
| `linkedGoalId` | UUID | ❌ | FK → goals.id |
| `accountId` | UUID | ❌ | FK → accounts.id |
| `transactionType` | enum | ❌ | `income`, `expense` |
| `currency` | string | ✅ | Valyuta |
| `limitAmount` | decimal | ✅ | Limit summa |
| `periodType` | enum | ✅ | Period turi |
| `startDate` | date | ❌ | Boshlanish sanasi |
| `endDate` | date | ❌ | Tugash sanasi |
| `spentAmount` | decimal | ✅ | Sarflangan (default: 0) |
| `remainingAmount` | decimal | ✅ | Qolgan (default: 0) |
| `percentUsed` | decimal | ✅ | Ishlatilgan % (default: 0) |
| `contributionTotal` | decimal | ✅ | Jami contribution (default: 0) |
| `currentBalance` | decimal | ✅ | Joriy balans (default: 0) |
| `isOverspent` | boolean | ✅ | Limitdan oshganmi (default: false) |
| `rolloverMode` | enum | ❌ | `none`, `carryover` |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `notifyOnExceed` | boolean | ✅ | Oshganda xabar berish (default: false) |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Budget Period Type Values
```
none         - Muddatsiz
weekly       - Haftalik
monthly      - Oylik
custom_range - Maxsus muddat
```

### Budget Entry (Related Table)
Transaction ↔ Budget bog'lanishi:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `budgetId` | UUID | ✅ | FK → budgets.id |
| `transactionId` | UUID | ✅ | FK → transactions.id |
| `appliedAmountBudgetCurrency` | decimal | ✅ | Budget valyutasidagi summa |
| `rateUsedTxnToBudget` | decimal | ✅ | Konvertatsiya kursi |
| `snapshottedAt` | timestamp | ✅ | |

### Budgets API Endpoints

#### GET /budgets
#### GET /budgets/:id
#### POST /budgets
```json
{
  "name": "Oziq-ovqat byudjeti",
  "budgetType": "category",
  "categoryIds": ["food", "groceries"],
  "transactionType": "expense",
  "currency": "UZS",
  "limitAmount": 2000000,
  "periodType": "monthly",
  "notifyOnExceed": true
}
```

#### PATCH /budgets/:id
#### DELETE /budgets/:id

#### GET /budgets/:id/transactions
Budget bilan bog'langan transactionlar

#### GET /budgets/:id/entries
Budget entries

#### POST /budgets/:id/recalculate
Byudjetni qayta hisoblash

---

## 4.4 DEBTS

### Debt Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `direction` | enum | ✅ | `i_owe`, `they_owe_me` |
| `counterpartyId` | UUID | ❌ | FK → counterparties.id |
| `counterpartyName` | string | ✅ | Counterparty ismi |
| `description` | text | ❌ | Tavsif |
| `principalAmount` | decimal | ✅ | Asosiy qarz summasi |
| `principalOriginalAmount` | decimal | ❌ | Original summa |
| `principalCurrency` | string | ✅ | Qarz valyutasi |
| `principalOriginalCurrency` | string | ❌ | Original valyuta |
| `baseCurrency` | string | ✅ | Base valyuta |
| `rateOnStart` | decimal | ✅ | Boshlang'ich kurs |
| `principalBaseValue` | decimal | ✅ | Base valyutadagi qiymat |
| `repaymentCurrency` | string | ❌ | To'lov valyutasi |
| `repaymentAmount` | decimal | ❌ | To'lov summasi |
| `repaymentRateOnStart` | decimal | ❌ | To'lov kursi |
| `isFixedRepaymentAmount` | boolean | ✅ | Fixed to'lov summasimi |
| `startDate` | date | ✅ | Boshlanish sanasi |
| `dueDate` | date | ❌ | Tugash sanasi |
| `interestMode` | enum | ❌ | `simple`, `compound` |
| `interestRateAnnual` | decimal | ❌ | Yillik foiz |
| `scheduleHint` | string | ❌ | To'lov jadvali hint |
| `linkedGoalId` | UUID | ❌ | FK → goals.id |
| `linkedBudgetId` | UUID | ❌ | FK → budgets.id |
| `fundingAccountId` | UUID | ❌ | FK → accounts.id |
| `fundingTransactionId` | UUID | ❌ | FK → transactions.id |
| `lentFromAccountId` | UUID | ❌ | Qarz berilgan hisob |
| `returnToAccountId` | UUID | ❌ | Qaytarilgan hisob |
| `receivedToAccountId` | UUID | ❌ | Olingan hisob |
| `payFromAccountId` | UUID | ❌ | To'lov hisobi |
| `customRateUsed` | decimal | ❌ | Custom kurs |
| `reminderEnabled` | boolean | ✅ | Eslatma (default: false) |
| `reminderTime` | string | ❌ | Eslatma vaqti |
| `status` | enum | ✅ | Qarz holati |
| `showStatus` | enum | ✅ | `active`, `archived`, `deleted` |
| `settledAt` | timestamp | ❌ | Yopilgan vaqt |
| `finalRateUsed` | decimal | ❌ | Yakuniy kurs |
| `finalProfitLoss` | decimal | ❌ | Foyda/zarar |
| `finalProfitLossCurrency` | string | ❌ | Foyda/zarar valyutasi |
| `totalPaidInRepaymentCurrency` | decimal | ❌ | Jami to'langan |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Debt Status Values
```
active   - Faol
paid     - To'langan
overdue  - Muddati o'tgan
canceled - Bekor qilingan
```

### Debt Payment (Related Table)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | |
| `debtId` | UUID | ✅ | FK → debts.id |
| `amount` | decimal | ✅ | To'lov summasi |
| `currency` | string | ✅ | To'lov valyutasi |
| `baseCurrency` | string | ✅ | Base valyuta |
| `rateUsedToBase` | decimal | ✅ | Base ga kurs |
| `convertedAmountToBase` | decimal | ✅ | Base dagi summa |
| `rateUsedToDebt` | decimal | ✅ | Debt valyutasiga kurs |
| `convertedAmountToDebt` | decimal | ✅ | Debt valyutasidagi summa |
| `paymentDate` | date | ✅ | To'lov sanasi |
| `accountId` | UUID | ❌ | FK → accounts.id |
| `note` | text | ❌ | Izoh |
| `relatedTransactionId` | UUID | ❌ | FK → transactions.id |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Debts API Endpoints

#### GET /debts
#### GET /debts/:id
#### POST /debts
**Men qarz bermoqdaman:**
```json
{
  "direction": "they_owe_me",
  "counterpartyName": "Alisher",
  "counterpartyId": "counterparty-uuid",
  "principalAmount": 500,
  "principalCurrency": "USD",
  "repaymentCurrency": "USD",
  "startDate": "2024-12-01",
  "dueDate": "2025-01-01",
  "lentFromAccountId": "account-uuid",
  "returnToAccountId": "account-uuid",
  "reminderEnabled": true
}
```

**Men qarz olmoqdaman:**
```json
{
  "direction": "i_owe",
  "counterpartyName": "Bank",
  "principalAmount": 10000000,
  "principalCurrency": "UZS",
  "repaymentCurrency": "UZS",
  "startDate": "2024-12-01",
  "dueDate": "2025-06-01",
  "interestMode": "simple",
  "interestRateAnnual": 24,
  "receivedToAccountId": "account-uuid",
  "payFromAccountId": "account-uuid"
}
```

#### PATCH /debts/:id
#### DELETE /debts/:id

#### POST /debts/:id/payments
To'lov qo'shish
```json
{
  "amount": 100,
  "currency": "USD",
  "paymentDate": "2024-12-17",
  "accountId": "account-uuid",
  "note": "Birinchi to'lov"
}
```

#### GET /debts/:id/payments
To'lovlar tarixi

#### POST /debts/:id/settle
Qarzni to'liq yopish

#### GET /debts/summary
```json
{
  "totalIOwe": {
    "UZS": 15000000,
    "USD": 0
  },
  "totalTheyOweMe": {
    "UZS": 0,
    "USD": 800
  },
  "netPosition": {
    "UZS": -15000000,
    "USD": 800
  },
  "overdueCount": 2,
  "activeCount": 5
}
```

---

## 4.5 FX RATES

### FxRate Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `date` | date | ✅ | Kurs sanasi |
| `fromCurrency` | string | ✅ | Asosiy valyuta |
| `toCurrency` | string | ✅ | Maqsad valyuta |
| `rate` | decimal | ✅ | Asosiy kurs |
| `rateMid` | decimal | ❌ | O'rta kurs |
| `rateBid` | decimal | ❌ | Sotib olish kursi |
| `rateAsk` | decimal | ❌ | Sotish kursi |
| `nominal` | integer | ✅ | Nominal (default: 1) |
| `spreadPercent` | decimal | ✅ | Spread % (default: 0.5) |
| `source` | enum | ✅ | Kurs manbai |
| `isOverridden` | boolean | ✅ | User tomonidan o'zgartirilgan |
| `effectiveFrom` | timestamp | ❌ | Kurs boshlanish vaqti |
| `effectiveUntil` | timestamp | ❌ | Kurs tugash vaqti |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### FX Rate Source Values
```
cbu        - O'zbekiston Markaziy Banki
cbr        - Rossiya Markaziy Banki
tcmb       - Turkiya Markaziy Banki
sama       - Saudiya Arabistoni Markaziy Banki
cbuae      - BAA Markaziy Banki
ecb        - Yevropa Markaziy Banki
fed        - Federal Reserve (AQSh)
boe        - Angliya Banki
market_api - Bozor kursi (API)
manual     - Qo'lda kiritilgan
```

### FxRates API Endpoints

#### GET /fx-rates
**Query:**
| Param | Type | Description |
|-------|------|-------------|
| `fromCurrency` | string | |
| `toCurrency` | string | |
| `date` | date | Ma'lum sana |
| `dateFrom` | date | |
| `dateTo` | date | |
| `source` | string | |

#### GET /fx-rates/latest
Eng so'nggi kurslar
```json
// Response
{
  "baseCurrency": "UZS",
  "rates": {
    "USD": { "rate": 12500, "rateBid": 12450, "rateAsk": 12550 },
    "EUR": { "rate": 13500, "rateBid": 13400, "rateAsk": 13600 },
    "RUB": { "rate": 125, "rateBid": 124, "rateAsk": 126 }
  },
  "updatedAt": "2024-12-17T10:00:00Z",
  "source": "cbu"
}
```

#### GET /fx-rates/convert
Valyuta konvertatsiyasi
```json
// Query: ?from=USD&to=UZS&amount=100
// Response
{
  "from": "USD",
  "to": "UZS",
  "amount": 100,
  "result": 1250000,
  "rate": 12500,
  "rateType": "mid"
}
```

#### POST /fx-rates
Qo'lda kurs qo'shish (admin)
```json
{
  "fromCurrency": "USD",
  "toCurrency": "UZS",
  "rate": 12500,
  "rateBid": 12450,
  "rateAsk": 12550,
  "date": "2024-12-17",
  "source": "manual",
  "isOverridden": true
}
```

#### GET /fx-rates/history
Kurs tarixi (grafik uchun)

---

## 4.6 COUNTERPARTIES

### Counterparty Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `displayName` | string | ✅ | Ko'rinadigan ism |
| `phoneNumber` | string | ❌ | Telefon raqami |
| `comment` | text | ❌ | Izoh |
| `searchKeywords` | string | ❌ | Qidiruv kalit so'zlari |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Counterparties API Endpoints

#### GET /counterparties
#### GET /counterparties/:id
#### POST /counterparties
```json
{
  "displayName": "Alisher Karimov",
  "phoneNumber": "+998901234567",
  "comment": "Do'stim"
}
```

#### PATCH /counterparties/:id
#### DELETE /counterparties/:id

#### GET /counterparties/:id/debts
Counterparty bilan bog'liq qarzlar

#### GET /counterparties/:id/transactions
Counterparty bilan bog'liq transactionlar

---

# 5. INSIGHTS MODULE

## 5.1 Insight Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `kind` | enum | ✅ | Insight turi |
| `level` | enum | ✅ | Muhimlik darajasi |
| `scope` | enum | ✅ | Qamrov |
| `category` | enum | ✅ | Kategoriya |
| `title` | string | ✅ | Sarlavha |
| `body` | text | ✅ | Asosiy matn |
| `explain` | text | ❌ | Batafsil tushuntirish |
| `push` | string | ❌ | Push notification matni |
| `tone` | enum | ❌ | `friend`, `strict`, `polite` |
| `priority` | integer | ✅ | Ustunlik (1-10) |
| `payload` | JSON | ❌ | Qo'shimcha ma'lumot |
| `related` | JSON | ❌ | Bog'langan entitylar |
| `actions` | JSON | ❌ | Tavsiya qilingan amallar |
| `status` | enum | ✅ | `new`, `viewed`, `completed`, `dismissed` |
| `source` | enum | ✅ | `chatgpt`, `local` |
| `validUntil` | timestamp | ❌ | Amal qilish muddati |
| `viewedAt` | timestamp | ❌ | Ko'rilgan vaqt |
| `createdAt` | timestamp | ✅ | |

### Insight Kind Values
```
finance   - Moliyaviy insight
planner   - Planner insight
habit     - Habit insight
focus     - Focus insight
combined  - Kombinatsiyalangan
wisdom    - Hikmatli so'zlar
```

### Insight Level Values
```
info        - Ma'lumot
warning     - Ogohlantirish
critical    - Muhim
celebration - Tabrik
```

### Insight Scope Values
```
daily   - Kunlik
weekly  - Haftalik
monthly - Oylik
custom  - Maxsus
```

### Insight Category Values
```
overview     - Umumiy
finance      - Moliya
productivity - Samaradorlik
wisdom       - Hikmat
```

### Insight Action (Embedded JSON)
```json
{
  "type": "review_budget",
  "label": "Byudjetni ko'rish",
  "action": "open_budgets",
  "targetId": "budget-uuid",
  "confidence": 0.85,
  "priority": "high"
}
```

### Insight Related (Embedded JSON)
```json
{
  "goalId": "goal-uuid",
  "budgetId": "budget-uuid",
  "debtId": "debt-uuid",
  "habitId": "habit-uuid",
  "taskId": "task-uuid"
}
```

## 5.2 Insight Question (Q&A)
| Field | Type | Required |
|-------|------|----------|
| `id` | UUID | ✅ |
| `userId` | UUID | ✅ |
| `prompt` | string | ✅ |
| `description` | text | ❌ |
| `options` | JSON | ❌ |
| `allowFreeText` | boolean | ✅ |
| `customLabel` | string | ❌ |
| `category` | enum | ✅ |
| `createdAt` | timestamp | ✅ |

## 5.3 Insight Question Answer
| Field | Type | Required |
|-------|------|----------|
| `id` | UUID | ✅ |
| `userId` | UUID | ✅ |
| `questionId` | UUID | ✅ |
| `optionId` | string | ❌ |
| `customAnswer` | text | ❌ |
| `answeredAt` | timestamp | ✅ |

### Insights API Endpoints

#### GET /insights
```json
// Query: ?category=finance&status=new&limit=10
```

#### GET /insights/:id
#### PATCH /insights/:id/view
Ko'rildi deb belgilash

#### PATCH /insights/:id/dismiss
Yashirish

#### PATCH /insights/:id/complete
Bajarildi deb belgilash

#### POST /insights/generate
AI insight yaratish (manual trigger)
```json
{
  "scope": "daily",
  "categories": ["finance", "productivity"]
}
```

#### GET /insights/history
Insight tarixi

#### GET /insights/questions
Savollar ro'yxati

#### POST /insights/questions/:id/answer
Savolga javob berish
```json
{
  "optionId": "option-1",
  "customAnswer": "Mening javobim"
}
```

#### POST /insights/ask
AI ga savol berish (Q&A)
```json
{
  "question": "Mening xarajatlarim qanday?",
  "context": {
    "dateRange": "last_30_days"
  }
}
```

---

# 6. SYNC & REALTIME

## 6.1 Sync Architecture

### Offline-First Strategy
1. Barcha o'zgarishlar avval local DB ga yoziladi
2. `syncStatus` = `pending` bo'ladi
3. Internet bo'lganda serverga sync qilinadi
4. Muvaffaqiyatli bo'lsa `syncStatus` = `synced`
5. Conflict bo'lsa `syncStatus` = `conflict`

### Sync Status Values
```
local    - Faqat localda
pending  - Sync kutilmoqda
synced   - Serverga sync qilingan
conflict - Conflict bor
```

## 6.2 Sync Endpoints

#### POST /sync/push
Local o'zgarishlarni serverga yuborish
```json
{
  "changes": [
    {
      "entity": "tasks",
      "id": "task-uuid",
      "operation": "create",
      "data": { ... },
      "localUpdatedAt": "2024-12-17T10:00:00Z",
      "idempotencyKey": "unique-key"
    }
  ],
  "lastSyncedAt": "2024-12-17T09:00:00Z"
}
```

#### GET /sync/pull
Serverdan o'zgarishlarni olish
```json
// Query: ?since=2024-12-17T09:00:00Z
// Response
{
  "changes": [
    {
      "entity": "tasks",
      "id": "task-uuid",
      "operation": "update",
      "data": { ... },
      "serverUpdatedAt": "2024-12-17T10:30:00Z"
    }
  ],
  "serverTime": "2024-12-17T11:00:00Z"
}
```

#### POST /sync/resolve-conflict
Conflictni hal qilish
```json
{
  "entity": "tasks",
  "id": "task-uuid",
  "resolution": "use_server" // yoki "use_local" yoki "merge"
}
```

## 6.3 WebSocket Events

### Connection
```javascript
// Client connects
ws://api.leora.app/ws?token=JWT_TOKEN
```

### Server → Client Events
```json
// Entity yangilandi
{
  "event": "entity:updated",
  "data": {
    "entity": "tasks",
    "id": "task-uuid",
    "operation": "update",
    "data": { ... }
  }
}

// Insight keldi
{
  "event": "insight:new",
  "data": {
    "id": "insight-uuid",
    "title": "...",
    "level": "warning"
  }
}

// Reminder
{
  "event": "reminder",
  "data": {
    "type": "task",
    "id": "task-uuid",
    "title": "...",
    "dueIn": 30
  }
}
```

### Client → Server Events
```json
// Presence
{
  "event": "presence:active"
}

// Subscribe to entity changes
{
  "event": "subscribe",
  "data": {
    "entities": ["tasks", "habits"]
  }
}
```

---

# 7. ENTITY RELATIONSHIPS

## 7.1 ER Diagram (Text)

```
                                    ┌─────────────┐
                                    │   USERS     │
                                    └──────┬──────┘
                                           │
         ┌──────────────┬──────────────────┼──────────────────┬──────────────┐
         │              │                  │                  │              │
         ▼              ▼                  ▼                  ▼              ▼
   ┌──────────┐  ┌──────────┐       ┌──────────┐       ┌──────────┐  ┌──────────┐
   │  TASKS   │  │  HABITS  │       │  GOALS   │       │ ACCOUNTS │  │ INSIGHTS │
   └────┬─────┘  └────┬─────┘       └────┬─────┘       └────┬─────┘  └──────────┘
        │             │                  │                  │
        │◄────────────┼──────────────────┤                  │
        │             │◄─────────────────┤                  │
        │             │                  │                  │
        ▼             ▼                  ▼                  ▼
   ┌──────────┐  ┌──────────┐       ┌──────────┐       ┌──────────┐
   │  FOCUS   │  │HABIT_COMP│       │MILESTONES│       │TRANSACTS │
   │ SESSIONS │  │  LETIONS │       │CHECK_INS │       └────┬─────┘
   └──────────┘  └──────────┘       └──────────┘            │
                                                            │
                                    ┌───────────────────────┼───────────────────────┐
                                    │                       │                       │
                                    ▼                       ▼                       ▼
                              ┌──────────┐            ┌──────────┐            ┌──────────┐
                              │ BUDGETS  │            │  DEBTS   │            │ COUNTER- │
                              └────┬─────┘            └────┬─────┘            │ PARTIES  │
                                   │                       │                  └──────────┘
                                   ▼                       ▼
                              ┌──────────┐            ┌──────────┐
                              │ BUDGET   │            │  DEBT    │
                              │ ENTRIES  │            │ PAYMENTS │
                              └──────────┘            └──────────┘
```

## 7.2 Relationships Summary

| From | To | Type | Description |
|------|-----|------|-------------|
| Task | User | Many-to-One | User owns tasks |
| Task | Goal | Many-to-One | Task linked to goal |
| Task | Habit | Many-to-One | Task linked to habit |
| Task | FocusSession | One-to-Many | Task has focus sessions |
| Goal | User | Many-to-One | User owns goals |
| Goal | Budget | One-to-One | Goal linked to budget |
| Goal | Debt | One-to-One | Goal linked to debt |
| Goal | Habit | Many-to-Many | Goal has linked habits |
| Goal | Task | Many-to-Many | Goal has linked tasks |
| Goal | Milestone | One-to-Many | Goal has milestones |
| Goal | CheckIn | One-to-Many | Goal has check-ins |
| Habit | User | Many-to-One | User owns habits |
| Habit | Goal | Many-to-Many | Habit linked to goals |
| Habit | Completion | One-to-Many | Habit has completions |
| Account | User | Many-to-One | User owns accounts |
| Account | Goal | Many-to-One | Account linked to goal |
| Transaction | User | Many-to-One | User owns transactions |
| Transaction | Account | Many-to-One | Transaction on account |
| Transaction | Goal | Many-to-One | Transaction for goal |
| Transaction | Budget | Many-to-One | Transaction in budget |
| Transaction | Debt | Many-to-One | Transaction for debt |
| Transaction | Habit | Many-to-One | Transaction triggers habit |
| Transaction | Counterparty | Many-to-One | Transaction with counterparty |
| Budget | User | Many-to-One | User owns budgets |
| Budget | Goal | One-to-One | Budget for goal |
| Budget | BudgetEntry | One-to-Many | Budget has entries |
| BudgetEntry | Transaction | Many-to-One | Entry for transaction |
| Debt | User | Many-to-One | User owns debts |
| Debt | Counterparty | Many-to-One | Debt with counterparty |
| Debt | Goal | One-to-One | Debt linked to goal |
| Debt | Payment | One-to-Many | Debt has payments |
| DebtPayment | Transaction | One-to-One | Payment creates transaction |
| Counterparty | User | Many-to-One | User owns counterparties |
| FocusSession | User | Many-to-One | User owns sessions |
| FocusSession | Task | Many-to-One | Session for task |
| FocusSession | Goal | Many-to-One | Session for goal |
| Insight | User | Many-to-One | User receives insights |

---

# 8. KATEGORIYALAR (Categories)

## 8.1 Expense Categories
```json
{
  "categories": [
    {
      "id": "food",
      "name": "Oziq-ovqat",
      "icon": "utensils",
      "subcategories": [
        { "id": "groceries", "name": "Oziq-ovqat do'koni" },
        { "id": "restaurants", "name": "Restoranlar" },
        { "id": "fast_food", "name": "Fast food" },
        { "id": "coffee", "name": "Qahva/choy" }
      ]
    },
    {
      "id": "transport",
      "name": "Transport",
      "icon": "car",
      "subcategories": [
        { "id": "fuel", "name": "Yoqilg'i" },
        { "id": "public_transport", "name": "Jamoat transporti" },
        { "id": "taxi", "name": "Taksi" },
        { "id": "parking", "name": "Parkovka" }
      ]
    },
    {
      "id": "housing",
      "name": "Uy-joy",
      "icon": "home",
      "subcategories": [
        { "id": "rent", "name": "Ijara" },
        { "id": "utilities", "name": "Kommunal xizmatlar" },
        { "id": "repairs", "name": "Ta'mirlash" },
        { "id": "furniture", "name": "Mebel" }
      ]
    },
    {
      "id": "health",
      "name": "Sog'liq",
      "icon": "heart-pulse",
      "subcategories": [
        { "id": "pharmacy", "name": "Dorixona" },
        { "id": "doctors", "name": "Shifokor" },
        { "id": "gym", "name": "Sport zal" },
        { "id": "insurance", "name": "Sug'urta" }
      ]
    },
    {
      "id": "education",
      "name": "Ta'lim",
      "icon": "graduation-cap",
      "subcategories": [
        { "id": "courses", "name": "Kurslar" },
        { "id": "books", "name": "Kitoblar" },
        { "id": "school", "name": "Maktab/universitet" }
      ]
    },
    {
      "id": "entertainment",
      "name": "Ko'ngil ochar",
      "icon": "gamepad-2",
      "subcategories": [
        { "id": "movies", "name": "Kino" },
        { "id": "games", "name": "O'yinlar" },
        { "id": "events", "name": "Tadbirlar" },
        { "id": "hobbies", "name": "Sevimli mashg'ulot" }
      ]
    },
    {
      "id": "shopping",
      "name": "Xaridlar",
      "icon": "shopping-bag",
      "subcategories": [
        { "id": "clothing", "name": "Kiyim" },
        { "id": "electronics", "name": "Elektronika" },
        { "id": "gifts", "name": "Sovg'alar" }
      ]
    },
    {
      "id": "communication",
      "name": "Aloqa",
      "icon": "phone",
      "subcategories": [
        { "id": "mobile", "name": "Mobil aloqa" },
        { "id": "internet", "name": "Internet" }
      ]
    },
    {
      "id": "personal_care",
      "name": "Shaxsiy parvarish",
      "icon": "sparkles",
      "subcategories": [
        { "id": "beauty", "name": "Go'zallik" },
        { "id": "barber", "name": "Sartaroshxona" }
      ]
    },
    {
      "id": "other_expense",
      "name": "Boshqa xarajatlar",
      "icon": "more-horizontal"
    }
  ]
}
```

## 8.2 Income Categories
```json
{
  "categories": [
    { "id": "salary", "name": "Ish haqi", "icon": "briefcase" },
    { "id": "freelance", "name": "Frilanc", "icon": "laptop" },
    { "id": "business", "name": "Biznes", "icon": "building" },
    { "id": "investments", "name": "Investitsiya", "icon": "trending-up" },
    { "id": "gifts_received", "name": "Sovg'alar", "icon": "gift" },
    { "id": "rental_income", "name": "Ijara daromadi", "icon": "home" },
    { "id": "refunds", "name": "Qaytarilgan pul", "icon": "rotate-ccw" },
    { "id": "other_income", "name": "Boshqa daromad", "icon": "more-horizontal" }
  ]
}
```

---

# 9. ERROR CODES

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `AUTH_INVALID_CREDENTIALS` | 401 | Noto'g'ri login/parol |
| `AUTH_TOKEN_EXPIRED` | 401 | Token muddati tugagan |
| `AUTH_TOKEN_INVALID` | 401 | Noto'g'ri token |
| `AUTH_UNAUTHORIZED` | 403 | Ruxsat yo'q |
| `VALIDATION_ERROR` | 400 | Validatsiya xatosi |
| `RESOURCE_NOT_FOUND` | 404 | Resurs topilmadi |
| `RESOURCE_ALREADY_EXISTS` | 409 | Resurs mavjud |
| `IDEMPOTENCY_CONFLICT` | 409 | Takroriy so'rov |
| `RATE_LIMIT_EXCEEDED` | 429 | So'rovlar limiti oshdi |
| `INTERNAL_ERROR` | 500 | Server xatosi |
| `SERVICE_UNAVAILABLE` | 503 | Servis ishlamayapti |

---

# 10. PAGINATION

Barcha list endpointlarida:

**Request:**
```
GET /tasks?page=2&limit=20&sortBy=createdAt&sortOrder=desc
```

**Response Meta:**
```json
{
  "meta": {
    "page": 2,
    "limit": 20,
    "total": 150,
    "totalPages": 8,
    "hasMore": true
  }
}
```

---

# 11. RATE LIMITING

| Endpoint Category | Limit |
|-------------------|-------|
| Auth endpoints | 10 req/min |
| Read endpoints | 100 req/min |
| Write endpoints | 50 req/min |
| Sync endpoints | 20 req/min |
| AI/Insights | 10 req/min |

---

# 12. PREMIUM & SUBSCRIPTION MODULE

## 12.1 Subscription Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `tier` | enum | ✅ | Obuna darajasi |
| `status` | enum | ✅ | Obuna holati |
| `provider` | enum | ✅ | To'lov provayderi |
| `providerSubscriptionId` | string | ❌ | Provider subscription ID |
| `providerCustomerId` | string | ❌ | Provider customer ID |
| `productId` | string | ✅ | Mahsulot ID (app store) |
| `priceId` | string | ❌ | Narx ID (Stripe) |
| `currency` | string | ✅ | Valyuta |
| `amount` | decimal | ✅ | Narx |
| `interval` | enum | ✅ | `monthly`, `yearly`, `lifetime` |
| `currentPeriodStart` | timestamp | ✅ | Joriy davr boshi |
| `currentPeriodEnd` | timestamp | ✅ | Joriy davr oxiri |
| `cancelAtPeriodEnd` | boolean | ✅ | Davr oxirida bekor qilish |
| `canceledAt` | timestamp | ❌ | Bekor qilingan vaqt |
| `trialStart` | timestamp | ❌ | Trial boshi |
| `trialEnd` | timestamp | ❌ | Trial oxiri |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Subscription Tier Values
```
free     - Bepul tarif
premium  - Premium tarif
pro      - Pro tarif (kelajakda)
```

### Subscription Status Values
```
active       - Faol
trialing     - Trial davrida
past_due     - To'lov kechikkan
canceled     - Bekor qilingan
unpaid       - To'lanmagan
incomplete   - To'liq emas
expired      - Muddati tugagan
```

### Payment Provider Values
```
apple       - Apple App Store
google      - Google Play Store
stripe      - Stripe (web)
manual      - Qo'lda (admin)
```

## 12.2 Subscription History Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `subscriptionId` | UUID | ✅ | FK → subscriptions.id |
| `event` | enum | ✅ | Event turi |
| `tier` | enum | ✅ | Obuna darajasi |
| `amount` | decimal | ❌ | Summa |
| `currency` | string | ❌ | Valyuta |
| `metadata` | JSON | ❌ | Qo'shimcha ma'lumot |
| `createdAt` | timestamp | ✅ | |

### Subscription Event Values
```
created          - Yangi obuna
renewed          - Yangilandi
upgraded         - Oshirildi
downgraded       - Tushirildi
canceled         - Bekor qilindi
expired          - Muddati tugadi
trial_started    - Trial boshlandi
trial_ended      - Trial tugadi
payment_failed   - To'lov muvaffaqiyatsiz
payment_succeeded - To'lov muvaffaqiyatli
refunded         - Qaytarildi
```

## 12.3 Premium Features Entity
Qaysi featurelar qaysi tarifga tegishli:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `featureKey` | string | ✅ | Feature identifikatori |
| `name` | string | ✅ | Feature nomi |
| `description` | text | ❌ | Tavsif |
| `tier` | enum | ✅ | Minimal kerakli tarif |
| `limitFree` | integer | ❌ | Free tarif limiti |
| `limitPremium` | integer | ❌ | Premium tarif limiti (-1 = cheksiz) |
| `isActive` | boolean | ✅ | Faolmi |
| `createdAt` | timestamp | ✅ | |

### Feature Keys & Limits
```json
{
  "features": [
    {
      "featureKey": "ai_daily_insights",
      "name": "AI Kunlik Tahlillar",
      "limitFree": 2,
      "limitPremium": -1
    },
    {
      "featureKey": "ai_questions",
      "name": "AI Savollar",
      "limitFree": 5,
      "limitPremium": -1
    },
    {
      "featureKey": "ai_voice_commands",
      "name": "Ovozli Buyruqlar",
      "limitFree": 10,
      "limitPremium": -1
    },
    {
      "featureKey": "accounts",
      "name": "Hisoblar soni",
      "limitFree": 3,
      "limitPremium": -1
    },
    {
      "featureKey": "budgets",
      "name": "Byudjetlar soni",
      "limitFree": 2,
      "limitPremium": -1
    },
    {
      "featureKey": "goals",
      "name": "Maqsadlar soni",
      "limitFree": 3,
      "limitPremium": -1
    },
    {
      "featureKey": "habits",
      "name": "Odatlar soni",
      "limitFree": 5,
      "limitPremium": -1
    },
    {
      "featureKey": "export_data",
      "name": "Ma'lumotlarni eksport qilish",
      "limitFree": 0,
      "limitPremium": -1
    },
    {
      "featureKey": "custom_categories",
      "name": "Maxsus kategoriyalar",
      "limitFree": 0,
      "limitPremium": -1
    },
    {
      "featureKey": "advanced_analytics",
      "name": "Kengaytirilgan analitika",
      "limitFree": 0,
      "limitPremium": -1
    },
    {
      "featureKey": "multi_currency",
      "name": "Ko'p valyuta",
      "limitFree": 1,
      "limitPremium": -1
    },
    {
      "featureKey": "cloud_backup",
      "name": "Cloud backup",
      "limitFree": 0,
      "limitPremium": -1
    },
    {
      "featureKey": "priority_support",
      "name": "Ustuvor yordam",
      "limitFree": 0,
      "limitPremium": -1
    },
    {
      "featureKey": "virtual_mentors",
      "name": "Virtual mentorlar",
      "limitFree": 1,
      "limitPremium": 10
    }
  ]
}
```

## 12.4 Subscription Plans (Product Catalog)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `name` | string | ✅ | Plan nomi |
| `tier` | enum | ✅ | Tarif darajasi |
| `interval` | enum | ✅ | `monthly`, `yearly`, `lifetime` |
| `prices` | JSON | ✅ | Narxlar (valyuta bo'yicha) |
| `appleProductId` | string | ❌ | Apple product ID |
| `googleProductId` | string | ❌ | Google product ID |
| `stripePriceId` | string | ❌ | Stripe price ID |
| `features` | string[] | ✅ | Feature keys ro'yxati |
| `isPopular` | boolean | ✅ | Mashhur plan |
| `discount` | decimal | ❌ | Chegirma % |
| `isActive` | boolean | ✅ | Faolmi |
| `createdAt` | timestamp | ✅ | |

### Subscription Plans Example
```json
{
  "plans": [
    {
      "id": "plan_monthly",
      "name": "Premium Oylik",
      "tier": "premium",
      "interval": "monthly",
      "prices": {
        "USD": 4.99,
        "UZS": 59900,
        "RUB": 449
      },
      "appleProductId": "com.leora.premium.monthly",
      "googleProductId": "premium_monthly",
      "isPopular": false
    },
    {
      "id": "plan_yearly",
      "name": "Premium Yillik",
      "tier": "premium",
      "interval": "yearly",
      "prices": {
        "USD": 39.99,
        "UZS": 479900,
        "RUB": 3599
      },
      "appleProductId": "com.leora.premium.yearly",
      "googleProductId": "premium_yearly",
      "discount": 33,
      "isPopular": true
    },
    {
      "id": "plan_lifetime",
      "name": "Premium Abadiy",
      "tier": "premium",
      "interval": "lifetime",
      "prices": {
        "USD": 99.99,
        "UZS": 1199900,
        "RUB": 8999
      },
      "appleProductId": "com.leora.premium.lifetime",
      "googleProductId": "premium_lifetime",
      "isPopular": false
    }
  ]
}
```

## 12.5 Subscription API Endpoints

### GET /subscriptions/me
Joriy user obunasi
```json
// Response
{
  "subscription": {
    "id": "sub-uuid",
    "tier": "premium",
    "status": "active",
    "currentPeriodEnd": "2025-01-17T00:00:00Z",
    "cancelAtPeriodEnd": false
  },
  "features": {
    "ai_daily_insights": { "limit": -1, "used": 5 },
    "ai_questions": { "limit": -1, "used": 12 },
    "accounts": { "limit": -1, "used": 4 }
  },
  "plan": {
    "name": "Premium Yillik",
    "interval": "yearly"
  }
}
```

### GET /subscriptions/plans
Mavjud planlar ro'yxati

### POST /subscriptions/checkout
Yangi obuna boshlash
```json
// Request
{
  "planId": "plan_yearly",
  "provider": "stripe"
}

// Response (Stripe)
{
  "checkoutUrl": "https://checkout.stripe.com/...",
  "sessionId": "cs_xxx"
}

// Response (Mobile)
{
  "productId": "com.leora.premium.yearly",
  "offerId": "offer_xxx"
}
```

### POST /subscriptions/verify-receipt
Receipt tekshirish (Mobile)
```json
// Request (Apple)
{
  "provider": "apple",
  "receipt": "base64_receipt_data",
  "productId": "com.leora.premium.yearly"
}

// Request (Google)
{
  "provider": "google",
  "purchaseToken": "token_xxx",
  "productId": "premium_yearly"
}
```

### POST /subscriptions/cancel
Obunani bekor qilish
```json
// Request
{
  "reason": "Too expensive",
  "feedback": "Optional feedback"
}
```

### POST /subscriptions/restore
Obunani tiklash (Mobile)
```json
// Request
{
  "provider": "apple"
}
```

### GET /subscriptions/history
Obuna tarixi

### POST /subscriptions/webhook/stripe
Stripe webhook handler

### POST /subscriptions/webhook/apple
Apple App Store webhook

### POST /subscriptions/webhook/google
Google Play webhook

---

# 13. USER SETTINGS MODULE

## 13.1 UserSettings Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id (unique) |
| `theme` | enum | ✅ | `light`, `dark`, `auto` |
| `language` | enum | ✅ | `en`, `ru`, `uz`, `ar`, `tr` |
| `notifications` | JSON | ✅ | Notification sozlamalari |
| `security` | JSON | ✅ | Security sozlamalari |
| `ai` | JSON | ✅ | AI sozlamalari |
| `focus` | JSON | ✅ | Focus mode sozlamalari |
| `privacy` | JSON | ✅ | Privacy sozlamalari |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Notifications Settings (JSON)
```json
{
  "enabled": true,
  "sound": true,
  "vibration": true,
  "showOnLockScreen": true,

  "finance": {
    "budgetOverspend": true,
    "debtReminder": true,
    "unusualSpends": true,
    "goalAchievements": true,
    "quietHoursStart": "21:00",
    "quietHoursEnd": "08:00"
  },

  "tasks": {
    "reminder": true,
    "reminderMinutesBefore": 15,
    "deadline": true,
    "deadlineDaysBefore": 1,
    "goalProgress": true,
    "reschedule": true
  },

  "habits": {
    "morningReminder": true,
    "morningTime": "07:00",
    "eveningReminder": true,
    "eveningTime": "21:00",
    "streakReminder": true,
    "motivational": true
  },

  "ai": {
    "smartRecommendations": true,
    "recommendationsPerDay": 3,
    "weeklyInsights": true,
    "mentorAdvices": true,
    "predictions": true
  },

  "doNotDisturb": {
    "enabled": false,
    "start": "22:00",
    "end": "07:00",
    "onWeekends": false
  }
}
```

### Security Settings (JSON)
```json
{
  "lockEnabled": true,
  "biometricsEnabled": true,
  "pinEnabled": false,
  "pinHash": null,
  "autoLockTimeoutMs": 60000,
  "unlockGraceMs": 30000,
  "askOnLaunch": true,

  "dataProtection": {
    "databaseEncryption": true,
    "hideBalances": false,
    "screenshotBlock": false,
    "fakeAccount": false
  },

  "backup": {
    "autoBackup": true,
    "frequency": "weekly",
    "lastBackup": "2024-12-15T10:00:00Z"
  },

  "privacy": {
    "anonymousAnalytics": true,
    "personalizedAds": false,
    "shareWithPartners": false
  }
}
```

### AI Settings (JSON)
```json
{
  "helpLevel": "medium",
  "assistantName": "Leora",
  "talkStyle": "friendly",
  "language": "uz",

  "features": {
    "voiceRecognition": true,
    "transactionCategories": true,
    "smartReminders": true,
    "predictions": true,
    "motivational": true,
    "scheduleOptimization": true
  },

  "speech": {
    "voiceTyping": true,
    "inputMode": "button",
    "language": "uz-UZ"
  },

  "mentors": [
    { "id": "buffett", "name": "Warren Buffett", "category": "financial", "enabled": true },
    { "id": "musk", "name": "Elon Musk", "category": "productivity", "enabled": true },
    { "id": "aurelius", "name": "Marcus Aurelius", "category": "balance", "enabled": false }
  ]
}
```

### Focus Settings (JSON)
```json
{
  "defaultDuration": 25,
  "breakDuration": 5,
  "longBreakDuration": 15,
  "sessionsBeforeLongBreak": 4,
  "technique": "pomodoro",
  "autoStartBreaks": false,
  "autoStartSessions": false,
  "soundEnabled": true,
  "tickingSound": false
}
```

## 13.2 User Settings API Endpoints

### GET /settings
Barcha sozlamalar

### PATCH /settings
Sozlamalarni yangilash
```json
// Request
{
  "theme": "dark",
  "language": "uz",
  "notifications": {
    "enabled": true,
    "sound": false
  }
}
```

### PATCH /settings/notifications
Notification sozlamalari

### PATCH /settings/security
Security sozlamalari

### PATCH /settings/ai
AI sozlamalari

### PATCH /settings/focus
Focus sozlamalari

### POST /settings/reset
Sozlamalarni default qiymatga qaytarish

---

# 14. ACHIEVEMENTS & GAMIFICATION MODULE

## 14.1 Achievement Entity (Predefined)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `key` | string | ✅ | Unique achievement key |
| `name` | string | ✅ | Achievement nomi |
| `description` | text | ✅ | Tavsif |
| `category` | enum | ✅ | Kategoriya |
| `icon` | string | ✅ | Icon nomi |
| `color` | string | ✅ | Rang (hex) |
| `xpReward` | integer | ✅ | XP mukofoti |
| `requirement` | JSON | ✅ | Talab (condition) |
| `isSecret` | boolean | ✅ | Yashirin achievement |
| `tier` | enum | ✅ | `bronze`, `silver`, `gold`, `platinum` |
| `order` | integer | ✅ | Tartib |

### Achievement Categories
```
finance      - Moliyaviy yutuqlar
tasks        - Task yutuqlari
habits       - Odat yutuqlari
focus        - Focus yutuqlari
goals        - Maqsad yutuqlari
streak       - Streak yutuqlari
social       - Ijtimoiy yutuqlar
special      - Maxsus yutuqlar
```

### Achievement Requirements (JSON Examples)
```json
// Task achievements
{ "type": "tasks_completed", "count": 100 }
{ "type": "tasks_streak", "days": 7 }

// Habit achievements
{ "type": "habits_streak", "days": 30 }
{ "type": "habit_completion_rate", "percent": 90, "days": 30 }

// Finance achievements
{ "type": "transactions_logged", "count": 500 }
{ "type": "budget_under_limit", "months": 3 }
{ "type": "savings_goal_reached", "count": 1 }

// Focus achievements
{ "type": "focus_minutes_total", "minutes": 1000 }
{ "type": "focus_sessions_completed", "count": 100 }

// Goal achievements
{ "type": "goals_completed", "count": 5 }

// Special
{ "type": "app_usage_days", "days": 365 }
```

## 14.2 UserAchievement Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `achievementId` | UUID | ✅ | FK → achievements.id |
| `progress` | decimal | ✅ | Joriy progress (0-100) |
| `unlockedAt` | timestamp | ❌ | Ochilgan vaqt |
| `notifiedAt` | timestamp | ❌ | Xabar berilgan vaqt |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

## 14.3 UserLevel Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id (unique) |
| `level` | integer | ✅ | Joriy level (default: 1) |
| `currentXP` | integer | ✅ | Joriy XP (default: 0) |
| `totalXP` | integer | ✅ | Jami XP (default: 0) |
| `title` | string | ❌ | Level title |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### XP Calculation Rules
```json
{
  "xpRules": {
    "task_completed": 50,
    "task_created": 10,
    "habit_completed": 30,
    "habit_streak_day": 20,
    "goal_progress_10_percent": 25,
    "goal_completed": 200,
    "focus_session_completed": 40,
    "transaction_logged": 5,
    "budget_created": 30,
    "debt_paid_off": 100,
    "achievement_unlocked": 100,
    "daily_login": 10,
    "weekly_streak": 50
  },
  "levelFormula": "floor(totalXP / 500) + 1",
  "xpForNextLevel": "level * 500"
}
```

### Level Titles
```json
{
  "titles": [
    { "minLevel": 1, "title": "Beginner" },
    { "minLevel": 5, "title": "Apprentice" },
    { "minLevel": 10, "title": "Achiever" },
    { "minLevel": 20, "title": "Expert" },
    { "minLevel": 35, "title": "Master" },
    { "minLevel": 50, "title": "Legend" },
    { "minLevel": 75, "title": "Champion" },
    { "minLevel": 100, "title": "Grandmaster" }
  ]
}
```

## 14.4 Achievements API Endpoints

### GET /achievements
Barcha achievementlar ro'yxati
```json
// Response
{
  "achievements": [...],
  "userProgress": {
    "totalUnlocked": 15,
    "totalAchievements": 50,
    "recentlyUnlocked": [...],
    "closeToUnlocking": [...]
  }
}
```

### GET /achievements/:id
Bitta achievement

### GET /achievements/categories
Kategoriyalar bo'yicha

### GET /users/me/level
User level ma'lumotlari
```json
// Response
{
  "level": 12,
  "title": "Achiever",
  "currentXP": 350,
  "totalXP": 5850,
  "xpForNextLevel": 500,
  "xpProgress": 70,
  "recentXPGains": [
    { "reason": "task_completed", "xp": 50, "createdAt": "..." },
    { "reason": "habit_completed", "xp": 30, "createdAt": "..." }
  ]
}
```

### GET /users/me/achievements
User achievementlari

### POST /achievements/:id/claim
Achievement mukofotini olish (agar claim kerak bo'lsa)

---

# 15. AI QUOTA & USAGE TRACKING MODULE

## 15.1 AIUsage Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `channel` | enum | ✅ | AI channel |
| `requestType` | string | ✅ | So'rov turi |
| `tokensUsed` | integer | ❌ | Ishlatilgan tokenlar |
| `responseTime` | integer | ❌ | Javob vaqti (ms) |
| `success` | boolean | ✅ | Muvaffaqiyatli |
| `errorCode` | string | ❌ | Xato kodi |
| `metadata` | JSON | ❌ | Qo'shimcha ma'lumot |
| `createdAt` | timestamp | ✅ | |

### AI Channel Values
```
daily    - Kunlik insight
period   - Haftalik/oylik summary
qa       - Savol-javob
voice    - Ovozli buyruqlar
```

## 15.2 AIQuota Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `channel` | enum | ✅ | AI channel |
| `periodStart` | timestamp | ✅ | Period boshi |
| `periodEnd` | timestamp | ✅ | Period oxiri |
| `limit` | integer | ✅ | Limit |
| `used` | integer | ✅ | Ishlatilgan (default: 0) |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Default Quotas (Daily)
```json
{
  "free": {
    "daily": 2,
    "period": 1,
    "qa": 5,
    "voice": 10
  },
  "premium": {
    "daily": -1,
    "period": -1,
    "qa": -1,
    "voice": -1
  }
}
```

## 15.3 AI Usage API Endpoints

### GET /ai/quota
Joriy quota holati
```json
// Response
{
  "tier": "free",
  "quotas": {
    "daily": { "limit": 2, "used": 1, "remaining": 1, "resetsAt": "..." },
    "qa": { "limit": 5, "used": 3, "remaining": 2, "resetsAt": "..." },
    "voice": { "limit": 10, "used": 5, "remaining": 5, "resetsAt": "..." }
  }
}
```

### GET /ai/usage/history
Foydalanish tarixi

### GET /ai/usage/stats
Foydalanish statistikasi
```json
// Response
{
  "thisMonth": {
    "totalRequests": 150,
    "byChannel": {
      "daily": 30,
      "qa": 80,
      "voice": 40
    },
    "avgResponseTime": 450
  },
  "allTime": {
    "totalRequests": 1200
  }
}
```

---

# 16. DATA MANAGEMENT MODULE

## 16.1 Backup Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `type` | enum | ✅ | `manual`, `auto`, `export` |
| `status` | enum | ✅ | `pending`, `completed`, `failed` |
| `storage` | enum | ✅ | `cloud`, `local`, `s3` |
| `fileUrl` | string | ❌ | Fayl URL |
| `fileSize` | integer | ❌ | Fayl hajmi (bytes) |
| `checksum` | string | ❌ | MD5/SHA256 hash |
| `entitiesIncluded` | string[] | ✅ | Qo'shilgan entitylar |
| `entityCounts` | JSON | ❌ | Entity sonlari |
| `expiresAt` | timestamp | ❌ | Tugash vaqti |
| `metadata` | JSON | ❌ | Qo'shimcha ma'lumot |
| `createdAt` | timestamp | ✅ | |

### Backup Entity Counts (JSON)
```json
{
  "tasks": 150,
  "goals": 10,
  "habits": 15,
  "accounts": 5,
  "transactions": 500,
  "budgets": 3,
  "debts": 2,
  "focusSessions": 200,
  "insights": 50
}
```

## 16.2 Export Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `format` | enum | ✅ | `json`, `csv`, `pdf` |
| `scope` | enum | ✅ | `all`, `finance`, `planner`, `custom` |
| `entities` | string[] | ❌ | Tanlangan entitylar |
| `dateFrom` | date | ❌ | Boshlanish sanasi |
| `dateTo` | date | ❌ | Tugash sanasi |
| `status` | enum | ✅ | `pending`, `processing`, `completed`, `failed` |
| `fileUrl` | string | ❌ | Fayl URL |
| `fileSize` | integer | ❌ | Fayl hajmi |
| `expiresAt` | timestamp | ❌ | Tugash vaqti |
| `createdAt` | timestamp | ✅ | |

## 16.3 Data Management API Endpoints

### POST /data/backup
Backup yaratish
```json
// Request
{
  "storage": "cloud",
  "entities": ["tasks", "goals", "habits", "transactions"]
}
```

### GET /data/backups
Backuplar ro'yxati

### GET /data/backups/:id
Bitta backup

### POST /data/backups/:id/restore
Backupdan tiklash

### DELETE /data/backups/:id
Backupni o'chirish

### POST /data/export
Export yaratish
```json
// Request
{
  "format": "csv",
  "scope": "finance",
  "dateFrom": "2024-01-01",
  "dateTo": "2024-12-31"
}
```

### GET /data/exports
Exportlar ro'yxati

### GET /data/exports/:id/download
Export faylni yuklab olish

### DELETE /data/account
Accountni o'chirish (GDPR)
```json
// Request
{
  "confirmation": "DELETE MY ACCOUNT",
  "reason": "optional reason"
}
```

### POST /data/cache/clear
Cache tozalash

---

# 17. INTEGRATIONS MODULE

## 17.1 Integration Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `provider` | enum | ✅ | Integration provider |
| `category` | enum | ✅ | `calendar`, `bank`, `app`, `device` |
| `status` | enum | ✅ | `connected`, `disconnected`, `error` |
| `accessToken` | string | ❌ | Access token (encrypted) |
| `refreshToken` | string | ❌ | Refresh token (encrypted) |
| `expiresAt` | timestamp | ❌ | Token tugash vaqti |
| `scope` | string[] | ❌ | Granted scopes |
| `accountId` | string | ❌ | Provider account ID |
| `accountName` | string | ❌ | Provider account name |
| `metadata` | JSON | ❌ | Qo'shimcha ma'lumot |
| `lastSyncAt` | timestamp | ❌ | Oxirgi sync vaqti |
| `syncError` | string | ❌ | Oxirgi xato |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

### Integration Providers
```
// Calendars
google_calendar
apple_calendar
outlook_calendar

// Banks (O'zbekiston)
uzcard
humo
kapitalbank
ipotekabank

// Apps
telegram
whatsapp
slack
spotify

// Devices
apple_watch
wear_os
fitbit
```

## 17.2 Integration Sync Log
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `integrationId` | UUID | ✅ | FK → integrations.id |
| `direction` | enum | ✅ | `push`, `pull`, `both` |
| `status` | enum | ✅ | `success`, `partial`, `failed` |
| `itemsSynced` | integer | ❌ | Sync qilingan elementlar |
| `errors` | JSON | ❌ | Xatolar ro'yxati |
| `startedAt` | timestamp | ✅ | |
| `completedAt` | timestamp | ❌ | |

## 17.3 Integrations API Endpoints

### GET /integrations
Barcha integrationlar

### GET /integrations/:provider/connect
OAuth flow boshlash
```json
// Response
{
  "authUrl": "https://accounts.google.com/oauth...",
  "state": "random_state_token"
}
```

### POST /integrations/:provider/callback
OAuth callback
```json
// Request
{
  "code": "auth_code",
  "state": "random_state_token"
}
```

### DELETE /integrations/:id
Integration uzish

### POST /integrations/:id/sync
Manual sync

### GET /integrations/:id/logs
Sync logs

---

# 18. DEVICE & SESSION MANAGEMENT

## 18.1 UserDevice Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `deviceId` | string | ✅ | Unique device identifier |
| `deviceName` | string | ❌ | Qurilma nomi |
| `deviceType` | enum | ✅ | `ios`, `android`, `web` |
| `osVersion` | string | ❌ | OS versiyasi |
| `appVersion` | string | ❌ | App versiyasi |
| `pushToken` | string | ❌ | Push notification token |
| `lastActiveAt` | timestamp | ✅ | Oxirgi faollik |
| `lastIp` | string | ❌ | Oxirgi IP |
| `location` | string | ❌ | Taxminiy joylashuv |
| `isTrusted` | boolean | ✅ | Ishonchli qurilma |
| `createdAt` | timestamp | ✅ | |
| `updatedAt` | timestamp | ✅ | |

## 18.2 UserSession Entity
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | ✅ | Primary key |
| `userId` | UUID | ✅ | FK → users.id |
| `deviceId` | UUID | ✅ | FK → user_devices.id |
| `tokenHash` | string | ✅ | Token hash |
| `isActive` | boolean | ✅ | Faol sessiya |
| `expiresAt` | timestamp | ✅ | Tugash vaqti |
| `lastUsedAt` | timestamp | ✅ | Oxirgi ishlatilgan |
| `createdAt` | timestamp | ✅ | |
| `revokedAt` | timestamp | ❌ | Bekor qilingan vaqt |

## 18.3 Device & Session API Endpoints

### GET /devices
Qurilmalar ro'yxati

### DELETE /devices/:id
Qurilmani o'chirish

### POST /devices/:id/trust
Qurilmani ishonchli qilish

### GET /sessions
Faol sessiyalar

### DELETE /sessions/:id
Sessiyani tugatish

### DELETE /sessions/all
Barcha sessiyalarni tugatish (joriy bundan mustasno)

---

# XULOSA

Bu dokumentatsiya Leora backend uchun to'liq API spesifikatsiyasini o'z ichiga oladi:

✅ **25+ ta asosiy entity** batafsil fieldlar bilan
✅ **150+ API endpoint** CRUD operatsiyalar bilan
✅ **Multi-tenant architecture** ko'p foydalanuvchilar uchun
✅ **Multi-currency support** turli valyutalar uchun
✅ **Offline-first sync** real-time yangilanishlar bilan
✅ **AI Insights** integratsiyasi
✅ **Cross-module relationships** planner ↔ finance
✅ **Premium/Subscription** tizimi (Apple, Google, Stripe)
✅ **Achievements & Gamification** XP va levellar
✅ **AI Quota tracking** Free vs Premium
✅ **Data Management** backup, export, GDPR
✅ **Integrations** calendar, bank, app, device
✅ **Security** 2FA, sessions, devices

Backend developerlarga bu hujjat asosida ishlash mumkin!

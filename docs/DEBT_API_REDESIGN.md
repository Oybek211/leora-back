# Debt API Redesign: Mismatch Report & Implementation Plan

## STEP 1: MISMATCH REPORT

### Current State Analysis

#### Backend Database Schema (Migration 009)
The `debts` table already has:
- `direction` TEXT (i_owe/they_owe_me) - **EXISTS**
- `counterparty_id` UUID - **EXISTS (nullable)**
- `counterparty_name` TEXT - **EXISTS**
- `principal_amount`, `principal_currency` - **EXISTS**
- `start_date`, `due_date` - **EXISTS**
- `status` (active/paid/overdue/canceled) - **EXISTS**
- Full multi-currency support with repayment fields - **EXISTS**

The `counterparties` table already exists:
- `id`, `user_id`, `display_name`, `phone_number`, `comment`, `search_keywords` - **EXISTS**
- `show_status`, `created_at`, `updated_at`, `deleted_at` - **EXISTS**

#### Backend Go Models (model.go)
- `Debt` struct: Has `CounterpartyID *string` and `CounterpartyName string` - **GOOD**
- `Counterparty` struct: Has all required fields - **GOOD**

#### Backend API Routes (routes.go)
- Debt CRUD: `/debts`, `/debts/:id` - **EXISTS**
- Counterparty CRUD: `/counterparties`, `/counterparties/:id` - **EXISTS**
- Extra endpoints: `/counterparties/:id/debts`, `/counterparties/:id/transactions` - **EXISTS**

### Identified Mismatches

| Issue | Current State | Required State | Severity |
|-------|---------------|----------------|----------|
| 1. Debt validation | Only validates `principalCurrency`, `startDate`, `principalAmount` | Must require either `counterpartyId` OR `counterpartyName` | HIGH |
| 2. Debt response | Returns flat `counterpartyId` and `counterpartyName` | Should embed full `counterparty` object for UI convenience | MEDIUM |
| 3. Inline counterparty creation | Not supported | Create debt can optionally create counterparty inline | MEDIUM |
| 4. Direction validation | No enum validation | Must validate `direction` is one of `i_owe`/`they_owe_me` | HIGH |
| 5. Cross-owner validation | Implicit via context | Explicit validation that counterparty_id belongs to same owner | HIGH |
| 6. Counterparty deletion | Can delete anytime | Should block if counterparty has linked debts | MEDIUM |
| 7. Error codes | Generic errors | Need specific codes for counterparty validation failures | LOW |

### What's Already Working

1. **Database schema is complete** - Both tables exist with correct fields
2. **Basic CRUD operations work** - List, create, get, update, delete for both entities
3. **Direction field exists** - Using `i_owe`/`they_owe_me` (mobile maps to `borrowed`/`lent`)
4. **Owner scoping** - Context-based user filtering in repository
5. **Counterparty picker in UI** - Mobile app has full counterparty selection UI
6. **Multi-currency support** - Repayment in different currency fully implemented

---

## STEP 2: DATA MODEL - NO CHANGES NEEDED

The existing schema satisfies all requirements:

```sql
-- Already exists in migrations/009_finance_alignment.sql
CREATE TABLE IF NOT EXISTS counterparties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    phone_number TEXT,
    comment TEXT,
    search_keywords TEXT,
    show_status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- debts table already has:
-- counterparty_id UUID
-- counterparty_name TEXT NOT NULL DEFAULT ''
-- direction TEXT NOT NULL DEFAULT 'i_owe'
```

---

## STEP 3: API CONTRACT

### Counterparties

```
GET    /counterparties              -> List (owner scoped, search filter)
POST   /counterparties              -> Create
GET    /counterparties/:id          -> Get one
PATCH  /counterparties/:id          -> Update
DELETE /counterparties/:id          -> Delete (conflict if has debts)
GET    /counterparties/:id/debts    -> List debts for counterparty
```

### Debts

```
GET    /debts                       -> List (owner scoped, filters: direction, status, linkedGoalId)
POST   /debts                       -> Create (with counterparty support)
GET    /debts/:id                   -> Get by ID (embed counterparty)
PUT    /debts/:id                   -> Full update
PATCH  /debts/:id                   -> Partial update
DELETE /debts/:id                   -> Delete
POST   /debts/:id/settle            -> Mark as settled
POST   /debts/:id/extend            -> Extend due date
```

### Create Debt Request Options

**Option A: Reference existing counterparty**
```json
{
  "direction": "i_owe",
  "counterpartyId": "uuid-of-existing-counterparty",
  "principalAmount": 120.50,
  "principalCurrency": "USD",
  "startDate": "2026-01-10",
  "dueDate": "2026-02-10",
  "description": "Borrowed for laptop"
}
```

**Option B: Create counterparty inline (UX friendly)**
```json
{
  "direction": "they_owe_me",
  "counterparty": {
    "displayName": "Ali Valiyev",
    "phoneNumber": "+998901234567"
  },
  "principalAmount": 500000,
  "principalCurrency": "UZS",
  "startDate": "2026-01-10",
  "dueDate": "2026-03-01"
}
```

**Option C: Quick create with just name (legacy support)**
```json
{
  "direction": "i_owe",
  "counterpartyName": "John Doe",
  "principalAmount": 100,
  "principalCurrency": "USD",
  "startDate": "2026-01-10"
}
```

### Debt Response (with embedded counterparty)

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "userId": "uuid",
    "direction": "i_owe",
    "counterpartyId": "uuid",
    "counterpartyName": "Ali Valiyev",
    "counterparty": {
      "id": "uuid",
      "displayName": "Ali Valiyev",
      "phoneNumber": "+998901234567",
      "comment": null
    },
    "principalAmount": 120.50,
    "principalCurrency": "USD",
    "baseCurrency": "UZS",
    "rateOnStart": 12500,
    "principalBaseValue": 1506250,
    "startDate": "2026-01-10",
    "dueDate": "2026-02-10",
    "status": "active",
    "remainingAmount": 120.50,
    "totalPaid": 0,
    "percentPaid": 0,
    "showStatus": "active",
    "createdAt": "2026-01-10T10:00:00Z",
    "updatedAt": "2026-01-10T10:00:00Z"
  }
}
```

---

## STEP 4: VALIDATION RULES

### Debt Creation
1. `direction` is required, must be `i_owe` or `they_owe_me`
2. At least one of: `counterpartyId`, `counterparty.displayName`, or `counterpartyName`
3. `principalAmount` must be > 0
4. `principalCurrency` is required
5. `startDate` is required
6. If `dueDate` provided, it must be >= `startDate`
7. If `counterpartyId` provided, it must belong to same owner (else 403/404)

### Counterparty Creation
1. `displayName` is required (min 2 characters)
2. `phoneNumber` is optional but if provided must be valid format

### Counterparty Deletion
1. Cannot delete if any debt references this counterparty (return CONFLICT)

---

## STEP 5: ERROR CODES

```go
// New error codes to add
var (
    CounterpartyRequired     = &Error{Code: -5010, Type: "VALIDATION", Message: "Counterparty is required for debt"}
    CounterpartyNameTooShort = &Error{Code: -5011, Type: "VALIDATION", Message: "Counterparty name must be at least 2 characters"}
    InvalidDebtDirection     = &Error{Code: -5012, Type: "VALIDATION", Message: "Direction must be 'i_owe' or 'they_owe_me'"}
    CounterpartyHasDebts     = &Error{Code: -5013, Type: "CONFLICT", Message: "Cannot delete counterparty that has linked debts"}
    CounterpartyForbidden    = &Error{Code: -5014, Type: "FORBIDDEN", Message: "Counterparty does not belong to this user"}
)
```

---

## STEP 6: IMPLEMENTATION CHANGES

### Files to Modify

1. `internal/errors/errors.go` - Add new error codes
2. `internal/modules/finance/model.go` - Add embedded counterparty to debt response
3. `internal/modules/finance/handler.go` - Update validation, add inline counterparty creation
4. `internal/modules/finance/service.go` - Add counterparty embedding logic

---

## STEP 7: POSTMAN EXAMPLES

See [postman_examples.md](./postman_examples.md) for complete request/response examples.

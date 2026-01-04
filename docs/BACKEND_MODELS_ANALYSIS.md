# LEORA BACKEND - TO'LIQ MODEL TAHLILI

Bu hujjat Finance, Planner va User Profile (More page) modellarining to'liq tahlilini o'z ichiga oladi.

---

## 1. USER / AUTH DOMAIN

### 1.1 User Model

```typescript
interface User {
  id: string
  email: string
  fullName: string
  username?: string
  phoneNumber?: string
  passwordHash: string  // Backend uchun qo'shiladi
  bio?: string
  birthday?: string

  // Profile visibility
  visibility: 'public' | 'friends' | 'private'

  // User preferences
  preferences: {
    showLevel?: boolean
    showAchievements?: boolean
    showStatistics?: boolean
  }

  // Region & Currency
  region: FinanceRegion
  primaryCurrency: FinanceCurrency

  // Verification status
  isEmailVerified: boolean
  isPhoneVerified: boolean

  // Profile
  profileImage?: string  // S3/R2 URL

  // Gamification
  xpPoints: number
  level: number

  // Timestamps
  lastLoginAt?: Date
  createdAt: Date
  updatedAt: Date
  deletedAt?: Date
}
```

### 1.2 Refresh Token Model

```typescript
interface RefreshToken {
  id: string
  userId: string
  token: string
  expiresAt: Date
  isRevoked: boolean
  deviceInfo?: {
    deviceName: string
    platform: string
    lastUsedAt: Date
  }
  createdAt: Date
}
```

### 1.3 User Settings Model (Backend uchun yangi)

```typescript
interface UserSettings {
  id: string
  userId: string

  // Theme
  theme: 'light' | 'dark' | 'auto'
  language: 'en' | 'ru' | 'uz' | 'ar' | 'tr'

  // Notifications
  notifications: {
    enabled: boolean
    sound: boolean
    vibration: boolean
    lockScreen: boolean

    // Finance notifications
    budgetOverspend: boolean
    debtReminder: boolean
    unusualSpends: boolean
    goalProgress: boolean

    // Planner notifications
    taskReminder: boolean
    deadlineReminder: boolean
    rescheduleSuggestion: boolean

    // Habit notifications
    morningHabits: boolean
    nightHabits: boolean
    streakReminder: boolean
    motivationalMessages: boolean

    // AI notifications
    smartRecommendations: boolean
    insights: boolean
    mentorAdvice: boolean
    predictions: boolean

    // Do Not Disturb
    dndEnabled: boolean
    dndStartTime?: string  // "22:00"
    dndEndTime?: string    // "07:00"
    dndWeekends: boolean
  }

  // Focus Mode settings
  focusMode: {
    defaultDuration: number     // minutes
    breakDuration: number       // minutes
    longBreakDuration: number   // minutes
    sessionsBeforeLongBreak: number
  }

  // Security
  security: {
    biometricEnabled: boolean
    pinEnabled: boolean
    pinHash?: string
    askOnLaunch: boolean
    autolockTimeout: number     // seconds (0, 30, 60, 300, 600, -1 for never)
    unlockGracePeriod: number   // seconds
    hideBalances: boolean
    screenshotBlock: boolean
    fakeAccountEnabled: boolean
  }

  // AI Settings
  ai: {
    helpLevel: 'minimal' | 'medium' | 'maximum'
    voiceRecognition: boolean
    transactionCategories: boolean
    smartReminders: boolean
    predictions: boolean
    motivationalMessages: boolean
    scheduleOptimization: boolean

    // Personalization
    assistantName: string       // default: "Leora"
    talkStyle: 'friendly' | 'formal' | 'casual'

    // Virtual Mentors
    activeMentors: string[]     // ["warren_buffett", "elon_musk", "marcus_aurelius"]
  }

  // Data & Sync
  data: {
    autoBackup: boolean
    backupFrequency: 'daily' | 'weekly' | 'monthly'
    backupLocation: 'icloud' | 'google_drive' | 'local'
    lastBackupAt?: Date
    lastSyncAt?: Date
  }

  // Privacy
  privacy: {
    anonymousAnalytics: boolean
    personalizedAds: boolean
    shareWithPartners: boolean
  }

  createdAt: Date
  updatedAt: Date
}
```

### 1.4 User Session Model (Backend uchun yangi)

```typescript
interface UserSession {
  id: string
  userId: string
  deviceName: string
  platform: 'ios' | 'android' | 'web'
  ipAddress: string
  isCurrent: boolean
  lastActiveAt: Date
  createdAt: Date
}
```

### 1.5 Achievement Model (Backend uchun yangi)

```typescript
type AchievementCategory = 'financial' | 'efficiency' | 'habits' | 'social' | 'special'

interface Achievement {
  id: string
  code: string              // unique identifier
  name: string
  description: string
  category: AchievementCategory
  iconUrl?: string
  xpReward: number
  isHidden: boolean         // hidden until unlocked
  requirements: {
    type: string            // e.g., "transaction_count", "streak_days"
    target: number
  }
  createdAt: Date
}

interface UserAchievement {
  id: string
  userId: string
  achievementId: string
  progress: number          // current progress towards target
  isUnlocked: boolean
  unlockedAt?: Date
  createdAt: Date
  updatedAt: Date
}
```

---

## 2. FINANCE DOMAIN

### 2.1 Account Model

```typescript
type AccountType = 'cash' | 'card' | 'savings' | 'investment' | 'credit' | 'debt' | 'other'
type ShowStatus = 'active' | 'archived' | 'deleted'

interface Account {
  id: string
  userId: string
  name: string
  accountType: AccountType
  currency: FinanceCurrency
  initialBalance: number
  currentBalance: number
  linkedGoalId?: string
  customTypeId?: string
  icon?: string
  color?: string
  showStatus: ShowStatus

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}
```

### 2.2 Transaction Model

```typescript
type TransactionType = 'income' | 'expense' | 'transfer'

interface Transaction {
  id: string
  userId: string
  type: TransactionType
  showStatus: ShowStatus

  // Account references
  accountId?: string
  fromAccountId?: string      // transfer uchun
  toAccountId?: string        // transfer uchun

  // Amount & Currency
  amount: number
  currency: FinanceCurrency
  baseCurrency: FinanceCurrency
  rateUsedToBase: number
  convertedAmountToBase: number

  // Transfer specific
  toAmount?: number
  toCurrency?: FinanceCurrency
  effectiveRateFromTo?: number

  // Fee
  feeAmount?: number
  feeCategoryId?: string

  // Categories
  categoryId?: string
  subcategoryId?: string

  // Details
  name?: string
  description?: string
  date: Date
  time?: string

  // Relationships
  goalId?: string
  budgetId?: string
  debtId?: string
  habitId?: string
  counterpartyId?: string

  // Split transactions
  splits?: TransactionSplit[]

  // Recurring
  recurringId?: string

  // Attachments
  attachments?: string[]    // S3 URLs
  tags?: string[]

  // Flags
  isBalanceAdjustment: boolean
  skipBudgetMatching: boolean

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface TransactionSplit {
  id: string
  transactionId: string
  categoryId: string
  amount: number
  createdAt: Date
}
```

### 2.3 Budget Model

```typescript
type BudgetType = 'category' | 'project'
type BudgetPeriodType = 'none' | 'weekly' | 'monthly' | 'custom_range'
type BudgetFlowType = 'income' | 'expense'

interface Budget {
  id: string
  userId: string
  name: string
  budgetType: BudgetType
  categoryIds?: string[]
  linkedGoalId?: string
  accountId?: string
  transactionType?: BudgetFlowType
  currency: FinanceCurrency
  limitAmount: number
  periodType: BudgetPeriodType
  startDate?: Date
  endDate?: Date

  // Calculated fields
  spentAmount: number
  remainingAmount: number
  percentUsed: number
  isOverspent: boolean

  // Settings
  rolloverMode?: 'none' | 'carryover'
  notifyOnExceed: boolean
  showStatus: ShowStatus

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface BudgetEntry {
  id: string
  budgetId: string
  transactionId: string
  appliedAmountBudgetCurrency: number
  rateUsedTxnToBudget: number
  snapshottedAt: Date

  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'
}
```

### 2.4 Debt Model

```typescript
type DebtDirection = 'i_owe' | 'they_owe_me'
type DebtStatus = 'active' | 'paid' | 'overdue' | 'canceled'

interface Debt {
  id: string
  userId: string
  direction: DebtDirection
  counterpartyId?: string
  counterpartyName: string
  description?: string

  // Principal amount
  principalAmount: number
  principalOriginalAmount?: number
  principalCurrency: FinanceCurrency
  principalOriginalCurrency?: FinanceCurrency
  baseCurrency: FinanceCurrency
  rateOnStart: number
  principalBaseValue: number

  // Repayment
  repaymentCurrency?: FinanceCurrency
  repaymentAmount?: number
  repaymentRateOnStart?: number
  isFixedRepaymentAmount: boolean

  // Dates
  startDate: Date
  dueDate?: Date

  // Interest
  interestMode?: 'simple' | 'compound'
  interestRateAnnual?: number
  scheduleHint?: string

  // Links
  linkedGoalId?: string
  linkedBudgetId?: string
  fundingAccountId?: string
  fundingTransactionId?: string

  // Dual account system
  lentFromAccountId?: string
  returnToAccountId?: string
  receivedToAccountId?: string
  payFromAccountId?: string

  customRateUsed?: number

  // Reminder
  reminderEnabled: boolean
  reminderTime?: string

  status: DebtStatus
  showStatus: ShowStatus

  // Settlement
  settledAt?: Date
  finalRateUsed?: number
  finalProfitLoss?: number
  finalProfitLossCurrency?: FinanceCurrency
  totalPaidInRepaymentCurrency?: number

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface DebtPayment {
  id: string
  debtId: string
  amount: number
  currency: FinanceCurrency
  baseCurrency: FinanceCurrency
  rateUsedToBase: number
  convertedAmountToBase: number
  rateUsedToDebt: number
  convertedAmountToDebt: number
  paymentDate: Date
  accountId?: string
  note?: string
  relatedTransactionId?: string

  createdAt: Date
  updatedAt: Date
}
```

### 2.5 FX Rate Model

```typescript
type FxRateSource = 'cbu' | 'cbr' | 'tcmb' | 'sama' | 'cbuae' | 'ecb' | 'fed' | 'boe' | 'market_api' | 'manual'

interface FxRate {
  id: string
  date: Date
  fromCurrency: FinanceCurrency
  toCurrency: FinanceCurrency
  rate: number
  rateMid?: number
  rateBid?: number
  rateAsk?: number
  nominal: number
  spreadPercent?: number
  source: FxRateSource
  isOverridden: boolean
  effectiveFrom?: Date
  effectiveUntil?: Date

  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}
```

### 2.6 Counterparty Model

```typescript
interface Counterparty {
  id: string
  userId: string
  displayName: string
  phoneNumber?: string
  comment?: string
  searchKeywords?: string

  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}
```

### 2.7 Finance Category Model (Backend uchun yangi)

```typescript
interface FinanceCategory {
  id: string
  userId?: string           // null = system default, string = user custom
  name: string
  icon: string
  color: string
  type: 'income' | 'expense' | 'both'
  parentCategoryId?: string // subcategory uchun
  isDefault: boolean
  isHidden: boolean
  sortOrder: number

  createdAt: Date
  updatedAt: Date
}
```

### 2.8 Recurring Transaction Model (Backend uchun yangi)

```typescript
type RecurringFrequency = 'daily' | 'weekly' | 'biweekly' | 'monthly' | 'quarterly' | 'yearly'

interface RecurringTransaction {
  id: string
  userId: string
  name: string
  type: TransactionType
  amount: number
  currency: FinanceCurrency
  accountId: string
  toAccountId?: string      // transfer uchun
  categoryId?: string
  frequency: RecurringFrequency
  dayOfWeek?: number        // 0-6 (weekly uchun)
  dayOfMonth?: number       // 1-31 (monthly uchun)
  startDate: Date
  endDate?: Date
  nextOccurrence: Date
  isActive: boolean

  // Auto-create settings
  autoCreate: boolean
  notifyBefore: number      // days before

  createdAt: Date
  updatedAt: Date
}
```

---

## 3. PLANNER DOMAIN

### 3.1 Goal Model

```typescript
type GoalType = 'financial' | 'health' | 'education' | 'productivity' | 'personal'
type GoalStatus = 'active' | 'paused' | 'completed' | 'archived'
type MetricKind = 'none' | 'amount' | 'weight' | 'count' | 'duration' | 'custom'
type FinanceMode = 'save' | 'spend' | 'debt_close'
type GoalDirection = 'increase' | 'decrease' | 'neutral'
type GoalProgressSource = 'manual' | 'task' | 'habit' | 'finance'

interface Goal {
  id: string
  userId: string
  title: string
  description?: string
  goalType: GoalType
  status: GoalStatus
  showStatus: ShowStatus

  // Metrics
  metricType: MetricKind
  direction: GoalDirection
  unit?: string
  initialValue?: number
  targetValue?: number
  progressTargetValue?: number
  currentValue: number

  // Finance link
  financeMode?: FinanceMode
  currency?: FinanceCurrency
  linkedBudgetId?: string
  linkedDebtId?: string

  // Dates
  startDate?: Date
  targetDate?: Date
  completedDate?: Date

  // Progress
  progressPercent: number

  // Stats (embedded as JSONB)
  stats: GoalStats

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface GoalStats {
  financialProgressPercent?: number
  habitsProgressPercent?: number
  tasksProgressPercent?: number
  focusMinutesLast30?: number
}

interface GoalMilestone {
  id: string
  goalId: string
  title: string
  description?: string
  targetPercent: number
  dueDate?: Date
  completedAt?: Date
  displayOrder: number

  createdAt: Date
}

interface GoalCheckIn {
  id: string
  goalId: string
  value: number
  note?: string
  sourceType: GoalProgressSource
  sourceId?: string
  dateKey: string           // "YYYY-MM-DD"

  createdAt: Date
}
```

### 3.2 Habit Model

```typescript
type HabitStatus = 'active' | 'paused' | 'archived'
type HabitType = 'health' | 'finance' | 'productivity' | 'education' | 'personal' | 'custom'
type Frequency = 'daily' | 'weekly' | 'custom'
type CompletionMode = 'boolean' | 'numeric'
type HabitCountingType = 'create' | 'quit'
type HabitDifficulty = 'easy' | 'medium' | 'hard'
type HabitPriority = 'low' | 'medium' | 'high'

interface Habit {
  id: string
  userId: string
  title: string
  description?: string
  iconId?: string
  habitType: HabitType
  status: HabitStatus
  showStatus: ShowStatus

  // Legacy single goal (deprecated)
  goalId?: string

  // Frequency
  frequency: Frequency
  daysOfWeek?: number[]     // 0-6
  timesPerWeek?: number
  customDates?: string[]
  timeOfDay?: string

  // Completion
  completionMode: CompletionMode
  targetPerDay?: number
  unit?: string

  // Finance rule (embedded as JSONB)
  financeRule?: HabitFinanceRule

  // Settings
  challengeLengthDays?: number
  countingType: HabitCountingType
  difficulty: HabitDifficulty
  priority: HabitPriority

  // Reminder
  reminderEnabled: boolean
  reminderTime?: string

  // Stats
  streakCurrent: number
  streakBest: number
  completionRate30d: number

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface HabitFinanceRule {
  rule: string
  categories: string[]
  thresholdAmount?: number
  currency?: FinanceCurrency
}

interface HabitCompletion {
  id: string
  habitId: string
  dateKey: string           // "YYYY-MM-DD"
  status: 'done' | 'miss'
  value?: number

  createdAt: Date
}
```

### 3.3 Task Model

```typescript
type TaskStatus = 'inbox' | 'planned' | 'in_progress' | 'completed' | 'canceled' | 'moved' | 'overdue' | 'active' | 'archived' | 'deleted'
type TaskPriority = 'low' | 'medium' | 'high'
type TaskFinanceLink = 'record_expenses' | 'pay_debt' | 'review_budget' | 'transfer_money' | 'none'

interface Task {
  id: string
  userId: string
  title: string
  status: TaskStatus
  showStatus: ShowStatus
  priority: TaskPriority

  // Links
  goalId?: string
  habitId?: string

  // Progress (goal contribution)
  progressValue?: number
  progressUnit?: string

  // Finance link
  financeLink?: TaskFinanceLink

  // Schedule
  dueDate?: Date
  startDate?: Date
  timeOfDay?: string
  estimatedMinutes?: number

  // Focus
  needFocus: boolean
  energyLevel?: number      // 1-3

  // Reminder
  reminderEnabled: boolean

  // Context
  context?: string          // @inbox, @home, @learning, etc.
  notes?: string

  // Focus tracking
  lastFocusSessionId?: string
  focusTotalMinutes: number

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}

interface TaskChecklistItem {
  id: string
  taskId: string
  title: string
  completed: boolean
  displayOrder: number

  createdAt: Date
}

interface TaskDependency {
  id: string
  taskId: string
  dependsOnTaskId: string
  status: 'pending' | 'met'

  createdAt: Date
}
```

### 3.4 Focus Session Model

```typescript
type FocusStatus = 'in_progress' | 'completed' | 'canceled' | 'paused'

interface FocusSession {
  id: string
  userId: string
  taskId?: string
  goalId?: string
  plannedMinutes: number
  actualMinutes?: number
  status: FocusStatus
  startedAt: Date
  endedAt?: Date
  interruptionsCount: number
  notes?: string

  // Sync
  idempotencyKey?: string
  syncStatus: 'local' | 'synced' | 'pending'

  createdAt: Date
  updatedAt: Date
}
```

---

## 4. MANY-TO-MANY JUNCTION TABLES

### 4.1 Goal-Habit Junction

```typescript
interface GoalHabit {
  goalId: string
  habitId: string
  createdAt: Date
}
```

### 4.2 Goal-Task Junction

```typescript
interface GoalTask {
  goalId: string
  taskId: string
  createdAt: Date
}
```

---

## 5. INTEGRATIONS DOMAIN (Yangi - More Page asosida)

### 5.1 Integration Model

```typescript
type IntegrationType = 'calendar' | 'bank' | 'app' | 'device'
type IntegrationStatus = 'connected' | 'disconnected' | 'error' | 'pending'

interface Integration {
  id: string
  userId: string
  type: IntegrationType
  providerId: string        // e.g., "google_calendar", "uzcard", "telegram"
  providerName: string
  status: IntegrationStatus

  // OAuth tokens (encrypted)
  accessToken?: string
  refreshToken?: string
  tokenExpiresAt?: Date

  // Settings
  settings: Record<string, any>

  // Sync
  lastSyncAt?: Date
  syncError?: string

  createdAt: Date
  updatedAt: Date
}
```

### 5.2 Integration Providers (hardcoded constants)

```typescript
const INTEGRATION_PROVIDERS = {
  calendars: [
    { id: 'google_calendar', name: 'Google Calendar' },
    { id: 'apple_calendar', name: 'Apple Calendar' },
    { id: 'outlook_calendar', name: 'Outlook Calendar' }
  ],
  banks: [
    { id: 'uzcard', name: 'Uzcard' },
    { id: 'humo', name: 'Humo' },
    { id: 'kapitalbank', name: 'Kapitalbank' },
    { id: 'ipoteka_bank', name: 'Ipoteka Bank' }
  ],
  apps: [
    { id: 'telegram', name: 'Telegram' },
    { id: 'whatsapp', name: 'WhatsApp' },
    { id: 'slack', name: 'Slack' },
    { id: 'notion', name: 'Notion' },
    { id: 'todoist', name: 'Todoist' },
    { id: 'spotify', name: 'Spotify' },
    { id: 'strava', name: 'Strava' },
    { id: 'myfitnesspal', name: 'MyFitnessPal' }
  ],
  devices: [
    { id: 'apple_watch', name: 'Apple Watch' },
    { id: 'wear_os', name: 'Wear OS' }
  ]
}
```

---

## 6. SUBSCRIPTION / PREMIUM DOMAIN

### 6.1 Subscription Model

```typescript
type SubscriptionPlan = 'free' | 'premium_monthly' | 'premium_yearly' | 'premium_lifetime'
type SubscriptionStatus = 'active' | 'canceled' | 'expired' | 'trial'

interface Subscription {
  id: string
  userId: string
  plan: SubscriptionPlan
  status: SubscriptionStatus

  // Billing
  priceAmount: number
  priceCurrency: string
  billingCycle?: 'monthly' | 'yearly'

  // Dates
  startDate: Date
  endDate?: Date
  trialEndDate?: Date
  canceledAt?: Date

  // Provider
  paymentProvider?: 'stripe' | 'apple' | 'google'
  externalSubscriptionId?: string

  createdAt: Date
  updatedAt: Date
}

interface PaymentHistory {
  id: string
  userId: string
  subscriptionId: string
  amount: number
  currency: string
  status: 'succeeded' | 'failed' | 'refunded'
  paymentMethod?: string
  externalPaymentId?: string

  createdAt: Date
}
```

---

## 7. NOTIFICATION DOMAIN

### 7.1 Notification Model

```typescript
type NotificationType =
  | 'budget_overspend' | 'debt_reminder' | 'goal_progress'
  | 'task_reminder' | 'habit_reminder' | 'streak_warning'
  | 'ai_insight' | 'system'

interface Notification {
  id: string
  userId: string
  type: NotificationType
  title: string
  body: string
  data?: Record<string, any>
  isRead: boolean
  readAt?: Date

  // Scheduling
  scheduledFor?: Date
  sentAt?: Date

  createdAt: Date
}
```

### 7.2 Push Token Model

```typescript
interface PushToken {
  id: string
  userId: string
  token: string
  platform: 'ios' | 'android' | 'web'
  deviceId: string
  isActive: boolean

  createdAt: Date
  updatedAt: Date
}
```

---

## 8. AI DOMAIN

### 8.1 AI Insight Model

```typescript
type InsightKind = 'finance' | 'planner' | 'habit' | 'focus' | 'combined' | 'wisdom'
type InsightLevel = 'info' | 'warning' | 'critical' | 'celebration'
type InsightScope = 'daily' | 'weekly' | 'monthly' | 'custom'

interface AiInsight {
  id: string
  userId: string
  kind: InsightKind
  level: InsightLevel
  scope: InsightScope
  title: string
  body: string

  // Related entities
  relatedGoalId?: string
  relatedBudgetId?: string
  relatedDebtId?: string
  relatedHabitId?: string
  relatedTaskId?: string

  // Actions
  actions?: AiInsightAction[]

  // Validity
  validUntil?: Date
  isRead: boolean
  isDismissed: boolean

  createdAt: Date
}

interface AiInsightAction {
  id: string
  insightId: string
  label: string
  actionType: string      // e.g., "navigate", "create_task", "adjust_budget"
  actionData: Record<string, any>
}
```

### 8.2 Virtual Mentor Model

```typescript
interface VirtualMentor {
  id: string
  code: string              // e.g., "warren_buffett"
  name: string
  description: string
  avatarUrl: string
  specialties: string[]     // e.g., ["finance", "investing"]
  isDefault: boolean

  createdAt: Date
}

interface UserMentor {
  id: string
  userId: string
  mentorId: string
  isActive: boolean

  createdAt: Date
}
```

---

## 9. SYNC DOMAIN

### 9.1 Sync Log Model

```typescript
interface SyncLog {
  id: string
  userId: string
  entityType: string        // e.g., "transaction", "task"
  entityId: string
  operation: 'create' | 'update' | 'delete'
  clientTimestamp: Date
  serverTimestamp: Date
  conflictResolution?: 'server_wins' | 'client_wins' | 'merged'

  createdAt: Date
}
```

---

## 10. RELATIONSHIPS SUMMARY

### Finance Domain
```
Account → Goal (linkedGoalId)
Transaction → Account, Goal, Budget, Debt, Habit, Counterparty
Budget → Goal, Account
BudgetEntry → Budget, Transaction
Debt → Goal, Budget, Counterparty, Account (5 fields)
DebtPayment → Debt, Account, Transaction
```

### Planner Domain
```
Goal ↔ Habit (many-to-many via GoalHabit)
Goal ↔ Task (many-to-many via GoalTask)
Goal → Budget, Debt (finance links)
Habit → Goal (legacy single link)
Task → Goal, Habit
FocusSession → Task, Goal
```

### User Domain
```
User → UserSettings (one-to-one)
User → RefreshToken (one-to-many)
User → UserSession (one-to-many)
User → UserAchievement (one-to-many)
User → Subscription (one-to-one active)
User → Integration (one-to-many)
User → Notification (one-to-many)
User → PushToken (one-to-many)
User → AiInsight (one-to-many)
User → UserMentor (one-to-many)
```

---

## 11. TOTAL MODEL COUNT

| Domain | Models | Junction Tables |
|--------|--------|-----------------|
| User/Auth | 6 | 0 |
| Finance | 8 | 0 |
| Planner | 8 | 2 |
| Integrations | 1 | 0 |
| Subscription | 2 | 0 |
| Notifications | 2 | 0 |
| AI | 4 | 0 |
| Sync | 1 | 0 |
| **TOTAL** | **32** | **2** |

---

## 12. SUPPORTED CURRENCIES

```typescript
type FinanceCurrency = 'UZS' | 'USD' | 'EUR' | 'GBP' | 'TRY' | 'SAR' | 'AED' | 'USDT' | 'RUB'
```

## 13. SUPPORTED REGIONS

```typescript
type FinanceRegion =
  | 'uzbekistan'
  | 'united-states'
  | 'eurozone'
  | 'united-kingdom'
  | 'turkey'
  | 'saudi-arabia'
  | 'united-arab-emirates'
  | 'russia'
  | 'other'
```

## 14. SUPPORTED LANGUAGES

```typescript
type SupportedLanguage = 'en' | 'ru' | 'uz' | 'ar' | 'tr'
```

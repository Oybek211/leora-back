# Categories Backend Prompt

## Mission
Provide canonical, localized category catalogs so finance modules (transactions, budgets, habits) and widgets can rely on consistent IDs, names, and icons.

## Entities & Data Relationships
- Categories are static records (e.g., `food`, `transport`, `housing` for expenses; `salary`, `freelance` for income) with optional subcategories. They include icons and may support translations.
- Finance modules reference these IDs in transactions, budgets, and habit finance rules, so any change must also update corresponding catalog versions consumed by clients.

## API Surface & Responsibilities
- Expose GET `/categories/expenses` and `/categories/income` (or a single `/categories?type=`) returning structured JSON with IDs, names, icons, and subcategory lists.
- Provide cache headers and versioning (e.g., `catalogVersion` or `ETag`) so clients can detect changes and avoid redundant downloads.
- Allow admins to add custom categories (if supported) by storing them per `userId` or `tenant` while falling back to defaults when none exist.

## Real-Time & WebSocket Notes
- Broadcast `entity:updated` or `category:updated` when a catalog entry changes so caches and widgets refresh.

## Operational Considerations
- Keep exports of categories small and pre-seeded; treat them as read-only for most clients to avoid consistency issues.

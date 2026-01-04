# Leora Backend Architecture

This repository follows a clean, modular structure tailored for a Go + Fiber backend that must grow over time while exposing the domains described in `BACKEND_API.md`.

```
leora-server/
├── cmd/
│   └── leora/
│       └── main.go                 # Application entry point (Fiber server, config bootstrapping)
├── configs/
│   └── config.yaml                 # Environment-driven configuration schema
├── internal/
│   ├── app/
│   │   ├── application.go           # High-level application wiring
│   │   ├── server.go                # Fiber app setup + middleware registration
│   │   └── routes.go                # Route registration per module
│   ├── config/
│   │   └── loader.go                # Config loading helpers
│   ├── domain/                      # Entity definitions & business rules
│   │   ├── auth/
│   │   ├── planner/
│   │   │   ├── tasks/
│   │   │   ├── goals/
│   │   │   ├── habits/
│   │   │   └── focus/
│   │   ├── finance/
│   │   ├── insights/
│   │   ├── sync/
│   │   ├── subscriptions/
│   │   ├── settings/
│   │   ├── achievements/
│   │   ├── ai/
│   │   ├── data/
│   │   ├── integrations/
│   │   └── devicesessions/
│   ├── ports/                       # Repository/service interfaces
│   └── transport/
│       └── http/
│           └── v1/
│               ├── handlers/       # Fiber handlers per module
│               ├── middleware/
│               └── dto/            # Request/response payloads
├── pkg/
│   └── middleware/                  # Reusable Fiber middleware (logging, auth, validation)
└── docs/
    └── ARCHITECTURE.md             # This document
```
```

- **Domains**: Each folder under `internal/domain` models the entities from `BACKEND_API.md` (users/auth, planner, finance, insights, and so on) and exposes value objects, enums, and validation rules.
- **Ports**: Interfaces in `internal/ports` describe services, repositories, and external integrations (e.g., token provider, database access) that connect the domain layer to infrastructure.
- **Transport**: API versioning and Fiber handlers live under `internal/transport/http/v1`. Handlers rely on DTOs and interact with ports. Response payloads follow the common `success/data/error` schema.
- **App wiring**: `internal/app` wires together config, Fiber server, middleware, and routes so cmd/main keeps minimal logic.
- **pkg/middleware**: Shared middleware (logging, request ID, error handling, auth guards) are reusable building blocks for each module.

This structure keeps the repository ready for additional modules described in `BACKEND_API.md` while keeping each responsibility isolated for easier testing and future migrations.

# Architecture Refactoring Plan: Vertical Slice Architecture (VSA)

## Goal
Refactor the current monolithic/flat Go application into a **Vertical Slice Architecture** to support future commercial growth (payments, subscriptions, etc.) and optimize for AI-assisted development.

## Core Philosophy
- **Features over Layers:** Code is organized by *what it does* (business feature), not *what it is* (technical layer).
- **Colocation:** Everything a feature needs (handlers, logic, database queries, HTML templates) should be as close together as possible.
- **AI Optimization:** Small, self-contained contexts minimize "hallucinations" and maximize the AI's context window efficiency.

## Target Structure

```text
/cmd
  /server
    main.go           # Entry point: Router wiring, Middleware setup, Config loading
/internal
  /domain             # Shared Kernel (Keep minimal!)
    models.go         # Shared structs like User, Wine (if used across many slices)
  /features
    /auth             # Slice: Authentication
      login_handler.go
      signup_handler.go
      auth_service.go # (Optional: only if logic is complex)
      login.html      # Template colocated with handler
      signup.html
    /wines
      /add            # Slice: Add Wine
        handler.go
        add.html
      /list           # Slice: List Wines
        handler.go
        list.html
      /details        # Slice: Wine Details
        handler.go
        details.html
    /settings         # Slice: User Settings
      handler.go
      settings.html
  /shared             # Infrastructure used by all slices
    /database         # DB connection setup
    /ui               # Shared UI components (Header, Footer, Layouts)
```

## Phase 1: Preparation & Infrastructure (Low Risk) - **COMPLETED**
1.  **Create Directory Structure:**
    - Create `cmd/server`, `internal/features`, `internal/shared/database`.
2.  **Move Database Setup:**
    - Extract the `InitDB` and global `DB` variable from `data.go` into `internal/shared/database/database.go`.
3.  **Move Shared Models:**
    - Move `User` and `Wine` structs from `data.go` to `internal/domain/models.go`.
    - *Note:* In pure VSA, we might duplicate models per slice, but for now, a shared domain model is a safer transition step.

## Phase 2: Feature Migration (Iterative)
*Migrate one feature at a time. The app should remain buildable between steps.*

1.  **Slice 1: Settings (Simple)** - **COMPLETED**
    - Create `internal/features/settings`.
    - Move `settingsHandler` from `main.go` to `internal/features/settings/handler.go`.
    - Move `templates/settings.html` to `internal/features/settings/settings.html`.
    - Update `main.go` to import the new handler.

2.  **Slice 2: Authentication (Medium)** - **COMPLETED**
    - Create `internal/features/auth`.
    - Move `loginHandler`, `signupHandler`, `logoutHandler` to `internal/features/auth`.
    - Move `auth.go` logic (hashing, middleware) into this package or a shared `internal/shared/auth` if needed by many slices.
    - Move related templates.

3.  **Slice 3: Wine Operations (Complex)** - **COMPLETED**
    - Create `internal/features/wines/add`, `internal/features/wines/list`, etc.
    - Move handlers and templates accordingly.
    - Ensure database queries in handlers use the shared `database.DB` instance.
    - **Status:** All wine operations (`list`, `add`, `details`, `edit`, `update`, `delete`) migrated.

4.  **Slice 4: Reviews & Actions** - **COMPLETED**
    - Create `internal/features/reviews/add`.
    - **Status:** `add-review` migrated.

## Phase 3: Cleanup & Entry Point
1.  **Refactor `main.go`:** - **COMPLETED**
    - `main.go` should now only contain:
        - Database initialization.
        - Router setup (mux/chi).
        - Route definitions mapping URLs to Feature Handlers.
        - Server start.
2.  **Template Loading Strategy:** - **COMPLETED**
    - *Challenge:* Go's `ParseFiles` usually expects paths relative to the CWD.
    - *Solution:* Handlers now use paths relative to project root (e.g., `internal/features/...`).

## Future Considerations for Commercial Features
- **Payments:** Create `internal/features/billing`.
- **Emails:** Create `internal/shared/email` (infrastructure) and call it from features.
- **Testing:** Add `handler_test.go` inside each feature folder.

## AI Workflow for Refactoring
When asking the AI to perform these tasks, use prompts like:
> "Refactor the 'Settings' feature. Move the handler from main.go and the HTML template into a new package `internal/features/settings`. Ensure imports are updated."

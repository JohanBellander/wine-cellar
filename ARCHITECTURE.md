# Architecture Documentation

This document provides a high-level overview of the **Winetrackr** project architecture. It is intended to guide future AI agents and developers in understanding the system's design, technology stack, and organizational patterns.

## 1. Overview

Winetrackr is a web application for managing wine collections. It allows users to track their wine inventory, add reviews, and manage their subscription. The application is built using Go for the backend and server-side rendered HTML with Tailwind CSS for the frontend.

## 2. Technology Stack

-   **Language**: Go (Golang) 1.25+
-   **Web Framework**: Standard library `net/http`
-   **Database**: PostgreSQL (hosted on **Neon**)
-   **ORM**: GORM
-   **Frontend**: Server-Side Rendered (SSR) HTML templates (`html/template`), **Tailwind CSS** (via CDN), Vanilla JavaScript
-   **Hosting**: **Render** (Dockerized deployment)
-   **Version Control**: **GitHub**
-   **Payments**: Stripe (for subscription management)

## 3. Architectural Pattern: Vertical Slice Architecture

The project follows the **Vertical Slice Architecture** pattern. Instead of organizing code by technical layers (controllers, services, repositories), the code is organized by **Features**.

Each feature is a self-contained slice that includes everything needed to implement that specific functionality (handlers, business logic, specific data models, and sometimes templates).

### Benefits
-   **High Cohesion**: Related code sits together.
-   **Low Coupling**: Changes in one feature rarely affect others.
-   **Scalability**: Easy to add new features without modifying shared layers.

## 4. Project Structure

```text
.
├── cmd/                    # Application entry points (if separated)
├── internal/               # Private application code
│   ├── features/           # VERTICAL SLICES
│   │   ├── auth/           # Authentication (Login, Signup, Middleware)
│   │   ├── reviews/        # Review management
│   │   ├── settings/       # User settings, GDPR, Export
│   │   ├── subscription/   # Stripe integration
│   │   └── wines/          # Wine management (CRUD operations)
│   │       ├── add/        # Add wine feature
│   │       ├── delete/     # Delete wine feature
│   │       ├── details/    # Wine details view
│   │       ├── edit/       # Edit wine feature
│   │       ├── list/       # Wine list/inventory view
│   │       └── update/     # Update logic (e.g., quantity)
│   └── shared/             # Shared cross-cutting concerns
│       ├── database/       # DB connection and global models
│       └── ui/             # Shared UI utilities
├── static/                 # Static assets (images, css, js)
├── templates/              # Shared HTML templates (layout, header, footer)
├── main.go                 # Application entry point, router setup
├── Dockerfile              # Docker build configuration
├── go.mod                  # Go module definition
└── README.md               # Project documentation
```

## 5. Key Components

### Authentication
Authentication is handled via a custom middleware located in `internal/features/auth`. It manages user sessions and protects routes that require login.

### Database Access
The project uses **GORM** for database interactions. The connection is initialized in `internal/shared/database`. While features may define their own specific data needs, shared models are often kept in the domain or shared packages to avoid circular dependencies.

### Frontend & UI
-   **Templates**: Go's `html/template` engine is used for rendering views.
-   **Styling**: Tailwind CSS is used for styling. It is currently loaded via CDN for simplicity.
-   **Components**: Shared UI parts like the header and footer are located in `templates/` and included via `{{template "name" .}}`.

## 6. Deployment & Infrastructure

-   **Source Code**: Hosted on **GitHub**.
-   **CI/CD**: Pushes to the `master` branch trigger a deployment on **Render**.
    -   **Secrets Management**: Sensitive configuration (Stripe keys, Database URL, Domain) is stored in **GitHub Secrets**.
    -   **Auto-Sync**: The GitHub Actions workflow (`.github/workflows/ci.yml`) automatically syncs these secrets to Render's environment variables during every deployment. This ensures the production environment is always up-to-date with the configuration in GitHub.
-   **Containerization**: The application is packaged using **Docker**. The `Dockerfile` builds a lightweight Alpine-based image.
    -   *Note*: The `static/` directory is copied into the image to serve assets in production.
-   **Database**: A managed PostgreSQL instance hosted on **Neon**. Connection string is provided via environment variables.

## 7. Development Workflow

1.  **Local Setup**:
    -   Ensure Go is installed.
    -   Create a `.env` file with necessary credentials (DB URL, Stripe keys).
2.  **Running the App**:
    -   `go run main.go`
3.  **Making Changes**:
    -   Identify the relevant feature slice in `internal/features/`.
    -   Modify the handler, logic, or template within that slice.
    -   If adding a new feature, create a new directory under `internal/features/`.

## 8. Future Considerations for AI Agents

-   **Context**: When working on a task, first identify which "Slice" (feature) it belongs to.
-   **Isolation**: Try to keep changes contained within the specific feature folder.
-   **Shared Code**: Only modify `internal/shared/` if the change truly applies globally.
-   **Templates**: Be aware that some templates are feature-specific (inside `features/`) while others are global (inside `templates/`).

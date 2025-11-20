# Authentication Implementation Plan

## Phase 1: Dependencies & Database
- [ ] Install dependencies:
  - `golang.org/x/crypto/bcrypt` (Password hashing)
  - `github.com/gorilla/sessions` (Session management)
- [ ] Update `data.go`:
  - Create `User` struct (`ID`, `Email`, `PasswordHash`).
  - Add `UserID` foreign key to `Wine` struct.
  - Update `InitDB` to auto-migrate the new `User` table.

## Phase 2: Authentication Logic
- [ ] Create `auth.go`:
  - Initialize `CookieStore` for sessions.
  - Implement `HashPassword(password string)`.
  - Implement `CheckPasswordHash(password, hash string)`.
  - Create `AuthMiddleware(next http.HandlerFunc)` to protect routes.
  - Create helper `GetUserIDFromSession(r *http.Request)`.

## Phase 3: User Interface
- [ ] Create `templates/signup.html`: Registration form.
- [ ] Create `templates/login.html`: Login form.
- [ ] Update `templates/header.html`:
  - Show "Log In" / "Sign Up" when logged out.
  - Show "Log Out" when logged in.

## Phase 4: Wiring It Up (`main.go`)
- [ ] Implement Handlers:
  - `signupHandler` (GET/POST)
  - `loginHandler` (GET/POST)
  - `logoutHandler` (POST)
- [ ] Protect Routes:
  - Wrap `listHandler`, `addHandler`, `editHandler`, `detailsHandler`, `deleteHandler` with `AuthMiddleware`.
- [ ] Scope Data to User:
  - Update `listHandler` to fetch only current user's wines.
  - Update `addHandler` to save `UserID` on creation.
  - Update `details/edit/delete` handlers to verify ownership.

## Phase 5: Deployment
- [ ] Ensure `SESSION_SECRET` environment variable is supported.
- [ ] Verify Docker build includes new dependencies.

# Subscription Model Implementation Plan

## Overview
Implement a "Freemium" model using Stripe.
- **Free Tier**: Limited to 10 wines.
- **Pro Tier**: Unlimited wines.

## 1. Database Schema Updates
We need to store subscription details in the `users` table.

### New Fields for `User` struct:
- `SubscriptionTier` (string): "free" or "pro" (default: "free").
- `StripeCustomerID` (string): The customer ID from Stripe.
- `SubscriptionStatus` (string): "active", "past_due", "canceled", "incomplete".
- `SubscriptionID` (string): The ID of the active subscription.

## 2. Logic & Enforcement (The "Free" Limit)
Enforce the limit at the application level before database insertion.

### `internal/features/wines/add/handler.go`
- Check user's current wine count.
- If `SubscriptionTier == "free"` AND `count >= 10`:
    - Prevent addition.
    - Return an error or redirect to an upgrade page.

## 3. Stripe Integration (Backend)
Use the official Stripe Go library.

### Dependencies
- `github.com/stripe/stripe-go/v74`

### Components
1.  **Checkout Session**: Create a handler to generate a Stripe Checkout session URL for the "Pro" plan.
2.  **Webhooks**: Create a handler (`/webhook/stripe`) to listen for:
    - `checkout.session.completed`: Upgrade user to "pro".
    - `customer.subscription.updated`: Handle status changes (e.g., past due).
    - `customer.subscription.deleted`: Downgrade user to "free".
3.  **Customer Portal**: Create a handler to generate a link to the Stripe Customer Portal for managing billing.

## 4. UI Updates

### Settings Page (`internal/features/settings/settings.html`)
- Display current plan ("Free" or "Pro").
- If Free: Show "Upgrade to Pro" button.
- If Pro: Show "Manage Subscription" button (links to Stripe Portal).

### Add Wine Page (`internal/features/wines/add/add.html`)
- Display usage: "You have used X of 10 free wines."
- If limit reached: Disable the form and show an upgrade prompt.

### List Page (`internal/features/wines/list/list.html`)
- Optional: Show a small banner or indicator of the current tier.

## 5. Implementation Roadmap

### Phase 1: Foundation & Enforcement
- [ ] Update `User` model in `internal/domain/models.go`.
- [ ] Run migrations (auto-migration in `database.go`).
- [ ] Implement the 10-wine check in `internal/features/wines/add/handler.go`.
- [ ] Update UI to show the limit warning.

### Phase 2: Stripe Setup
- [ ] Sign up for Stripe (Dev mode).
- [ ] Get API Keys (Publishable, Secret, Webhook Secret).
- [ ] Create the "Pro" product in Stripe Dashboard.

### Phase 3: Backend Integration
- [ ] Install Stripe Go library.
- [ ] Create `internal/features/subscription` package.
- [ ] Implement Checkout Session handler.
- [ ] Implement Webhook handler.
- [ ] Implement Portal handler.

### Phase 4: Frontend Integration
- [ ] Connect "Upgrade" button to Checkout handler.
- [ ] Connect "Manage" button to Portal handler.
- [ ] Test the full flow using Stripe Test Cards.

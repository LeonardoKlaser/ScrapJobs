# MVP Polish & Launch Readiness — Design Document

**Date:** 2026-02-27
**Status:** Approved

## Overview

Final polishing pass before ScrapJobs MVP launch. Covers bug fixes, UI/UX improvements, subscription control, password recovery, account deletion, curriculum workflow refactoring, and legal pages.

---

## 1. Bug Fixes & UI/UX

### 1.1 Credit Card Disabled on Checkout
- Visually disable (opacity-50, pointer-events-none) the credit card option in `checkout.tsx`
- Add "Em breve" badge overlay
- Default `paymentMethod` to `'pix'`
- Keep card logic in code for future activation

### 1.2 Back Button on Checkout
- Add `← Voltar para a Página Inicial` link at top of checkout page
- Uses React Router `<Link to={PATHS.landing}>`

### 1.3 Site Counter Bug (Account Page)
- **Root cause:** `PlanSection` receives `currentUsage` but parent doesn't fetch real count
- **Fix backend:** Include `monitored_sites_count` in `GET /api/me` response using `GetUserSiteCount(userId)`
- **Fix frontend:** `PlanSection` reads count from user API response

### 1.4 Dashboard Table Overflow
- Fixed column widths: Title 40%, Company 20%, Location 15%, Link 10%, Action 15%
- Text truncation via `max-w-0 truncate` with native `title` tooltip
- Remove `min-w-[600px]` from `<Table>` to prevent unnecessary horizontal scroll

### 1.5 Login Page Redesign
- Desktop: split layout — brand hero (left) + login form (right)
- Left panel: ScrapJobs branding, headline, value proposition stats
- Right panel: clean form with generous spacing
- Mobile: form only, centered (current layout, polished)
- Add "Esqueci minha senha" link below password field

---

## 2. Curriculum & AI Flow Refactoring

### 2.1 Delete Curriculum
- **Backend:** `DELETE /curriculum/:id` — validates ownership, enforces minimum 1 curriculum
- **Frontend:** Trash icon in `CurriculumList` with confirmation dialog

### 2.2 Remove `is_active` Concept
- **Migration:** `ALTER TABLE curriculums DROP COLUMN is_active`, drop `partial_unique_active_curriculum` constraint
- **Backend:** Remove `PATCH /curriculum/:id/active` route and `SetActiveCurriculum` logic
- **Frontend:** Remove all `is_active` references from curriculum components

### 2.3 Curriculum Selection Modal for AI Analysis
- **Backend:** `POST /api/analyze-job` now accepts `{ job_id, curriculum_id }`
  - Validates curriculum belongs to authenticated user
  - Calls OpenAI with selected curriculum + job
  - Saves analysis result as JSONB in `job_notifications`
- **Frontend:** `AnalysisDialog` becomes 2-step flow:
  1. Step 1: List user's curricula as selectable cards → "Gerar Análise" button
  2. Step 2: Loading → Result display (existing behavior)

### 2.4 Analysis History & "Redo Analysis"
- **Migration:** Add `analysis_result JSONB NULL` and `curriculum_id INT NULL` to `job_notifications`
- **Backend:** `GET /api/analyze-job/history?job_id=X` returns previous analysis if exists
- **Frontend:** On modal open:
  - If previous analysis exists → show result + "Refazer Análise" button
  - If not → go to curriculum selection

### 2.5 Gemini → OpenAI Migration
- Already completed — backend uses `infra/openai/openai_client.go`
- Model configurable via `OPENAI_MODEL` env var (default: `gpt-4o-mini`)
- Remove any residual Gemini references

---

## 3. Core Business Features

### 3.1 Forgot Password
- **Migration:** New table `password_reset_tokens`:
  - `id SERIAL PRIMARY KEY`
  - `user_id INT REFERENCES users(id)`
  - `token UUID DEFAULT gen_random_uuid()`
  - `expires_at TIMESTAMP NOT NULL` (NOW() + 1h)
  - `used_at TIMESTAMP NULL`
  - `created_at TIMESTAMP DEFAULT NOW()`
- **Backend:**
  - `POST /api/forgot-password` (public, 3 req/min): lookup user → generate token → send SES email
  - `POST /api/reset-password` (public): validate token + TTL + unused → update password → mark used
  - Email template: branded HTML with reset link `{FRONTEND_URL}/reset-password?token=xxx`
- **Frontend:**
  - `/forgot-password` route: email input form
  - `/reset-password` route: new password + confirmation (token from query param)
  - "Esqueci minha senha" link on login page

### 3.2 Delete Account (Soft Delete)
- **Migration:** `ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL`
- **Backend:**
  - `DELETE /api/user/account` (protected): sets `deleted_at = NOW()`, requires password confirmation
  - Auth middleware: reject login if `deleted_at IS NOT NULL` → "Conta desativada"
  - All user queries filter `WHERE deleted_at IS NULL`
- **Frontend:**
  - "Excluir minha conta" button in Security tab of account page
  - Confirmation dialog requiring password
  - On success: logout → redirect to landing with confirmation toast

### 3.3 Subscription Control (`expires_at`)
- **Migration:** `ALTER TABLE users ADD COLUMN expires_at TIMESTAMP NULL`
  - Backfill: `UPDATE users SET expires_at = NOW() + INTERVAL '30 days' WHERE plan_id IS NOT NULL`
- **Backend:**
  - Webhook handler: on payment confirmed → `SET expires_at = NOW() + 30 days` (or 365 for annual)
  - New middleware `RequireActiveSubscription`: checks `expires_at > NOW()`, returns 403 `{"error": "subscription_expired"}`
  - `GetActiveUserIDs()`: filter `WHERE expires_at > NOW()`
  - `GET /api/me` includes `expires_at` in response
- **Frontend:**
  - `authLoader` intercepts 403 `subscription_expired` → redirect to renewal page
  - Renewal page: shows plans + direct checkout link
  - **Full access block** when expired — redirect to renewal

### 3.4 Manage Plan
- Account page "Plano" tab: "Gerenciar Plano" button opens dialog with:
  - Current plan + expiration date
  - Available plans (from `GET /api/plans`)
  - "Renovar" button per plan → redirect to `/checkout/:planId`
- No refund logic

---

## 4. Legal Pages

### 4.1 Terms of Service & Privacy Policy
- New public routes: `/terms`, `/privacy`
- Static page components with structured lorem ipsum sections
- Footer links updated from `href="#"` to actual routes
- Consistent styling with existing page layouts

---

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Subscription enforcement | Backend middleware | Prevents API bypass, single source of truth |
| Password reset token | UUID in DB table | Allows invalidation, audit trail, 1h TTL |
| Account deletion | Soft delete (`deleted_at`) | LGPD compliance, 30-day recovery window |
| AI provider | OpenAI (already migrated) | Configurable model via `OPENAI_MODEL` env |
| Analysis history | JSONB in `job_notifications` | Avoids new table, natural association with notification |
| Curriculum delete | Enforce min 1 | Prevents user from losing ability to analyze |
| Expired user behavior | Full access block | Clean UX, forces renewal |

## Database Migrations Required

1. `033_add_expires_at_to_users.up.sql` — `expires_at TIMESTAMP NULL` + backfill
2. `034_add_deleted_at_to_users.up.sql` — `deleted_at TIMESTAMP NULL`
3. `035_create_password_reset_tokens.up.sql` — New table
4. `036_drop_is_active_from_curriculums.up.sql` — Remove column + constraint
5. `037_add_analysis_result_to_notifications.up.sql` — `analysis_result JSONB`, `curriculum_id INT`

## Execution Order

**Backend first, then Frontend** — as requested by user.

1. Database migrations (all 5)
2. Backend: subscription middleware, forgot password, delete account, curriculum delete, analysis refactor
3. Frontend: checkout fixes, login redesign, dashboard table, account page fixes, curriculum UI, legal pages

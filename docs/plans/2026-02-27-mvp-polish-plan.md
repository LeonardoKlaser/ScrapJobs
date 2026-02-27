# MVP Polish & Launch Readiness — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Prepare ScrapJobs for MVP launch with subscription control, password recovery, account deletion, curriculum workflow refactoring, UI/UX fixes, and legal pages.

**Architecture:** Backend-first approach — all database migrations and API changes land first, then frontend adapts. Each task is self-contained with its own migration, tests, and commit.

**Tech Stack:** Go 1.24 (Gin, Asynq, lib/pq), React 19 (TypeScript, TanStack Query, shadcn/ui, Tailwind v4, Zod, react-i18next)

---

## Task 1: Migration — Add `expires_at` and `deleted_at` to Users

**Files:**
- Create: `migrations/033_add_expires_at_deleted_at_to_users.up.sql`
- Create: `migrations/033_add_expires_at_deleted_at_to_users.down.sql`
- Modify: `model/user.go`

**Step 1: Write the up migration**

```sql
-- migrations/033_add_expires_at_deleted_at_to_users.up.sql
ALTER TABLE users ADD COLUMN expires_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Backfill: give existing paid users 30 days from now
UPDATE users SET expires_at = NOW() + INTERVAL '30 days' WHERE plan_id IS NOT NULL;

CREATE INDEX idx_users_expires_at ON users (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_users_deleted_at ON users (deleted_at) WHERE deleted_at IS NULL;
```

**Step 2: Write the down migration**

```sql
-- migrations/033_add_expires_at_deleted_at_to_users.down.sql
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS expires_at;
```

**Step 3: Update User model**

In `model/user.go`, add `ExpiresAt` and `DeletedAt` fields:

```go
type User struct {
	Id           int        `json:"id"`
	Name         string     `json:"user_name"`
	Email        string     `json:"email"`
	Password     string     `json:"-"`
	Tax          *string    `json:"tax,omitempty" db:"tax"`
	Cellphone    *string    `json:"cellphone,omitempty" db:"cellphone"`
	IsAdmin      bool       `json:"is_admin"`
	CurriculumId *int       `json:"curriculum_id,omitempty"`
	PlanID       *int       `json:"plan_id,omitempty"`
	Plan         *Plan      `json:"plan,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	DeletedAt    *time.Time `json:"-"`
}
```

Add `"time"` to the import block.

**Step 4: Update UserRepository to read/write new columns**

In `repository/user_repository.go`:

- Update `GetUserByEmail` query: add `u.expires_at, u.deleted_at` to SELECT, add `AND u.deleted_at IS NULL` to WHERE, add scan variables:
  ```go
  var expiresAt sql.NullTime
  var deletedAt sql.NullTime
  // add to Scan: &expiresAt, &deletedAt
  // after scan:
  if expiresAt.Valid {
      userToReturn.ExpiresAt = &expiresAt.Time
  }
  ```

- Update `GetUserById` query identically: add columns to SELECT, add `AND u.deleted_at IS NULL` to WHERE, scan new fields.

- Update `CreateUser` query: add `expires_at` to INSERT:
  ```go
  query := `INSERT INTO users (user_name, email, user_password, cellphone, tax, plan_id, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_name, email`
  // Pass user.ExpiresAt as 7th parameter
  ```

- Add new method `SoftDeleteUser`:
  ```go
  func (usr *UserRepository) SoftDeleteUser(userId int) error {
      query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
      result, err := usr.db.Exec(query, userId)
      if err != nil {
          return fmt.Errorf("error soft-deleting user %d: %w", userId, err)
      }
      rows, _ := result.RowsAffected()
      if rows == 0 {
          return fmt.Errorf("user %d not found or already deleted", userId)
      }
      return nil
  }
  ```

- Add new method `UpdateExpiresAt`:
  ```go
  func (usr *UserRepository) UpdateExpiresAt(userId int, expiresAt time.Time) error {
      query := `UPDATE users SET expires_at = $1 WHERE id = $2`
      _, err := usr.db.Exec(query, expiresAt, userId)
      if err != nil {
          return fmt.Errorf("error updating expires_at for user %d: %w", userId, err)
      }
      return nil
  }
  ```

**Step 5: Update UserRepositoryInterface**

In `interfaces/user_interface.go`, add:
```go
SoftDeleteUser(userId int) error
UpdateExpiresAt(userId int, expiresAt time.Time) error
```

Add `"time"` to imports.

**Step 6: Update `GetActiveUserIDs` to filter by `expires_at`**

In `repository/user_site_repository.go`, change query in `GetActiveUserIDs`:
```go
query := `SELECT DISTINCT us.user_id FROM user_sites us INNER JOIN users u ON us.user_id = u.id WHERE u.expires_at > NOW() AND u.deleted_at IS NULL ORDER BY us.user_id`
```

**Step 7: Run tests**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...`
Expected: BUILD SUCCESS

**Step 8: Commit**

```bash
cd "/Users/erickschaedler/Documents/Scrap Jobs/ScrapJobs"
git add migrations/033_* model/user.go repository/user_repository.go interfaces/user_interface.go repository/user_site_repository.go
git commit -m "feat: add expires_at and deleted_at to users table

Adds subscription expiration tracking and soft delete support.
Backfills existing paid users with 30-day expiration.
Filters GetActiveUserIDs by expires_at > NOW()."
```

---

## Task 2: Migration — Create `password_reset_tokens` Table

**Files:**
- Create: `migrations/034_create_password_reset_tokens.up.sql`
- Create: `migrations/034_create_password_reset_tokens.down.sql`

**Step 1: Write the up migration**

```sql
-- migrations/034_create_password_reset_tokens.up.sql
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token UUID NOT NULL DEFAULT gen_random_uuid(),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_password_reset_tokens_token ON password_reset_tokens (token);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
```

**Step 2: Write the down migration**

```sql
-- migrations/034_create_password_reset_tokens.down.sql
DROP TABLE IF EXISTS password_reset_tokens;
```

**Step 3: Create model**

Create `model/password_reset_token.go`:
```go
package model

import "time"

type PasswordResetToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
```

**Step 4: Create repository**

Create `repository/password_reset_repository.go`:
```go
package repository

import (
	"database/sql"
	"fmt"
	"time"
	"web-scrapper/model"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) CreateToken(userID int, ttl time.Duration) (string, error) {
	query := `INSERT INTO password_reset_tokens (user_id, expires_at) VALUES ($1, NOW() + $2::interval) RETURNING token`
	var token string
	err := r.db.QueryRow(query, userID, fmt.Sprintf("%d seconds", int(ttl.Seconds()))).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("error creating password reset token: %w", err)
	}
	return token, nil
}

func (r *PasswordResetRepository) FindValidToken(token string) (*model.PasswordResetToken, error) {
	query := `SELECT id, user_id, token, expires_at, used_at, created_at FROM password_reset_tokens WHERE token = $1 AND expires_at > NOW() AND used_at IS NULL`
	var t model.PasswordResetToken
	err := r.db.QueryRow(query, token).Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding password reset token: %w", err)
	}
	return &t, nil
}

func (r *PasswordResetRepository) MarkUsed(tokenID int) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, tokenID)
	if err != nil {
		return fmt.Errorf("error marking token as used: %w", err)
	}
	return nil
}
```

**Step 5: Create interface**

Create `interfaces/password_reset_interface.go`:
```go
package interfaces

import (
	"time"
	"web-scrapper/model"
)

type PasswordResetRepositoryInterface interface {
	CreateToken(userID int, ttl time.Duration) (string, error)
	FindValidToken(token string) (*model.PasswordResetToken, error)
	MarkUsed(tokenID int) error
}
```

**Step 6: Commit**

```bash
git add migrations/034_* model/password_reset_token.go repository/password_reset_repository.go interfaces/password_reset_interface.go
git commit -m "feat: add password_reset_tokens table and repository

UUID-based tokens with 1h TTL for password recovery flow."
```

---

## Task 3: Migration — Drop `is_active` from Curriculum, Add Analysis Fields to Notifications

**Files:**
- Create: `migrations/035_drop_is_active_add_analysis_fields.up.sql`
- Create: `migrations/035_drop_is_active_add_analysis_fields.down.sql`
- Modify: `model/curriculum.go`
- Modify: `model/notification.go`

**Step 1: Write the up migration**

```sql
-- migrations/035_drop_is_active_add_analysis_fields.up.sql

-- Drop is_active from curriculum
ALTER TABLE curriculum DROP CONSTRAINT IF EXISTS partial_unique_active_curriculum;
ALTER TABLE curriculum DROP COLUMN IF EXISTS is_active;

-- Add analysis result and curriculum_id to job_notifications
ALTER TABLE job_notifications ADD COLUMN analysis_result JSONB NULL;
ALTER TABLE job_notifications ADD COLUMN curriculum_id INT NULL;
```

**Step 2: Write the down migration**

```sql
-- migrations/035_drop_is_active_add_analysis_fields.down.sql
ALTER TABLE job_notifications DROP COLUMN IF EXISTS curriculum_id;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS analysis_result;
ALTER TABLE curriculum ADD COLUMN is_active BOOLEAN DEFAULT FALSE;
CREATE UNIQUE INDEX partial_unique_active_curriculum ON curriculum (user_id) WHERE is_active = TRUE;
```

**Step 3: Update Curriculum model**

In `model/curriculum.go`, remove `IsActive`:
```go
type Curriculum struct {
	Id          int          `json:"id"`
	Title       string       `json:"title"`
	UserID      int          `json:"user_id"`
	Experiences []Experience `json:"experiences"`
	Skills      string       `json:"skills"`
	Summary     string       `json:"summary"`
	Educations  []Education  `json:"educations"`
	Languages   string       `json:"languages"`
}
```

**Step 4: Update CurriculumRepository**

In `repository/curriculum_repository.go`:
- `FindCurriculumByUserID`: remove `is_active` from SELECT and from `rows.Scan`
- `SetActiveCurriculum` method: delete entirely
- `CreateCurriculum`: no change needed (is_active was not in INSERT)

**Step 5: Update CurriculumRepositoryInterface**

In `interfaces/curriculum_interface.go`, remove `SetActiveCurriculum`:
```go
type CurriculumRepositoryInterface interface {
	CreateCurriculum(curriculum model.Curriculum) (model.Curriculum, error)
	FindCurriculumByUserID(userId int) ([]model.Curriculum, error)
	UpdateCurriculum(curriculum model.Curriculum) (model.Curriculum, error)
	DeleteCurriculum(userId int, curriculumId int) error
	CountCurriculumsByUserID(userId int) (int, error)
}
```

**Step 6: Add DeleteCurriculum and CountCurriculumsByUserID to repository**

In `repository/curriculum_repository.go`, add:
```go
func (cur *CurriculumRepository) DeleteCurriculum(userId int, curriculumId int) error {
	query := `DELETE FROM curriculum WHERE id = $1 AND user_id = $2`
	result, err := cur.connection.Exec(query, curriculumId, userId)
	if err != nil {
		return fmt.Errorf("error deleting curriculum: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("curriculum not found or does not belong to user")
	}
	return nil
}

func (cur *CurriculumRepository) CountCurriculumsByUserID(userId int) (int, error) {
	query := `SELECT COUNT(*) FROM curriculum WHERE user_id = $1`
	var count int
	err := cur.connection.QueryRow(query, userId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting curriculums: %w", err)
	}
	return count, nil
}
```

**Step 7: Update CurriculumUsecase**

In `usecase/curriculum_usecase.go`:
- Remove `SetActiveCurriculum` method
- Add `DeleteCurriculum`:
  ```go
  func (cur *CurriculumUsecase) DeleteCurriculum(userId int, curriculumId int) error {
      count, err := cur.CurriculumRepository.CountCurriculumsByUserID(userId)
      if err != nil {
          return err
      }
      if count <= 1 {
          return fmt.Errorf("não é possível excluir o único currículo")
      }
      return cur.CurriculumRepository.DeleteCurriculum(userId, curriculumId)
  }
  ```
  Add `"fmt"` to imports if not already present.

**Step 8: Update CurriculumController**

In `controller/curriculum_controller.go`:
- Remove `SetActiveCurriculum` method entirely
- Add `DeleteCurriculum`:
  ```go
  func (c *CurriculumController) DeleteCurriculum(ctx *gin.Context) {
      userInterface, exists := ctx.Get("user")
      if !exists {
          ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
          return
      }
      user, ok := userInterface.(model.User)
      if !ok {
          ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido no contexto"})
          return
      }
      curriculumID, err := strconv.Atoi(ctx.Param("id"))
      if err != nil {
          ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de currículo inválido"})
          return
      }
      err = c.curriculumUsecase.DeleteCurriculum(user.Id, curriculumID)
      if err != nil {
          if err.Error() == "não é possível excluir o único currículo" {
              ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
              return
          }
          ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
          return
      }
      ctx.JSON(http.StatusOK, gin.H{"message": "Currículo excluído com sucesso"})
  }
  ```

**Step 9: Update routes in `cmd/api/main.go`**

- Remove: `privateRoutes.PATCH("/curriculum/:id/active", curriculumController.SetActiveCurriculum)`
- Add: `privateRoutes.DELETE("/curriculum/:id", curriculumController.DeleteCurriculum)`

**Step 10: Update NotificationRepository to save/fetch analysis results**

In `repository/notification_repository.go`, add:
```go
func (db *NotificationRepository) InsertNotificationWithAnalysis(jobId int, userId int, curriculumId int, analysisResult []byte) error {
	query := `INSERT INTO job_notifications (user_id, job_id, curriculum_id, analysis_result, status) VALUES ($1, $2, $3, $4, 'SENT') ON CONFLICT (user_id, job_id) DO UPDATE SET analysis_result = $4, curriculum_id = $3, notified_at = NOW()`
	_, err := db.connection.Exec(query, userId, jobId, curriculumId, analysisResult)
	if err != nil {
		return fmt.Errorf("error inserting notification with analysis: %w", err)
	}
	return nil
}

func (db *NotificationRepository) GetAnalysisHistory(userId int, jobId int) ([]byte, *int, error) {
	query := `SELECT analysis_result, curriculum_id FROM job_notifications WHERE user_id = $1 AND job_id = $2 AND analysis_result IS NOT NULL ORDER BY notified_at DESC LIMIT 1`
	var result []byte
	var curriculumID sql.NullInt64
	err := db.connection.QueryRow(query, userId, jobId).Scan(&result, &curriculumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("error fetching analysis history: %w", err)
	}
	var cvID *int
	if curriculumID.Valid {
		id := int(curriculumID.Int64)
		cvID = &id
	}
	return result, cvID, nil
}
```

**Step 11: Update NotificationRepositoryInterface**

In `interfaces/notification_interface.go`, add:
```go
InsertNotificationWithAnalysis(jobId int, userId int, curriculumId int, analysisResult []byte) error
GetAnalysisHistory(userId int, jobId int) ([]byte, *int, error)
```

**Step 12: Build**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...`
Expected: BUILD SUCCESS (fix any compile errors from removed is_active references)

**Step 13: Commit**

```bash
git add migrations/035_* model/curriculum.go model/notification.go repository/curriculum_repository.go repository/notification_repository.go interfaces/curriculum_interface.go interfaces/notification_interface.go usecase/curriculum_usecase.go controller/curriculum_controller.go cmd/api/main.go
git commit -m "feat: remove is_active from curriculum, add delete + analysis history

- Drop is_active column and constraint from curriculum table
- Add DELETE /curriculum/:id with min-1 enforcement
- Add analysis_result JSONB and curriculum_id to job_notifications
- Add analysis history retrieval for 'redo analysis' flow"
```

---

## Task 4: Backend — Subscription Middleware + Webhook Update

**Files:**
- Create: `middleware/subscription.go`
- Modify: `controller/payment_controller.go` (or `usecase/payment_usecase.go`)
- Modify: `cmd/api/main.go`

**Step 1: Create subscription middleware**

Create `middleware/subscription.go`:
```go
package middleware

import (
	"net/http"
	"time"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

func RequireActiveSubscription() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userInterface, exists := ctx.Get("user")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			return
		}
		user, ok := userInterface.(model.User)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
			return
		}
		if user.ExpiresAt == nil || user.ExpiresAt.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "subscription_expired"})
			return
		}
		ctx.Next()
	}
}
```

**Step 2: Update CompleteRegistration to set expires_at**

In `usecase/payment_usecase.go`, in `CompleteRegistration`, before creating user, set `ExpiresAt`:
```go
// After unmarshalling pendingData, before creating user:
expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

userToCreate := model.User{
    Name:      pendingData.Name,
    Email:     pendingData.Email,
    Password:  pendingData.Password,
    Tax:       &pendingData.Tax,
    Cellphone: &pendingData.Cellphone,
    PlanID:    &pendingData.PlanID,
    ExpiresAt: &expiresAt,
}
```

**Step 3: Add renewal endpoint**

In `controller/payment_controller.go`, add method to handle subscription renewal webhook or direct renewal. However, since renewal goes through the same AbacatePay checkout flow, we need to update the webhook handler to update `expires_at` for existing users.

In `usecase/payment_usecase.go`, update `CompleteRegistration`:
- When user already exists (duplicate key), update their `expires_at`:
```go
if strings.Contains(err.Error(), "user already exists") || strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
    // ... existing logic to find user ...
    // Add: update expires_at for renewal
    newExpiry := time.Now().Add(30 * 24 * time.Hour)
    if updateErr := uc.userUsecase.userRepository.UpdateExpiresAt(existingUser.Id, newExpiry); updateErr != nil {
        log.Error().Err(updateErr).Int("user_id", existingUser.Id).Msg("Erro ao renovar assinatura do usuário")
    } else {
        log.Info().Int("user_id", existingUser.Id).Time("expires_at", newExpiry).Msg("Assinatura renovada com sucesso")
    }
    // ... rest of existing logic ...
}
```

Note: This requires `UserUsecase` to expose `UpdateExpiresAt`. Add to `usecase/user_usecase.go`:
```go
func (uuc *UserUsecase) UpdateExpiresAt(userId int, expiresAt time.Time) error {
    return uuc.userRepository.UpdateExpiresAt(userId, expiresAt)
}
```

**Step 4: Wire subscription middleware in `cmd/api/main.go`**

Add `middleware.RequireActiveSubscription()` to private routes that require active subscription. Place it after `RequireAuth` but only on routes that need it (not on `GET /api/me` or logout):

```go
// Routes that DON'T need subscription check (allow expired users to see their status):
privateRoutes.GET("api/me", checkAuthController.CheckAuthUser)
privateRoutes.POST("/api/logout", userController.Logout)

// Routes that DO need subscription check:
subscribedRoutes := server.Group("/")
subscribedRoutes.Use(logging.GinMiddleware())
subscribedRoutes.Use(metrics.GinPrometheus())
subscribedRoutes.Use(csrfMiddleware)
subscribedRoutes.Use(middlewareAuth.RequireAuth)
subscribedRoutes.Use(middleware.RequireActiveSubscription())
subscribedRoutes.Use(privateRateLimiter)
{
    // All existing private routes that need active subscription
    subscribedRoutes.GET("api/dashboard", dashboardController.GetDashboardDataByUserId)
    subscribedRoutes.GET("api/dashboard/jobs", dashboardController.GetLatestJobs)
    subscribedRoutes.GET("api/getSites", siteCareerController.GetAllSites)
    subscribedRoutes.GET("api/notifications", notificationController.GetNotificationsByUser)
    subscribedRoutes.POST("/curriculum", curriculumController.CreateCurriculum)
    // ... all other subscription-gated routes ...
}
```

**Step 5: Update CheckAuthUser to include monitored_sites_count**

In `controller/check_auth_controller.go` (or wherever `GET /api/me` handler is), add `monitored_sites_count` to the response by injecting `UserSiteRepository`:

Actually, looking at the codebase, `CheckAuthController` simply returns the user from context. We need to enrich the response. Modify the handler to also query `GetUserSiteCount`:

```go
// In the CheckAuthUser handler, after getting user from context:
siteCount, _ := ac.userSiteRepo.GetUserSiteCount(user.Id)
ctx.JSON(http.StatusOK, gin.H{
    "id":                    user.Id,
    "user_name":             user.Name,
    "email":                 user.Email,
    "tax":                   user.Tax,
    "cellphone":             user.Cellphone,
    "is_admin":              user.IsAdmin,
    "plan":                  user.Plan,
    "expires_at":            user.ExpiresAt,
    "monitored_sites_count": siteCount,
})
```

Update `NewCheckAuthController` to accept `UserSiteRepositoryInterface` as dependency, and wire it in `main.go`.

**Step 6: Build and verify**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 7: Commit**

```bash
git add middleware/subscription.go usecase/payment_usecase.go usecase/user_usecase.go controller/check_auth_controller.go cmd/api/main.go
git commit -m "feat: add subscription control middleware and expires_at flow

- RequireActiveSubscription middleware returns 403 subscription_expired
- CompleteRegistration sets expires_at on new users
- Renewal updates expires_at for existing users
- GET /api/me returns monitored_sites_count and expires_at"
```

---

## Task 5: Backend — Forgot Password Flow

**Files:**
- Modify: `cmd/api/main.go`
- Create: `controller/password_reset_controller.go`
- Modify: `usecase/emailAdapter.go` (add password reset email template)
- Modify: `interfaces/email_interface.go`

**Step 1: Create password reset controller**

Create `controller/password_reset_controller.go`:
```go
package controller

import (
	"net/http"
	"os"
	"time"
	"web-scrapper/interfaces"
	"web-scrapper/logging"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetController struct {
	resetRepo  interfaces.PasswordResetRepositoryInterface
	userRepo   interfaces.UserRepositoryInterface
	emailSvc   interfaces.EmailService
}

func NewPasswordResetController(
	resetRepo interfaces.PasswordResetRepositoryInterface,
	userRepo interfaces.UserRepositoryInterface,
	emailSvc interfaces.EmailService,
) *PasswordResetController {
	return &PasswordResetController{
		resetRepo: resetRepo,
		userRepo:  userRepo,
		emailSvc:  emailSvc,
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (c *PasswordResetController) ForgotPassword(ctx *gin.Context) {
	var body forgotPasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "E-mail inválido"})
		return
	}

	// Always return success to prevent email enumeration
	user, err := c.userRepo.GetUserByEmail(body.Email)
	if err != nil || user.Id == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instruções de recuperação."})
		return
	}

	token, err := c.resetRepo.CreateToken(user.Id, 1*time.Hour)
	if err != nil {
		logging.Logger.Error().Err(err).Str("email", body.Email).Msg("Erro ao criar token de recuperação")
		ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instruções de recuperação."})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := frontendURL + "/reset-password?token=" + token

	if err := c.emailSvc.SendPasswordResetEmail(ctx.Request.Context(), user.Email, user.Name, resetLink); err != nil {
		logging.Logger.Error().Err(err).Str("email", body.Email).Msg("Erro ao enviar email de recuperação")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Se o e-mail existir, enviaremos instruções de recuperação."})
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (c *PasswordResetController) ResetPassword(ctx *gin.Context) {
	var body resetPasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token e nova senha (mínimo 8 caracteres) são obrigatórios"})
		return
	}

	tokenRecord, err := c.resetRepo.FindValidToken(body.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}
	if tokenRecord == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token inválido ou expirado"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar nova senha"})
		return
	}

	if err := c.userRepo.UpdateUserPassword(tokenRecord.UserID, string(hashed)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha"})
		return
	}

	if err := c.resetRepo.MarkUsed(tokenRecord.ID); err != nil {
		logging.Logger.Error().Err(err).Int("token_id", tokenRecord.ID).Msg("Erro ao marcar token como usado")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Senha atualizada com sucesso"})
}
```

**Step 2: Add SendPasswordResetEmail to EmailService interface and adapter**

In `interfaces/email_interface.go`, add:
```go
SendPasswordResetEmail(ctx context.Context, email, userName, resetLink string) error
```

In `usecase/emailAdapter.go`, add the implementation with HTML template (follow existing pattern of `SendWelcomeEmail`):
```go
func (s *SESSenderAdapter) SendPasswordResetEmail(ctx context.Context, email, userName, resetLink string) error {
	subject := "ScrapJobs — Redefinição de Senha"
	bodyHTML := generatePasswordResetEmailHTML(userName, resetLink)
	bodyText := fmt.Sprintf("Olá %s, clique no link para redefinir sua senha: %s (válido por 1 hora)", userName, resetLink)
	return s.sender.SendEmail(ctx, email, subject, bodyText, bodyHTML)
}
```

Add `generatePasswordResetEmailHTML` function following existing email template style (use `html/template`).

**Step 3: Register routes in `cmd/api/main.go`**

```go
passwordResetRepo := repository.NewPasswordResetRepository(dbConnection)
passwordResetController := controller.NewPasswordResetController(passwordResetRepo, userRepository, emailService)

// Add to publicRoutes (with rate limiter):
forgotPasswordLimiter := rateLimiterFn(3, 60)
publicRoutes.POST("/api/forgot-password", forgotPasswordLimiter, passwordResetController.ForgotPassword)
publicRoutes.POST("/api/reset-password", forgotPasswordLimiter, passwordResetController.ResetPassword)
```

**Step 4: Build**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 5: Commit**

```bash
git add controller/password_reset_controller.go usecase/emailAdapter.go interfaces/email_interface.go repository/password_reset_repository.go cmd/api/main.go
git commit -m "feat: add forgot/reset password flow

- POST /api/forgot-password generates token, sends SES email
- POST /api/reset-password validates token, updates password
- Prevents email enumeration (always returns success)
- 1h token TTL, 3 req/min rate limit"
```

---

## Task 6: Backend — Delete Account + Refactored Analysis Endpoint

**Files:**
- Modify: `controller/analysis_controller.go`
- Create: `controller/account_controller.go`
- Modify: `cmd/api/main.go`
- Modify: `middleware/requiredAuth.go`

**Step 1: Create account deletion controller**

Create `controller/account_controller.go`:
```go
package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AccountController struct {
	userRepo interfaces.UserRepositoryInterface
}

func NewAccountController(userRepo interfaces.UserRepositoryInterface) *AccountController {
	return &AccountController{userRepo: userRepo}
}

type deleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

func (ac *AccountController) DeleteAccount(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
		return
	}

	var body deleteAccountRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Senha é obrigatória para excluir conta"})
		return
	}

	// Fetch full user with password hash
	fullUser, err := ac.userRepo.GetUserByEmail(user.Email)
	if err != nil || fullUser.Id == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar usuário"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fullUser.Password), []byte(body.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	if err := ac.userRepo.SoftDeleteUser(user.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir conta"})
		return
	}

	// Clear auth cookie
	ctx.SetCookie("Authorization", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "Conta excluída com sucesso"})
}
```

**Step 2: Update auth middleware to reject soft-deleted users**

In `middleware/requiredAuth.go`, after fetching user from DB:
```go
if user.DeletedAt != nil {
    ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Conta desativada"})
    return
}
```

**Step 3: Refactor AnalysisController.AnalyzeJob to accept curriculum_id**

In `controller/analysis_controller.go`, update the request struct and handler:

```go
type analyzeJobRequest struct {
	JobID        int `json:"job_id" binding:"required"`
	CurriculumID int `json:"curriculum_id" binding:"required"`
}
```

Replace the active curriculum lookup (lines 77-94) with direct curriculum fetch:
```go
// Fetch specified curriculum (validate ownership)
curricula, err := ac.curriculumRepo.FindCurriculumByUserID(user.Id)
if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar currículo"})
    return
}
var selectedCurriculum *model.Curriculum
for i := range curricula {
    if curricula[i].Id == body.CurriculumID {
        selectedCurriculum = &curricula[i]
        break
    }
}
if selectedCurriculum == nil {
    ctx.JSON(http.StatusBadRequest, gin.H{"error": "Currículo não encontrado ou não pertence ao usuário"})
    return
}
```

Then use `selectedCurriculum` instead of `activeCurriculum`:
```go
analysis, err := ac.analysisService.Analyze(ctx.Request.Context(), *selectedCurriculum, *job)
```

After analysis, save with curriculum_id:
```go
analysisJSON, _ := json.Marshal(analysis)
if err := ac.notificationRepo.InsertNotificationWithAnalysis(job.ID, user.Id, body.CurriculumID, analysisJSON); err != nil {
    logging.Logger.Error().Err(err).Msg("Erro ao registrar análise")
}
```

Add `"encoding/json"` to imports.

**Step 4: Add GetAnalysisHistory endpoint**

In `controller/analysis_controller.go`, add:
```go
func (ac *AnalysisController) GetAnalysisHistory(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
		return
	}

	jobIDStr := ctx.Query("job_id")
	if jobIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id é obrigatório"})
		return
	}
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id inválido"})
		return
	}

	result, cvID, err := ac.notificationRepo.GetAnalysisHistory(user.Id, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar histórico"})
		return
	}
	if result == nil {
		ctx.JSON(http.StatusOK, gin.H{"has_analysis": false})
		return
	}

	var analysis model.ResumeAnalysis
	if err := json.Unmarshal(result, &analysis); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar análise"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"has_analysis":  true,
		"analysis":      analysis,
		"curriculum_id": cvID,
	})
}
```

Add `"strconv"` to imports.

**Step 5: Register new routes in `cmd/api/main.go`**

```go
accountController := controller.NewAccountController(userRepository)

// In private routes (no subscription check needed):
privateRoutes.DELETE("/api/user/account", accountController.DeleteAccount)

// In subscribed routes (analysis):
if analysisController != nil {
    analyzeRateLimiter := rateLimiterFn(3, 60)
    subscribedRoutes.POST("/api/analyze-job", analyzeRateLimiter, analysisController.AnalyzeJob)
    subscribedRoutes.POST("/api/analyze-job/send-email", analyzeRateLimiter, analysisController.SendAnalysisEmail)
    subscribedRoutes.GET("/api/analyze-job/history", analysisController.GetAnalysisHistory)
}
```

**Step 6: Build**

Run: `go build ./...`
Expected: BUILD SUCCESS

**Step 7: Commit**

```bash
git add controller/account_controller.go controller/analysis_controller.go middleware/requiredAuth.go cmd/api/main.go
git commit -m "feat: add account deletion and refactored analysis endpoint

- DELETE /api/user/account with password confirmation (soft delete)
- POST /api/analyze-job now requires curriculum_id
- GET /api/analyze-job/history returns previous analysis
- Auth middleware rejects soft-deleted users"
```

---

## Task 7: Backend — Clean Up Gemini References

**Files:**
- Check: `infra/gemini/` directory

**Step 1: Verify and clean**

Search for any remaining Gemini references in the codebase. The AI integration is already using OpenAI (`infra/openai/`). If `infra/gemini/` directory exists and is unused, delete it. Remove any unused imports.

Run: `grep -r "gemini" --include="*.go" .` in the ScrapJobs directory.

**Step 2: Commit if changes made**

```bash
git add -A
git commit -m "chore: remove unused Gemini references"
```

---

## Task 8: Frontend — Update Models and Services

**Files:**
- Modify: `src/models/user.ts`
- Modify: `src/models/curriculum.ts`
- Modify: `src/services/curriculumService.ts`
- Modify: `src/services/analysisService.ts`
- Modify: `src/services/authService.ts`
- Modify: `src/hooks/useCurriculum.ts`
- Modify: `src/hooks/useAnalysis.ts`

**Step 1: Update User model**

In `src/models/user.ts`:
```typescript
import type { Plan } from './plan'

export interface User {
  id: string
  user_name: string
  email: string
  cellphone?: string
  tax?: string
  is_admin?: boolean
  plan: Plan | undefined
  expires_at?: string
  monitored_sites_count?: number
}
```

**Step 2: Update Curriculum model (remove is_active)**

In `src/models/curriculum.ts`:
```typescript
export interface Experience {
  id: string
  company: string
  title: string
  description: string
}

export interface Education {
  id: string
  institution: string
  degree: string
  year: string
}

export interface Curriculum {
  id: number
  title: string
  summary: string
  skills: string
  languages: string
  experiences: Experience[]
  educations: Education[]
}
```

**Step 3: Update curriculumService — remove setActive, add delete**

In `src/services/curriculumService.ts`:
- Remove `setActiveCurriculum` method
- Add:
```typescript
deleteCurriculum: async (curriculumId: number): Promise<void> => {
  await api.delete(`/curriculum/${curriculumId}`)
}
```

**Step 4: Update analysisService — add curriculum_id and history**

In `src/services/analysisService.ts`:
```typescript
export const analysisService = {
  analyzeJob: async (jobId: number, curriculumId: number): Promise<ResumeAnalysis> => {
    const { data } = await api.post('/api/analyze-job', { job_id: jobId, curriculum_id: curriculumId })
    return data
  },

  getAnalysisHistory: async (jobId: number): Promise<{
    has_analysis: boolean
    analysis?: ResumeAnalysis
    curriculum_id?: number
  }> => {
    const { data } = await api.get('/api/analyze-job/history', { params: { job_id: jobId } })
    return data
  },

  sendAnalysisEmail: async (jobId: number, analysis: ResumeAnalysis): Promise<void> => {
    await api.post('/api/analyze-job/send-email', { job_id: jobId, analysis })
  }
}
```

**Step 5: Update authService — add forgotPassword, resetPassword, deleteAccount**

In `src/services/authService.ts`, add:
```typescript
forgotPassword: async (email: string) => {
  const { data } = await api.post('/api/forgot-password', { email })
  return data
},

resetPassword: async (token: string, newPassword: string) => {
  const { data } = await api.post('/api/reset-password', { token, new_password: newPassword })
  return data
},

deleteAccount: async (password: string) => {
  const { data } = await api.delete('/api/user/account', { data: { password } })
  return data
}
```

**Step 6: Update hooks**

In `src/hooks/useCurriculum.ts`:
- Remove `useSetActiveCurriculum`
- Add:
```typescript
export function useDeleteCurriculum() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (curriculumId: number) => curriculumService.deleteCurriculum(curriculumId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['curriculumList'] })
    }
  })
}
```

In `src/hooks/useAnalysis.ts`:
```typescript
export function useAnalyzeJob() {
  return useMutation({
    mutationFn: ({ jobId, curriculumId }: { jobId: number; curriculumId: number }) =>
      analysisService.analyzeJob(jobId, curriculumId)
  })
}

export function useAnalysisHistory(jobId: number | null) {
  return useQuery({
    queryKey: ['analysisHistory', jobId],
    queryFn: () => analysisService.getAnalysisHistory(jobId!),
    enabled: jobId !== null
  })
}
```

Add `import { useQuery } from '@tanstack/react-query'` to the import.

**Step 7: Update API interceptor for subscription_expired**

In `src/services/api.ts`, add handling for 403 subscription_expired:
```typescript
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const isLoginPage = window.location.pathname === '/login'
      const isPublicPage = window.location.pathname === '/'
      if (!isLoginPage && !isPublicPage) {
        window.location.href = `/login?from=${encodeURIComponent(window.location.pathname)}`
      }
    }
    if (error.response?.status === 403 && error.response?.data?.error === 'subscription_expired') {
      window.location.href = '/app/renew'
    }
    return Promise.reject(error)
  }
)
```

**Step 8: Commit**

```bash
cd "/Users/erickschaedler/Documents/Scrap Jobs/FrontScrapJobs"
git add src/models/ src/services/ src/hooks/
git commit -m "feat: update models and services for MVP polish

- User model: add expires_at, monitored_sites_count
- Curriculum: remove is_active, add delete
- Analysis: add curriculum_id selection, history endpoint
- Auth: add forgot/reset password, delete account
- API interceptor handles subscription_expired"
```

---

## Task 9: Frontend — Checkout Fixes + Back Button

**Files:**
- Modify: `src/pages/checkout.tsx`

**Step 1: Disable credit card option**

Find the credit card `<div>` in checkout.tsx and add disabled styling:
```tsx
{/* Credit Card Option — Disabled (Em breve) */}
<div className="group relative flex items-center gap-4 rounded-lg border-2 p-4 opacity-50 pointer-events-none border-border/50">
  <input type="radio" name="paymentMethod" value="card" disabled className="sr-only" />
  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
    <CreditCardIcon className="h-5 w-5" />
  </div>
  <div>
    <p className="font-semibold text-foreground">{t('paymentForm.creditCard')}</p>
    <p className="text-xs text-muted-foreground">{t('paymentForm.creditCardDescription')}</p>
  </div>
  <span className="absolute right-3 top-3 rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
    Em breve
  </span>
</div>
```

Remove the `onClick` handler from the credit card div.

**Step 2: Default payment method to PIX**

Ensure `formData` initial state has `paymentMethod: 'pix'`.

**Step 3: Add back button**

At the top of the checkout page, before the form:
```tsx
<Link
  to={PATHS.landing}
  className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors mb-6"
>
  <ArrowLeft className="h-4 w-4" />
  {t('paymentForm.backToHome')}
</Link>
```

Import `ArrowLeft` from lucide-react and `Link` from react-router.

Add i18n key `paymentForm.backToHome`: "Voltar para a Página Inicial" in `pt-BR/plans.json`.

**Step 4: Commit**

```bash
git add src/pages/checkout.tsx src/i18n/locales/
git commit -m "fix: disable credit card option, add back button on checkout

Credit card visually disabled with 'Em breve' badge.
PIX is default. Back to home link added."
```

---

## Task 10: Frontend — Dashboard Table Truncation

**Files:**
- Modify: `src/pages/Home.tsx`

**Step 1: Fix table column widths and truncation**

Replace the table header and cells with fixed widths:
```tsx
<Table className="text-sm">
  <TableHeader className="bg-muted/40">
    <TableRow>
      <TableHead className="w-[40%] font-medium">{t('latestJobs.jobTitle')}</TableHead>
      <TableHead className="w-[20%] font-medium">{t('latestJobs.company')}</TableHead>
      <TableHead className="w-[15%] font-medium">{t('latestJobs.location')}</TableHead>
      <TableHead className="w-[10%] font-medium">{t('latestJobs.link')}</TableHead>
      <TableHead className="w-[15%] font-medium text-right">{t('latestJobs.action')}</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    {jobs.map((job) => (
      <TableRow key={job.id} className="group/row hover:bg-muted/30">
        <TableCell className="max-w-0 font-medium text-foreground">
          <span className="block truncate" title={job.title}>{job.title}</span>
        </TableCell>
        <TableCell className="max-w-0 text-muted-foreground">
          <span className="block truncate" title={job.company}>{job.company}</span>
        </TableCell>
        <TableCell className="max-w-0 text-muted-foreground">
          <span className="block truncate" title={job.location}>{job.location}</span>
        </TableCell>
        {/* Link and Action cells remain unchanged */}
      </TableRow>
    ))}
  </TableBody>
</Table>
```

Remove `min-w-[600px]` from the `<Table>` class.

**Step 2: Commit**

```bash
git add src/pages/Home.tsx
git commit -m "fix: dashboard table column widths with text truncation

Fixed percentages prevent horizontal scroll. Truncated text
shows full content via native title tooltip."
```

---

## Task 11: Frontend — Login Page Redesign

**Files:**
- Modify: `src/pages/Login.tsx`
- Modify: `src/components/forms/Auth.tsx`
- Modify: `src/i18n/locales/pt-BR/auth.json`

**Step 1: Redesign Login page with split layout**

Replace `Login.tsx` with split layout (hero left + form right):
```tsx
export default function Login() {
  const { data: user } = useUser()
  const navigate = useNavigate()
  const { t } = useTranslation('auth')

  useEffect(() => {
    if (user) navigate(PATHS.app.home)
  }, [user, navigate])

  return (
    <div className="flex min-h-screen">
      {/* Left Panel — Brand Hero (hidden on mobile) */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-center px-12 xl:px-20 bg-card border-r border-border/50 relative overflow-hidden">
        <div className="pointer-events-none absolute -left-24 -top-24 h-[400px] w-[400px] rounded-full bg-primary/5 blur-[100px]" />
        <div className="relative z-10">
          <img src={Logo} alt="ScrapJobs" className="h-20 w-20 mb-8 select-none" draggable={false} />
          <h1 className="font-display text-3xl font-bold tracking-tight text-foreground mb-4">
            {t('hero.title')}
          </h1>
          <p className="text-lg text-muted-foreground mb-8 max-w-md">
            {t('hero.subtitle')}
          </p>
          <div className="flex gap-8">
            <div>
              <p className="font-display text-2xl font-bold text-primary">500+</p>
              <p className="text-sm text-muted-foreground">{t('hero.jobsMonitored')}</p>
            </div>
            <div>
              <p className="font-display text-2xl font-bold text-primary">50+</p>
              <p className="text-sm text-muted-foreground">{t('hero.companiesTracked')}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Right Panel — Form */}
      <div className="flex w-full lg:w-1/2 items-center justify-center px-4 bg-background">
        <div className="w-full max-w-sm animate-fade-in-up">
          <div className="mb-8 flex items-center justify-center lg:hidden">
            <img src={Logo} alt="ScrapJobs" className="h-32 w-32 select-none" draggable={false} />
          </div>
          <h2 className="text-xl font-semibold text-foreground mb-1 lg:mb-2">
            {t('login.welcome')}
          </h2>
          <p className="text-sm text-muted-foreground mb-8">
            {t('login.subtitle')}
          </p>
          <AuthForm />
          <p className="mt-6 text-center text-sm text-muted-foreground">
            {t('login.noAccount')}{' '}
            <Link to="/#pricing" className="font-medium text-primary hover:underline">
              {t('login.choosePlan')}
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
```

**Step 2: Add "Esqueci minha senha" link in AuthForm**

In `src/components/forms/Auth.tsx`, after the password field:
```tsx
<div className="flex items-center justify-between">
  <Label htmlFor="password" className="text-muted-foreground">
    {t('password')}
  </Label>
  <Link to="/forgot-password" className="text-xs text-primary hover:underline">
    {t('forgotPassword')}
  </Link>
</div>
```

Import `Link` from react-router.

**Step 3: Add i18n keys**

In `src/i18n/locales/pt-BR/auth.json`, add:
```json
"hero": {
  "title": "Encontre as vagas certas, automaticamente.",
  "subtitle": "O ScrapJobs monitora páginas de carreiras e analisa compatibilidade com seu currículo usando IA.",
  "jobsMonitored": "vagas monitoradas",
  "companiesTracked": "empresas rastreadas"
},
"login": {
  "welcome": "Bem-vindo de volta",
  "subtitle": "Entre na sua conta para continuar",
  "noAccount": "Primeira missão por aqui?",
  "choosePlan": "Escolha um plano"
},
"forgotPassword": "Esqueci minha senha"
```

Add equivalent keys in `en-US/auth.json`.

**Step 4: Commit**

```bash
git add src/pages/Login.tsx src/components/forms/Auth.tsx src/i18n/locales/
git commit -m "feat: redesign login page with split layout

Desktop: brand hero left + form right.
Mobile: centered form.
Added 'Forgot password' link."
```

---

## Task 12: Frontend — Forgot/Reset Password Pages

**Files:**
- Create: `src/pages/ForgotPassword.tsx`
- Create: `src/pages/ResetPassword.tsx`
- Modify: `src/router/routes.tsx`
- Modify: `src/router/paths.ts`

**Step 1: Create ForgotPassword page**

Simple form with email input, calls `authService.forgotPassword`. Shows success message regardless.

**Step 2: Create ResetPassword page**

Form with new password + confirm password. Reads `token` from URL query params. Calls `authService.resetPassword`. On success, redirects to login.

**Step 3: Add routes**

In `src/router/paths.ts`:
```typescript
forgotPassword: '/forgot-password',
resetPassword: '/reset-password',
```

In `src/router/routes.tsx`, add to public routes:
```typescript
{ path: 'forgot-password', element: <ForgotPassword /> },
{ path: 'reset-password', element: <ResetPassword /> },
```

**Step 4: Add i18n keys**

Add PT-BR translations for both pages.

**Step 5: Commit**

```bash
git add src/pages/ForgotPassword.tsx src/pages/ResetPassword.tsx src/router/ src/i18n/locales/
git commit -m "feat: add forgot/reset password pages

- /forgot-password: email input, always shows success (no enumeration)
- /reset-password?token=xxx: new password + confirmation form"
```

---

## Task 13: Frontend — Account Page Fixes (Site Count + Delete Account + Manage Plan)

**Files:**
- Modify: `src/components/accountPage/plan-section.tsx`
- Modify: `src/components/accountPage/security-section.tsx`

**Step 1: Fix site counter in PlanSection**

Use `user.monitored_sites_count` from the API response instead of hardcoded 0:
```tsx
const currentUsage = user?.monitored_sites_count ?? 0
const maxUsage = user?.plan?.max_sites ?? 0
```

Also display `expires_at`:
```tsx
{user?.expires_at && (
  <p className="text-xs text-muted-foreground mt-2">
    {t('plan.expiresAt')}: {new Date(user.expires_at).toLocaleDateString('pt-BR')}
  </p>
)}
```

**Step 2: Add "Gerenciar Plano" dialog**

In PlanSection, add a dialog that lists plans (use `usePlans` hook) with "Renovar" buttons that link to `/checkout/:planId`.

**Step 3: Add "Excluir minha conta" to SecuritySection**

Add a destructive button at the bottom of the security tab:
```tsx
<div className="border-t border-destructive/20 pt-6 mt-8">
  <h3 className="text-sm font-semibold text-destructive mb-2">{t('security.dangerZone')}</h3>
  <p className="text-xs text-muted-foreground mb-4">{t('security.deleteWarning')}</p>
  <Button variant="destructive" size="sm" onClick={() => setShowDeleteDialog(true)}>
    {t('security.deleteAccount')}
  </Button>
</div>
```

Dialog asks for password confirmation, calls `authService.deleteAccount`, then logs out + redirects to landing.

**Step 4: Add i18n keys**

**Step 5: Commit**

```bash
git add src/components/accountPage/
git commit -m "fix: account page — real site count, manage plan, delete account

- PlanSection uses monitored_sites_count from API
- Manage Plan dialog with available plans
- Delete Account with password confirmation in Security tab"
```

---

## Task 14: Frontend — Curriculum Page Updates (Delete + Remove is_active)

**Files:**
- Modify: `src/components/curriculum/curriculum-list.tsx`
- Modify: `src/components/curriculum/curriculum-form.tsx`
- Modify: `src/pages/Curriculum.tsx`

**Step 1: Remove is_active references**

Remove any `is_active` badge/indicator from `curriculum-list.tsx`. Remove `useSetActiveCurriculum` hook usage.

**Step 2: Add delete button to curriculum list**

In `curriculum-list.tsx`, add a trash icon button per curriculum card:
```tsx
<Button
  variant="ghost"
  size="icon"
  className="h-7 w-7 text-muted-foreground hover:text-destructive"
  onClick={(e) => {
    e.stopPropagation()
    setDeleteId(cv.id)
  }}
>
  <Trash2 className="h-3.5 w-3.5" />
</Button>
```

Add confirmation dialog using shadcn `AlertDialog`. On confirm, call `useDeleteCurriculum` mutation.

**Step 3: Commit**

```bash
git add src/components/curriculum/ src/pages/Curriculum.tsx
git commit -m "feat: curriculum delete + remove is_active

- Delete button with confirmation (min 1 enforced by backend)
- Removed all is_active UI references"
```

---

## Task 15: Frontend — Refactored Analysis Dialog (Curriculum Selection + History)

**Files:**
- Modify: `src/components/analysis/analysis-dialog.tsx`

**Step 1: Refactor to multi-step flow**

Replace auto-analyze behavior with 2-step flow:

1. **Step 1 (Curriculum Selection):** On dialog open, check analysis history. If exists, show previous result + "Refazer" button. If not, show curriculum list with "Gerar Análise" button.

2. **Step 2 (Result):** Loading state → Result display (existing AnalysisResult component).

Key changes:
- Use `useCurriculum()` to fetch curriculum list
- Use `useAnalysisHistory(jobId)` to check for existing analysis
- `useAnalyzeJob` mutation now sends `{ jobId, curriculumId }`
- State machine: `'select' | 'loading' | 'result' | 'history'`

**Step 2: Commit**

```bash
git add src/components/analysis/
git commit -m "feat: refactored analysis dialog with curriculum selection

- 2-step flow: select curriculum → generate analysis
- Shows previous analysis if exists + 'Redo' button
- Sends curriculum_id with analysis request"
```

---

## Task 16: Frontend — Legal Pages (Terms + Privacy)

**Files:**
- Create: `src/pages/TermsOfService.tsx`
- Create: `src/pages/PrivacyPolicy.tsx`
- Modify: `src/router/routes.tsx`
- Modify: `src/router/paths.ts`
- Modify: `src/components/landingPage/footer.tsx`

**Step 1: Create static pages**

Both pages follow a simple layout: centered container with heading + structured sections using lorem ipsum.

**Step 2: Add routes**

In `paths.ts`: `terms: '/terms'`, `privacy: '/privacy'`

In `routes.tsx`, add to public routes:
```typescript
{ path: 'terms', element: <TermsOfService /> },
{ path: 'privacy', element: <PrivacyPolicy /> },
```

**Step 3: Update footer links**

In `src/components/landingPage/footer.tsx`, replace `href="#"` with:
```tsx
<Link to={PATHS.terms} className="hover:text-foreground transition-colors duration-150">
  {t('footer.terms')}
</Link>
<Link to={PATHS.privacy} className="hover:text-foreground transition-colors duration-150">
  {t('footer.privacy')}
</Link>
```

Import `Link` from react-router and `PATHS` from router.

**Step 4: Commit**

```bash
git add src/pages/TermsOfService.tsx src/pages/PrivacyPolicy.tsx src/router/ src/components/landingPage/footer.tsx
git commit -m "feat: add Terms of Service and Privacy Policy pages

Static pages with lorem ipsum placeholder content.
Footer links updated from # to actual routes."
```

---

## Task 17: Frontend — Subscription Renewal Page

**Files:**
- Create: `src/pages/RenewSubscription.tsx`
- Modify: `src/router/routes.tsx`

**Step 1: Create renewal page**

Page shown when subscription expires. Displays plan cards (reuses `usePlans` hook) with CTA to checkout. Friendly message explaining subscription expired.

**Step 2: Add route**

In protected routes:
```typescript
{ path: 'renew', element: <RenewSubscription /> }
```

Note: This route should NOT have `RequireActiveSubscription` middleware (it's the page expired users see).

**Step 3: Commit**

```bash
git add src/pages/RenewSubscription.tsx src/router/
git commit -m "feat: add subscription renewal page

Shown when subscription expires (403 subscription_expired).
Lists available plans with checkout links."
```

---

## Task 18: Final — Build Verification + i18n Check

**Step 1: Build backend**

```bash
cd "/Users/erickschaedler/Documents/Scrap Jobs/ScrapJobs"
go build ./...
go vet ./...
```

**Step 2: Build frontend**

```bash
cd "/Users/erickschaedler/Documents/Scrap Jobs/FrontScrapJobs"
npm run build
npm run lint
```

**Step 3: Fix any issues found**

**Step 4: Final commit if needed**

---

## Execution Order Summary

| # | Task | Repo | Type |
|---|------|------|------|
| 1 | Migration: expires_at + deleted_at | Backend | DB + Model |
| 2 | Migration: password_reset_tokens | Backend | DB + Model |
| 3 | Migration: drop is_active + analysis fields | Backend | DB + Model |
| 4 | Subscription middleware + webhook | Backend | Middleware |
| 5 | Forgot/reset password | Backend | Controller |
| 6 | Delete account + analysis refactor | Backend | Controller |
| 7 | Clean up Gemini refs | Backend | Cleanup |
| 8 | Update models + services | Frontend | Services |
| 9 | Checkout fixes | Frontend | UI |
| 10 | Dashboard table | Frontend | UI |
| 11 | Login redesign | Frontend | UI |
| 12 | Forgot/reset password pages | Frontend | Pages |
| 13 | Account page fixes | Frontend | UI |
| 14 | Curriculum updates | Frontend | UI |
| 15 | Analysis dialog refactor | Frontend | UI |
| 16 | Legal pages | Frontend | Pages |
| 17 | Renewal page | Frontend | Pages |
| 18 | Build verification | Both | QA |

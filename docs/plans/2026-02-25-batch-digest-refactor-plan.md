# Batch Digest Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Decouple the scraping/match/notification pipeline into 3 independent scheduler-driven processes with ID-only Asynq payloads and a digest email pattern.

**Architecture:** Three independent cronjobs in the Scheduler (2h scraping, 4h match, 8h digest). Scraping only persists jobs. Match creates PENDING notification records. Digest sends consolidated emails and marks SENT. All Asynq payloads carry only user_id (int).

**Tech Stack:** Go 1.24, Asynq, PostgreSQL, AWS SES, testify/mock

**Design doc:** `docs/plans/2026-02-25-batch-digest-refactor-design.md`

---

### Task 1: Database Migration — Add `status` Column to `job_notifications`

**Files:**
- Create: `migrations/031_add_status_to_job_notifications.up.sql`
- Create: `migrations/031_add_status_to_job_notifications.down.sql`

**Step 1: Create the UP migration**

```sql
ALTER TABLE job_notifications
ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'SENT';

CREATE INDEX idx_job_notifications_status
ON job_notifications(user_id, status)
WHERE status = 'PENDING';
```

**Step 2: Create the DOWN migration**

```sql
DROP INDEX IF EXISTS idx_job_notifications_status;
ALTER TABLE job_notifications DROP COLUMN IF EXISTS status;
```

**Step 3: Run migration**

Run: `cd ScrapJobs && migrate -path migrations/ -database "postgres://${USER_DB}:${PASSWORD_DB}@${HOST_DB}:${PORT_DB}/${DBNAME}?sslmode=${DB_SSLMODE}" up`
Expected: migration 031 applied successfully

**Step 4: Commit**

```bash
git add migrations/031_add_status_to_job_notifications.up.sql migrations/031_add_status_to_job_notifications.down.sql
git commit -m "feat: add status column to job_notifications for digest pattern"
```

---

### Task 2: Add New Task Types and Model — Payloads for Match and Digest

**Files:**
- Modify: `tasks/payloads.go`
- Modify: `model/notification.go`

**Step 1: Add new constants and payload structs to `tasks/payloads.go`**

Add these new constants alongside existing ones:

```go
const TypeMatchUser  = "match:user"
const TypeSendDigest = "digest:send"
```

Add these new payload structs (ID-only, no domain structs):

```go
type MatchUserPayload struct {
	UserID int `json:"user_id"`
}

type SendDigestPayload struct {
	UserID int `json:"user_id"`
}
```

**Step 2: Add `JobWithFilters` struct to `model/notification.go`**

This is used by the match query that returns jobs with their per-site filters:

```go
type JobWithFilters struct {
	JobID    int      `json:"job_id"`
	Title    string   `json:"title"`
	Location string   `json:"location"`
	Company  string   `json:"company"`
	JobLink  string   `json:"job_link"`
	Filters  []string `json:"-"`
}
```

**Step 3: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles without errors

**Step 4: Commit**

```bash
git add tasks/payloads.go model/notification.go
git commit -m "feat: add MatchUser and SendDigest task types and JobWithFilters model"
```

---

### Task 3: New Repository Methods — Interfaces and Mocks

**Files:**
- Modify: `interfaces/notification_interface.go`
- Modify: `interfaces/user_site_interface.go`
- Modify: `interfaces/user_interface.go`
- Modify: `repository/mocks/notification_repository.go`
- Modify: `repository/mocks/user_site_repository.go`
- Modify: `repository/mocks/user_repository.go`

**Step 1: Extend `NotificationRepositoryInterface` in `interfaces/notification_interface.go`**

Add these 4 new methods to the interface:

```go
BulkInsertPendingNotifications(userID int, jobIDs []int) error
GetUserIDsWithPendingNotifications() ([]int, error)
GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error)
BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error
```

**Step 2: Extend `UserSiteRepositoryInterface` in `interfaces/user_site_interface.go`**

Add this method:

```go
GetActiveUserIDs() ([]int, error)
```

**Step 3: Extend `UserRepositoryInterface` in `interfaces/user_interface.go`**

Add this method (needed by digest to fetch user name/email):

```go
GetUserBasicInfo(userID int) (string, string, error)
```

**Step 4: Add mock methods to `repository/mocks/notification_repository.go`**

```go
func (m *MockNotificationRepository) BulkInsertPendingNotifications(userID int, jobIDs []int) error {
	args := m.Called(userID, jobIDs)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUserIDsWithPendingNotifications() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockNotificationRepository) GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.NotificationWithJob), args.Error(1)
}

func (m *MockNotificationRepository) BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error {
	args := m.Called(userID, jobIDs, status)
	return args.Error(0)
}
```

**Step 5: Add mock method to `repository/mocks/user_site_repository.go`**

```go
func (m *MockUserSiteRepository) GetActiveUserIDs() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}
```

**Step 6: Add mock method to `repository/mocks/user_repository.go`**

Read the file first to understand the existing mock structure, then add:

```go
func (m *MockUserRepository) GetUserBasicInfo(userID int) (string, string, error) {
	args := m.Called(userID)
	return args.String(0), args.String(1), args.Error(2)
}
```

**Step 7: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles (implementations will be added next task)

**Step 8: Commit**

```bash
git add interfaces/ repository/mocks/
git commit -m "feat: add interfaces and mocks for match/digest repository methods"
```

---

### Task 4: Repository Implementations — New Query Methods

**Files:**
- Modify: `repository/notification_repository.go`
- Modify: `repository/user_site_repository.go`
- Modify: `repository/user_repository.go`

**Step 1: Add `GetUnnotifiedJobsForUser` to `repository/notification_repository.go`**

This is the optimized query for the match engine. Note: this method is on NotificationRepository because it joins job_notifications:

```go
func (db *NotificationRepository) GetUnnotifiedJobsForUser(userID int) ([]model.JobWithFilters, error) {
	query := `
		SELECT j.id, j.title, j.location, j.company, j.job_link, us.filters
		FROM jobs j
		INNER JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1
		WHERE j.last_seen_at >= NOW() - INTERVAL '24 hours'
		  AND NOT EXISTS (
			  SELECT 1 FROM job_notifications jn
			  WHERE jn.user_id = $1 AND jn.job_id = j.id
		  )`

	rows, err := db.connection.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching unnotified jobs for user %d: %w", userID, err)
	}
	defer rows.Close()

	var jobs []model.JobWithFilters
	for rows.Next() {
		var j model.JobWithFilters
		var filtersJSON sql.NullString
		if err := rows.Scan(&j.JobID, &j.Title, &j.Location, &j.Company, &j.JobLink, &filtersJSON); err != nil {
			return nil, fmt.Errorf("error scanning job with filters: %w", err)
		}
		if filtersJSON.Valid {
			if err := json.Unmarshal([]byte(filtersJSON.String), &j.Filters); err != nil {
				return nil, fmt.Errorf("error unmarshalling filters: %w", err)
			}
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
```

Add import `"encoding/json"` to the file if not already present.

Also add the `GetUnnotifiedJobsForUser` method to `NotificationRepositoryInterface` in `interfaces/notification_interface.go`:

```go
GetUnnotifiedJobsForUser(userID int) ([]model.JobWithFilters, error)
```

And add the mock in `repository/mocks/notification_repository.go`:

```go
func (m *MockNotificationRepository) GetUnnotifiedJobsForUser(userID int) ([]model.JobWithFilters, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.JobWithFilters), args.Error(1)
}
```

**Step 2: Add `BulkInsertPendingNotifications` to `repository/notification_repository.go`**

```go
func (db *NotificationRepository) BulkInsertPendingNotifications(userID int, jobIDs []int) error {
	if len(jobIDs) == 0 {
		return nil
	}

	query := `INSERT INTO job_notifications (user_id, job_id, status) SELECT $1, unnest($2::int[]), 'PENDING' ON CONFLICT (user_id, job_id) DO NOTHING`

	_, err := db.connection.Exec(query, userID, pq.Array(jobIDs))
	if err != nil {
		return fmt.Errorf("error bulk inserting pending notifications for user %d: %w", userID, err)
	}
	return nil
}
```

**Step 3: Add `GetUserIDsWithPendingNotifications` to `repository/notification_repository.go`**

```go
func (db *NotificationRepository) GetUserIDsWithPendingNotifications() ([]int, error) {
	query := `SELECT DISTINCT user_id FROM job_notifications WHERE status = 'PENDING'`

	rows, err := db.connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching users with pending notifications: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}
```

**Step 4: Add `GetPendingJobsForUser` to `repository/notification_repository.go`**

```go
func (db *NotificationRepository) GetPendingJobsForUser(userID int) ([]model.NotificationWithJob, error) {
	query := `
		SELECT jn.id, jn.job_id, jn.user_id, jn.notified_at,
			   j.title, j.company, j.location, j.job_link
		FROM job_notifications jn
		INNER JOIN jobs j ON jn.job_id = j.id
		WHERE jn.user_id = $1 AND jn.status = 'PENDING'
		ORDER BY j.company, j.title`

	rows, err := db.connection.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching pending jobs for user %d: %w", userID, err)
	}
	defer rows.Close()

	var notifications []model.NotificationWithJob
	for rows.Next() {
		var n model.NotificationWithJob
		if err := rows.Scan(&n.ID, &n.JobID, &n.UserID, &n.NotifiedAt,
			&n.JobTitle, &n.JobCompany, &n.JobLocation, &n.JobLink); err != nil {
			return nil, fmt.Errorf("error scanning pending notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}
```

**Step 5: Add `BulkUpdateNotificationStatus` to `repository/notification_repository.go`**

```go
func (db *NotificationRepository) BulkUpdateNotificationStatus(userID int, jobIDs []int, status string) error {
	if len(jobIDs) == 0 {
		return nil
	}

	query := `UPDATE job_notifications SET status = $1, notified_at = NOW() WHERE user_id = $2 AND job_id = ANY($3) AND status = 'PENDING'`

	_, err := db.connection.Exec(query, status, userID, pq.Array(jobIDs))
	if err != nil {
		return fmt.Errorf("error bulk updating notification status for user %d: %w", userID, err)
	}
	return nil
}
```

**Step 6: Add `GetActiveUserIDs` to `repository/user_site_repository.go`**

```go
func (dep *UserSiteRepository) GetActiveUserIDs() ([]int, error) {
	query := `SELECT DISTINCT user_id FROM user_sites`

	rows, err := dep.connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching active user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}
```

**Step 7: Add `GetUserBasicInfo` to `repository/user_repository.go`**

```go
func (usr *UserRepository) GetUserBasicInfo(userID int) (string, string, error) {
	query := `SELECT user_name, email FROM users WHERE id = $1`

	var name, email string
	err := usr.connection.QueryRow(query, userID).Scan(&name, &email)
	if err != nil {
		return "", "", fmt.Errorf("error fetching basic info for user %d: %w", userID, err)
	}
	return name, email, nil
}
```

**Step 8: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles without errors

**Step 9: Commit**

```bash
git add repository/ interfaces/
git commit -m "feat: add repository implementations for match and digest queries"
```

---

### Task 5: Match Use Case — `MatchJobsForUser`

**Files:**
- Modify: `usecase/notifications_usecase.go`
- Test: `usecase/notification_usecase_test.go`

**Step 1: Write the failing tests**

Add these test cases to `usecase/notification_usecase_test.go`:

```go
func TestNotificationsUsecase_MatchJobsForUser(t *testing.T) {
	mockUserSiteRepo := new(mocks.MockUserSiteRepository)
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockEmailService := new(mocks.MockEmailService)
	mockPlanRepo := new(mocks.MockPlanRepository)

	notificationUsecase := NewNotificationUsecase(
		mockUserSiteRepo,
		nil,
		mockEmailService,
		mockNotificationRepo,
		nil,
		mockPlanRepo,
	)

	t.Run("should bulk insert PENDING for matching jobs", func(t *testing.T) {
		userID := 10
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 1, Title: "Go Developer", Company: "Acme", Filters: []string{"developer"}},
			{JobID: 2, Title: "Python Engineer", Company: "Acme", Filters: []string{"developer"}},
			{JobID: 3, Title: "Go Developer Senior", Company: "Beta", Filters: []string{"developer"}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()
		// Jobs 1 and 3 match "developer" in title, job 2 does not
		mockNotificationRepo.On("BulkInsertPendingNotifications", userID, []int{1, 3}).Return(nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should skip bulk insert when no jobs match filters", func(t *testing.T) {
		userID := 20
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 5, Title: "Python Engineer", Company: "Acme", Filters: []string{"java"}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkInsertPendingNotifications")
	})

	t.Run("should match all jobs when user has no filters", func(t *testing.T) {
		userID := 30
		jobsWithFilters := []model.JobWithFilters{
			{JobID: 10, Title: "Any Job", Company: "Acme", Filters: []string{}},
			{JobID: 11, Title: "Another Job", Company: "Beta", Filters: []string{}},
		}

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return(jobsWithFilters, nil).Once()
		mockNotificationRepo.On("BulkInsertPendingNotifications", userID, []int{10, 11}).Return(nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("should return nil when no unnotified jobs exist", func(t *testing.T) {
		userID := 40

		mockNotificationRepo.On("GetUnnotifiedJobsForUser", userID).Return([]model.JobWithFilters{}, nil).Once()

		err := notificationUsecase.MatchJobsForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkInsertPendingNotifications")
	})
}
```

Add `"context"` to the test file imports.

**Step 2: Run tests to verify they fail**

Run: `cd ScrapJobs && go test ./usecase/ -run TestNotificationsUsecase_MatchJobsForUser -v`
Expected: FAIL — `MatchJobsForUser` method does not exist yet

**Step 3: Implement `MatchJobsForUser` in `usecase/notifications_usecase.go`**

```go
func (s *NotificationsUsecase) MatchJobsForUser(ctx context.Context, userID int) error {
	jobs, err := s.notificationRepository.GetUnnotifiedJobsForUser(userID)
	if err != nil {
		return fmt.Errorf("error fetching unnotified jobs for user %d: %w", userID, err)
	}

	if len(jobs) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No unnotified jobs found for user")
		return nil
	}

	var matchedJobIDs []int
	for _, job := range jobs {
		if matchJobWithFiltersFromList(job.Title, job.Filters) {
			matchedJobIDs = append(matchedJobIDs, job.JobID)
		}
	}

	if len(matchedJobIDs) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No jobs matched user filters")
		return nil
	}

	if err := s.notificationRepository.BulkInsertPendingNotifications(userID, matchedJobIDs); err != nil {
		return fmt.Errorf("error inserting pending notifications for user %d: %w", userID, err)
	}

	logging.Logger.Info().Int("user_id", userID).Int("matched_count", len(matchedJobIDs)).Msg("Pending notifications created for user")
	return nil
}
```

Also add the standalone filter function (reuses the same logic as `matchJobWithFilters` but takes filters directly instead of from a user struct):

```go
func matchJobWithFiltersFromList(jobTitle string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	jobTitleLower := strings.ToLower(jobTitle)
	for _, filter := range filters {
		if strings.Contains(jobTitleLower, strings.ToLower(filter)) {
			return true
		}
	}
	return false
}
```

**Step 4: Run tests to verify they pass**

Run: `cd ScrapJobs && go test ./usecase/ -run TestNotificationsUsecase_MatchJobsForUser -v`
Expected: all 4 tests PASS

**Step 5: Run all existing tests to verify no regressions**

Run: `cd ScrapJobs && go test ./usecase/ -v`
Expected: all tests pass

**Step 6: Commit**

```bash
git add usecase/notifications_usecase.go usecase/notification_usecase_test.go
git commit -m "feat: add MatchJobsForUser use case with batch filter and bulk insert"
```

---

### Task 6: Digest Use Case — `SendDigestForUser`

**Files:**
- Modify: `usecase/notifications_usecase.go`
- Test: `usecase/notification_usecase_test.go`

**Step 1: Add `userRepository` dependency to `NotificationsUsecase`**

The digest needs to fetch user name/email by ID. Modify the struct and constructor:

In the struct, add:

```go
userRepository interfaces.UserRepositoryInterface
```

In `NewNotificationUsecase`, add the parameter and assignment:

```go
func NewNotificationUsecase(
	userSiteRepo interfaces.UserSiteRepositoryInterface,
	analysisService interfaces.AnalysisService,
	emailService interfaces.EmailService,
	notificationRepository interfaces.NotificationRepositoryInterface,
	asynqClient *asynq.Client,
	planRepository interfaces.PlanRepositoryInterface,
	userRepository interfaces.UserRepositoryInterface,
) *NotificationsUsecase {
```

**IMPORTANT:** Update ALL call sites of `NewNotificationUsecase`:
- `cmd/worker/main.go` — add `userRepository` as last argument
- `cmd/api/main.go` (if it creates this usecase) — add the argument
- `usecase/notification_usecase_test.go` — add mock user repo in all existing tests

**Step 2: Write the failing tests**

Add to `usecase/notification_usecase_test.go`:

```go
func TestNotificationsUsecase_SendDigestForUser(t *testing.T) {
	mockNotificationRepo := new(mocks.MockNotificationRepository)
	mockEmailService := new(mocks.MockEmailService)
	mockUserRepo := new(mocks.MockUserRepository)

	notificationUsecase := NewNotificationUsecase(
		nil,
		nil,
		mockEmailService,
		mockNotificationRepo,
		nil,
		nil,
		mockUserRepo,
	)

	t.Run("should send digest email and mark notifications as SENT", func(t *testing.T) {
		userID := 10
		pendingJobs := []model.NotificationWithJob{
			{ID: 1, JobID: 100, UserID: userID, JobTitle: "Go Dev", JobCompany: "Acme", JobLocation: "Remote", JobLink: "https://acme.com/1"},
			{ID: 2, JobID: 101, UserID: userID, JobTitle: "Go Senior", JobCompany: "Acme", JobLocation: "SP", JobLink: "https://acme.com/2"},
		}

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return(pendingJobs, nil).Once()
		mockUserRepo.On("GetUserBasicInfo", userID).Return("Test User", "test@example.com", nil).Once()
		mockEmailService.On("SendNewJobsEmail", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("[]*model.Job")).Return(nil).Once()
		mockNotificationRepo.On("BulkUpdateNotificationStatus", userID, []int{100, 101}, "SENT").Return(nil).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockNotificationRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockEmailService.AssertExpectations(t)
	})

	t.Run("should return nil when no pending notifications exist", func(t *testing.T) {
		userID := 20

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return([]model.NotificationWithJob{}, nil).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.NoError(t, err)
		mockEmailService.AssertNotCalled(t, "SendNewJobsEmail")
	})

	t.Run("should NOT mark SENT if email fails", func(t *testing.T) {
		userID := 30
		pendingJobs := []model.NotificationWithJob{
			{ID: 3, JobID: 200, UserID: userID, JobTitle: "Dev", JobCompany: "Beta", JobLocation: "RJ", JobLink: "https://beta.com/1"},
		}

		mockNotificationRepo.On("GetPendingJobsForUser", userID).Return(pendingJobs, nil).Once()
		mockUserRepo.On("GetUserBasicInfo", userID).Return("Fail User", "fail@example.com", nil).Once()
		mockEmailService.On("SendNewJobsEmail", mock.Anything, "fail@example.com", "Fail User", mock.AnythingOfType("[]*model.Job")).Return(fmt.Errorf("SES error")).Once()

		err := notificationUsecase.SendDigestForUser(context.Background(), userID)

		assert.Error(t, err)
		mockNotificationRepo.AssertNotCalled(t, "BulkUpdateNotificationStatus")
	})
}
```

Add `"fmt"` and `"github.com/stretchr/testify/mock"` to imports if not already present.

**Step 2b: Run tests to verify they fail**

Run: `cd ScrapJobs && go test ./usecase/ -run TestNotificationsUsecase_SendDigestForUser -v`
Expected: FAIL — `SendDigestForUser` method does not exist yet

**Step 3: Implement `SendDigestForUser` in `usecase/notifications_usecase.go`**

```go
func (s *NotificationsUsecase) SendDigestForUser(ctx context.Context, userID int) error {
	pendingNotifications, err := s.notificationRepository.GetPendingJobsForUser(userID)
	if err != nil {
		return fmt.Errorf("error fetching pending notifications for user %d: %w", userID, err)
	}

	if len(pendingNotifications) == 0 {
		logging.Logger.Debug().Int("user_id", userID).Msg("No pending notifications for user")
		return nil
	}

	userName, userEmail, err := s.userRepository.GetUserBasicInfo(userID)
	if err != nil {
		return fmt.Errorf("error fetching user info for user %d: %w", userID, err)
	}

	jobs := make([]*model.Job, len(pendingNotifications))
	jobIDs := make([]int, len(pendingNotifications))
	for i, n := range pendingNotifications {
		jobs[i] = &model.Job{
			ID:       n.JobID,
			Title:    n.JobTitle,
			Company:  n.JobCompany,
			Location: n.JobLocation,
			JobLink:  n.JobLink,
		}
		jobIDs[i] = n.JobID
	}

	if err := s.emailService.SendNewJobsEmail(ctx, userEmail, userName, jobs); err != nil {
		return fmt.Errorf("error sending digest email for user %d: %w", userID, err)
	}

	if err := s.notificationRepository.BulkUpdateNotificationStatus(userID, jobIDs, "SENT"); err != nil {
		return fmt.Errorf("error marking notifications as SENT for user %d: %w", userID, err)
	}

	logging.Logger.Info().Int("user_id", userID).Int("job_count", len(jobs)).Msg("Digest email sent and notifications marked as SENT")
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd ScrapJobs && go test ./usecase/ -run TestNotificationsUsecase_SendDigestForUser -v`
Expected: all 3 tests PASS

**Step 5: Run all usecase tests**

Run: `cd ScrapJobs && go test ./usecase/ -v`
Expected: all tests pass (update existing tests that call `NewNotificationUsecase` to pass the new `userRepository` argument — use `nil` or a mock)

**Step 6: Commit**

```bash
git add usecase/notifications_usecase.go usecase/notification_usecase_test.go
git commit -m "feat: add SendDigestForUser use case with email + bulk status update"
```

---

### Task 7: New Task Handlers — Match and Digest in Processor

**Files:**
- Modify: `processor/processor.go`

**Step 1: Add `HandleMatchUserTask` handler**

```go
func (p *TaskProcessor) HandleMatchUserTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.MatchUserPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleMatchUserTask")
		return fmt.Errorf("error decoding MatchUserPayload: %w", err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Processing match job for user")

	if err := p._notifier.MatchJobsForUser(ctx, payload.UserID); err != nil {
		logging.Logger.Error().Err(err).Int("user_id", payload.UserID).Msg("MatchJobsForUser failed")
		return fmt.Errorf("error matching jobs for user %d: %w", payload.UserID, err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Match job completed for user")
	return nil
}
```

**Step 2: Add `HandleSendDigestTask` handler**

```go
func (p *TaskProcessor) HandleSendDigestTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.SendDigestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleSendDigestTask")
		return fmt.Errorf("error decoding SendDigestPayload: %w", err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Processing digest email for user")

	if err := p._notifier.SendDigestForUser(ctx, payload.UserID); err != nil {
		logging.Logger.Error().Err(err).Int("user_id", payload.UserID).Msg("SendDigestForUser failed")
		return fmt.Errorf("error sending digest for user %d: %w", payload.UserID, err)
	}

	logging.Logger.Info().Int("user_id", payload.UserID).Msg("Digest email sent for user")
	return nil
}
```

**Step 3: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles without errors

**Step 4: Commit**

```bash
git add processor/processor.go
git commit -m "feat: add HandleMatchUserTask and HandleSendDigestTask handlers"
```

---

### Task 8: Refactor Scraping Handler — Remove Pipeline Chaining

**Files:**
- Modify: `processor/processor.go`

**Step 1: Simplify `HandleScrapeSiteTask`**

Replace the current implementation (which enqueues `TypeProcessResults` after scraping) with:

```go
func (p *TaskProcessor) HandleScrapeSiteTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ScrapeSitePayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Logger.Error().Err(err).Msg("Falha ao decodificar payload HandleScrapeSiteTask")
		return fmt.Errorf("error to get payload: %w", err)
	}

	logging.Logger.Info().Int("site_id", payload.SiteID).Msg("Processing task to scrap site")

	_, err := p._scraper.ScrapeAndStoreJobs(ctx, payload.SiteScrapingConfig)
	if err != nil {
		logging.Logger.Warn().Err(err).Int("site_id", payload.SiteID).Msg("ScrapeAndStoreJobs failed but task will not be retried")
		if p.dashboardRepo != nil {
			recErr := p.dashboardRepo.RecordScrapingError(payload.SiteID, payload.SiteScrapingConfig.SiteName, err.Error(), t.ResultWriter().TaskID())
			if recErr != nil {
				logging.Logger.Error().Err(recErr).Msg("Failed to record scraping error")
			}
		}
	}

	logging.Logger.Info().Int("site_id", payload.SiteID).Msg("Scraping task completed")
	return nil
}
```

Key changes: removed the `newJobs` check, removed the marshal+enqueue of `TypeProcessResults`, handler now just scrapes and dies.

**Step 2: Delete `HandleFindMatchesTask` method**

Remove the entire `HandleFindMatchesTask` function (was lines 89-118 approximately).

**Step 3: Delete `HandleNotifyNewJobsTask` method**

Remove the entire `HandleNotifyNewJobsTask` function (was lines 120-136 approximately).

**Step 4: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles without errors

**Step 5: Commit**

```bash
git add processor/processor.go
git commit -m "refactor: decouple scraping handler - remove pipeline chaining to match/notify"
```

---

### Task 9: Clean Up Dead Code — Payloads, Constants, and Old Use Case Methods

**Files:**
- Modify: `tasks/payloads.go`
- Modify: `usecase/notifications_usecase.go`

**Step 1: Remove old constants from `tasks/payloads.go`**

Delete these lines:

```go
TypeProcessResults = "process:results"
TypeNotifyNewJobs  = "notify:new_jobs"
```

**Step 2: Remove old payload structs from `tasks/payloads.go`**

Delete these structs:

```go
type ProcessResultsPayload struct {
	SiteID int
	Jobs   []*model.Job
}

type NotifyNewJobsPayload struct {
	User model.UserSiteCurriculum
	Jobs []*model.Job
}
```

**Step 3: Remove old use case methods from `usecase/notifications_usecase.go`**

Delete `FindMatches` method (the old pipeline matcher).
Delete `ProcessNewJobsNotification` method (the old per-batch email sender).

**Step 4: Remove old `matchJobWithFilters` method if fully replaced**

The old `matchJobWithFilters(job model.Job, filters []string)` takes a `model.Job`. It may still be used by manual analysis endpoints. Check if any other code references it. If not, delete it — the new `matchJobWithFiltersFromList` replaces it.

**Step 5: Update old tests that reference removed methods**

In `usecase/notification_usecase_test.go`, delete the `TestNotificationsUsecase_FindMatchesAndNotify` test function (it tests the removed `FindMatches` method).

**Step 6: Clean unused imports in modified files**

Remove any imports that are no longer used after deletions (e.g., `tasks` import from `notifications_usecase.go` if no longer needed).

**Step 7: Verify compilation and tests**

Run: `cd ScrapJobs && go build ./... && go test ./usecase/ -v`
Expected: compiles and all remaining tests pass

**Step 8: Commit**

```bash
git add tasks/payloads.go usecase/notifications_usecase.go usecase/notification_usecase_test.go
git commit -m "refactor: remove dead code from old coupled pipeline"
```

---

### Task 10: Wire Everything — Worker Registration and Scheduler Cronjobs

**Files:**
- Modify: `cmd/worker/main.go`
- Modify: `cmd/scheduler/main.go`

**Step 1: Update `cmd/worker/main.go` — register new handlers, remove old ones**

In the mux registration section, change from:

```go
mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
mux.HandleFunc(tasks.TypeProcessResults, taskProcessor.HandleFindMatchesTask)
mux.HandleFunc(tasks.TypeNotifyNewJobs, taskProcessor.HandleNotifyNewJobsTask)
mux.HandleFunc(tasks.TypeCompleteRegistration, taskProcessor.HandleCompleteRegistrationTask)
```

To:

```go
mux.HandleFunc(tasks.TypeScrapSite, taskProcessor.HandleScrapeSiteTask)
mux.HandleFunc(tasks.TypeMatchUser, taskProcessor.HandleMatchUserTask)
mux.HandleFunc(tasks.TypeSendDigest, taskProcessor.HandleSendDigestTask)
mux.HandleFunc(tasks.TypeCompleteRegistration, taskProcessor.HandleCompleteRegistrationTask)
```

Also update the `NewNotificationUsecase` call to pass `userRepository`:

```go
notificationUsecase := usecase.NewNotificationUsecase(userSiteRepository, nil, emailService, notificationRepository, clientAsynq, planRepository, userRepository)
```

**Step 2: Update `cmd/scheduler/main.go` — add match and digest tickers**

Add `userSiteRepo` and `notificationRepo` initialization (they already have `dbConnection`):

```go
userSiteRepo := repository.NewUserSiteRepository(dbConnection)
notificationRepo := repository.NewNotificationRepository(dbConnection)
```

Add two new tickers:

```go
tickerMatch := time.NewTicker(4 * time.Hour)
defer tickerMatch.Stop()

tickerDigest := time.NewTicker(8 * time.Hour)
defer tickerDigest.Stop()
```

Update the select loop to handle new tickers:

```go
for {
    select {
    case <-ticker.C:
        go enqueueScrapingTasks(context.Background(), siteRepo, client)
    case <-tickerMatch.C:
        go enqueueMatchTasks(context.Background(), userSiteRepo, client)
    case <-tickerDigest.C:
        go enqueueDigestTasks(context.Background(), notificationRepo, client)
    case <-tickerDeleteJobs.C:
        go func() {
            if err := jobRepo.DeleteOldJobs(); err != nil {
                logging.Logger.Error().Err(err).Msg("ERROR: failed to delete old jobs")
            }
        }()
    case sig := <-sigCh:
        logging.Logger.Info().Str("signal", sig.String()).Msg("Scheduler shutting down")
        return
    }
}
```

Add initial runs on startup (like scraping already does):

```go
go enqueueScrapingTasks(context.Background(), siteRepo, client)
go enqueueMatchTasks(context.Background(), userSiteRepo, client)
go enqueueDigestTasks(context.Background(), notificationRepo, client)
```

**Step 3: Implement `enqueueMatchTasks` function in `cmd/scheduler/main.go`**

```go
func enqueueMatchTasks(ctx context.Context, userSiteRepo *repository.UserSiteRepository, client *asynq.Client) {
	userIDs, err := userSiteRepo.GetActiveUserIDs()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Scheduler can't get active user IDs")
		return
	}
	logging.Logger.Info().Int("count", len(userIDs)).Msg("Active users for match")

	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()
			payload, err := json.Marshal(tasks.MatchUserPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal match task")
				return
			}

			task := asynq.NewTask(tasks.TypeMatchUser, payload, asynq.MaxRetry(3))
			info, err := client.EnqueueContext(ctx, task)
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not enqueue match task")
			} else {
				logging.Logger.Info().Int("user_id", uid).Str("task_id", info.ID).Msg("Match task enqueued")
			}
		}(userID)
	}
	wg.Wait()
}
```

**Step 4: Implement `enqueueDigestTasks` function in `cmd/scheduler/main.go`**

```go
func enqueueDigestTasks(ctx context.Context, notificationRepo *repository.NotificationRepository, client *asynq.Client) {
	userIDs, err := notificationRepo.GetUserIDsWithPendingNotifications()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Scheduler can't get users with pending notifications")
		return
	}
	logging.Logger.Info().Int("count", len(userIDs)).Msg("Users with pending digest notifications")

	if len(userIDs) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()
			payload, err := json.Marshal(tasks.SendDigestPayload{UserID: uid})
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not marshal digest task")
				return
			}

			task := asynq.NewTask(tasks.TypeSendDigest, payload, asynq.MaxRetry(3))
			info, err := client.EnqueueContext(ctx, task)
			if err != nil {
				logging.Logger.Error().Err(err).Int("user_id", uid).Msg("Could not enqueue digest task")
			} else {
				logging.Logger.Info().Int("user_id", uid).Str("task_id", info.ID).Msg("Digest task enqueued")
			}
		}(userID)
	}
	wg.Wait()
}
```

**Step 5: Add missing imports to `cmd/scheduler/main.go`**

Ensure `repository` is imported (it's already there for `repository.NewSiteCareerRepository`).

**Step 6: Update any other call sites of `NewNotificationUsecase`**

Search for all call sites:

Run: `cd ScrapJobs && grep -rn "NewNotificationUsecase" --include="*.go"`

Update each call site to pass the `userRepository` parameter.

**Step 7: Verify compilation**

Run: `cd ScrapJobs && go build ./...`
Expected: compiles without errors

**Step 8: Run all tests**

Run: `cd ScrapJobs && go test ./... -v`
Expected: all tests pass

**Step 9: Commit**

```bash
git add cmd/worker/main.go cmd/scheduler/main.go
git commit -m "feat: wire match and digest cronjobs in scheduler and worker"
```

---

### Task 11: Final Verification

**Step 1: Run full test suite**

Run: `cd ScrapJobs && go test ./... -v`
Expected: all tests pass

**Step 2: Build all binaries**

Run: `cd ScrapJobs && go build ./cmd/api/ && go build ./cmd/worker/ && go build ./cmd/scheduler/ && go build ./cmd/archive-monitor/`
Expected: all 4 binaries compile

**Step 3: Verify no dead references**

Run: `cd ScrapJobs && grep -rn "TypeProcessResults\|TypeNotifyNewJobs\|ProcessResultsPayload\|NotifyNewJobsPayload\|HandleFindMatchesTask\|HandleNotifyNewJobsTask\|FindMatches\b\|ProcessNewJobsNotification" --include="*.go"`
Expected: no matches (all dead code removed)

**Step 4: Final commit if any cleanup needed**

```bash
git add -A
git commit -m "chore: final cleanup after batch digest refactor"
```

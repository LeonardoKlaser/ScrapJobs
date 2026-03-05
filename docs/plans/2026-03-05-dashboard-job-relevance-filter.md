# Dashboard Job Relevance Filter — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Show only matched jobs by default in the dashboard, with a toggle to see all jobs and a badge to distinguish matched ones.

**Architecture:** Backend computes a `matched` boolean per job via SQL (comparing job title against user's `filters` from `user_sites`). A `matched_only` query param (default true) controls filtering. Frontend adds a Switch toggle and a "Match" badge.

**Tech Stack:** Go (Gin, database/sql), PostgreSQL (json_array_elements_text), React (TypeScript, TanStack Query, shadcn/ui Switch + Badge), i18next

---

### Task 1: Backend — Add `JobWithMatch` model

**Files:**
- Modify: `ScrapJobs/model/dashboardData.go`

**Step 1: Add `JobWithMatch` struct and update `PaginatedJobs`**

In `model/dashboardData.go`, add a new struct and change the `Jobs` field type:

```go
type JobWithMatch struct {
	Job
	Matched bool `json:"matched"`
}

type PaginatedJobs struct {
	Jobs       []JobWithMatch `json:"jobs"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}
```

Remove the old `PaginatedJobs` definition (which used `[]Job`).

**Step 2: Verify build compiles**

Run: `cd ScrapJobs && go build ./...`
Expected: Compilation errors in `dashboard_repository.go` (scanning into `model.Job` but `PaginatedJobs` now expects `JobWithMatch`). This is expected — Task 2 fixes it.

**Step 3: Commit**

```bash
git add model/dashboardData.go
git commit -m "feat: add JobWithMatch model for dashboard relevance filter"
```

---

### Task 2: Backend — Update repository query with matched field

**Files:**
- Modify: `ScrapJobs/repository/dashboard_repository.go:96-158` (the `GetLatestJobsPaginated` method)

**Step 1: Add `matchedOnly` parameter and rewrite query**

Change the method signature to:

```go
func (dr *DashboardRepository) GetLatestJobsPaginated(userID, page, limit, days int, search string, matchedOnly bool) (model.PaginatedJobs, error) {
```

Replace the entire method body with:

```go
func (dr *DashboardRepository) GetLatestJobsPaginated(userID, page, limit, days int, search string, matchedOnly bool) (model.PaginatedJobs, error) {
	var result model.PaginatedJobs
	result.Page = page
	result.Limit = limit

	offset := (page - 1) * limit

	// matched = true when job title contains any of the user's filter words (case-insensitive)
	// If user has no filters for that site, matched = true (empty filters = match all)
	matchedExpr := `
		CASE WHEN us.filters IS NULL OR us.filters::text = '[]' THEN TRUE
		ELSE EXISTS (
			SELECT 1 FROM json_array_elements_text(us.filters) AS f
			WHERE LOWER(j.title) LIKE '%%' || LOWER(f.value) || '%%'
		) END`

	baseFrom := fmt.Sprintf(`
		FROM jobs j
		JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1
		WHERE 1=1`)

	args := []interface{}{userID}
	argIdx := 2

	if days > 0 {
		baseFrom += fmt.Sprintf(" AND j.created_at >= NOW() - INTERVAL '1 day' * $%d", argIdx)
		args = append(args, days)
		argIdx++
	}

	if search != "" {
		baseFrom += fmt.Sprintf(" AND LOWER(j.title) LIKE '%%' || LOWER($%d) || '%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	if matchedOnly {
		baseFrom += fmt.Sprintf(` AND (%s)`, matchedExpr)
	}

	// Count query
	countQuery := "SELECT COUNT(DISTINCT j.id) " + baseFrom
	err := dr.connection.QueryRow(countQuery, args...).Scan(&result.TotalCount)
	if err != nil {
		return result, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(
		`SELECT DISTINCT j.id, j.site_id, j.title, j.location, j.company, j.job_link, j.requisition_id, COALESCE(j.description, ''), (%s) AS matched %s ORDER BY j.created_at DESC LIMIT $%d OFFSET $%d`,
		matchedExpr, baseFrom, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := dr.connection.Query(dataQuery, args...)
	if err != nil {
		return result, fmt.Errorf("erro ao buscar vagas paginadas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job model.JobWithMatch
		if err := rows.Scan(&job.ID, &job.SiteID, &job.Title, &job.Location, &job.Company, &job.JobLink, &job.RequisitionID, &job.Description, &job.Matched); err != nil {
			return result, fmt.Errorf("erro ao ler vaga: %w", err)
		}
		result.Jobs = append(result.Jobs, job)
	}

	if result.Jobs == nil {
		result.Jobs = []model.JobWithMatch{}
	}

	return result, rows.Err()
}
```

Key details:
- `JOIN user_sites us ON j.site_id = us.site_id AND us.user_id = $1` — moves user filter into JOIN so we access `us.filters`
- `matchedExpr` uses `json_array_elements_text` (not `jsonb_`) because `filters` column is `JSON` type
- Empty filters (`[]`) means matched = true (same logic as Go's `matchJobWithFiltersFromList`)
- `DISTINCT` prevents duplicate rows when a user is subscribed to the same site via multiple entries
- When `matchedOnly=true`, the WHERE clause filters; when false, `matched` is just computed for display

**Step 2: Verify build compiles**

Run: `cd ScrapJobs && go build ./...`
Expected: Compilation error in `controller/dashboardDataController.go` because `GetLatestJobsPaginated` now requires the `matchedOnly` parameter. This is expected — Task 3 fixes it.

**Step 3: Commit**

```bash
git add repository/dashboard_repository.go
git commit -m "feat: compute matched field in dashboard jobs query"
```

---

### Task 3: Backend — Update controller to pass `matched_only` param

**Files:**
- Modify: `ScrapJobs/controller/dashboardDataController.go:68-99` (the `GetLatestJobs` method)

**Step 1: Read and pass the new query param**

After the existing `search := ctx.Query("search")` line (line 84), add:

```go
matchedOnly := ctx.DefaultQuery("matched_only", "true") != "false"
```

Then update the repository call (line 93) to pass it:

```go
data, err := repo.repo.GetLatestJobsPaginated(user.Id, page, limit, days, search, matchedOnly)
```

**Step 2: Verify full build compiles**

Run: `cd ScrapJobs && go build ./...`
Expected: exit 0 — everything compiles.

**Step 3: Run existing tests**

Run: `cd ScrapJobs && go test ./...`
Expected: All tests pass (no existing dashboard tests that would break).

**Step 4: Commit**

```bash
git add controller/dashboardDataController.go
git commit -m "feat: add matched_only query param to dashboard jobs endpoint"
```

---

### Task 4: Frontend — Update model and service

**Files:**
- Modify: `FrontScrapJobs/src/models/dashboard.ts`
- Modify: `FrontScrapJobs/src/services/dashboardService.ts`
- Modify: `FrontScrapJobs/src/hooks/useDashboard.ts`

**Step 1: Add `matched` field to `LatestJob` interface**

In `src/models/dashboard.ts`, add to `LatestJob`:

```typescript
export interface LatestJob {
  id: number
  title: string
  location: string
  company: string
  job_link: string
  matched: boolean
}
```

**Step 2: Add `matched_only` param to service and hook**

In `src/services/dashboardService.ts`, update the `getLatestJobs` params type:

```typescript
getLatestJobs: async (params: {
  days?: number
  search?: string
  page?: number
  limit?: number
  matched_only?: boolean
}): Promise<PaginatedJobsResponse> => {
```

In `src/hooks/useDashboard.ts`, update the `useLatestJobs` params type:

```typescript
export function useLatestJobs(params: {
  days?: number
  search?: string
  page?: number
  limit?: number
  matched_only?: boolean
}) {
```

**Step 3: Verify build**

Run: `cd FrontScrapJobs && npm run build`
Expected: exit 0

**Step 4: Commit**

```bash
cd FrontScrapJobs
git add src/models/dashboard.ts src/services/dashboardService.ts src/hooks/useDashboard.ts
git commit -m "feat: add matched_only param to dashboard jobs API layer"
```

---

### Task 5: Frontend — Add i18n keys

**Files:**
- Modify: `FrontScrapJobs/src/i18n/locales/pt-BR/dashboard.json`
- Modify: `FrontScrapJobs/src/i18n/locales/en-US/dashboard.json`

**Step 1: Add translation keys**

In `pt-BR/dashboard.json`, inside the `"latestJobs"` object, add these keys:

```json
"relevantOnly": "Só relevantes",
"matchBadge": "Match"
```

In `en-US/dashboard.json`, inside the `"latestJobs"` object, add:

```json
"relevantOnly": "Relevant only",
"matchBadge": "Match"
```

**Step 2: Commit**

```bash
cd FrontScrapJobs
git add src/i18n/locales/pt-BR/dashboard.json src/i18n/locales/en-US/dashboard.json
git commit -m "feat: add i18n keys for job relevance toggle"
```

---

### Task 6: Frontend — Add Switch toggle and Match badge to Home.tsx

**Files:**
- Modify: `FrontScrapJobs/src/pages/Home.tsx`

**Step 1: Add state and imports**

Add `Switch` and `Badge` to imports:

```typescript
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
```

Inside the `Home` component, after the existing state declarations (around line 88), add:

```typescript
const [matchedOnly, setMatchedOnly] = useState(true)
```

**Step 2: Pass `matched_only` to the hook**

Update the `useLatestJobs` call (around line 94) to include `matched_only`:

```typescript
const {
  data: jobsData,
  isLoading: isJobsLoading,
  isError: isJobsError
} = useLatestJobs({ page, limit: LIMIT, search, days: days || undefined, matched_only: matchedOnly })
```

**Step 3: Reset page when toggling**

Add a handler after `handleDaysChange`:

```typescript
const handleMatchedOnlyChange = (checked: boolean) => {
  setMatchedOnly(checked)
  setPage(1)
}
```

**Step 4: Add Switch next to SectionHeader**

Replace the `SectionHeader` line (around line 168):

```tsx
<SectionHeader title={t('latestJobs.title')} icon={Sparkles}>
  <div className="flex items-center gap-2">
    <label htmlFor="matched-only" className="text-sm text-muted-foreground cursor-pointer">
      {t('latestJobs.relevantOnly')}
    </label>
    <Switch
      id="matched-only"
      checked={matchedOnly}
      onCheckedChange={handleMatchedOnlyChange}
    />
  </div>
</SectionHeader>
```

Note: `SectionHeader` already accepts `children` (used in the monitored URLs section with a button). The Switch will render on the right side of the header.

**Step 5: Add Match badge in the table**

In the job title `TableCell` (around line 229), add the badge:

```tsx
<TableCell className="max-w-0 font-medium text-foreground">
  <span className="flex items-center gap-2">
    <span className="truncate" title={job.title}>
      {job.title}
    </span>
    {!matchedOnly && job.matched && (
      <Badge variant="default" className="shrink-0 text-xs">
        {t('latestJobs.matchBadge')}
      </Badge>
    )}
  </span>
</TableCell>
```

The badge only shows when `matchedOnly` is false (viewing all jobs) and the job `matched` is true.

**Step 6: Verify build**

Run: `cd FrontScrapJobs && npm run build`
Expected: exit 0

**Step 7: Verify lint**

Run: `cd FrontScrapJobs && npm run lint`
Expected: No new errors.

**Step 8: Commit**

```bash
cd FrontScrapJobs
git add src/pages/Home.tsx
git commit -m "feat: add relevance toggle and match badge to dashboard jobs"
```

---

### Task 7: End-to-end verification

**Step 1: Build backend**

Run: `cd ScrapJobs && go build ./...`
Expected: exit 0

**Step 2: Run backend tests**

Run: `cd ScrapJobs && go test ./...`
Expected: All pass

**Step 3: Build frontend**

Run: `cd FrontScrapJobs && npm run build`
Expected: exit 0

**Step 4: Lint frontend**

Run: `cd FrontScrapJobs && npm run lint`
Expected: No errors

High Priority

1. Authorization issue in submissions
  [submission.go (line 35)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/internal/handlers/submission.go:35) accepts user_id from request body. A user can submit as another user. Take user ID from JWT context instead. `Done`
2. Broken route for submission update
  Route is registered as /api/submission/update, but handler expects ctx.Param("id").
   See [main.go (line 177)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/cmd/api/main.go:177) and [submission.go (line 98)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/internal/handlers/submission.go:98).
   It should likely be PUT /api/submission/:id/update.
3. Role naming mismatch
  Some routes use problem-setter, others use problem-setter. `Done`  
   See [main.go (line 119)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/cmd/api/main.go:119) and [main.go (line 138)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/cmd/api/main.go:138).  
   Domain validation uses problem-setter, so contest routes may fail for problem setters.
4. Config design needs cleanup
  Config uses global vars and weak defaults like default JWT secret/database password.
   See [configuration.go (line 11)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/internal/conf/configuration.go:11).
   Better: create a typed Config struct, validate required env vars, fail fast in production.
5. AutoMigrate in app startup
  [main.go (line 43)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/cmd/api/main.go:43) runs DB migration on every server start. Fine for local, risky for production. Use proper migration files/tooling.

Backend Design Gaps

1. main.go does too much
  Dependency wiring, routes, migration, DB, Swagger, server start all live in one file. Split into app, router, config, server.
2. No centralized error model
  Handlers return mixed 400/404/500 and raw errors. Add typed service errors like ErrNotFound, ErrForbidden, ErrConflict, then map them consistently.
3. Submission judging lacks transaction/idempotency
  Judge0 callbacks can arrive out of order or duplicate. Current update flow can double-count passed tests.
   See [submission_service.go (line 205)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/internal/services/submission_service.go:205).
   Add transaction + token status check + idempotent callback handling.
4. External Judge0 client is mixed inside service
  [submission_service.go (line 60)](/Users/harsh.g/Documents/GitHub/Algoforces-backend/internal/services/submission_service.go:60) mixes business logic, HTTP calls, callback formatting, scoring. Extract JudgeClient.
5. Data model needs stronger constraints
  Add DB-level foreign keys, indexes, unique constraints, and consistent UUID defaults. Current models mostly rely on app logic.
6. Swagger/docs are stale
  README only documents auth/user while backend has contest/problem/testcase/submission APIs. Also docs appear duplicated under docs/ and cmd/api/docs/.
7. Testing coverage is thin for core flows
  Need service tests for contests, problem permissions, submissions, Judge0 callback edge cases, duplicate callbacks, unauthorized access.
8. Repo contains built binaries
  api and main binaries are committed in the root. These should be removed from git and added to .gitignore.
9. No CI/dev hygiene visible
  Add GitHub Actions or similar for go test, go vet, formatting, Swagger generation check, and Docker build.


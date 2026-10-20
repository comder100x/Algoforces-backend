# Algoforces Backend Low-Level Design Notes

These notes explain the current backend implementation in this repository. The goal is to help you learn the code in detail, from the HTTP request entering Gin to the database/Judge0 side effects.

## 1. Project Purpose

Algoforces is a coding-contest backend. It supports:

- user signup, login, JWT authentication, and profile updates
- role-based access for `user`, `admin`, and `problem-setter`
- contest creation and registration
- problem and test case management
- mapping problems into contests with order and max points
- code submission to Judge0
- asynchronous Judge0 callbacks that aggregate testcase results into a final verdict
- Swagger API documentation

The implementation uses Go, Gin, GORM, PostgreSQL, JWT, bcrypt, zerolog, and Judge0.

## 2. High-Level Package Layout

```text
cmd/api/main.go
  Application entry point.

internal/app
  Wires configuration, logger, database, migration, seed data, repositories,
  services, handlers, and router.

internal/conf
  Reads environment configuration with defaults.

internal/router
  Defines all HTTP routes and attaches middleware.

internal/handlers
  Gin HTTP controllers. They parse request input, read auth context, call
  services, and write JSON responses.

internal/services
  Business logic layer. This is where validation, authorization checks,
  default values, scoring, and Judge0 orchestration live.

internal/domain
  Domain entities, request/response DTOs, and repository/use-case interfaces.

internal/repository/postgres
  PostgreSQL persistence implemented with GORM.

internal/middleware
  JWT authentication and role authorization middleware.

internal/utils
  JWT, password hashing, logging, and response helpers.

pkg/database
  PostgreSQL connection setup and DB health helpers.

docs and cmd/api/docs
  Generated Swagger files.
```

## 3. Runtime Startup Flow

The application starts in `cmd/api/main.go`.

```text
main()
  app.New()
    utils.Init(conf.ENV)
    database.NewPostgresConnection()
    migrate(db)
    seed(db), unless ENV is prod/production
    buildHandlers(db)
    router.New(handlers)
  server.Run(router, ":8080")
```

Important files:

- `cmd/api/main.go`
- `internal/app/app.go`
- `internal/app/seed.go`
- `internal/server/server.go`
- `pkg/database/postgres.go`

The app hardcodes the server address as `:8080`.

## 4. Dependency Wiring

The app manually wires dependencies in `internal/app/app.go`.

Repositories are created first:

```text
userRepo
adminRepo
contestRepo
contestRegisterRepo
contestProblemsRepo
problemRepo
testCaseRepo
submissionRepo
```

Services are created next:

```text
authService(userRepo)
adminService(adminRepo)
contestService(contestRepo, userRepo)
contestRegisterService(contestRegisterRepo, contestRepo, userRepo)
contestProblemsService(contestProblemsRepo, contestRepo, problemRepo)
problemService(problemRepo, userRepo)
testCaseService(testCaseRepo)
submissionService(submissionRepo, JUDGE0_API_KEY, JUDGE0_URL)
```

Handlers receive services:

```text
AuthHandler(authService)
UserHandler(authService)
AdminHandler(adminService)
ContestHandler(contestService)
ContestRegisterHandler(contestRegisterService)
ContestProblemsHandler(contestProblemsService)
ProblemHandler(problemService)
TestCaseHandler(testCaseService)
SubmissionHandler(submissionService)
```

This is manual dependency injection. There is no DI framework.

## 5. Configuration Design

Configuration lives in `internal/conf/configuration.go`.

Defaults:

```text
JWT_SECRET       dev-secret-key-change-in-production
DB_HOST          host.docker.internal
DB_PORT          5432
DB_USER          postgres
DB_PASSWORD      postgres
DB_NAME          algoforces
DB_SSLMODE       disable
REDIS_URL        localhost:6379
JUDGE0_URL       http://localhost:2358
APP_URL          http://localhost:8080
JUDGE0_API_KEY   empty
ENV              dev
PAGINATION_LIMIT 10
```

Each value can be overridden by an environment variable with the same name.

Design note: `REDIS_URL` exists, but this codebase currently does not use Redis in the business logic.

## 6. Database Design

The app connects to PostgreSQL using GORM in `pkg/database/postgres.go`.

Connection behavior:

- builds a Postgres DSN from `conf`
- uses GORM PostgreSQL driver
- sets GORM query logger to silent
- sets timestamps using UTC time
- configures connection pool:
  - max idle connections: `10`
  - max open connections: `100`
  - connection max lifetime: `1 hour`
- pings the database before returning

Auto-migrated models:

```text
User
Contest
ContestRegistration
ContestProblems
Problem
TestCase
Submission
SubmissionTestCaseMapping
```

Migration is done through `db.AutoMigrate(...)` during app startup.

## 7. Main Data Models

### User

File: `internal/domain/user.go`

Fields:

```text
Id        uuid primary key
Email     unique, required
Password  bcrypt hash, omitted from JSON
Role      default user
Username
CreatedAt
UpdatedAt
```

Roles currently used:

```text
user
admin
problem-setter
```

### Contest

File: `internal/domain/contest.go`

Fields:

```text
Id
Name
Description
StartTime
EndTime
Duration
Visible
CreatedBy
IsActive
ProblemSetters text[]
CreatedAt
UpdatedAt
```

Design meaning:

- contests have a time window
- overlapping contest windows are rejected by the service
- problem setters are stored as a Postgres text array of user IDs

### Problem

File: `internal/domain/problem.go`

Fields:

```text
UniqueID
Title
Statement
Difficulty: easy | medium | hard
TimeLimitInSeconds
MemoryLimitInMB
CreatedBy
CreatedAt
UpdatedAt
```

Default limits applied by service:

```text
time limit: 1 second
memory limit: 256 MB
```

### TestCase

File: `internal/domain/testcase.go`

Fields:

```text
UniqueID
ProblemID
Input
ExpectedOutput
IsHidden
OrderPosition
```

`IsHidden` controls whether users can see testcase input/output in submission results.

### ContestProblems

File: `internal/domain/contest_problems.go`

This model maps a problem into a contest.

Fields:

```text
UniqueID
ContestID
ProblemID
OrderPosition
MaxPoints
CreatedAt
UpdatedAt
```

This allows the same problem conceptually to be attached to contests with contest-specific ordering and scoring.

### ContestRegistration

File: `internal/domain/contest_register.go`

Fields:

```text
ID
UserID
ContestID
RegisteredAt
Status: registered | unregistered
CreatedAt
UpdatedAt
```

The system soft-unregisters by changing `Status` instead of deleting the registration row.

### Submission

File: `internal/domain/submission.go`

Core fields:

```text
UniqueID
UserId
ContestID
ProblemID
Code
Language: python | cpp | java
SubmittedAt
QueuedAt
Verdict
Score
TestCasesPassed
TotalTestCases
ExecutionTimeInMS
MemoryUsedInKB
CompilationError
RuntimeError
TestCaseResults text[]
FailedTestCase
JudgeCompletedAt
MaxPoints
TokenList text[]
```

### SubmissionTestCaseMapping

File: `internal/domain/submission.go`

This maps one Judge0 token to one testcase.

Fields:

```text
UniqueID
Token
SubmissionID
TestCaseID
Status
TestOrderPosition
TestCaseInput
TestExpectedOutput
IsHidden
CreatedAt
UpdatedAt
```

This table is critical because Judge0 callbacks return only a token. The backend uses the token to recover which submission and testcase the callback belongs to.

## 8. Interfaces and Layering

Each domain file usually contains:

- database model
- request DTOs
- response DTOs
- repository interface
- use-case/service interface

Example from `User`:

```text
UserRepository
  Create
  GetByEmail
  GetByID
  UpdateByID

UserUseCase
  Signup
  Login
  GetUserProfile
  UpdateUserProfile
```

This keeps contracts close to the model. It also means the `domain` package is larger because it includes both persistence models and HTTP DTOs.

## 9. Router Design

File: `internal/router/router.go`

Public routes:

```text
GET  /swagger/*any
GET  /api/health
POST /api/auth/signup
POST /api/auth/login
POST /api/submission/callback
PUT  /api/submission/callback
```

Authenticated route groups use `middleware.AuthMiddleware()`.

Role-protected routes additionally use `middleware.RoleMiddleware(...)`.

Route groups:

```text
/api/auth
/api/user
/api/admin
/api/contest
/api/contest-problem
/api/problem
/api/testcase
/api/submission
```

## 10. Authentication Flow

Signup:

```text
POST /api/auth/signup
  AuthHandler.Signup
    bind SignupRequest
    authService.Signup
      userRepo.GetByEmail
      utils.HashPassword
      userRepo.Create
      utils.GenerateToken
    SendSuccess
```

Login:

```text
POST /api/auth/login
  AuthHandler.Login
    bind LoginRequest
    authService.Login
      userRepo.GetByEmail
      utils.VerifyPassword
      utils.GenerateToken
    SendSuccess
```

JWT details:

- file: `internal/utils/jwt.go`
- signing method: HS256
- claims:
  - `user_id`
  - `role`
  - `email`
  - issued at
  - expires at
- token lifetime: `24 hours`

Password hashing:

- file: `internal/utils/hash.go`
- uses bcrypt

## 11. Auth Middleware Flow

File: `internal/middleware/auth.go`

`AuthMiddleware`:

```text
read Authorization header
accept "Bearer <token>" or raw "<token>"
ValidateToken
set Gin context:
  user_id
  user_email
  user_role
continue request
```

`RoleMiddleware`:

```text
read user_role from context
compare against allowed roles
allow request if any role matches
otherwise return 403
```

Handlers then retrieve current user data with:

```text
middleware.GetUserID(ctx)
middleware.GetUserEmail(ctx)
middleware.GetUserRole(ctx)
```

## 12. Admin Flow

Admin routes require:

```text
AuthMiddleware + RoleMiddleware("admin")
```

Admin operations:

```text
PUT /api/admin/addrole
PUT /api/admin/removerole
GET /api/admin/users
GET /api/admin/admins
GET /api/admin/problem-setters
GET /api/admin/contests
GET /api/admin/registrations
```

Business rules:

- admin can assign a user's role
- admin can fetch all users
- admin can fetch users by role
- role values are validated by request binding

Design note: despite the name `RemoveRole`, the request still requires a `role` value and then sets the user to that role. It does not strictly “remove” unless the caller passes a role that represents the intended result.

## 13. Contest Flow

Contest routes:

```text
POST   /api/contest/create       admin, problem-setter
GET    /api/contest/:id          authenticated
PUT    /api/contest/update       admin, problem-setter
DELETE /api/contest/:id          admin
```

Create contest service rules:

```text
CheckContestInTimeWindow(start, end)
reject overlapping contests
validate every problem setter user exists
validate every problem setter has role problem-setter or admin
create contest with generated UUID
```

Overlap query:

```sql
start_time < requested_end AND end_time > requested_start
```

This catches any existing contest whose time range intersects the requested range.

## 14. Contest Registration Flow

Routes:

```text
POST /api/contest/register       user, admin
POST /api/contest/unregister     user, admin
GET  /api/contest/registrations  user, admin
GET  /api/admin/registrations    admin
```

Register service rules:

```text
verify user exists
verify contest exists
reject if contest starts in less than 5 minutes
if active registration exists, reject duplicate
if previous registration is unregistered, set back to registered
otherwise create new registration
```

Unregister service rules:

```text
verify registration exists
verify contest exists
reject if contest starts in less than 2 minutes
set registration status to unregistered
```

## 15. Problem Flow

Routes:

```text
POST   /api/problem/create   admin, problem-setter
POST   /api/problem/bulk     admin, problem-setter
GET    /api/problem/all      authenticated
GET    /api/problem/:id      authenticated
PUT    /api/problem/update   admin, problem-setter
DELETE /api/problem/:id      admin, problem-setter
```

Create rules:

```text
verify creator exists
verify creator role is admin or problem-setter
apply default time/memory limits when omitted
create problem
```

Update/delete rules:

```text
verify actor exists
verify actor is admin or problem-setter
load existing problem
if actor is not admin, require actor to be the problem creator
update/delete
```

Pagination:

```text
GET /api/problem/all?page=<page>
```

The service reads `PAGINATION_LIMIT`, and repository applies:

```text
Offset(pageOffset * limit)
Limit(limit)
Order(created_at DESC)
```

## 16. Test Case Flow

Routes:

```text
POST   /api/testcase/create              admin, problem-setter
POST   /api/testcase/bulk                admin, problem-setter
GET    /api/testcase/problem/:problemId  authenticated
GET    /api/testcase/:id                 authenticated
PUT    /api/testcase/update              admin, problem-setter
DELETE /api/testcase/:id                 admin, problem-setter
```

Test cases belong to problems through `ProblemID`.

Important fields:

- `Input`
- `ExpectedOutput`
- `IsHidden`
- `OrderPosition`

`OrderPosition` is used later when displaying testcase results in deterministic order.

## 17. Contest Problem Flow

Routes:

```text
POST   /api/contest-problem/create             admin, problem-setter
POST   /api/contest-problem/bulk               admin, problem-setter
GET    /api/contest-problem/:id                authenticated
GET    /api/contest-problem/contest/:contestId authenticated
PUT    /api/contest-problem/update             admin, problem-setter
DELETE /api/contest-problem/:id                admin
```

Create rules:

```text
validate contest exists
validate problem exists
create mapping with order_position and max_points
```

Bulk create:

- validates contest once
- validates each problem
- collects per-item errors
- returns success and failure counts

## 18. Submission Flow

Submission routes:

```text
POST /api/submission/create       user, admin
GET  /api/submission/system/:id   admin, problem-setter
GET  /api/submission/:id          user, admin
PUT  /api/submission/update       user, admin
POST /api/submission/callback     public Judge0 callback
PUT  /api/submission/callback     public Judge0 callback
```

### Create Submission

File: `internal/services/submission_service.go`

Flow:

```text
CreateSubmission handler
  bind CreateSubmissionRequest
  read user_id from JWT context
  service.CreateNewSubmission
    load all test cases for problem
    load problem to get time and memory limits
    load contest-problem mapping to get max points
    create Submission row with Pending verdict
    map language string to Judge0 language ID
    build callback URL from APP_URL + /api/submission/callback
    create Judge0 batch submission, one item per testcase
    create SubmissionTestCaseMapping row per returned Judge0 token
    return "Submission queued successfully for judging"
```

Language mapping:

```text
python -> 71
cpp    -> 54
java   -> 62
```

Judge0 request uses:

```text
source_code
language_id
stdin
expected_output
cpu_time_limit
memory_limit
callback_url
```

Memory is converted from MB to KB:

```text
problem.MemoryLimitInMB * 1024
```

### Judge0 Callback

Judge0 calls:

```text
POST or PUT /api/submission/callback
```

Flow:

```text
JudgeSubmissionCallback handler
  bind JudgeSubmissionCallbackRequest
  service.JudgeSubmissionCallback
    find SubmissionTestCaseMapping by token
    load parent submission
    convert Judge0 time seconds string to milliseconds
    update max execution time and max memory
    map Judge0 status ID to internal verdict
    format testcase result JSON
    ignore duplicate token aggregation if token already seen
    append result to submission.TestCaseResults
    append token to submission.TokenList
    update mapping status
    parse all results
    count accepted results
    find first failed result
    if all testcase callbacks arrived:
      final verdict = Accepted unless first failure exists
      score = maxPoints only if all tests passed, otherwise 0
      set JudgeCompletedAt
      persist final result
    else:
      persist Processing result
```

### Verdict Mapping

```text
Judge0 accepted              -> Accepted
Judge0 wrong answer          -> Wrong Answer
Judge0 time limit exceeded   -> Time Limit Exceeded
Judge0 compilation error     -> Compilation Error
Judge0 runtime statuses      -> Runtime Error
anything else                -> System Error
```

### Submission Result Privacy

`GetSubmissionDetails(id, includeHidden)` controls visibility.

Normal user endpoint:

```text
GET /api/submission/:id
includeHidden = false
```

System endpoint:

```text
GET /api/submission/system/:id
includeHidden = true
```

When `includeHidden` is false:

- hidden testcase results are removed
- token list is hidden
- failed hidden testcase details are hidden
- visible testcase input is sanitized out

User access rule:

- normal users can view only their own submissions
- admin can view user submissions
- system endpoint is for admin/problem-setter and includes hidden testcase data

## 19. Response Format

File: `internal/utils/reponse.go`

Success response:

```json
{
  "data": {},
  "message": "string",
  "status_code": 200
}
```

Error response:

```json
{
  "error": "string",
  "message": "string",
  "status_code": 400
}
```

Design caveat: `SendError` calls `err.Error()`. Some handlers pass `nil` as the error value in certain branches, which can panic. A safer helper would handle `err == nil`.

## 20. Repository Pattern

Repositories live in `internal/repository/postgres`.

They are thin GORM wrappers:

```text
Create -> db.Create
Read   -> db.Where(...).First / Find
Update -> db.Save or db.Model(...).Updates
Delete -> db.Delete
```

Examples:

- `user_repository.go`: create/get/update users
- `problem_repository.go`: problem CRUD and pagination
- `contest_repository.go`: contest CRUD and overlap checks
- `submission_repository.go`: submission CRUD, testcase loading, token mapping

The service layer owns most business decisions. The repository layer mostly owns SQL/GORM data access.

## 21. Seeding Design

File: `internal/app/seed.go`

Seeding runs unless:

```text
ENV=prod
ENV=production
```

Seed data:

- admin user
- problem setter user
- normal user
- demo contest starting 24 hours from startup
- two easy problems
- visible and hidden test cases
- contest-problem mappings

All seed IDs are fixed UUIDs. Inserts use `OnConflict DoNothing`, so restarts do not duplicate seed rows.

Default password:

```text
password123
```

## 22. Swagger Design

Swagger annotations are present in handlers and `cmd/api/main.go`.

Swagger route:

```text
GET /swagger/index.html
```

Generated docs exist in:

```text
docs/
cmd/api/docs/
```

Design caveat: some Swagger comments appear to be older than the router. When learning the API, prefer `internal/router/router.go` as the source of truth for actual routes.

## 23. End-to-End Request Examples

### User Login

```text
Client
  POST /api/auth/login
Gin Router
  AuthHandler.Login
Service
  authService.Login
Repository
  userRepo.GetByEmail
Utils
  VerifyPassword
  GenerateToken
Response
  access_token + user
```

### Create Problem

```text
Client
  POST /api/problem/create with JWT
Middleware
  AuthMiddleware validates token
  RoleMiddleware allows admin/problem-setter
Handler
  ProblemHandler.CreateProblem
Service
  verify current user exists and role is valid
  set default limits
  generate problem UUID
Repository
  problemRepo.CreateProblem
Response
  created problem details
```

### Register Contest

```text
Client
  POST /api/contest/register with JWT
Middleware
  AuthMiddleware
  RoleMiddleware allows user/admin
Handler
  ContestRegisterHandler.RegisterContest
Service
  validate user
  validate contest
  check registration deadline
  create or reactivate registration
Repository
  contestRegisterRepo
Response
  registration details
```

### Submit Code

```text
Client
  POST /api/submission/create with JWT
Middleware
  AuthMiddleware
  RoleMiddleware allows user/admin
Handler
  SubmissionHandler.CreateSubmission
Service
  create DB submission
  send batch request to Judge0
  store token mappings
Response
  submission id and Processing verdict

Judge0 later calls callback once per testcase:
  POST /api/submission/callback
Service
  aggregate testcase result
  finalize submission when all tokens arrived
```

## 24. Authorization Matrix

```text
Public:
  health
  signup
  login
  Judge0 callback
  Swagger

Authenticated any role:
  profile read/update
  contest details
  problem reads
  testcase reads
  contest-problem reads

User/admin:
  contest registration/unregistration
  user submission creation
  normal submission details
  submission status update

Admin/problem-setter:
  contest create/update
  problem create/update/delete
  testcase create/update/delete
  contest-problem create/update/bulk
  system submission details

Admin only:
  role management
  user/admin/problem-setter lists
  all contests via admin route
  all registrations via admin route
  contest deletion
  contest-problem deletion
```

## 25. Important Current Design Observations

These are not blockers for understanding the project, but they are useful learning points.

1. `domain` contains models, DTOs, and interfaces together. This is simple for a small project, but it can become crowded as the system grows.

2. `SendError` is unsafe when `err` is nil. Several handlers pass nil in permission or validation branches.

3. `ValidateToken` parses claims but does not explicitly check `token.Valid` before type assertion. JWT libraries usually validate registered claims during parsing, but production code should still be defensive.

4. Contest update checks overlapping contests but does not appear to exclude the contest being updated. Updating a contest without changing its time can be rejected because it overlaps itself.

5. Submission callback aggregation reads and writes array fields. Concurrent callbacks for the same submission could race unless database-level locking or transactional updates are added.

6. `RemoveRole` acts more like “set role” than “remove role”.

7. `REDIS_URL` is configured but unused.

8. Some Swagger comments do not match actual router paths.

9. The README describes an earlier smaller API surface. The source code is the more accurate reference.

## 26. How To Study This Codebase

Suggested reading order:

1. Start with `cmd/api/main.go`.
2. Read `internal/app/app.go` to understand wiring.
3. Read `internal/router/router.go` to understand API surface.
4. Read `internal/middleware/auth.go` and `internal/utils/jwt.go`.
5. Read `internal/domain/*.go` to understand data shapes.
6. Pick one feature and follow it end to end:
   - `auth.go` handler -> `auth_service.go` -> `user_repository.go`
   - `problem.go` handler -> `problem_service.go` -> `problem_repository.go`
   - `submission.go` handler -> `submission_service.go` -> `submission_repository.go`
7. Read tests under `internal/**/test` to see expected behavior.

Best feature to study first:

```text
Auth
```

Best feature to study for architecture:

```text
Problem CRUD
```

Best feature to study for real backend complexity:

```text
Submission + Judge0 callback aggregation
```

## 27. Mental Model

Think of the backend as five rings:

```text
HTTP layer
  Gin routes, handlers, request binding, responses

Security layer
  JWT middleware and role checks

Use-case layer
  business rules and orchestration

Persistence layer
  GORM repositories and PostgreSQL

External system layer
  Judge0 HTTP API and callbacks
```

Most changes should follow this pattern:

```text
define/update domain DTO
add handler method
add service method with business rules
add repository method if data access is needed
register route
add/update tests
update Swagger docs
```


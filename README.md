# Algoforces Services

A robust Golang backend service providing authentication and user management APIs. Built with clean architecture principles using Gin web framework, GORM ORM, and PostgreSQL database.

## 🚀 Features

- **User Authentication**: JWT-based authentication system
- **User Management**: Complete user profile management
- **Clean Architecture**: Well-structured codebase following domain-driven design
- **Database Integration**: PostgreSQL with GORM ORM
- **Middleware Support**: Authentication middleware for protected routes
- **Docker Support**: Containerized deployment ready
- **Health Checks**: Built-in health monitoring endpoint

## 🛠️ Tech Stack

- **Language**: Go 1.23+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt
- **Containerization**: Docker
- **UUID Generation**: Google UUID

## 📁 Project Structure

```
algoforces/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── conf/
│   │   └── configuration.go     # Configuration management
│   ├── domain/
│   │   └── user.go             # Domain models and interfaces
│   ├── handlers/
│   │   ├── auth.go             # Authentication handlers
│   │   ├── health.go           # Health check handlers
│   │   └── user.go             # User management handlers
│   ├── middleware/
│   │   └── auth.go             # Authentication middleware
│   ├── repository/
│   │   └── postgres/
│   │       └── user_repository.go # Database operations
│   ├── services/
│   │   └── auth_service.go     # Business logic
│   └── utils/
│       ├── hash.go             # Password hashing utilities
│       ├── jwt.go              # JWT token utilities
│       └── response.go         # Response formatting utilities
├── pkg/
│   └── database/
│       └── postgres.go         # Database connection
├── build/
│   └── Dockerfile              # Docker configuration
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
└── README.md                   # Project documentation
```

## 🚦 API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| POST | `/api/auth/signup` | User registration |
| POST | `/api/auth/login` | User authentication |

### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/user/profile` | Get user profile |
| PUT | `/api/user/profile` | Update user profile |

## 📋 Prerequisites

- Go 1.23 or higher
- PostgreSQL database
- Docker (optional, for containerized deployment)

## ⚙️ Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/Algoforces-Services.git
cd Algoforces-Services
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Start Required Services

The API needs PostgreSQL. For local development, you can start the provided devcontainer database stack:

```bash
docker compose -f .devcontainer/docker-compose.yml up -d postgres redis judge0-server judge0-worker
```

This exposes the application database on host port `5433`, while the container database port remains `5432`.

If you already have PostgreSQL running locally, create the `algoforces` database yourself and use your own connection values.

### 4. Configure Environment Variables

Create a local `.env` from the example file:

```bash
cp .env.example .env
```

For the provided Docker Compose database, use:

```bash
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=algoforces
DB_SSLMODE=disable

REDIS_URL=localhost:6379
JUDGE0_URL=http://localhost:2358
APP_URL=http://localhost:8080
JWT_SECRET=your-local-development-secret
```

### 5. Run the Application

Export the env file, then start the API:

```bash
set -a
source .env
set +a
go run cmd/api/main.go
```

The server starts on `http://localhost:8080`.

Check it with:

```bash
curl http://localhost:8080/api/health
```

Swagger docs are available at `http://localhost:8080/swagger/index.html`.

### 6. Seed Data

In local development, the app automatically creates starter data after database migration. Seeding is skipped when `ENV=prod` or `ENV=production`.

Default login accounts use password `password123`:

| Role | Email |
|------|-------|
| Admin | `admin@algoforces.local` |
| Problem Setter | `setter@algoforces.local` |
| User | `user@algoforces.local` |

The seed also creates a demo contest, two easy problems, test cases, and contest-problem mappings. It is safe to restart the app because seed rows use fixed IDs and are inserted only if missing.

## 🏁 Contest flow

End-to-end flow for a contest, from creation to final results. The seeded
demo contest (`d4e5f6a7-4444-4444-8444-456789abcdef`, "Algoforces Demo
Contest") already has the two seeded problems attached, so the local flow
below can be run against it without creating anything first. All requests
except health/login need `Authorization: Bearer $TOKEN`.

1. **Create a contest** (admin / problem-setter): `POST /api/contest/create`
   with name, description, start/end times, duration, and problem setters.
2. **Add problems and test cases** (admin / problem-setter):
   attach problems with `POST /api/contest-problem/create`, then add test
   cases with `POST /api/testcase/create`. The seed attaches
   `e5f6a7b8-5555-4555-8555-56789abcdef1` ("A Plus B") and
   `f6a7b8c9-6666-4666-8666-6789abcdef12` ("Maximum of Three"), each with
   one visible and one hidden test case.
3. **Register** (user): `POST /api/contest/register`.
4. **Start the contest** (PR1 lifecycle): starting a contest opens its live
   window. Starting/ending is provided by the contest lifecycle work (PR1);
   this PR does not change it.
5. **Submit during the live window** (user): `POST /api/submission/create`.
   Submissions outside the live window are rejected by the submission
   gate (PR2); this PR does not change that gate.
6. **List contest submissions**: `GET /api/contest/:id/submissions`
   (this PR). Submissions are returned newest first. A regular user sees
   only their own submissions and never sees hidden test case input/output;
   admin and problem-setter callers see all submissions for the contest,
   including hidden test case data.
7. **Standings** (PR4): standings are computed from the live submission
   records (PR4 ranking); this PR does not change ranking.
8. **End the contest** (PR1 lifecycle): after `EndContest`, submissions
   are rejected by the live-window gate (PR2) and standings show the
   contest as ended/final. There is no snapshot table: finalization relies
   on the PR1 end state plus the PR2 gate, and the submissions list keeps
   reading the live submissions table, so it reflects the final state
   without a separate frozen copy.

```bash
# Register for the seeded demo contest
curl -X POST http://localhost:8080/api/contest/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "contest_id": "d4e5f6a7-4444-4444-8444-456789abcdef"
  }'

# Submit a solution for "A Plus B" during the live window
curl -X POST http://localhost:8080/api/submission/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "contest_id": "d4e5f6a7-4444-4444-8444-456789abcdef",
    "problem_id": "e5f6a7b8-5555-4555-8555-56789abcdef1",
    "code": "print(sum(map(int, input().split())))",
    "language": "python"
  }'

# List submissions for the seeded demo contest (own submissions for a
# regular user; all submissions for admin / problem-setter)
curl -X GET http://localhost:8080/api/contest/d4e5f6a7-4444-4444-8444-456789abcdef/submissions \
  -H "Authorization: Bearer $TOKEN"
```

## 🐳 Docker Deployment

### Build Docker Image

```bash
docker build -f build/Dockerfile -t algoforces-api .
```

### Run with Docker

```bash
docker run -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e DB_PORT=5432 \
  -e DB_USER=your_db_user \
  -e DB_PASSWORD=your_db_password \
  -e DB_NAME=algoforces \
  -e JWT_SECRET=your_jwt_secret \
  algoforces-api
```

### Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'
services:
  api:
    build:
      context: .
      dockerfile: build/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=algoforces
      - DB_PASSWORD=password
      - DB_NAME=algoforces
      - JWT_SECRET=your_jwt_secret_key
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=algoforces
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=algoforces
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Run with:
```bash
docker-compose up -d
```

## 📝 API Usage Examples

### User Registration

```bash
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### User Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Get User Profile (Protected)

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Update User Profile (Protected)

```bash
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "username": "newusername",
    "email": "newemail@example.com"
  }'
```

## 🧪 Testing

### Run Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

## 🤝 Contributing

We welcome contributions! Please follow these steps:

### 1. Fork the Repository

Click the "Fork" button on GitHub to create your own copy.

### 2. Clone Your Fork

```bash
git clone https://github.com/your-username/Algoforces-Services.git
cd Algoforces-Services
```

### 3. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 4. Make Your Changes

- Follow Go best practices and conventions
- Write clear, concise commit messages
- Add tests for new functionality
- Update documentation as needed

### 5. Code Quality Guidelines

- **Formatting**: Use `go fmt` to format your code
- **Linting**: Run `go vet` and consider using `golangci-lint`
- **Testing**: Ensure all tests pass with `go test ./...`
- **Documentation**: Add comments for exported functions and types

### 6. Commit Your Changes

```bash
git add .
git commit -m "feat: add your feature description"
```

### 7. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 8. Create a Pull Request

1. Go to the original repository on GitHub
2. Click "New Pull Request"
3. Select your fork and feature branch
4. Fill out the PR template with:
   - Clear description of changes
   - Any breaking changes
   - Testing instructions
   - Screenshots (if applicable)

### Code Style Guidelines

- Follow standard Go conventions
- Use meaningful variable and function names
- Keep functions small and focused
- Write comprehensive error handling
- Add appropriate logging
- Follow the existing project structure

### Commit Message Format

Use conventional commits format:

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(auth): add password reset functionality`
- `fix(user): resolve profile update validation issue`
- `docs(readme): update installation instructions`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/your-username/Algoforces-Services/issues) page
2. Create a new issue with detailed information
3. Join our community discussions

## 🔄 Changelog

### v1.0.0 (Current)
- Initial release
- User authentication system
- User profile management
- JWT-based authorization
- PostgreSQL integration
- Docker support

## 🚧 Roadmap

- [ ] Add comprehensive test coverage
- [ ] Implement rate limiting
- [ ] Add API documentation with Swagger
- [ ] Implement email verification
- [ ] Add password reset functionality
- [ ] Performance optimization
- [ ] Add monitoring and logging
- [ ] Implement caching layer

## 👥 Contributors

Thanks to all contributors who have helped make this project better!

---

**Made with ❤️ by the Algoforces Team**

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

### 3. Database Setup

Ensure PostgreSQL is running and create a database for the application. Update your database configuration in the environment variables or configuration file.

### 4. Environment Variables

Set up the following environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=algoforces

# JWT Configuration
JWT_SECRET=your_jwt_secret_key

# Server Configuration
PORT=8080
```

### 5. Run the Application

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

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
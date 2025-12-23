package main

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"algoforces/internal/middleware"
	"algoforces/internal/repository/postgres"
	"algoforces/internal/services"
	"algoforces/pkg/database"
	"fmt"
	"log"

	_ "algoforces/docs" // Import generated docs for Swagger

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Algoforces API
// @version         1.0
// @description     API for Algoforces application

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your token only (without Bearer prefix)

func main() {
	// 1. Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 2. Initialize dependencies
	userRepo := postgres.NewUserRepository(db.DB)
	adminRepo := postgres.NewAdminRepository(db.DB)
	authService := services.NewAuthService(userRepo)
	adminService := services.NewAdminService(adminRepo)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(authService)
	adminHandler := handlers.NewAdminHandler(adminService)
	// 3. Setup router
	r := gin.Default()

	//swagger Registration
	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	r.GET("/api/health", handlers.GetHealth)

	// Auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
	}

	// User routes (protected)
	user := r.Group("/api/user")
	user.Use(middleware.AuthMiddleware())
	{
		user.GET("/profile", userHandler.GetUserProfile)
		user.PUT("/profile", userHandler.UpdateUserProfile)
	}

	// Admin routes (protected + admin role required)
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	{
		admin.PUT("/addrole", adminHandler.AddRole)
		admin.PUT("/removerole", adminHandler.RemoveRole)
		admin.GET("/users", adminHandler.GetAllUsers)
		admin.GET("/admins", adminHandler.GetAdmins)
		admin.GET("/problem-setters", adminHandler.GetProblemSetters)
	}

	// 5. Start the Server
	fmt.Println("Starting Algoforces API on :8080...")
	err = r.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

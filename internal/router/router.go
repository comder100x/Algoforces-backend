package router

import (
	"algoforces/internal/handlers"
	"algoforces/internal/middleware"

	_ "algoforces/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth            *handlers.AuthHandler
	User            *handlers.UserHandler
	Admin           *handlers.AdminHandler
	Contest         *handlers.ContestHandler
	ContestRegister *handlers.ContestRegisterHandler
	ContestProblems *handlers.ContestProblemsHandler
	Problem         *handlers.ProblemHandler
	TestCase        *handlers.TestCaseHandler
	Submission      *handlers.SubmissionHandler
}

func New(h Handlers) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/api/health", handlers.GetHealth)

	registerAuthRoutes(r, h)
	registerUserRoutes(r, h)
	registerAdminRoutes(r, h)
	registerContestRoutes(r, h)
	registerContestRegistrationRoutes(r, h)
	registerContestProblemRoutes(r, h)
	registerProblemRoutes(r, h)
	registerTestCaseRoutes(r, h)
	registerSubmissionRoutes(r, h)

	return r
}

func registerAuthRoutes(r *gin.Engine, h Handlers) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/signup", h.Auth.Signup)
		auth.POST("/login", h.Auth.Login)
	}
}

func registerUserRoutes(r *gin.Engine, h Handlers) {
	user := r.Group("/api/user")
	user.Use(middleware.AuthMiddleware())
	{
		user.GET("/profile", h.User.GetUserProfile)
		user.PUT("/profile", h.User.UpdateUserProfile)
	}
}

func registerAdminRoutes(r *gin.Engine, h Handlers) {
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	{
		admin.PUT("/addrole", h.Admin.AddRole)
		admin.PUT("/removerole", h.Admin.RemoveRole)
		admin.GET("/users", h.Admin.GetAllUsers)
		admin.GET("/admins", h.Admin.GetAdmins)
		admin.GET("/problem-setters", h.Admin.GetProblemSetters)
		admin.GET("/contests", h.Contest.GetAllContests)
		admin.GET("/registrations", h.ContestRegister.GetAllRegistrationsForAdmin)
	}
}

func registerContestRoutes(r *gin.Engine, h Handlers) {
	contest := r.Group("/api/contest")
	contest.Use(middleware.AuthMiddleware())
	{
		contest.POST("/create", middleware.RoleMiddleware("admin", "problem-setter"), h.Contest.CreateContest)
		contest.GET("/:id", h.Contest.GetContestDetails)
		contest.PUT("/update", middleware.RoleMiddleware("admin", "problem-setter"), h.Contest.UpdateContest)
		contest.DELETE("/:id", middleware.RoleMiddleware("admin"), h.Contest.DeleteContest)
	}
}

func registerContestRegistrationRoutes(r *gin.Engine, h Handlers) {
	contestRegistration := r.Group("/api/contest")
	contestRegistration.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("user", "admin"))
	{
		contestRegistration.POST("/register", h.ContestRegister.RegisterContest)
		contestRegistration.POST("/unregister", h.ContestRegister.UnregisterContest)
		contestRegistration.GET("/registrations", h.ContestRegister.GetAllRegistrations)
	}
}

func registerContestProblemRoutes(r *gin.Engine, h Handlers) {
	contestProblems := r.Group("/api/contest-problem")
	contestProblems.Use(middleware.AuthMiddleware())
	{
		contestProblems.POST("/create", middleware.RoleMiddleware("admin", "problem-setter"), h.ContestProblems.CreateContestProblem)
		contestProblems.POST("/bulk", middleware.RoleMiddleware("admin", "problem-setter"), h.ContestProblems.BulkCreateContestProblems)
		contestProblems.GET("/:id", h.ContestProblems.GetContestProblem)
		contestProblems.GET("/contest/:contestId", h.ContestProblems.GetContestProblems)
		contestProblems.PUT("/update", middleware.RoleMiddleware("admin", "problem-setter"), h.ContestProblems.UpdateContestProblem)
		contestProblems.DELETE("/:id", middleware.RoleMiddleware("admin"), h.ContestProblems.DeleteContestProblem)
	}
}

func registerProblemRoutes(r *gin.Engine, h Handlers) {
	problem := r.Group("/api/problem")
	problem.Use(middleware.AuthMiddleware())
	{
		problem.POST("/create", middleware.RoleMiddleware("admin", "problem-setter"), h.Problem.CreateProblem)
		problem.POST("/bulk", middleware.RoleMiddleware("admin", "problem-setter"), h.Problem.CreateProblemsInBulk)
		problem.GET("/all", h.Problem.GetAllProblems)
		problem.GET("/:id", h.Problem.GetProblemByID)
		problem.PUT("/update", middleware.RoleMiddleware("admin", "problem-setter"), h.Problem.UpdateProblem)
		problem.DELETE("/:id", middleware.RoleMiddleware("admin", "problem-setter"), h.Problem.DeleteProblem)
	}
}

func registerTestCaseRoutes(r *gin.Engine, h Handlers) {
	testCase := r.Group("/api/testcase")
	testCase.Use(middleware.AuthMiddleware())
	{
		testCase.POST("/create", middleware.RoleMiddleware("admin", "problem-setter"), h.TestCase.CreateTestCase)
		testCase.POST("/bulk", middleware.RoleMiddleware("admin", "problem-setter"), h.TestCase.UploadTestCasesInBulk)
		testCase.GET("/problem/:problemId", h.TestCase.GetAllTestCasesForProblem)
		testCase.GET("/:id", h.TestCase.GetTestCaseDetails)
		testCase.PUT("/update", middleware.RoleMiddleware("admin", "problem-setter"), h.TestCase.UpdateTestCase)
		testCase.DELETE("/:id", middleware.RoleMiddleware("admin", "problem-setter"), h.TestCase.DeleteTestCase)
	}
}

func registerSubmissionRoutes(r *gin.Engine, h Handlers) {
	submission := r.Group("/api/submission")
	submission.PUT("/callback", h.Submission.JudgeSubmissionCallback)
	submission.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("user", "admin"))
	{
		submission.POST("/create", h.Submission.CreateSubmission)
		submission.GET("/:id", h.Submission.GetSubmissionDetails)
		submission.PUT("/update", h.Submission.UpdateSubmissionStatus)
	}
}

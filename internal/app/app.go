package app

import (
	"algoforces/internal/conf"
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"algoforces/internal/repository/postgres"
	"algoforces/internal/router"
	"algoforces/internal/services"
	"algoforces/internal/utils"
	"algoforces/pkg/database"

	"github.com/gin-gonic/gin"
)

type App struct {
	DB     *database.Database
	Router *gin.Engine
}

func New() (*App, error) {
	utils.Init(conf.ENV)

	db, err := database.NewPostgresConnection()
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &App{
		DB:     db,
		Router: router.New(buildHandlers(db)),
	}, nil
}

func (a *App) Close() error {
	if a.DB == nil {
		return nil
	}
	return a.DB.Close()
}

func migrate(db *database.Database) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Contest{},
		&domain.ContestRegistration{},
		&domain.ContestProblems{},
		&domain.Problem{},
		&domain.TestCase{},
		&domain.Submission{},
		&domain.SubmissionTestCaseMapping{},
	)
}

func buildHandlers(db *database.Database) router.Handlers {
	userRepo := postgres.NewUserRepository(db.DB)
	adminRepo := postgres.NewAdminRepository(db.DB)
	contestRepo := postgres.NewContestRepository(db.DB)
	contestRegisterRepo := postgres.NewContestRegisterRepository(db.DB)
	contestProblemsRepo := postgres.NewContestProblemsRepository(db.DB)
	problemRepo := postgres.NewProblemRepository(db.DB)
	testCaseRepo := postgres.NewTestCaseRepository(db.DB)
	submissionRepo := postgres.NewSubmissionRepository(db.DB)

	authService := services.NewAuthService(userRepo)
	adminService := services.NewAdminService(adminRepo)
	contestService := services.NewContestService(contestRepo, userRepo)
	contestRegisterService := services.NewContestRegisterService(contestRegisterRepo, contestRepo, userRepo)
	contestProblemsService := services.NewContestProblemsService(contestProblemsRepo, contestRepo, problemRepo)
	problemService := services.NewProblemService(problemRepo, userRepo)
	testCaseService := services.NewTestCaseService(testCaseRepo)
	submissionService := services.NewSubmissionService(submissionRepo, conf.JUDGE0_API_KEY, conf.JUDGE0_URL)

	return router.Handlers{
		Auth:            handlers.NewAuthHandler(authService),
		User:            handlers.NewUserHandler(authService),
		Admin:           handlers.NewAdminHandler(adminService),
		Contest:         handlers.NewContestHandler(contestService),
		ContestRegister: handlers.NewContestRegisterHandler(contestRegisterService),
		ContestProblems: handlers.NewContestProblemsHandler(contestProblemsService),
		Problem:         handlers.NewProblemHandler(problemService),
		TestCase:        handlers.NewTestCaseHandler(testCaseService),
		Submission:      handlers.NewSubmissionHandler(submissionService),
	}
}

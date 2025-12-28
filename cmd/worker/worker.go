package worker

import (
	"algoforces/internal/domain"
	"algoforces/pkg/database"
	"log"
)

func main() {
	// connect to the database
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	err = db.AutoMigrate(&domain.User{}, &domain.Contest{}, &domain.ContestRegistration{}, &domain.Problem{}, &domain.TestCase{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository
    submissionRepo := postgres.NewSubmissionRepository(db.DB)

    // Initialize Judge Worker
    judgeWorker := worker.NewJudgeWorker(submissionRepo, conf.JUDGE0_URL)

	 // Setup Asynq Server
    redisOpt := asynq.RedisClientOpt{Addr: conf.REDIS_ADDR}

	rv := asynq.NewServer(redisOpt, asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "submission": 10,
        },
    })

    // Register task handlers
    mux := asynq.NewServeMux()
    mux.HandleFunc(queue.TypeSubmissionJudge, judgeWorker.JudgeSubmission)

    // Start the server
    log.Println("Starting Judge Worker...")
    if err := srv.Run(mux); err != nil {
        log.Fatal("Failed to start worker:", err)
    }


}

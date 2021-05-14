package main

import (
	"log"
	"os"

	"cowin-emailer/tasks"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	redisAddr := os.Getenv("REDIS_HOST")

	r := asynq.RedisClientOpt{Addr: redisAddr}
	srv := asynq.NewServer(r, asynq.Config{
		Concurrency: 10,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc("slots:fetch", tasks.HandleFetchSlotsTask)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}

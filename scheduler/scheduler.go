package main

import (
	"context"
	"cowin-emailer/db"
	"cowin-emailer/tasks"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	redisAddr := os.Getenv("REDIS_HOST")
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}
	r := asynq.RedisClientOpt{Addr: redisAddr}
	scheduler := asynq.NewScheduler(
		r,
		&asynq.SchedulerOpts{
			Location: loc,
		},
	)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	go registerTasks(scheduler, db.DB, rdb)

	scheduler.Run()

}

func registerTasks(scheduler *asynq.Scheduler, db_handle *gorm.DB, rdb *redis.Client) {
	ctx := context.Background()
	for {
		var districts []string
		db_handle.Model(&db.User{}).Distinct("district").Find(&districts)
		for _, district := range districts {
			exists, err := rdb.SIsMember(ctx, "districts", district).Result()
			if err != nil {
				continue
			}
			if !exists {
				week := 0
				for i := 0; i < 4; i++ {
					task := tasks.FetchSlotsTask(district, week)
					entryID, e := scheduler.Register("@every 150s", task)
					if e != nil {
						log.Fatal(e)
					}
					log.Printf("registered an entry: %q\n", entryID)
					week += 1
				}
				rdb.SAdd(ctx, "districts", district)
			}
		}
		time.Sleep(15 * time.Minute)
	}
}

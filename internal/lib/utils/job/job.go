package job

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	zerolog "github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"
)

type JobService struct {
	Client    *asynq.Client
	Server    *asynq.Server
	Scheduler *asynq.Scheduler
	Redis     *redis.Client
	Logger    *zerolog.Logger
}

func NewJobService(logger *zerolog.Logger, cfg *config.Config) *JobService {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	}

	client := asynq.NewClient(redisOpt)
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	scheduler := asynq.NewScheduler(redisOpt, nil)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	return &JobService{
		Client:    client,
		Server:    server,
		Scheduler: scheduler,
		Redis:     redisClient,
		Logger:    logger,
	}
}

func (j *JobService) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskWelcome, j.handleWelcomeEmailTask)

	go func() {
		if err := j.Scheduler.Run(); err != nil {
			j.Logger.Error().Err(err).Msg("scheduler stopped")
		}
	}()

	j.Logger.Info().Msg("starting background job server")
	if err := j.Server.Start(mux); err != nil {
		return err
	}
	return nil
}

func (j *JobService) Stop() {
	j.Logger.Info().Msg("stopping background job server")
	j.Scheduler.Shutdown()
	j.Server.Shutdown()
	if err := j.Client.Close(); err != nil {
		j.Logger.Error().Err(err).Msg("failed to close asynq client")
	}
	if err := j.Redis.Close(); err != nil {
		j.Logger.Error().Err(err).Msg("failed to close redis client")
	}
}

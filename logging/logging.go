package logging

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func init(){
	isProduction := os.Getenv("GIN_MODE") == "release"
	

	log.Logger = log.With().Caller().Logger()

	if isProduction {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		Logger = log.Level(zerolog.InfoLevel)
	}else{
		output := zerolog.ConsoleWriter{Out : os.Stderr, TimeFormat : time.Kitchen}
		Logger = log.Output(output).Level(zerolog.DebugLevel)
	}
}

func GinMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		requestID := uuid.New().String()

		reqLogger := Logger.With().Str("request_id", requestID).Logger()

		c.Set("logger", reqLogger)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		reqLogger.Info().Int("status_code", c.Writer.Status()).
						Str("method", c.Request.Method).
						Str("path", c.Request.URL.Path).
						Str("client_ip", c.ClientIP()).
						Dur("latency_ms", latency).
						Msg("request completed")
	}
}

func AsynqMiddleware(h asynq.Handler) asynq.Handler{
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		taskID := t.ResultWriter().TaskID()
		retryCount, _ := asynq.GetRetryCount(ctx)

		taskLogger := Logger.With().
				Str("task_id", taskID).
				Str("task_type", t.Type()).
				Str("queue", "default").
				Int("retry_count", retryCount).
				Logger()

		ctx = taskLogger.WithContext(ctx)

		log.Ctx(ctx).Info().Msg("Starting to process task")
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Task processing failed")
		}else {
			log.Ctx(ctx).Info().Msg("Task processed successfully")
		}

		return err
	})
}
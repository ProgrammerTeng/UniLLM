package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init configures the global zerolog logger.
func Init(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "dev" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).
			With().Timestamp().Caller().Logger()
	} else {
		log.Logger = zerolog.New(os.Stdout).
			With().Timestamp().Str("service", "unillm").Logger()
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if env == "dev" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// GinLogger returns a gin middleware that logs requests with zerolog.
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Int("status", status).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Str("request_id", c.GetString("request_id")).
			Int("body_size", c.Writer.Size()).
			Msg("request")
	}
}

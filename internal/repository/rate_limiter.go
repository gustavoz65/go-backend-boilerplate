package repository

import (
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	Max         int
	Duration    time.Duration
	KeyFunc     func(c *gin.Context) string
	Endpoint    string
	SkipOnerror bool
}

package repository

import (
	"time"

	"github.com/labstack/echo/v4"
)

type RateLimiter struct {
	Max         int
	Duration    time.Duration
	KeyFunc     func(c echo.Context) string
	Endpoint    string
	SkipOnerror bool
}

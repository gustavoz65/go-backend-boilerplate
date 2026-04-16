package service

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
)

type RateLimiterService struct {
	defaultConfig repository.RateLimiter
}

func NewRateLimiterService(max int, duration time.Duration, endpoint string) *RateLimiterService {
	return &RateLimiterService{
		defaultConfig: repository.RateLimiter{
			Max:         max,
			Duration:    duration,
			Endpoint:    endpoint,
			SkipOnerror: true,
		},
	}
}

func (s *RateLimiterService) GetDefaultConfig() repository.RateLimiter {
	return s.defaultConfig
}

func (s *RateLimiterService) CreateConfig(max int, duration time.Duration, endpoint string, keyFunc func(echo.Context) string) repository.RateLimiter {
	return repository.RateLimiter{
		Max:         max,
		Duration:    duration,
		Endpoint:    endpoint,
		KeyFunc:     keyFunc,
		SkipOnerror: true,
	}
}

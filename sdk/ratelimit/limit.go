package ratelimit

import (
	"time"

	"go.uber.org/ratelimit"
)

var limiters = map[string]ratelimit.Limiter{
	"cam":     ratelimit.New(20),
	"sts":     ratelimit.New(600),
	"default": ratelimit.New(20),
}

func Take(service string) time.Time {
	limiter := limiters[service]
	if limiter == nil {
		limiter = limiters["default"]
	}

	return limiter.Take()
}

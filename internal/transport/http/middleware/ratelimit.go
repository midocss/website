package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/midocss/website/internal/transport/http/response"
	"github.com/midocss/website/pkg/apperr"
)

// RateLimitConfig describes a per-client token bucket.
type RateLimitConfig struct {
	// RequestsPerMinute is the sustained rate allowed per client.
	RequestsPerMinute int
	// Burst is how many requests may arrive back to back.
	Burst int
	// TTL is how long an idle client bucket is kept in memory.
	TTL time.Duration
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterStore struct {
	mu       sync.Mutex
	clients  map[string]*clientLimiter
	limit    rate.Limit
	burst    int
	ttl      time.Duration
	lastGC   time.Time
	nowFn    func() time.Time
	gcPeriod time.Duration
}

// RateLimit throttles requests per client IP. It is applied to sensitive
// endpoints (login, register, public forms) rather than globally.
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.RequestsPerMinute
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}

	store := &limiterStore{
		clients:  make(map[string]*clientLimiter),
		limit:    rate.Every(time.Minute / time.Duration(cfg.RequestsPerMinute)),
		burst:    cfg.Burst,
		ttl:      cfg.TTL,
		nowFn:    time.Now,
		gcPeriod: time.Minute,
	}

	return func(c *gin.Context) {
		if !store.allow(c.ClientIP()) {
			response.Fail(c, apperr.TooManyRequests("too many requests, please slow down"))
			return
		}
		c.Next()
	}
}

func (s *limiterStore) allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFn()
	s.collect(now)

	client, ok := s.clients[key]
	if !ok {
		client = &clientLimiter{limiter: rate.NewLimiter(s.limit, s.burst)}
		s.clients[key] = client
	}
	client.lastSeen = now
	return client.limiter.Allow()
}

func (s *limiterStore) collect(now time.Time) {
	if now.Sub(s.lastGC) < s.gcPeriod {
		return
	}
	for key, client := range s.clients {
		if now.Sub(client.lastSeen) > s.ttl {
			delete(s.clients, key)
		}
	}
	s.lastGC = now
}

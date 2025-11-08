package interceptors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors:  make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}

	go rl.resetVisitorCount()
	return rl
}

func (rl *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(rl.resetTime)
		rl.mu.Lock()
		rl.visitors = make(map[string]int)
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) RateLimiterInterceprot(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println("Rate limiter Middleware being retruned ...")
	rl.mu.Lock()
	defer rl.mu.Unlock()

	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to get client IP")
	}

	visitorsIP := p.Addr.String()
	rl.visitors[visitorsIP]++
	log.Printf("+++++++++++ Vistor count from IP: %s: %d\n", visitorsIP, rl.visitors[visitorsIP])

	if rl.visitors[visitorsIP] > rl.limit {
		return nil, status.Error(codes.ResourceExhausted, "Too many requests")
	}

	fmt.Println("Rate Limiter ends....")
	return handler(ctx, req)
}

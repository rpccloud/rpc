package base

import (
	"sync"
	"sync/atomic"
	"time"
)

// SpeedCounter count the speed
type SpeedCounter struct {
	total     int64
	lastCount int64
	lastTime  time.Time
	mu        sync.Mutex
}

// NewSpeedCounter ...
func NewSpeedCounter() *SpeedCounter {
	return &SpeedCounter{
		total:     0,
		lastCount: 0,
		lastTime:  time.Now(),
	}
}

// Count ...
func (p *SpeedCounter) Count() int64 {
	return atomic.AddInt64(&p.total, 1)
}

// Total ...
func (p *SpeedCounter) Total() int64 {
	return atomic.LoadInt64(&p.total)
}

// Calculate ...
func (p *SpeedCounter) Calculate(
	now time.Time,
) (int64, time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	deltaCount := atomic.LoadInt64(&p.total) - p.lastCount
	deltaTime := now.Sub(p.lastTime)

	if deltaTime <= 0 {
		return 0, 0
	} else if deltaCount < 0 {
		return 0, 0
	} else {
		p.lastCount += deltaCount
		p.lastTime = now
		return (deltaCount * int64(time.Second)) / int64(deltaTime), deltaTime
	}
}

package base

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSpeedCounter(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	assert(sc1).IsNotNil()
	assert(sc1.total).Equal(int64(0))
	assert(sc1.lastCount).Equal(int64(0))
	assert(time.Since(sc1.lastTime) < 10*time.Millisecond).IsTrue()
	assert(time.Since(sc1.lastTime) > -10*time.Millisecond).IsTrue()
}

func TestSpeedCounter_Count(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	waitCH := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			sc1.Count()
			waitCH <- true
		}()
	}
	for i := 0; i < 100; i++ {
		<-waitCH
	}
	assert(sc1.Total()).Equal(int64(100))
}

func TestSpeedCounter_Total(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()
	waitCH := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			sc1.Count()
			waitCH <- true
		}()
	}
	for i := 0; i < 100; i++ {
		<-waitCH
	}
	assert(sc1.Total()).Equal(int64(100))
}

func TestSpeedCounter_CalculateSpeed(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	sc1 := NewSpeedCounter()

	for n := 0; n < 100; n++ {
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				sc1.Count()
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
		assert(sc1.Calculate(sc1.lastTime)).
			Equal(int64(0), time.Duration(0))
		assert(sc1.Calculate(sc1.lastTime.Add(time.Second))).
			Equal(int64(100), time.Second)
	}

	// Test(1)
	sc2 := NewSpeedCounter()
	atomic.StoreInt64(&sc2.lastCount, 10)
	assert(sc2.Calculate(sc2.lastTime.Add(time.Second))).
		Equal(int64(0), time.Duration(0))
}
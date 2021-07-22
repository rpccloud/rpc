package base

import (
	"testing"
	"time"
)

func TestNewSpeedCounter(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		assert(v1).IsNotNil()
		assert(v1.total).Equals(int64(0))
		assert(v1.lastCount).Equals(int64(0))
		assert(IsTimeApproximatelyEqual(TimeNow(), v1.lastTime)).IsTrue()
	})
}

func TestSpeedCounter_Count(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					v1.Count()
					time.Sleep(time.Millisecond)
				}
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
		assert(v1.Total()).Equals(int64(1000))
	})
}

func TestSpeedCounter_Total(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					v1.Count()
					time.Sleep(time.Millisecond)
				}
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
		assert(v1.Total()).Equals(int64(1000))
	})
}

func TestSpeedCounter_Calculate(t *testing.T) {
	t.Run("delta time equal zero", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		assert(v1.Calculate(v1.lastTime)).Equals(int64(0), time.Duration(0))
	})

	t.Run("lastCount less than zero", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		v1.lastCount = 1
		assert(v1.Calculate(v1.lastTime.Add(time.Second))).
			Equals(int64(0), time.Duration(0))
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewSpeedCounter()
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				for k := 0; k < 10; k++ {
					v1.Count()
					v1.Count()
					time.Sleep(time.Millisecond)
				}
				waitCH <- true
			}(i)
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
		assert(v1.Calculate(v1.lastTime.Add(time.Second))).
			Equals(int64(2000), time.Second)
	})
}

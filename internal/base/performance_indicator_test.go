package base

import (
	"testing"
	"time"
)

func TestNewPerformanceIndicator(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewPerformanceIndicator()
		assert(len(v1.successArray)).Equals(8)
		for i := 0; i < len(v1.successArray); i++ {
			assert(v1.successArray[i]).Equals(int64(0))
		}
		assert(len(v1.failArray)).Equals(8)
		for i := 0; i < len(v1.failArray); i++ {
			assert(v1.failArray[i]).Equals(int64(0))
		}
		assert(v1.lastTotal).Equals(int64(0))
		assert(IsTimeApproximatelyEqual(TimeNow(), v1.lastTime)).IsTrue()
	})
}

func TestPerformanceIndicator_Count(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewPerformanceIndicator()
		for i := 0; i < 2000; i++ {
			v1.Count(time.Duration(i)*time.Millisecond, true)
			v1.Count(time.Duration(i)*time.Millisecond, false)
		}
		assert(v1.successArray).
			Equals([8]int64{10, 10, 30, 50, 100, 300, 500, 1000})
		assert(v1.failArray).
			Equals([8]int64{10, 10, 30, 50, 100, 300, 500, 1000})
	})
}

func TestPerformanceIndicator_Calculate(t *testing.T) {
	t.Run("delta time equal zero", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewPerformanceIndicator()
		assert(v1.Calculate(v1.lastTime)).Equals(int64(0), time.Duration(0))
	})

	t.Run("deltaCount less than zero", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewPerformanceIndicator()
		v1.lastTotal = 1
		assert(v1.Calculate(v1.lastTime.Add(time.Second))).
			Equals(int64(0), time.Duration(0))
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := NewPerformanceIndicator()
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				for k := 0; k < 10; k++ {
					v1.Count(time.Duration(idx)*time.Millisecond, true)
					v1.Count(time.Duration(idx)*time.Millisecond, false)
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

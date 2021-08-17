package base

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestORCManagerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(orcStatusNone).Equals(uint32(0))
		assert(orcStatusOpened).Equals(uint32(1))
		assert(orcStatusRunning).Equals(uint32(2))
		assert(orcStatusDone).Equals(uint32(3))
		assert(orcStatusClosing).Equals(uint32(4))
		assert(orcStatusClosed).Equals(uint32(5))
	})
}

func TestORCManager_NewORCManager(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o).IsNotNil()
		assert(o.getStatus()).Equals(uint32(orcStatusNone))
		assert(o.statusCond.L).IsNil()
		assert(o.runningCond.L).IsNil()
	})
}

func TestORCManager_getStatus(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for _, status := range []uint32{
			orcStatusNone, orcStatusOpened, orcStatusRunning,
			orcStatusDone, orcStatusClosing, orcStatusClosed,
		} {
			o.setStatus(status)
			assert(o.getStatus()).Equals(status)
		}
	})
}

func TestORCManager_setStatus(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for _, status := range []uint32{
			orcStatusNone, orcStatusOpened, orcStatusRunning,
			orcStatusDone, orcStatusClosing, orcStatusClosed,
		} {
			o.setStatus(status)
			assert(o.getStatus()).Equals(status)
		}
	})
}

func TestORCManager_waitForStatusChange(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		o := NewORCManager()
		waitCH := make(chan bool)

		o.mu.Lock()
		defer o.mu.Unlock()

		go func() {
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(orcStatusClosed)
		}()

		go func() {
			time.Sleep(100 * time.Millisecond)
			o.waitForStatusChange()
			waitCH <- true
		}()

		<-waitCH
	})
}

func TestORCManager_decRuningCount(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		v := int64(10000)

		waitCH := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					o.decRuningCount(&v)
				}
				waitCH <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-waitCH
		}

		assert(v).Equals(int64(0))
	})
}

func TestORCManager_waitForRunningCountChange(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		v := int64(10000)

		o.mu.Lock()
		defer o.mu.Unlock()

		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					o.decRuningCount(&v)
				}
			}()
		}

		for atomic.LoadInt64(&v) > 0 {
			o.waitForRunningCountChange()
		}

		assert(v).Equals(int64(0))
	})
}

func TestORCManager_IsRunning(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		o.setStatus(orcStatusRunning)
		assert(o.IsRunning()).IsTrue()

		for _, status := range []uint32{
			orcStatusNone, orcStatusOpened,
			orcStatusDone, orcStatusClosing, orcStatusClosed,
		} {
			o.setStatus(status)
			assert(o.IsRunning()).IsFalse()
		}
	})
}

func TestORCManager_IsClosed(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		o.setStatus(orcStatusClosed)
		assert(o.IsClosed()).IsTrue()

		for _, status := range []uint32{
			orcStatusNone, orcStatusOpened, orcStatusRunning,
			orcStatusDone, orcStatusClosing,
		} {
			o.setStatus(status)
			assert(o.IsClosed()).IsFalse()
		}
	})
}

func TestORCManager_Open(t *testing.T) {
	t.Run("status error", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for _, status := range []uint32{
			orcStatusOpened, orcStatusRunning,
			orcStatusDone, orcStatusClosing, orcStatusClosed,
		} {
			isCalled := false
			o.setStatus(status)
			assert(o.Open(func() bool {
				isCalled = true
				return true
			})).IsFalse()
			assert(isCalled).IsFalse()
			assert(o.status).Equals(status)
		}
	})

	t.Run("onOpen is nil", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusNone)
		assert(o.Open(nil)).IsTrue()
		assert(o.status).Equals(orcStatusOpened)
	})

	t.Run("onOpen return false", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		isCalled := false
		o.setStatus(orcStatusNone)
		assert(o.Open(func() bool {
			isCalled = true
			return false
		})).IsFalse()
		assert(isCalled).IsTrue()
		assert(o.status).Equals(orcStatusClosed)
	})

	t.Run("onOpen return true", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		isCalled := false
		o.setStatus(orcStatusNone)
		assert(o.Open(func() bool {
			isCalled = true
			return true
		})).IsTrue()
		assert(isCalled).IsTrue()
		assert(o.status).Equals(orcStatusOpened)
	})
}

func TestORCManager_Run(t *testing.T) {
	t.Run("status error", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for _, status := range []uint32{
			orcStatusNone, orcStatusRunning,
			orcStatusDone, orcStatusClosing, orcStatusClosed,
		} {
			isCalled := false
			o.setStatus(status)
			o.Run(func(isRunning func() bool) {
				isCalled = true
			})
			assert(isCalled).IsFalse()
			assert(o.status).Equals(status)
		}
	})

	t.Run("without Close", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		sum := int64(0)
		fn := func(isRunning func() bool) {
			i := 0
			for isRunning() && i < 100 {
				i++
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&sum, 1)
			}
		}
		fnArray := make([]func(isRunning func() bool), 100)
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				fnArray[i] = nil
			} else {
				fnArray[i] = fn
			}
		}
		o.Open(nil)
		o.Run(fnArray...)
		assert(sum).Equals(int64(50 * 100))
		assert(o.getStatus()).Equals(orcStatusDone)
	})

	t.Run("with Close", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.Open(nil)
		fnArray := make([]func(isRunning func() bool), 100)
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				fnArray[i] = nil
			} else {
				fnArray[i] = func(isRunning func() bool) {
					for isRunning() {
						go func() {
							o.Close(nil, nil)
						}()
					}
				}
			}
		}
		o.Run(fnArray...)
		assert(o.getStatus()).Equals(orcStatusClosed)
	})
}

func TestORCManager_Close(t *testing.T) {
	t.Run("status error", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for _, status := range []uint32{
			orcStatusClosing, orcStatusClosed,
		} {
			calledCount := 0
			o.setStatus(status)
			assert(o.Close(func() {
				calledCount++
			}, func() {
				calledCount++
			})).IsFalse()
			assert(calledCount).Equals(0)
			assert(o.status).Equals(status)
		}
	})

	t.Run("status == orcStatusNone", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		calledCount := 0
		o.setStatus(orcStatusNone)
		assert(o.Close(func() {
			calledCount++
		}, func() {
			calledCount++
		})).IsTrue()
		assert(calledCount).Equals(0)
		assert(o.status).Equals(orcStatusClosed)
	})

	t.Run("status == orcStatusOpened", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		calledCount := 0
		o.Open(nil)
		assert(o.Close(func() {
			calledCount++
		}, func() {
			calledCount++
		})).IsTrue()
		assert(calledCount).Equals(2)
		assert(o.status).Equals(orcStatusClosed)
	})

	t.Run("status == orcStatusRunning", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		calledCount := 0
		o.Open(nil)
		go func() {
			o.Run(func(isRunning func() bool) {
				for isRunning() {
					time.Sleep(100 * time.Millisecond)
				}
			})
		}()

		for !o.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		assert(o.Close(func() {
			calledCount++
		}, func() {
			calledCount++
		})).IsTrue()
		assert(calledCount).Equals(2)
		assert(o.status).Equals(orcStatusClosed)
	})
}

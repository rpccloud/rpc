package base

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestORCManagerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(orcLockBit).Equals(4)
		assert(orcStatusMask).Equals(3)
		assert(orcStatusClosed).Equals(0)
		assert(orcStatusReady).Equals(1)
		assert(orcStatusClosing).Equals(2)
	})
}

func TestORCManager_NewORCManager(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o).IsNotNil()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
		assert(o.isWaitChange).Equals(false)
		assert(o.cond.L == nil).IsTrue()
	})
}

func TestORCManager_getBaseSequence(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for i := uint64(0); i < 10; i++ {
			for j := uint64(0); j < 8; j++ {
				atomic.StoreUint64(&o.sequence, i*8+j)
				assert(o.getBaseSequence()).Equals(i * 8)
			}
		}
	})
}

func TestORCManager_getStatus(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()

		for i := uint64(0); i < 10; i++ {
			for j := uint64(0); j < 8; j++ {
				atomic.StoreUint64(&o.sequence, i*8+j)
				assert(o.getStatus()).Equals(j)
			}
		}
	})
}

func TestORCManager_GetRunningFn(t *testing.T) {
	t.Run("orcStatusReady after lock", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		waitCH := make(chan bool)

		isRunning := o.GetRunningFn()

		go func() {
			<-waitCH
			assert(isRunning()).Equals(true)
		}()

		o.Open(func() bool {
			waitCH <- true
			return true
		})
	})

	t.Run("orcStatusReady｜orcLockBit after lock", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		waitCH := make(chan bool)

		o.Open(func() bool {
			return true
		})

		go func() {
			o.Close(func() bool {
				waitCH <- true
				time.Sleep(200 * time.Millisecond)
				return false
			}, func() {

			})
		}()

		o.Run(func(isRunning func() bool) bool {
			<-waitCH
			assert(isRunning()).Equals(true)
			return true
		})
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		atomic.StoreUint64(&o.sequence, 80)
		isRunning := o.GetRunningFn()
		o.setStatus(orcStatusReady)
		assert(isRunning()).IsTrue()
		o.setStatus(orcStatusReady | orcLockBit)
		assert(isRunning()).IsTrue()
		o.setStatus(orcStatusClosing)
		assert(isRunning()).IsFalse()
		o.setStatus(orcStatusClosing | orcLockBit)
		assert(isRunning()).IsFalse()
		o.setStatus(orcStatusClosed)
		assert(isRunning()).IsFalse()
		o.setStatus(orcStatusReady)
		assert(isRunning()).IsFalse()
		o.setStatus(orcStatusReady | orcLockBit)
		assert(isRunning()).IsFalse()
	})
}

func TestORCManager_setStatus(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)

		for i := uint64(0); i < 1000; i++ {
			for _, status := range []uint64{
				orcStatusClosed,
				orcStatusReady | orcLockBit,
				orcStatusClosing | orcLockBit,
			} {
				o := NewORCManager()
				o.sequence = i
				o.isWaitChange = true

				o.setStatus(status)
				assert(o.getStatus()).Equals(status)

				if i%8 != status {
					if status != orcStatusClosed {
						assert(o.sequence).Equals(i - i%8 + status)
					} else {
						assert(o.sequence).Equals(i - i%8 + 8 + status)
					}
					assert(o.isWaitChange).IsFalse()
				} else {
					assert(o.sequence).Equals(i)
					assert(o.isWaitChange).IsTrue()
				}
			}
		}
	})
}

func TestORCManager_waitStatusChange(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		waitCH := make(chan bool)
		o := NewORCManager()
		assert(o.cond.L).IsNil()

		go func() {
			<-waitCH
			o.mu.Lock()
			defer o.mu.Unlock()
			assert(o.isWaitChange).Equals(true)
			o.setStatus(orcStatusReady)
			assert(o.isWaitChange).Equals(false)
		}()

		o.mu.Lock()
		waitCH <- true
		o.waitStatusChange()
		o.mu.Unlock()
		assert(o.cond.L).IsNotNil()
	})
}

func TestORCManager_Open(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		assert(o.Open(nil)).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady)
		assert(o.Open(func() bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady | orcLockBit)
		assert(o.Open(func() bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusReady | orcLockBit))
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosed)
		callCount := 0
		assert(o.Open(func() bool {
			callCount++
			return true
		})).IsTrue()
		assert(callCount).Equals(1)
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
		o.setStatus(orcStatusClosed)
		assert(o.Open(func() bool {
			callCount++
			return false
		})).IsFalse()
		assert(callCount).Equals(2)
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosing | orcLockBit)

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(orcStatusClosed)
		}()

		assert(o.Open(func() bool {
			return true
		})).IsTrue()

		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})
}

func TestORCManager_Run(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady)
		assert(o.Run(nil)).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosed)
		assert(o.Run(func(isRunning func() bool) bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosed | orcLockBit)
		assert(o.Run(func(isRunning func() bool) bool {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed | orcLockBit))
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady)
		callCount := 0
		assert(o.Run(func(isRunning func() bool) bool {
			callCount++
			return true
		})).IsTrue()
		assert(callCount).Equals(1)
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady | orcLockBit)

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(o.getStatus() & 0x3)
		}()

		assert(o.Run(func(isRunning func() bool) bool {
			return true
		})).IsTrue()
		assert(o.Run(func(isRunning func() bool) bool {
			return false
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})
}

func TestORCManager_Close(t *testing.T) {
	t.Run("nil func", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady)
		assert(o.Close(nil, nil)).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusReady))
	})

	t.Run("status orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosed)
		assert(o.Close(func() bool {
			panic("illegal call here")
		}, func() {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})

	t.Run("status orcBitLock | orcStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusClosed | orcLockBit)
		assert(o.Close(func() bool {
			panic("illegal call here")
		}, func() {
			panic("illegal call here")
		})).IsFalse()
		assert(o.getStatus()).Equals(uint64(orcStatusClosed | orcLockBit))
	})

	t.Run("status orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady)
		callCount := 0
		assert(o.Close(func() bool {
			callCount++
			return true
		}, func() {
			callCount++
		})).IsTrue()
		assert(callCount).Equals(2)
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})

	t.Run("status orcBitLock | orcStatusReady", func(t *testing.T) {
		assert := NewAssert(t)
		o := NewORCManager()
		o.setStatus(orcStatusReady | orcLockBit)
		callCount := 0

		go func() {
			time.Sleep(300 * time.Millisecond)
			o.mu.Lock()
			defer o.mu.Unlock()
			o.setStatus(o.getStatus() & orcStatusMask)
		}()

		assert(o.Close(func() bool {
			callCount++
			return true
		}, func() {
			callCount++
		})).IsTrue()
		assert(callCount).Equals(2)
		assert(o.getStatus()).Equals(uint64(orcStatusClosed))
	})
}

func TestORCManagerParallels(t *testing.T) {
	t.Run("test open and close", func(t *testing.T) {
		fnTest := func() (int64, int64, int64) {
			waitCH := make(chan bool)
			o := NewORCManager()

			testCount := 1000
			parallels := 4

			openOK := int64(0)
			runOK := int64(0)
			closeOK := int64(0)

			for n := 0; n < parallels; n++ {
				go func() {
					for i := 0; i < testCount; i++ {
						if o.Open(func() bool {
							time.Sleep(10 * time.Microsecond)
							return true
						}) {
							atomic.AddInt64(&openOK, 1)
						} else {
							time.Sleep(10 * time.Microsecond)
						}
					}
					waitCH <- true
				}()

				go func() {
					for i := 0; i < testCount; i++ {
						if o.Run(func(isRunning func() bool) bool {
							time.Sleep(10 * time.Microsecond)
							return true
						}) {
							atomic.AddInt64(&runOK, 1)
						} else {
							time.Sleep(10 * time.Microsecond)
						}
					}
					waitCH <- true
				}()

				go func() {
					for i := 0; i < testCount; i++ {
						if o.Close(func() bool {
							time.Sleep(10 * time.Microsecond)
							return true
						}, nil) {
							atomic.AddInt64(&closeOK, 1)
						} else {
							time.Sleep(10 * time.Microsecond)
						}
					}
					waitCH <- true
				}()
			}

			for n := 0; n < parallels; n++ {
				<-waitCH
				<-waitCH
				<-waitCH
			}

			if o.Close(func() bool {
				return true
			}, nil) {
				closeOK++
			}

			return openOK, runOK, closeOK
		}

		waitCH := make(chan bool)
		for i := 0; i < 20; i++ {
			go func() {
				assert := NewAssert(t)
				openOK, runOK, closeOK := fnTest()
				assert(openOK).Equals(closeOK)
				assert(runOK > 0).Equals(true)
				waitCH <- true
			}()
		}

		for i := 0; i < 20; i++ {
			<-waitCH
		}
	})
}

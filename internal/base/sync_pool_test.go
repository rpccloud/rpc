package base

import (
	"testing"
)

func TestSyncPoolDebug_Put(t *testing.T) {
	t.Run("put nil value", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(RunWithCatchPanic(func() {
			v1.Put(nil)
		})).Equals("value is nil")
	})

	t.Run("put unmanaged value", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(RunWithCatchPanic(func() {
			ret := make([]byte, 512)
			v1.Put(&ret)
		})).Equals("check failed")
	})

	t.Run("put managed value", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}
		assert(RunWithCatchPanic(func() {
			v1.Put(v1.Get())
		})).IsNil()
	})
}

func TestSyncPoolDebug_Get(t *testing.T) {
	t.Run("get value twice", func(t *testing.T) {
		assert := NewAssert(t)
		outString := captureStdout(func() {
			assert(RunWithCatchPanic(func() {
				v1 := &SyncPoolDebug{
					New: func() interface{} {
						return 3
					},
				}
				v1.Get()
				v1.Get()
			})).Equals("check failed")
		})
		assert(outString).Equals(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	})

	t.Run("get value ok", func(t *testing.T) {
		assert := NewAssert(t)

		v1 := &SyncPoolDebug{
			New: func() interface{} {
				ret := make([]byte, 512)
				return &ret
			},
		}

		outString := captureStdout(func() {
			assert(v1.Get()).IsNotNil()
		})

		assert(outString).Equals(
			"Warn: SyncPool is in debug mode, which may slow down the program",
		)
	})
}

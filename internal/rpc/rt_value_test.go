package rpc

import (
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func testWithRTValue(fn func(v RTValue), v interface{}) {
	testRuntime.thread.Reset()
	rtArray := testRuntime.NewRTArray(1)
	_ = rtArray.Append(v)
	fn(rtArray.Get(0))
	testRuntime.thread.Reset()
}

func TestRTValue_ToBool(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToBool()).
			Equals(false, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBool()).
			Equals(false, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBool()).Equals(false, base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBool()).Equals(true, nil)
		}, true)
	})
}

func TestRTValue_ToInt64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToInt64()).
			Equals(int64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToInt64()).
			Equals(int64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToInt64()).Equals(int64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToInt64()).Equals(int64(12), nil)
		}, 12)
	})
}

func TestRTValue_ToUint64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToUint64()).
			Equals(uint64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToUint64()).
			Equals(uint64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToUint64()).Equals(uint64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToUint64()).Equals(uint64(12), nil)
		}, uint64(12))
	})
}

func TestRTValue_ToFloat64(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToFloat64()).
			Equals(float64(0), base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToFloat64()).
			Equals(float64(0), base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToFloat64()).Equals(float64(0), base.ErrStream)
		}, "kitty")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToFloat64()).Equals(float64(12), nil)
		}, float64(12))
	})
}

func TestRTValue_ToString(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToString()).
			Equals("", base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToString()).
			Equals("", base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToString()).Equals("", base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToString()).Equals("kitty", nil)
		}, "kitty")
	})
}

func TestRTValue_ToBytes(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToBytes()).
			Equals([]byte{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToBytes()).
			Equals([]byte{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBytes()).Equals([]byte{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToBytes()).Equals([]byte{1, 2, 3, 4, 5}, nil)
		}, []byte{1, 2, 3, 4, 5})
	})
}

func TestRTValue_ToArray(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToArray()).
			Equals(Array{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToArray()).
			Equals(Array{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToArray()).Equals(Array{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToArray()).Equals(Array{"1", true, int64(3)}, nil)
		}, Array{"1", true, int64(3)})
	})
}

func TestRTValue_ToRTArray(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToRTArray()).
			Equals(RTArray{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToRTArray()).
			Equals(RTArray{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToRTArray()).Equals(RTArray{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			rtArray, err := v.ToRTArray()
			assert(err).IsNil()
			assert(rtArray.Size()).Equals(3)
			assert(rtArray.Get(0).ToString()).Equals("1", nil)
			assert(rtArray.Get(1).ToBool()).Equals(true, nil)
			assert(rtArray.Get(2).ToInt64()).Equals(int64(3), nil)
		}, Array{"1", true, int64(3)})
	})
}

func TestRTValue_ToAMap(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToMap()).
			Equals(Map{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToMap()).
			Equals(Map{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToMap()).Equals(Map{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToMap()).Equals(Map{"name": "kitty", "age": int64(18)}, nil)
		}, Map{"name": "kitty", "age": int64(18)})
	})
}

func TestRTValue_ToRTMap(t *testing.T) {
	t.Run("err is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{err: base.ErrStream}.ToRTMap()).
			Equals(RTMap{}, base.ErrStream)
	})

	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(RTValue{}.ToRTMap()).
			Equals(RTMap{}, base.ErrRuntimeIllegalInCurrentGoroutine)
	})

	t.Run("type error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			assert(v.ToRTMap()).Equals(RTMap{}, base.ErrStream)
		}, true)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithRTValue(func(v RTValue) {
			rtMap, err := v.ToRTMap()
			assert(err).IsNil()
			assert(rtMap.Size()).Equals(2)
			assert(rtMap.Get("name").ToString()).Equals("kitty", nil)
			assert(rtMap.Get("age").ToInt64()).Equals(int64(18), nil)
		}, Map{"name": "kitty", "age": int64(18)})
	})
}

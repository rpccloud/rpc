package rpc

import (
	"runtime"
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestRTArray(t *testing.T) {
	t.Run("test thread safe", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(7)
		wait := make(chan bool)
		for i := 0; i < 40; i++ {
			go func(idx int) {
				for j := 0; j < 100; j++ {
					assert(v.Append(idx)).IsNil()
				}
				wait <- true
			}(i)
			runtime.GC()
		}
		for i := 0; i < 40; i++ {
			<-wait
		}
		assert(v.Size()).Equals(4000)
		sum := int64(0)
		for i := 0; i < 10000; i++ {
			v, _ := v.Get(i).ToInt64()
			sum += v
		}
		assert(sum).Equals(int64(78000))
	})
}

func TestRTArray_Get(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Get(0).err).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(1).err).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(-1).err).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("index ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Get(0).ToBool()).Equals(true, nil)
	})
}

func TestRTArray_Set(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Set(0, true)).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("unsupported value", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(0, make(chan bool))).Equals(
			base.ErrUnsupportedValue.AddDebug(
				"value type(chan bool) is not supported",
			),
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(1, true)).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Set(-1, true)).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		_ = v.Append("kitty")
		assert(v.Get(0).ToBool()).Equals(true, nil)
		assert(v.Get(1).ToString()).Equals("kitty", nil)
		assert(v.Set(0, "doggy")).Equals(nil)
		assert(v.Set(1, false)).Equals(nil)
		assert(v.Get(0).ToString()).Equals("doggy", nil)
		assert(v.Get(1).ToBool()).Equals(false, nil)
	})
}

func TestRTArray_Append(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Append(true)).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("unsupported value", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Append(make(chan bool))).Equals(
			base.ErrUnsupportedValue.AddDebug(
				"value type(chan bool) is not supported",
			),
		)
		assert(v.Size()).Equals(0)
	})

	t.Run("test ok (string)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Append("kitty")).Equals(nil)
		assert(v.Size()).Equals(1)
		assert(v.Get(0).ToString()).Equals("kitty", nil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(0)
			for j := int64(0); j < 100; j++ {
				assert(v.Append(j)).IsNil()
			}

			assert(v.Size()).Equals(100)
			for j := 0; j < 100; j++ {
				assert(v.Get(j).ToInt64()).Equals(int64(j), nil)
			}
		}
	})
}

func TestRTArray_Delete(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Delete(0)).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("index overflow 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Delete(1)).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index 1 out of range"))
	})

	t.Run("index overflow 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(true)
		assert(v.Delete(-1)).Equals(base.ErrRTArrayIndexOverflow.
			AddDebug("RTArray index -1 out of range"))
	})

	t.Run("delete first elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(0)).Equals(nil)
		assert(v.Size()).Equals(2)
		assert(v.Get(0).ToInt64()).Equals(int64(1), nil)
		assert(v.Get(1).ToInt64()).Equals(int64(2), nil)
	})

	t.Run("delete middle elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(1)).Equals(nil)
		assert(v.Size()).Equals(2)
		assert(v.Get(0).ToInt64()).Equals(int64(0), nil)
		assert(v.Get(1).ToInt64()).Equals(int64(2), nil)
	})

	t.Run("delete last elem", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(0)
		_ = v.Append(1)
		_ = v.Append(2)
		assert(v.Delete(2)).Equals(nil)
		assert(v.Size()).Equals(2)
		assert(v.Get(0).ToInt64()).Equals(int64(0), nil)
		assert(v.Get(1).ToInt64()).Equals(int64(1), nil)
	})
}

func TestRTArray_DeleteAll(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.DeleteAll()).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine,
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		rtArray := testRuntime.NewRTArray(0)

		for i := 0; i < 100; i++ {
			for j := 0; j < 100; j++ {
				_ = rtArray.Append(j)
			}
			assert(rtArray.Size()).Equals(100)
			preCap := cap(*rtArray.items)
			assert(rtArray.DeleteAll()).Equals(nil)
			assert(rtArray.Size()).Equals(0)
			assert(len(*rtArray.items), cap(*rtArray.items)).Equals(0, preCap)
		}
	})
}

func TestRTArray_Size(t *testing.T) {
	t.Run("invalid RTArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		rtArray := RTArray{}
		assert(rtArray.Size()).Equals(-1)
	})

	t.Run("valid RTArray 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		assert(v.Size()).Equals(0)
	})

	t.Run("valid RTArray 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := testRuntime.NewRTArray(0)
		_ = v.Append(1)
		assert(v.Size()).Equals(1)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(0)
			for j := 0; j < i; j++ {
				assert(v.Append(true)).Equals(nil)
			}
			assert(v.Size()).Equals(i)
			for j := 0; j < i; j++ {
				assert(v.Delete(0)).Equals(nil)
			}
			assert(v.Size()).Equals(0)
		}
	})
}

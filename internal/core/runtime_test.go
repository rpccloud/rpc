package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
	"unsafe"
)

var testRuntime = Runtime{id: 0, thread: testThread}

func TestRuntime_lock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(Runtime{}.lock()).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := testThread
		v := Runtime{id: thread.top.lockStatus, thread: thread}
		assert(v.lock()).Equal(thread)
		assert(v.unlock()).IsTrue()
	})
}

func TestRuntime_unlock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			Runtime{}.unlock()
		})).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := testThread
		v := Runtime{id: thread.top.lockStatus, thread: thread}
		assert(base.RunWithCatchPanic(func() {
			v.lock()
			v.unlock()
		})).IsNil()
		// if unlock is success, the thread can lock again
		assert(v.lock()).Equal(thread)
		assert(v.unlock()).Equal(true)
	})
}

func TestRuntime_Reply(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			v := Runtime{}
			_, source = v.Reply(true), base.GetFileLine(0)
		})).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				return rt.Reply(true)
			}, nil),
		)).Equal(true, nil)
	})

	t.Run("test ok (Error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				err := errors.ErrStream.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}, nil),
		)).Equal(nil, errors.ErrStream.AddDebug("error").AddDebug(source))
	})

	t.Run("test ok (error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				err := errors.ErrActionCustom.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}, nil),
		)).Equal(nil, errors.ErrActionCustom.AddDebug("error").AddDebug(source))
	})

	t.Run("argument is (*Error)(nil)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}, nil),
		)).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source),
		)
	})

	t.Run("argument is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equal(emptyReturn)
				return ret
			}, nil),
		)).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source),
		)
	})
}

func TestRuntime_Call(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		ret, source := Runtime{}.Call("#"), base.GetFileLine(0)
		assert(ret).Equal(RTValue{
			err: errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		})
	})

	t.Run("make stream error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source1 := ""
		source2 := ""

		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				errArg := make(chan bool)
				v, s1 := rt.Call("#.test.SayHello", errArg), base.GetFileLine(0)
				source1 = rt.thread.GetActionNode().path + " " + s1
				_, err := v.ToString()
				ret, s2 := rt.Reply(err), base.GetFileLine(0)
				source2 = rt.thread.GetActionNode().path + " " + s2
				return ret
			}, nil),
		)).Equal(nil, errors.ErrUnsupportedValue.
			AddDebug("2nd argument: value type(chan bool) is not supported").
			AddDebug(source1).AddDebug(source2),
		)
	})

	t.Run("depth overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		source1 := ""
		source2 := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(
				func(processor *Processor, rt Runtime) Return {
					// set max call depth to 1
					oldMaxCallDepth := processor.maxCallDepth
					processor.maxCallDepth = 1
					defer func() {
						processor.maxCallDepth = oldMaxCallDepth
					}()

					rtValue, s1 := rt.Call("#.test:SayHello", "ts"), base.GetFileLine(0)
					source1 = rt.thread.GetActionNode().path + " " + s1
					_, err := rtValue.ToString()
					ret, s2 := rt.Reply(err), base.GetFileLine(0)
					source2 = rt.thread.GetActionNode().path + " " + s2
					return ret
				},
				nil,
			),
		)).Equal(nil, errors.ErrCallOverflow.
			AddDebug("call #.test:SayHello level(1) overflows").
			AddDebug(source1).AddDebug(source2),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(
				func(processor *Processor, rt Runtime) Return {
					rtValue := rt.Call("#.test:SayHello", "ts")
					return rt.Reply(rtValue)
				},
				nil,
			),
		)).Equal("hello ts", nil)
	})
}

func TestRuntime_NewRTArray(t *testing.T) {
	t.Run("runtime error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := Runtime{}.NewRTArray(i)
			assert(v.rt).Equal(Runtime{})
			assert(v.items).Equal(nil)
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestRuntime_NewRTMap(t *testing.T) {
	t.Run("runtime error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := Runtime{}.NewRTMap(i)
			assert(v.rt).Equal(Runtime{})
			assert(v.items).Equal(nil)
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})
}

func TestRuntime_GetServiceConfig(t *testing.T) {
	t.Run("runtime is invalid", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Runtime{}
		assert(v.GetServiceConfig("name")).Equal(nil, false)
	})

	t.Run("action node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testRuntime.GetServiceConfig("name")).Equal(nil, false)
	})

	t.Run("service node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.top.actionNode = unsafe.Pointer(&rpcActionNode{})
		assert(testRuntime.GetServiceConfig("name")).Equal(nil, false)
		testRuntime.thread.top.actionNode = nil
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equal(nil, false)
				return rt.Reply(true)
			},
			nil,
		)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equal("kitty", true)
				return rt.Reply(true)
			},
			Map{"name": "kitty"},
		)
	})
}

func TestRuntime_SetServiceConfig(t *testing.T) {
	t.Run("runtime is invalid", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Runtime{}
		assert(v.SetServiceConfig("name", "kitty")).Equal(false)
	})

	t.Run("action node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testRuntime.SetServiceConfig("name", "kitty")).Equal(false)
	})

	t.Run("service node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.top.actionNode = unsafe.Pointer(&rpcActionNode{})
		assert(testRuntime.SetServiceConfig("name", "kitty")).Equal(false)
		testRuntime.thread.top.actionNode = nil
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.SetServiceConfig("name", "kitty")).Equal(true)
				assert(rt.GetServiceConfig("name")).Equal("kitty", true)
				return rt.Reply(true)
			},
			nil,
		)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equal("doggy", true)
				assert(rt.SetServiceConfig("name", "kitty")).Equal(true)
				assert(rt.GetServiceConfig("name")).Equal("kitty", true)
				return rt.Reply(true)
			},
			Map{"name": "doggy"},
		)
	})
}

func TestRuntime_parseResponseStream(t *testing.T) {
	t.Run("errCode format error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteInt64(3)
		assert(testRuntime.parseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("Read ret error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(0)
		assert(testRuntime.parseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("Read ret ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(0)
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v).ToBool()).Equal(true, nil)
	})

	t.Run("error message Read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(uint64(base.ErrorTypeSecurity))
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("error stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})

	t.Run("error stream ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.WriteUint64(errors.ErrStream.GetCode())
		v.WriteString(errors.ErrStream.GetMessage())
		assert(testRuntime.parseResponseStream(v)).Equal(
			RTValue{err: errors.ErrStream},
		)
	})
}

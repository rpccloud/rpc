package rpc

import (
	"testing"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
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
		assert(v.lock()).Equals(thread)
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
		assert(v.lock()).Equals(thread)
		assert(v.unlock()).Equals(true)
	})
}

func TestRuntime_Reply(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			v := Runtime{}
			_, source = v.Reply(true), base.GetFileLine(0)
		})).Equals(
			base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				return rt.Reply(true)
			}, nil),
		)).Equals(true, nil)
	})

	t.Run("test ok (Error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				err := base.ErrStream.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equals(emptyReturn)
				return ret
			}, nil),
		)).Equals(
			nil,
			base.ErrStream.AddDebug("error").AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok (error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				err := base.ErrStream.AddDebug("error")
				ret, s := rt.Reply(err), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equals(emptyReturn)
				return ret
			}, nil),
		)).Equals(
			nil,
			base.ErrStream.AddDebug("error").AddDebug(source).Standardize(),
		)
	})

	t.Run("argument is (*Error)(nil)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equals(emptyReturn)
				return ret
			}, nil),
		)).Equals(
			nil,
			base.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source).
				Standardize(),
		)
	})

	t.Run("argument is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(ParseResponseStream(
			testWithProcessorAndRuntime(func(_ *Processor, rt Runtime) Return {
				ret, s := rt.Reply((*base.Error)(nil)), base.GetFileLine(0)
				source = rt.thread.GetActionNode().path + " " + s
				assert(ret).Equals(emptyReturn)
				return ret
			}, nil),
		)).Equals(
			nil,
			base.ErrUnsupportedValue.
				AddDebug("value is nil").
				AddDebug(source).
				Standardize(),
		)
	})
}

func TestRuntime_Post(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		ret, source := Runtime{}.Post("", "Msg", "HI"), base.GetFileLine(0)
		assert(ret).
			Equals(base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source))
	})

	t.Run("Post EndPoint error", func(t *testing.T) {
		assert := base.NewAssert(t)
		var e error
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				e = rt.Post("error endpoint", "Msg", "HI")
				return rt.Reply("ok")
			},
			nil,
		)
		assert(e).Equals(base.ErrRuntimePostEndpoint)
	})

	t.Run("Post value not supported", func(t *testing.T) {
		assert := base.NewAssert(t)
		var e error
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				e = rt.Post(rt.GetPostEndPoint(), "Msg", make(chan bool))
				return rt.Reply("ok")
			},
			nil,
		)
		assert(e).Equals(base.ErrUnsupportedValue.AddDebug(base.ConcatString(
			"value type(chan bool) is not supported",
		)))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		var e error
		stream := testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				e = rt.Post(rt.GetPostEndPoint(), "Msg", "HI")
				return emptyReturn
			},
			nil,
		)
		assert(e).IsNil()
		assert(stream.GetKind()).Equals(uint8(StreamKindRPCBoardCast))
		assert(stream.GetGatewayID()).Equals(uint64(1234))
		assert(stream.GetSessionID()).Equals(uint64(5678))
		assert(stream.Read()).Equals("#.test%Msg", nil)
		assert(stream.Read()).Equals("HI", nil)
		assert(stream.IsReadFinish()).IsTrue()
	})
}

func TestRuntime_Call(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		ret, source := Runtime{}.Call("#"), base.GetFileLine(0)
		assert(ret).Equals(RTValue{
			err: base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
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
		)).Equals(nil, base.ErrUnsupportedValue.
			AddDebug("2nd argument: value type(chan bool) is not supported").
			AddDebug(source1).AddDebug(source2).Standardize(),
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

					getFL := base.GetFileLine
					rtValue, s1 := rt.Call("#.test:SayHello", "ts"), getFL(0)
					source1 = rt.thread.GetActionNode().path + " " + s1
					_, err := rtValue.ToString()
					ret, s2 := rt.Reply(err), base.GetFileLine(0)
					source2 = rt.thread.GetActionNode().path + " " + s2
					return ret
				},
				nil,
			),
		)).Equals(nil, base.ErrCallOverflow.
			AddDebug("call #.test:SayHello level(1) overflows").
			AddDebug(source1).AddDebug(source2).Standardize(),
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
		)).Equals("hello ts", nil)
	})
}

func TestRuntime_NewRTArray(t *testing.T) {
	t.Run("runtime error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := Runtime{}.NewRTArray(i)
			assert(v.rt).Equals(Runtime{})
			assert(v.items).Equals(nil)
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTArray(i)
			assert(v.rt).Equals(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equals(0, i)
		}
	})
}

func TestRuntime_NewRTMap(t *testing.T) {
	t.Run("runtime error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := Runtime{}.NewRTMap(i)
			assert(v.rt).Equals(Runtime{})
			assert(v.items).Equals(nil)
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := testRuntime.NewRTMap(i)
			assert(v.rt).Equals(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equals(0, i)
		}
	})
}

func TestRuntime_GetPostEndPoint(t *testing.T) {
	t.Run("lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Runtime{}
		assert(v.GetPostEndPoint()).Equals("")
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Runtime{thread: testThread}
		v.thread.top.stream = NewStream()
		v.thread.top.stream.SetGatewayID(13)
		v.thread.top.stream.SetSessionID(15)
		assert(base.DecryptSessionEndpoint(v.GetPostEndPoint())).
			Equals(uint64(13), uint64(15), true)
		v.thread.top.stream = nil
	})
}

func TestRuntime_GetServiceConfig(t *testing.T) {
	t.Run("runtime is invalid", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := Runtime{}
		assert(v.GetServiceConfig("name")).Equals(nil, false)
	})

	t.Run("action node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testRuntime.GetServiceConfig("name")).Equals(nil, false)
	})

	t.Run("service node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.top.actionNode = unsafe.Pointer(&rpcActionNode{})
		assert(testRuntime.GetServiceConfig("name")).Equals(nil, false)
		testRuntime.thread.top.actionNode = nil
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equals(nil, false)
				return rt.Reply(true)
			},
			nil,
		)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equals("kitty", true)
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
		assert(v.SetServiceConfig("name", "kitty")).Equals(false)
	})

	t.Run("action node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testRuntime.SetServiceConfig("name", "kitty")).Equals(false)
	})

	t.Run("service node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.top.actionNode = unsafe.Pointer(&rpcActionNode{})
		assert(testRuntime.SetServiceConfig("name", "kitty")).Equals(false)
		testRuntime.thread.top.actionNode = nil
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.SetServiceConfig("name", "kitty")).Equals(true)
				assert(rt.GetServiceConfig("name")).Equals("kitty", true)
				return rt.Reply(true)
			},
			nil,
		)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		testWithProcessorAndRuntime(
			func(processor *Processor, rt Runtime) Return {
				assert(rt.GetServiceConfig("name")).Equals("doggy", true)
				assert(rt.SetServiceConfig("name", "kitty")).Equals(true)
				assert(rt.GetServiceConfig("name")).Equals("kitty", true)
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
		v.SetKind(StreamKindRPCResponseError)
		v.WriteInt64(3)
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("errCode overflows", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseError)
		v.WriteUint64(1 << 32)
		v.WriteString(base.ErrStream.GetMessage())
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("Read ret error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseError)
		v.WriteUint64(0)
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("Read ret ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseOK)
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v).ToBool()).Equals(true, nil)
	})

	t.Run("error message Read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseError)
		v.WriteUint64(uint64(base.ErrorTypeSecurity))
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("error stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseError)
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		v.WriteBool(true)
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("error stream ok (StreamKindRPCResponseError)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindRPCResponseError)
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("error stream ok (StreamKindSystemErrorReport)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindSystemErrorReport)
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})

	t.Run("error stream kind error", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := NewStream()
		v.SetKind(StreamKindConnectResponse)
		v.WriteUint64(uint64(base.ErrStream.GetCode()))
		v.WriteString(base.ErrStream.GetMessage())
		assert(testRuntime.parseResponseStream(v)).Equals(
			RTValue{err: base.ErrStream},
		)
	})
}

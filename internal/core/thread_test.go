package core

import (
	"fmt"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

var (
	fnEvalBack   = func(stream *Stream) {}
	fnEvalFinish = func(thread *rpcThread) {}
	testThread   = newThread(
		testProcessor,
		5*time.Second,
		2048,
		func(stream *Stream) {},
		func(thread *rpcThread) {},
	)
)

func TestNewRPCThreadFrame(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 100; i++ {
			v := newRPCThreadFrame()
			assert(v).IsNotNil()
			v.Release()
		}
	})
}

func TestRpcThreadFrame_Reset(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newRPCThreadFrame()
		v.stream = NewStream()
		v.actionNode = unsafe.Pointer(&rpcActionNode{})
		v.from = "#"
		v.depth = 13
		v.cacheArrayItemsPos = 16
		v.cacheMapItemsPos = 16
		v.cacheArrayEntryPos = 7
		v.cacheMapEntryPos = 7
		v.retStatus = 1
		v.lockStatus = 82737243243
		v.parentRTWritePos = 100
		v.next = &rpcThreadFrame{}
		v.Reset()
		assert(v.stream).Equal(nil)
		assert(v.actionNode).Equal(nil)
		assert(v.from).Equal("")
		assert(v.depth).Equal(uint16(13))
		assert(v.cacheArrayItemsPos).Equal(uint32(0))
		assert(v.cacheMapItemsPos).Equal(uint32(0))
		assert(v.cacheArrayEntryPos).Equal(uint32(0))
		assert(v.cacheMapEntryPos).Equal(uint32(0))
		assert(v.retStatus).Equal(uint32(1))
		assert(v.lockStatus).Equal(uint64(82737243243))
		assert(v.parentRTWritePos).Equal(streamPosBody)
		assert(v.next).Equal(nil)
		v.Release()
	})
}

func TestRpcThreadFrame_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := newRPCThreadFrame()
			assert(v.cacheArrayItemsPos).Equal(uint32(0))
			v.cacheArrayItemsPos = 32
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}

func TestRTArrayNewRTArray(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTArray(Runtime{}, 2)).Equal(RTArray{})
	})

	t.Run("size is less than zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		assert(newRTArray(testRuntime, -1)).Equal(RTArray{})
	})

	t.Run("size is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := newRTArray(testRuntime, 0)
		assert(v.rt).Equal(testRuntime)
		assert(v.items != nil)
		assert(len(*v.items), cap(*v.items)).Equal(0, 0)
	})

	t.Run("test ok 1", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := newRTArray(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, i)
		}
	})

	t.Run("test ok 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()

		for i := 0; i < 500; i++ {
			v := newRTArray(testRuntime, 1)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, 1)
		}
	})
}

func TestNewRTMap(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newRTMap(Runtime{}, 2)).Equal(RTMap{})
	})

	t.Run("size is less than zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		assert(newRTMap(testRuntime, -1)).Equal(RTMap{})
	})

	t.Run("size is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()
		v := newRTMap(testRuntime, 0)
		assert(v.rt).Equal(testRuntime)
		assert(v.items != nil)
		assert(len(*v.items), cap(*v.items), *v.length).Equal(0, 0, uint32(0))
	})

	t.Run("test ok 1", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			testRuntime.thread.Reset()
			v := newRTMap(testRuntime, i)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items), *v.length).Equal(0, i, uint32(0))
		}
	})

	t.Run("test ok 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRuntime.thread.Reset()

		for i := 0; i < 500; i++ {
			v := newRTMap(testRuntime, 1)
			assert(v.rt).Equal(testRuntime)
			assert(len(*v.items), cap(*v.items)).Equal(0, 1)
		}
	})
}

func TestNewThread(t *testing.T) {
	t.Run("processor is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(nil, 5*time.Second, 2048, fnEvalBack, nil)).IsNil()
	})

	t.Run("onEvalBack is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(testProcessor, 5*time.Second, 2048, nil, fnEvalFinish)).
			IsNil()
	})

	t.Run("onEvalFinish is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, nil)).
			IsNil()
	})

	t.Run("test ok (timeout 1s)", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 32; i++ {
			v := newThread(testProcessor, 100*time.Millisecond, 2048, fnEvalBack, fnEvalFinish)
			assert(v.processor).Equal(testProcessor)
			assert(v.inputCH).IsNotNil()
			assert(v.closeCH).IsNotNil()
			assert(v.closeTimeout).Equal(time.Second)
			assert(v.top).Equal(&v.rootFrame)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.rtStream).IsNotNil()
			assert(len(v.cacheEntry)).Equal(8)
			assert(len(v.cacheArrayItems), cap(v.cacheArrayItems)).Equal(128, 128)
			assert(len(v.cacheMapItems), cap(v.cacheMapItems)).Equal(32, 32)
			v.Close()
			assert(v.closeCH).IsNil()
		}
	})

	t.Run("test ok (timeout 5s)", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 32; i++ {
			v := newThread(
				testProcessor,
				5*time.Second,
				2048,
				fnEvalBack,
				fnEvalFinish,
			)
			assert(v.processor).Equal(testProcessor)
			assert(v.inputCH).IsNotNil()
			assert(v.closeCH).IsNotNil()
			assert(v.closeTimeout).Equal(5 * time.Second)
			assert(v.top).Equal(&v.rootFrame)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.rtStream).IsNotNil()
			assert(len(v.cacheEntry)).Equal(8)
			assert(len(v.cacheArrayItems), cap(v.cacheArrayItems)).Equal(128, 128)
			assert(len(v.cacheMapItems), cap(v.cacheMapItems)).Equal(32, 32)
			v.Close()
			assert(v.closeCH).IsNil()
		}
	})

	t.Run("test eval", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := 0; i < 100; i++ {
			chBack := make(chan bool, 1)
			chFinish := make(chan bool, 1)
			v := newThread(
				testProcessor,
				5*time.Second,
				2048,
				func(stream *Stream) {
					chBack <- true
				},
				func(thread *rpcThread) {
					chFinish <- true
				},
			)
			v.PutStream(NewStream())
			assert(<-chBack).Equal(true)
			assert(<-chFinish).Equal(true)
			assert(v.sequence > 0).Equal(true)
			assert(v.sequence % 2).Equal(uint64(0))
			assert(v.top).Equal(&v.rootFrame)
			v.Close()
			assert(v.top.stream).Equal(nil)
		}
	})
}

func TestRpcThread_Reset(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(
			testProcessor,
			5*time.Second,
			2048,
			func(stream *Stream) {},
			func(thread *rpcThread) {},
		)

		v.rtStream.Write(3)
		v.rootFrame.stream = NewStream()
		v.cacheEntry[0].mapLength = 3

		assert(v.rtStream.GetWritePos() == streamPosBody).IsFalse()
		assert(v.rootFrame.stream == nil).IsFalse()
		assert(v.cacheEntry[0].mapLength == 0).IsFalse()
		v.Reset()
		assert(v.rtStream.GetWritePos() == streamPosBody).IsTrue()
		assert(v.rootFrame.stream == nil).IsTrue()
		assert(v.cacheEntry[0].mapLength == 0).IsTrue()
		v.Close()
	})
}

func TestRpcThread_Close(t *testing.T) {
	t.Run("close twice", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.Close()).IsTrue()
		assert(v.Close()).IsFalse()
	})

	t.Run("cannot close", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime, testThread bool) Return {
			if testThread {
				v := newThread(
					rt.thread.processor,
					3*time.Second,
					2048,
					fnEvalBack,
					fnEvalFinish,
				)
				s, _ := MakeRequestStream(true, 0, "#.test:Eval", "", false)
				v.PutStream(s)
				assert(v.Close()).IsFalse()
			} else {
				time.Sleep(3500 * time.Millisecond)
			}
			return rt.Reply(true)
		}, true)).Equal(true, nil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.Close()).IsTrue()
	})
}

func TestRpcThread_lock(t *testing.T) {
	t.Run("concurrent lock", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v Int64) Return {
				if v == 0 {
					return rt.Reply("OK")
				}

				wait := make(chan bool)

				for i := 0; i < 5; i++ {
					go func() {
						t := rt.thread.lock(rt.id)
						assert(t).IsNotNil()
						time.Sleep(100 * time.Millisecond)
						rt.thread.unlock(rt.id)
						wait <- true
					}()
				}

				for i := 0; i < 5; i++ {
					<-wait
				}

				return rt.Reply(rt.Call("#.test:Eval", v-1))
			}, 3)).Equal("OK", nil)
	})

	t.Run("lock failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v Int64) Return {
				if v == 0 {
					return rt.Reply("OK")
				}
				assert(rt.thread.lock(rt.id + 2)).IsNil()
				return rt.Reply(rt.Call("#.test:Eval", v-1))
			}, 10)).Equal("OK", nil)
	})

	t.Run("lock ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v Int64) Return {
				if v == 0 {
					return rt.Reply("OK")
				}
				assert(rt.thread.lock(rt.id)).IsNotNil()
				rt.thread.unlock(rt.id)
				return rt.Reply(rt.Call("#.test:Eval", v-1))
			}, 10)).Equal("OK", nil)
	})
}

func TestRpcThread_unlock(t *testing.T) {
	t.Run("unlock failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v Int64) Return {
				if v == 0 {
					return rt.Reply("OK")
				}
				assert(rt.thread.lock(rt.id)).IsNotNil()
				assert(rt.thread.unlock(rt.id + 2)).IsFalse()
				rt.thread.unlock(rt.id)
				return rt.Reply(rt.Call("#.test:Eval", v-1))
			}, 10)).Equal("OK", nil)
	})

	t.Run("unlock ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v Int64) Return {
				if v == 0 {
					return rt.Reply("OK")
				}
				assert(rt.thread.lock(rt.id)).IsNotNil()
				assert(rt.thread.unlock(rt.id)).IsTrue()
				return rt.Reply(rt.Call("#.test:Eval", v-1))
			}, 10)).Equal("OK", nil)
	})
}

func TestRpcThread_pushFrame(t *testing.T) {
	t.Run("unlock ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime) Return {
				curFrame := rt.thread.top
				rtArray := rt.NewRTArray(7)
				rtMap := rt.NewRTMap(7)
				_ = rtArray.Append("hello")
				_ = rtMap.Set("name", "kitty")

				parentRTWritePos := rt.thread.rtStream.GetWritePos()
				rt.thread.pushFrame()
				newFrame := rt.thread.top
				assert(newFrame.next).Equal(curFrame)
				assert(newFrame.cacheArrayItemsPos).Equal(uint32(7))
				assert(newFrame.cacheMapItemsPos).Equal(uint32(7))
				assert(newFrame.cacheArrayEntryPos).Equal(uint32(1))
				assert(newFrame.cacheMapEntryPos).Equal(uint32(1))
				assert(newFrame.parentRTWritePos).Equal(parentRTWritePos)
				rt.thread.popFrame()
				return rt.Reply("OK")
			})).Equal("OK", nil)
	})
}

func TestRpcThread_popFrame(t *testing.T) {
	t.Run("unlock ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime, v int64) Return {
				rtArray := rt.NewRTArray(7)
				rtMap := rt.NewRTMap(7)
				_ = rtArray.Append("hello")
				_ = rtMap.Set("name", "kitty")

				if v == 0 {
					return rt.Reply(true)
				}

				curFrame := rt.thread.top
				cacheArrayItemsPos := curFrame.cacheArrayItemsPos
				cacheMapItemsPos := curFrame.cacheMapItemsPos
				cacheArrayEntryPos := curFrame.cacheArrayEntryPos
				cacheMapEntryPos := curFrame.cacheMapEntryPos
				parentRTWritePos := rt.thread.rtStream.GetWritePos()
				ret := rt.Reply(rt.Call("#.test:Eval", v-1))
				assert(rt.thread.top).Equal(curFrame)
				assert(rt.thread.top.cacheArrayItemsPos).Equal(cacheArrayItemsPos)
				assert(rt.thread.top.cacheMapItemsPos).Equal(cacheMapItemsPos)
				assert(rt.thread.top.cacheArrayEntryPos).Equal(cacheArrayEntryPos)
				assert(rt.thread.top.cacheMapEntryPos).Equal(cacheMapEntryPos)
				assert(rt.thread.rtStream.GetWritePos()).Equal(parentRTWritePos + 1)
				return ret
			}, 10)).Equal(true, nil)
	})
}

func TestRpcThread_GetActionNode(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetActionNode()).Equal(nil)
		v.Close()
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			assert(rt.thread.GetActionNode()).IsNotNil()
			return rt.Reply(true)
		})).Equal(true, nil)
	})
}

func TestRpcThread_GetExecActionNodePath(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetExecActionNodePath()).Equal("")
		v.Close()
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			assert(rt.thread.GetExecActionNodePath()).Equal("#.test:Eval")
			return rt.Reply(true)
		})).Equal(true, nil)
	})
}

func TestRpcThread_GetExecActionDebug(t *testing.T) {
	t.Run("node is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 3*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.GetExecActionDebug()).Equal("")
		v.Close()
	})

	t.Run("node is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			(*rpcActionNode)(rt.thread.top.actionNode).meta.fileLine = "/file:001"
			assert(rt.thread.GetExecActionDebug()).Equal("#.test:Eval /file:001")
			return rt.Reply(true)
		})).Equal(true, nil)
	})
}

func TestRpcThread_Write(t *testing.T) {
	t.Run("value is endless loop", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			v := make(Map)
			v["v"] = v
			ret, s := thread.Write(v, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrUnsupportedValue.AddDebug(
				"value[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"]"+
					"[\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"][\"v\"] overflows").
				AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("value is not supported", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			ret, s := thread.Write(make(chan bool), 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrUnsupportedValue.
				AddDebug("value type(chan bool) is not supported").
				AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("value is (*Error)(nil)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			ret, s := thread.Write((*base.Error)(nil), 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrUnsupportedValue.AddDebug("value is nil").
				AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("value is RTValue (not available)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			ret, s := thread.Write(RTValue{}, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrUnsupportedValue.AddDebug("value is not available").
				AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("value is RTValue (with err)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			v := RTValue{err: errors.ErrStream}
			ret, s := thread.Write(v, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrStream.AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("reply has already benn called (debug true)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			thread.Write(true, 0, true)
			ret, s := thread.Write(true, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(
			nil,
			errors.ErrRuntimeReplyHasBeenCalled.AddDebug("#.test:Eval "+source),
		)
	})

	t.Run("reply has already benn called (debug false)", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			thread.Write(true, 0, false)
			return thread.Write(true, 0, false)
		})).Equal(
			nil,
			errors.ErrRuntimeReplyHasBeenCalled,
		)
	})

	t.Run("test ok (*Error, message is empty)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			e := errors.ErrUnsupportedValue
			ret, s := thread.Write(e, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(nil, errors.ErrUnsupportedValue.AddDebug("#.test:Eval "+source))
	})

	t.Run("test ok (*Error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			ret, s := thread.Write(errors.ErrStream, 0, true), base.GetFileLine(0)
			source = s
			return ret
		})).Equal(nil, errors.ErrStream.AddDebug("#.test:Eval "+source))
	})

	t.Run("test ok (int64)", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			thread := rt.thread
			return thread.Write(int64(12), 0, true)
		})).Equal(int64(12), nil)
	})
}

func TestRpcThread_PutStream(t *testing.T) {
	t.Run("thread is close", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, fnEvalFinish)
		v.Close()
		assert(v.PutStream(NewStream())).IsFalse()
	})

	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.PutStream(nil)).IsFalse()
		v.Close()
	})

	t.Run("thread has internal error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, fnEvalFinish)
		v.Close()
		atomic.StorePointer(&v.closeCH, unsafe.Pointer(v))
		assert(v.PutStream(NewStream())).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := newThread(testProcessor, 5*time.Second, 2048, fnEvalBack, fnEvalFinish)
		assert(v.PutStream(NewStream())).IsTrue()
		v.Close()
	})
}

func TestRpcThread_Eval(t *testing.T) {
	t.Run("action path type is not string", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetDepth(3)
		stream.WriteBytes([]byte("#.test:Eval"))
		stream.WriteString("")
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			return rt.Reply(true)
		}, stream)).Equal(nil, errors.ErrStream)
	})

	t.Run("action path is not mounted", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetDepth(3)
		stream.WriteString("#.system:Eval")
		stream.WriteString("")
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			return rt.Reply(true)
		}, stream)).Equal(
			nil,
			errors.ErrTargetNotExist.
				AddDebug("rpc-call: #.system:Eval does not exist"),
		)
	})

	t.Run("depth is overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetDepth(17)
		stream.WriteString("#.test:Eval")
		stream.WriteString("")
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			return rt.Reply(true)
		}, stream)).Equal(
			nil,
			errors.ErrCallOverflow.AddDebug("call #.test:Eval level(17) overflows"),
		)
	})

	t.Run("execFrom data format error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetDepth(3)
		stream.WriteString("#.test:Eval")
		stream.WriteBool(true)
		assert(testReply(true, nil, nil, func(rt Runtime) Return {
			return rt.Reply(true)
		}, stream)).Equal(nil, errors.ErrStream)
	})

	t.Run("call with all type value", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			assert(testReply(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"},
			)).Equal(true, nil)
		}
		fnTest(true, nil)
		fnTest(true, &testFuncCache{})
	})

	t.Run("1st param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				int64(3), int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 1st argument does not match. "+
							"want: rpc.Bool got: rpc.Int64",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("2nd param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, false, uint64(3), float64(3), "hello", []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 2nd argument does not match. "+
							"want: rpc.Int64 got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("3rd param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), true, float64(3), "hello", []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 3rd argument does not match. "+
							"want: rpc.Uint64 got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("4th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), true, "hello", []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 4th argument does not match. "+
							"want: rpc.Float64 got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("5th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), true, []byte("hello"),
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 5th argument does not match. "+
							"want: rpc.String got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("6th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", true,
				Array{1}, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 6th argument does not match. "+
							"want: rpc.Bytes got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("7th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				true, Map{"name": "kitty"}, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 7th argument does not match. "+
							"want: rpc.Array got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("8th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{2}, true, true, Array{2}, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 8th argument does not match. "+
							"want: rpc.Map got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("9th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			sendStream, _ := MakeRequestStream(
				dbg, 0, "#.test:Eval", "",
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{2}, Map{"name": "doggy"},
			)
			sendStream.PutBytes([]byte{0})
			sendStream.Write(Array{2})
			sendStream.Write(Map{"name": "doggy"})

			assert(testReply(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				}, sendStream)).
				Equal(nil, errors.ErrStream)
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("10th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{2}, Map{"name": "doggy"}, true, false, Map{"name": "doggy"})

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 10th argument does not match. "+
							"want: rpc.RTArray got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("11th param error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				},
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{2}, Map{"name": "doggy"}, true, Array{2}, false)

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 11th argument does not match. "+
							"want: rpc.RTMap got: rpc.Bool",
					).AddDebug("#.test:Eval "+source))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("arguments length not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			sendStream, _ := MakeRequestStream(
				dbg, 0, "#.test:Eval", "",
				true, int64(3), uint64(3), float64(3), "hello", []byte("hello"),
				Array{2}, Map{"name": "doggy"}, true, Array{2}, Map{"name": "doggy"},
				false,
			)
			sendStream.SetWritePos(sendStream.GetWritePos() + 1)
			assert(testReply(dbg, fnCache, nil,
				func(rt Runtime, b Bool, i Int64, u Uint64, f Float64, s String,
					x Bytes, a Array, m Map, v RTValue, y RTArray, z RTMap) Return {
					return rt.Reply(true)
				}, sendStream)).
				Equal(nil, errors.ErrStream)
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("unsupported type", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnTest := func(dbg bool, fnCache ActionCache) {
			source1 := ""
			source2 := ""
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime, hook Bool) Return {
					actionNode := rt.thread.GetActionNode()
					actionNode.argTypes[1] = reflect.ValueOf(int16(0)).Type()
					rtValue, s1 := rt.Call("#.test:Eval", false), base.GetFileLine(0)
					ret, s2 := rt.Reply(rtValue), base.GetFileLine(0)
					source1 = s1
					source2 = s2
					return ret
				}, true)

			if dbg {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval 1st argument does not match. "+
							"want: int16 got: rpc.Bool",
					).AddDebug("#.test:Eval "+source).
						AddDebug("#.test:Eval "+source1).
						AddDebug("#.test:Eval "+source2))
			} else {
				assert(ParseResponseStream(stream)).
					Equal(nil, errors.ErrArgumentsNotMatch.AddDebug(
						"rpc-call: #.test:Eval arguments does not match",
					).AddDebug("#.test:Eval "+source1).AddDebug("#.test:Eval "+source2))
			}
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("call function error", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(dbg bool, fnCache ActionCache) {
			stream, source := testReplyWithSource(dbg, fnCache, nil,
				func(rt Runtime) Return {
					panic("error")
				})
			ret, err := ParseResponseStream(stream)
			expectErr := errors.ErrActionPanic.
				AddDebug("runtime error: error").
				AddDebug("#.test:Eval " + source)

			assert(ret).Equal(nil)
			assert(err).IsNotNil()
			assert(err.GetCode()).Equal(expectErr.GetCode())
			assert(strings.HasPrefix(
				err.GetMessage(),
				expectErr.GetMessage(),
			)).IsTrue()
			assert(strings.Contains(err.GetMessage(), "[running]:")).IsTrue()
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("onEvalFinish panic", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnTest := func(dbg bool, fnCache ActionCache) {
			processor, _ := NewProcessor(
				1,
				16,
				16,
				2048,
				fnCache,
				3*time.Second,
				[]*ServiceMeta{{
					name: "test",
					service: NewService().On("Eval", func(rt Runtime) Return {
						return rt.Reply(true)
					}),
					fileLine: "",
					data:     nil,
				}},
				func(stream *Stream) {
					panic("error")
				},
			)
			defer processor.Close()
			sendStream, _ := MakeRequestStream(dbg, 0, "#.test:Eval", "")
			processor.PutStream(sendStream)

			errCH := make(chan *base.Error)
			subscription := base.SubscribePanic(func(e *base.Error) {
				errCH <- e
			})
			defer subscription.Close()

			err := <-errCH
			expectErr := errors.ErrThreadEvalFatal.AddDebug("error")
			assert(err).IsNotNil()
			assert(strings.HasPrefix(
				err.GetMessage(),
				expectErr.GetMessage(),
			)).IsTrue()
			assert(strings.Contains(err.GetMessage(), "[running]:")).IsTrue()
		}
		fnTest(true, nil)
		fnTest(false, nil)
		fnTest(true, &testFuncCache{})
		fnTest(false, &testFuncCache{})
	})

	t.Run("return without runtime", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream, source := testReplyWithSource(true, nil, nil,
			func(rt Runtime) Return {
				return emptyReturn
			})
		assert(ParseResponseStream(stream)).Equal(
			nil,
			errors.ErrRuntimeExternalReturn.AddDebug("#.test:Eval "+source))
	})

	t.Run("thread is locked for a while", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testReply(true, nil, nil,
			func(rt Runtime) Return {
				wait := make(chan bool)
				ret := rt.Reply(true)
				go func() {
					rt.lock()
					wait <- true
					time.Sleep(500 * time.Millisecond)
					rt.unlock()
				}()
				<-wait
				return ret
			}),
		).Equal(true, nil)
	})
}

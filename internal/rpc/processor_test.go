package rpc

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

var (
	testProcessor = NewProcessor(
		1,
		32,
		32,
		2048,
		nil,
		5*time.Second,
		nil,
		NewTestStreamReceiver(),
	)
)

func init() {
	testProcessor.Close()
}

func testProcessorMountError(services []*ServiceMeta) *base.Error {
	streamReceiver := NewTestStreamReceiver()
	processor := NewProcessor(
		freeGroups,
		2,
		3,
		2048,
		nil,
		time.Second,
		services,
		streamReceiver,
	)
	defer processor.Close()

	if stream := streamReceiver.GetStream(); stream != nil {
		_, err := ParseResponseStream(stream)
		return err
	}

	return nil
}

func TestRpcActionNode_GetConfig(t *testing.T) {
	t.Run("data is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{}
		assert(v.GetConfig("name")).Equals(nil, false)
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 18}}
		assert(v.GetConfig("name")).Equals(nil, false)
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 18}}
		assert(v.GetConfig("age")).Equals(18, true)
	})
}

func TestRpcActionNode_SetConfig(t *testing.T) {
	t.Run("data is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{}
		v.SetConfig("age", 3)
		assert(v.data).Equals(nil)
	})

	t.Run("key does not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{}}
		v.SetConfig("age", 3)
		assert(v.data).Equals(Map{"age": 3})
	})

	t.Run("key exists", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &rpcServiceNode{data: Map{"age": 5}}
		v.SetConfig("age", 3)
		assert(v.data).Equals(Map{"age": 3})
	})
}

func TestProcessor(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		emptyEvalBack(nil)
		emptyEvalFinish(nil)
		assert(actionNameRegex.MatchString("$onMount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUnmount")).IsTrue()
		assert(actionNameRegex.MatchString("$onUpdateConfig")).IsTrue()
		assert(actionNameRegex.MatchString("onMount")).IsTrue()
		assert(actionNameRegex.MatchString("sayHello")).IsTrue()
		assert(actionNameRegex.MatchString("$sayHello")).IsFalse()
		assert(rootName).Equals("#")
		assert(freeGroups).Equals(256)
		assert(processorStatusClosed).Equals(0)
		assert(processorStatusRunning).Equals(1)
	})
}

func TestNewProcessor(t *testing.T) {
	t.Run("streamReceiver is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewProcessor(
				256, 16, 16, 2048, nil, 5*time.Second, nil, nil,
			)
		})).Equals("streamReceiver is nil")
	})

	t.Run("numOfThreads <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		assert(NewProcessor(
			0, 16, 16, 2048, nil, 5*time.Second, nil, streamReceiver,
		)).Equals(nil)
		assert(ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrNumOfThreadsIsWrong)
	})

	t.Run("maxNodeDepth <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		assert(NewProcessor(
			256, 0, 16, 2048, nil, 5*time.Second, nil, streamReceiver,
		)).Equals(nil)
		assert(ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrMaxNodeDepthIsWrong)
	})

	t.Run("maxCallDepth <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		assert(NewProcessor(
			256, 16, 0, 2048, nil, 5*time.Second, nil, streamReceiver,
		)).Equals(nil)
		assert(ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrProcessorMaxCallDepthIsWrong)
	})

	t.Run("mount service error", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		assert(NewProcessor(
			256, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{nil}, streamReceiver,
		)).Equals(nil)
		assert(ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrProcessorNodeMetaIsNil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			NewTestStreamReceiver(),
		)
		assert(processor).IsNotNil()
		assert(len(processor.threads)).Equals(freeGroups)
		_ = processor.Close()
	})

	t.Run("test ok (subscribe error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		processor := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			streamReceiver,
		)
		assert(processor).IsNotNil()
		base.PublishPanic(base.ErrStream)
		assert(ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrStream)
		_ = processor.Close()
	})

	t.Run("test ok (system action)", func(t *testing.T) {
		assert := base.NewAssert(t)
		wait := make(chan string, 3)
		service := NewService().
			On("$onMount", func(rt Runtime) Return {
				wait <- "$onMount called"
				return rt.Reply(true)
			}).
			On("$onUpdateConfig", func(rt Runtime) Return {
				wait <- "$onUpdateConfig called"
				return rt.Reply(true)
			}).
			On("$onUnmount", func(rt Runtime) Return {
				wait <- "$onUnmount called"
				return rt.Reply(true)
			})
		processor := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  service,
				fileLine: "",
			}},
			NewTestStreamReceiver(),
		)
		assert(processor).IsNotNil()
		assert(<-wait).Equals("$onMount called")
		assert(<-wait).Equals("$onUpdateConfig called")
		processor.Close()
		assert(<-wait).Equals("$onUnmount called")
	})

	t.Run("test ok (10K calls)", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := NewTestStreamReceiver()
		service := NewService().
			On("Eval", func(rt Runtime) Return {
				return rt.Reply(true)
			})
		processor := NewProcessor(
			1, 16, 16, 2048, nil, 5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  service,
				fileLine: "",
			}},
			streamReceiver,
		)

		go func() {
			for i := 0; i < 10000; i++ {
				stream, _ := MakeInternalRequestStream(
					true, 0, "#.test:Eval", "",
				)
				processor.PutStream(stream)
			}
		}()

		// wait for finish
		for streamReceiver.TotalStreams() < 10000 {
			time.Sleep(10 * time.Millisecond)
		}

		for i := 0; i < 10000; i++ {
			assert(ParseResponseStream(streamReceiver.GetStream())).
				Equals(true, nil)
		}

		processor.Close()
	})
}

func TestProcessor_Close(t *testing.T) {
	t.Run("processor is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := NewProcessor(
			1,
			32,
			32,
			2048,
			nil,
			5*time.Second,
			nil,
			NewTestStreamReceiver(),
		)
		processor.Close()
		assert(processor.Close()).Equals(false)
	})

	t.Run("close timeout", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnTest := func(count int) {
			mutex := &sync.Mutex{}
			source := ""
			waitCH := make(chan bool)
			streamReceiver := NewTestStreamReceiver()
			processor := NewProcessor(
				256,
				2,
				3,
				2048,
				nil,
				time.Second,
				[]*ServiceMeta{{
					name: "test",
					service: NewService().On("Eval", func(rt Runtime) Return {
						waitCH <- true
						mutex.Lock()
						source = rt.thread.GetExecActionDebug()
						mutex.Unlock()
						time.Sleep(3 * time.Second)
						return rt.Reply(true)
					}),
					fileLine: "",
				}},
				streamReceiver,
			)

			for i := 0; i < count; i++ {
				stream, _ := MakeInternalRequestStream(
					true, 0, "#.test:Eval", "",
				)
				processor.PutStream(stream)
				<-waitCH
			}

			assert(processor.Close()).IsFalse()

			mutex.Lock()
			if count == 1 {
				assert(ParseResponseStream(streamReceiver.GetStream())).Equals(
					nil,
					base.ErrActionCloseTimeout.AddDebug(fmt.Sprintf(
						"the following actions can not close: \n\t%s "+
							"(1 goroutine)",
						source,
					)).Standardize(),
				)
			} else {
				assert(ParseResponseStream(streamReceiver.GetStream())).Equals(
					nil,
					base.ErrActionCloseTimeout.AddDebug(fmt.Sprintf(
						"the following actions can not close: \n\t%s "+
							"(%d goroutines)",
						source, count,
					)).Standardize(),
				)
			}
			mutex.Unlock()
		}

		fnTest(1)
		fnTest(100)
	})
}

func TestProcessor_PutStream(t *testing.T) {
	t.Run("processor is closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := NewProcessor(
			256,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			NewTestStreamReceiver(),
		)
		processor.Close()

		for i := 0; i < 2048; i++ {
			assert(processor.PutStream(NewStream())).IsFalse()
		}
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testCount := 10240
		streamReceiver := NewTestStreamReceiver()
		processor := NewProcessor(
			256,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			streamReceiver,
		)

		defer processor.Close()

		for i := 0; i < testCount; i++ {
			assert(processor.PutStream(NewStream())).IsTrue()
		}

		for streamReceiver.TotalStreams() < testCount {
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(100 * time.Millisecond)

		freeSum := 0
		for i := 0; i < len(processor.freeCHArray); i++ {
			freeSum += len(processor.freeCHArray[i])
		}
		assert(freeSum).Equals(256)
	})
}

func TestProcessor_BuildCache(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	curDir := path.Dir(file)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
	}()

	t.Run("services is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		tmpFile := path.Join(curDir, "_tmp_/test-processor-01.go")
		snapshotFile := path.Join(
			curDir,
			"_snapshot_/test-processor-01.snapshot",
		)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			NewTestStreamReceiver(),
		)
		defer processor.Close()
		assert(processor.BuildCache("pkgName", tmpFile)).IsNil()
		assert(base.ReadFromFile(tmpFile)).
			Equals(base.ReadFromFile(snapshotFile))
	})

	t.Run("service is not empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		tmpFile := path.Join(curDir, "_tmp_/test-processor-02.go")
		snapshotFile := path.Join(
			curDir,
			"_snapshot_/test-processor-02.snapshot",
		)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().On("Eval", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "",
			}},
			NewTestStreamReceiver(),
		)
		defer processor.Close()
		assert(processor.BuildCache("pkgName", tmpFile)).IsNil()
		assert(base.ReadFromFile(tmpFile)).
			Equals(base.ReadFromFile(snapshotFile))
	})

	t.Run("processor is closed", func(t *testing.T) {
		assert := base.NewAssert(t)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			nil,
			NewTestStreamReceiver(),
		)
		processor.Close()
		assert(processor.BuildCache("pkgName", "")).
			Equals(base.ErrProcessorIsNotRunning)
	})
}

func TestProcessor_onUpdateConfig(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool, 2)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			[]*ServiceMeta{{
				name: "test1",
				service: NewService().On(
					"$onUpdateConfig", func(rt Runtime) Return {
						waitCH <- true
						return rt.Reply(true)
					},
				),
				fileLine: "",
			}, {
				name: "test2",
				service: NewService().On(
					"$onUpdateConfig", func(rt Runtime) Return {
						waitCH <- true
						return rt.Reply(true)
					},
				),
				fileLine: "",
			}},
			NewTestStreamReceiver(),
		)
		processor.onUpdateConfig()
		assert(<-waitCH).Equals(true)
		assert(<-waitCH).Equals(true)
		processor.Close()
	})
}

func TestProcessor_invokeSystemAction(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool, 3)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			[]*ServiceMeta{{
				name: "test",
				service: NewService().
					On("$onMount", func(rt Runtime) Return {
						waitCH <- true
						return rt.Reply(true)
					}).
					On("$onUpdateConfig", func(rt Runtime) Return {
						waitCH <- true
						return rt.Reply(true)
					}).
					On("$onUnmount", func(rt Runtime) Return {
						waitCH <- true
						return rt.Reply(true)
					}),
				fileLine: "",
			}},
			NewTestStreamReceiver(),
		)

		// for default onMount
		assert(<-waitCH).Equals(true)
		assert(processor.invokeSystemAction("#.test", "$onMount")).IsTrue()
		assert(processor.invokeSystemAction("#.test", "$onUpdateConfig")).
			IsTrue()
		assert(processor.invokeSystemAction("#.test", "$onUnmount")).IsTrue()
		assert(<-waitCH).Equals(true)
		assert(<-waitCH).Equals(true)
		assert(<-waitCH).Equals(true)
		processor.Close()
	})
}

func TestProcessor_mountNode(t *testing.T) {
	t.Run("nodeMeta is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{
			nil,
		})).Equals(base.ErrProcessorNodeMetaIsNil)
	})

	t.Run("service name is illegal", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name:     "+",
			service:  NewService(),
			fileLine: "dbg",
		}})).Equals(base.ErrServiceName.
			AddDebug("service name + is illegal").
			AddDebug("dbg").
			Standardize(),
		)
	})

	t.Run("service is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name:     "abc",
			service:  nil,
			fileLine: "dbg",
		}})).Equals(base.ErrServiceIsNil.AddDebug("dbg").Standardize())
	})

	t.Run("depth overflows", func(t *testing.T) {
		assert := base.NewAssert(t)
		embedService, source := NewService().
			AddChildService("s", NewService(), nil), base.GetFileLine(0)
		assert(testProcessorMountError([]*ServiceMeta{{
			name:     "s",
			service:  NewService().AddChildService("s", embedService, nil),
			fileLine: "dbg",
		}})).Equals(base.ErrServiceOverflow.AddDebug(
			"service path #.s.s.s overflows (max depth: 2, current depth:3)",
		).AddDebug(source).Standardize())
	})

	t.Run("duplicated service name", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name:     "user",
			service:  NewService(),
			fileLine: "Debug1",
		}, {
			name:     "user",
			service:  NewService(),
			fileLine: "Debug2",
		}})).Equals(base.ErrServiceName.
			AddDebug("duplicated service name user").
			AddDebug("current:\n\tDebug2\nconflict:\n\tDebug1").
			Standardize(),
		)
	})

	t.Run("duplicated service name", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name:     "user",
			service:  NewService(),
			fileLine: "dbg1",
		}, {
			name:     "user",
			service:  NewService(),
			fileLine: "dbg2",
		}})).Equals(base.ErrServiceName.
			AddDebug("duplicated service name user").
			AddDebug("current:\n\tdbg2\nconflict:\n\tdbg1").
			Standardize(),
		)
	})

	t.Run("mount actions error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions:  []*ActionMeta{nil},
			},
			fileLine: "",
		}})).Equals(base.ErrProcessorActionMetaIsNil)
	})

	t.Run("mount children error", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{nil},
				actions:  []*ActionMeta{},
			},
			fileLine: "",
		}})).Equals(base.ErrProcessorNodeMetaIsNil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			nil,
			time.Second,
			[]*ServiceMeta{{
				name: "user",
				service: NewService().On("Login", func(rt Runtime) Return {
					return rt.Reply(true)
				}),
				fileLine: "dbg",
				data:     Map{"name": "kitty", "age": 18},
			}},
			NewTestStreamReceiver(),
		)
		assert(processor).IsNotNil()
		assert(processor.servicesMap["#.user"]).Equals(&rpcServiceNode{
			path: "#.user",
			addMeta: &ServiceMeta{
				name:     "user",
				service:  processor.servicesMap["#.user"].addMeta.service,
				fileLine: "dbg",
				data:     Map{"name": "kitty", "age": 18},
			},
			depth:   1,
			data:    Map{"name": "kitty", "age": 18},
			isMount: true,
		})
	})
}

func TestProcessor_mountAction(t *testing.T) {
	t.Run("meta is nil", func(t *testing.T) {
		assert := base.NewAssert(t)

		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions:  []*ActionMeta{nil},
			},
			fileLine: "nodeDebug",
		}})).Equals(base.ErrProcessorActionMetaIsNil)
	})

	t.Run("name is illegal", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "+",
					handler:  func(rt Runtime) Return { return rt.Reply(true) },
					fileLine: "actionDebug",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionName.
				AddDebug("action name + is illegal").
				AddDebug("actionDebug").
				Standardize(),
		)
	})

	t.Run("handler is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "login",
					handler:  nil,
					fileLine: "actionDebug",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionHandler.
				AddDebug("handler is nil").
				AddDebug("actionDebug").
				Standardize(),
		)
	})

	t.Run("handler is not function", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "login",
					handler:  3,
					fileLine: "actionDebug",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionHandler.AddDebug(
				"handler must be func(rt rpc.Runtime, ...) rpc.Return",
			).AddDebug("actionDebug").Standardize(),
		)
	})

	t.Run("handler is error 1", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "login",
					handler:  func() {},
					fileLine: "actionDebug",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionHandler.
				AddDebug("handler 1st argument type must be rpc.Runtime").
				AddDebug("actionDebug").
				Standardize(),
		)
	})

	t.Run("handler is error 2", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "login",
					handler:  func(rt Runtime) {},
					fileLine: "actionDebug",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionHandler.
				AddDebug("handler return type must be rpc.Return").
				AddDebug("actionDebug").
				Standardize(),
		)
	})

	t.Run("duplicated name", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testProcessorMountError([]*ServiceMeta{{
			name: "user",
			service: &Service{
				children: []*ServiceMeta{},
				actions: []*ActionMeta{{
					name:     "login",
					handler:  func(rt Runtime) Return { return rt.Reply(true) },
					fileLine: "actionDebug1",
				}, {
					name:     "login",
					handler:  func(rt Runtime) Return { return rt.Reply(true) },
					fileLine: "actionDebug2",
				}},
			},
			fileLine: "nodeDebug",
		}})).Equals(
			base.ErrActionName.
				AddDebug("duplicated action name login").
				AddDebug("current:\n\tactionDebug2\nconflict:\n\tactionDebug1").
				Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		handler := func(rt Runtime, v Bool) Return { return rt.Reply(v) }
		fnCache := &testFuncCache{}

		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			fnCache,
			time.Second,
			[]*ServiceMeta{{
				name: "user",
				service: &Service{
					children: []*ServiceMeta{},
					actions: []*ActionMeta{{
						name:     "login",
						handler:  handler,
						fileLine: "actionDebug",
					}},
				},
				fileLine: "nodeDebug",
			}},
			NewTestStreamReceiver(),
		)
		assert(processor).IsNotNil()
		assert(processor.actionsMap["#.user:login"]).Equals(&rpcActionNode{
			path:       "#.user:login",
			meta:       processor.actionsMap["#.user:login"].meta,
			service:    processor.servicesMap["#.user"],
			cacheFN:    fnCache.Get("B"),
			reflectFn:  reflect.ValueOf(handler),
			callString: "#.user:login(rpc.Runtime, rpc.Bool) rpc.Return",
			argTypes:   []reflect.Type{runtimeType, boolType},
			indicator:  processor.actionsMap["#.user:login"].indicator,
		})
		processor.Close()
	})
}

func TestProcessor_unmount(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		handler := func(rt Runtime, v Bool) Return { return rt.Reply(v) }
		fnCache := &testFuncCache{}
		waitCH := make(chan string, 2)
		processor := NewProcessor(
			freeGroups,
			2,
			3,
			2048,
			fnCache,
			time.Second,
			[]*ServiceMeta{{
				name: "test1",
				service: NewService().
					On("action", handler).
					On("$onUnmount", func(rt Runtime) Return {
						waitCH <- "test1"
						return rt.Reply(true)
					}),
				fileLine: "nodeDebug",
			}, {
				name: "test2",
				service: NewService().
					On("action", handler).
					On("$onUnmount", func(rt Runtime) Return {
						waitCH <- "test2"
						return rt.Reply(true)
					}),
				fileLine: "nodeDebug",
			}, {
				name: "test3",
				service: NewService().
					On("action", handler).
					On("$onUnmount", func(rt Runtime) Return {
						waitCH <- "test3"
						return rt.Reply(true)
					}),
				fileLine: "nodeDebug",
			}},
			NewTestStreamReceiver(),
		)

		processor.unmount("#.test1")
		assert(<-waitCH).Equals("test1")
		assert(len(processor.servicesMap)).Equals(3)
		assert(len(processor.actionsMap)).Equals(4)
		processor.unmount("#.test2")
		assert(<-waitCH).Equals("test2")
		assert(len(processor.servicesMap)).Equals(2)
		assert(len(processor.actionsMap)).Equals(2)
		processor.unmount("#")
		assert(<-waitCH).Equals("test3")
		assert(len(processor.servicesMap)).Equals(0)
		assert(len(processor.actionsMap)).Equals(0)
		processor.Close()
	})
}

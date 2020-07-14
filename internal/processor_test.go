package internal

import (
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func TestNewRPCProcessor(t *testing.T) {
	assert := NewRPCAssert(t)

	callbackFn := func(stream *RPCStream, success bool) {}

	processor := NewRPCProcessor(16, 32, callbackFn, nil)
	assert(processor).IsNotNil()
	assert(processor.isRunning).IsFalse()
	assert(processor.callback).IsNotNil()
	assert(len(processor.echosMap)).Equals(0)
	assert(len(processor.nodesMap)).Equals(1)
	assert(processor.threadPools).IsNotNil()
	lenThreadPool := len(processor.threadPools)
	capThreadPool := cap(processor.threadPools)
	assert(lenThreadPool >= numOfMinThreadPool).IsTrue()
	assert(lenThreadPool <= numOfMaxThreadPool).IsTrue()
	assert(capThreadPool == lenThreadPool).IsTrue()
	for i := 0; i < lenThreadPool; i++ {
		assert(processor.threadPools[i]).IsNil()
	}
	assert(processor.maxNodeDepth).Equals(uint64(16))
	assert(processor.maxCallDepth).Equals(uint64(32))
	processor.Stop()

	// save fnGetRuntimeNumberOfCPU
	var oldFnGetRuntimeNumberOfCPU = fnGetRuntimeNumberOfCPU

	// mock fnGetRuntimeNumberOfCPU to zero
	fnGetRuntimeNumberOfCPU = func() int {
		return 0
	}
	processor1 := NewRPCProcessor(16, 32, callbackFn, nil)
	assert(len(processor1.threadPools)).Equals(numOfMinThreadPool)
	processor1.Stop()

	// mock fnGetRuntimeNumberOfCPU to max
	fnGetRuntimeNumberOfCPU = func() int {
		return 999999999
	}
	processor2 := NewRPCProcessor(16, 32, callbackFn, nil)
	assert(len(processor2.threadPools)).Equals(numOfMaxThreadPool)
	processor2.Stop()

	// restore fnGetRuntimeNumberOfCPU
	fnGetRuntimeNumberOfCPU = oldFnGetRuntimeNumberOfCPU
}

func TestRPCProcessor_Start_Stop(t *testing.T) {
	assert := NewRPCAssert(t)

	processor := NewRPCProcessor(16, 32, nil, nil)
	assert(processor.Stop()).IsNotNil()
	assert(processor.isRunning).IsFalse()
	for i := 0; i < len(processor.threadPools); i++ {
		assert(processor.threadPools[i]).IsNil()
	}
	assert(processor.Start()).IsTrue()
	assert(processor.isRunning).IsTrue()
	for i := 0; i < len(processor.threadPools); i++ {
		assert(processor.threadPools[i]).IsNotNil()
	}
	assert(processor.Start()).IsFalse()
	assert(processor.isRunning).IsTrue()
	for i := 0; i < len(processor.threadPools); i++ {
		assert(processor.threadPools[i]).IsNotNil()
	}
	assert(processor.Stop()).IsNil()
	assert(processor.isRunning).IsFalse()
	for i := 0; i < len(processor.threadPools); i++ {
		assert(processor.threadPools[i]).IsNil()
	}
	assert(processor.Stop()).IsNotNil()
	assert(processor.isRunning).IsFalse()
	for i := 0; i < len(processor.threadPools); i++ {
		assert(processor.threadPools[i]).IsNil()
	}
}

func TestRPCProcessor_PutStream(t *testing.T) {
	assert := NewRPCAssert(t)
	processor := NewRPCProcessor(16, 32, nil, nil)
	assert(processor.PutStream(NewRPCStream())).IsFalse()
	processor.Start()
	assert(processor.PutStream(NewRPCStream())).IsTrue()
	for i := 0; i < len(processor.threadPools); i++ {
		processor.threadPools[i].stop()
	}
	assert(processor.PutStream(NewRPCStream())).IsFalse()
}

func TestRPCProcessor_AddService(t *testing.T) {
	assert := NewRPCAssert(t)

	processor := NewRPCProcessor(16, 32, nil, nil)
	assert(processor.AddService("test", nil, "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	service := NewRPCService()
	assert(processor.AddService("test", service, "")).IsNil()
}

func TestRPCProcessor_BuildCache(t *testing.T) {
	assert := NewRPCAssert(t)
	_, file, _, _ := runtime.Caller(0)

	processor0 := NewRPCProcessor(16, 32, nil, nil)
	assert(processor0.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-0.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-0.go")))

	processor1 := NewRPCProcessor(16, 32, nil, nil)
	_ = processor1.AddService("abc", NewRPCService().
		Echo("sayHello", true, func(ctx *RPCContext, name string) *RPCReturn {
			return ctx.OK("hello " + name)
		}), "")
	assert(processor1.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/processor-build-cache-1.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/processor-build-cache-1.go")))

	_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
}

func TestRPCProcessor_mountNode(t *testing.T) {
	assert := NewRPCAssert(t)

	processor := NewRPCProcessor(16, 16, nil, nil)

	assert(processor.mountNode(rootName, nil).GetMessage()).
		Equals("rpc: mountNode: nodeMeta is nil")
	assert(processor.mountNode(rootName, nil).GetDebug()).
		Equals("")

	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "+",
		serviceMeta: NewRPCService().(*rpcService),
		debug:       "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service name \"+\" is illegal",
		"DebugMessage",
	))

	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "abc",
		serviceMeta: nil,
		debug:       "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service is nil",
		"DebugMessage",
	))

	assert(processor.mountNode("123", &rpcNodeMeta{
		name:        "abc",
		serviceMeta: NewRPCService().(*rpcService),
		debug:       "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"rpc: mountNode: parentNode is nil",
		"DebugMessage",
	))

	processor.maxNodeDepth = 0
	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "abc",
		serviceMeta: NewRPCService().(*rpcService),
		debug:       "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service path depth $.abc is too long, it must be less or equal than 0",
		"DebugMessage",
	))
	processor.maxNodeDepth = 16

	_ = processor.mountNode(rootName, &rpcNodeMeta{
		name:        "abc",
		serviceMeta: NewRPCService().(*rpcService),
		debug:       "DebugMessage",
	})
	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "abc",
		serviceMeta: NewRPCService().(*rpcService),
		debug:       "DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Service name \"abc\" is duplicated",
		"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
	))

	// mount echo error
	service := NewRPCService()
	service.Echo("abc", true, nil)
	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "test",
		serviceMeta: service.(*rpcService),
		debug:       "DebugMessage",
	}).GetMessage()).Equals("Echo handler is nil")

	// mount children error
	service1 := NewRPCService()
	service1.AddService("abc", NewRPCService())
	assert(len(service1.(*rpcService).children)).Equals(1)
	service1.(*rpcService).children[0] = nil
	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "003",
		serviceMeta: service1.(*rpcService),
		debug:       "DebugMessage",
	}).GetMessage()).Equals("rpc: mountNode: nodeMeta is nil")

	// OK
	service2 := NewRPCService()
	service2.AddService("user", NewRPCService().
		Echo("sayHello", true, func(ctx *RPCContext) *RPCReturn {
			return ctx.OK(true)
		}))
	assert(processor.mountNode(rootName, &rpcNodeMeta{
		name:        "system",
		serviceMeta: service2.(*rpcService),
		debug:       "DebugMessage",
	})).IsNil()
}

func TestRPCProcessor_mountEcho(t *testing.T) {
	assert := NewRPCAssert(t)

	processor := NewRPCProcessor(16, 16, nil, &TestFuncCache{})
	rootNode := processor.nodesMap[rootName]

	// check the node is nil
	assert(processor.mountEcho(nil, nil)).Equals(NewRPCErrorByDebug(
		"rpc: mountEcho: node is nil",
		"",
	))

	// check the echoMeta is nil
	assert(processor.mountEcho(rootNode, nil)).Equals(NewRPCErrorByDebug(
		"rpc: mountEcho: echoMeta is nil",
		"",
	))

	// check the name
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"###",
		true,
		nil,
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo name ### is illegal",
		"DebugMessage",
	))

	// check the echo path is not occupied
	_ = processor.mountEcho(rootNode, &rpcEchoMeta{
		"testOccupied",
		true,
		func(ctx *RPCContext) *RPCReturn { return ctx.OK(true) },
		"DebugMessage",
	})
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testOccupied",
		true,
		func(ctx *RPCContext) *RPCReturn { return ctx.OK(true) },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo name testOccupied is duplicated",
		"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
	))

	// check the echo handler is nil
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerIsNil",
		true,
		nil,
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler is nil",
		"DebugMessage",
	))

	// Check echo handler is Func
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerIsFunction",
		true,
		make(chan bool),
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler must be func(ctx rpc.Context, ...) rpc.Return",
		"DebugMessage",
	))

	// Check echo handler arguments types
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerArguments",
		true,
		func(ctx bool) *RPCReturn { return nilReturn },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler 1st argument type must be rpc.Context",
		"DebugMessage",
	))

	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerArguments",
		true,
		func(ctx *RPCContext, ch chan bool) *RPCReturn { return nilReturn },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler 2nd argument type <chan bool> not supported",
		"DebugMessage",
	))

	// Check return type
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerReturn",
		true,
		func(ctx *RPCContext) (*RPCReturn, bool) { return nilReturn, true },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler return type must be rpc.Return",
		"DebugMessage",
	))

	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testEchoHandlerReturn",
		true,
		func(ctx RPCContext) bool { return true },
		"DebugMessage",
	})).Equals(NewRPCErrorByDebug(
		"Echo handler return type must be rpc.Return",
		"DebugMessage",
	))

	// ok
	assert(processor.mountEcho(rootNode, &rpcEchoMeta{
		"testOK",
		true,
		func(ctx *RPCContext, _ bool, _ RPCMap) *RPCReturn { return nilReturn },
		GetStackString(0),
	})).IsNil()

	assert(processor.echosMap["$:testOK"].serviceNode).
		Equals(processor.nodesMap[rootName])
	assert(processor.echosMap["$:testOK"].path).Equals("$:testOK")
	assert(processor.echosMap["$:testOK"].echoMeta.name).Equals("testOK")
	assert(processor.echosMap["$:testOK"].reflectFn).IsNotNil()
	assert(processor.echosMap["$:testOK"].callString).
		Equals("$:testOK(rpc.Context, rpc.Bool, rpc.Map) rpc.Return")
	assert(
		strings.Contains(processor.echosMap["$:testOK"].debugString, "$:testOK"),
	).IsTrue()
	assert(processor.echosMap["$:testOK"].argTypes[0]).
		Equals(reflect.ValueOf(nilContext).Type())
	assert(processor.echosMap["$:testOK"].argTypes[1]).Equals(boolType)
	assert(processor.echosMap["$:testOK"].argTypes[2]).Equals(mapType)
	assert(processor.echosMap["$:testOK"].indicator).IsNotNil()
}

func TestRPCProcessor_OutPutErrors(t *testing.T) {
	assert := NewRPCAssert(t)

	processor := NewRPCProcessor(16, 16, nil, nil)

	// Service is nil
	assert(processor.AddService("", nil, "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	assert(processor.AddService("abc", (*rpcService)(nil), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service is nil",
			"DebugMessage",
		))

	// Service name %s is illegal
	assert(processor.AddService("\"\"", NewRPCService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service name \"\"\"\" is illegal",
			"DebugMessage",
		))

	processor.maxNodeDepth = 0
	assert(processor.AddService("abc", NewRPCService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service path depth $.abc is too long, it must be less or equal than 0",
			"DebugMessage",
		))
	processor.maxNodeDepth = 16

	_ = processor.AddService("abc", NewRPCService(), "DebugMessage")
	assert(processor.AddService("abc", NewRPCService(), "DebugMessage")).
		Equals(NewRPCErrorByDebug(
			"Service name \"abc\" is duplicated",
			"Current:\n\tDebugMessage\nConflict:\n\tDebugMessage",
		))
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	processor := NewRPCProcessor(
		16,
		16,
		func(stream *RPCStream, success bool) {
			stream.Release()
		},
		&TestFuncCache{},
	)
	processor.Start()
	_ = processor.AddService(
		"user",
		NewRPCService().
			Echo("sayHello", true, func(
				ctx *RPCContext,
				name string,
			) *RPCReturn {
				return ctx.OK(name)
			}),
		"",
	)
	file, _ := os.Create("../cpu.prof")

	time.Sleep(5000 * time.Millisecond)
	_ = pprof.StartCPUProfile(file)

	b.ReportAllocs()
	b.N = 50000000
	b.SetParallelism(1024)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stream := NewRPCStream()
			stream.WriteString("$.user:sayHello")
			stream.WriteUint64(3)
			stream.WriteString("#")
			stream.WriteString("world")
			processor.PutStream(stream)
		}
	})
	b.StopTimer()

	pprof.StopCPUProfile()
}

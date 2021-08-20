package rpc

import (
	"strconv"
	"testing"
	"time"
)

type blackHoleStreamReceiver struct {
}

func (p *blackHoleStreamReceiver) OnReceiveStream(stream *Stream) {
	stream.Release()
}

func testWithRPCBenchmark(
	numOfThreads int,
	threadBufferSize uint32,
	fnCache ActionCache,
	handler interface{},
	serviceData Map,
	testCount int,
	b *testing.B,
	args ...interface{},
) bool {
	if stream, err := MakeInternalRequestStream(
		true, 0, "#.test:bench", "", args...,
	); err == nil {
		defer stream.Release()

		if processor := NewProcessor(
			numOfThreads,
			64,
			64,
			threadBufferSize,
			fnCache,
			5*time.Second,
			[]*ServiceMeta{{
				name:     "test",
				service:  NewService().On("bench", handler),
				fileLine: "",
				data:     serviceData,
			}},
			&blackHoleStreamReceiver{},
		); processor != nil {
			b.ResetTimer()
			b.ReportAllocs()
			bNPtr := &b.N
			*bNPtr = testCount
			b.SetParallelism(32)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					processor.PutStream(stream.Clone())
				}
			})
			return processor.Close()
		}
	}

	return false
}

func BenchmarkRPC_basic(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime) Return {
			return rt.Reply(true)
		},
		nil,
		20000000,
		b,
	)
}

func BenchmarkRPC_string(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, rtName RTValue) Return {
			if name, err := rtName.ToString(); err != nil {
				panic("error")
			} else if name != "kitty" {
				panic("error")
			} else {
				return rt.Reply(true)
			}
		},
		nil,
		20000000,
		b,
		"kitty",
	)
}

func BenchmarkRPC_array(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, rtArray RTArray) Return {
			if rtArray.Size() != 3 {
				panic("error")
			}
			if v, err := rtArray.Get(0).ToString(); err != nil || v != "hello" {
				panic("error")
			}
			if v, err := rtArray.Get(1).ToString(); err != nil || v != "world" {
				panic("error")
			}
			if v, err := rtArray.Get(2).ToBool(); err != nil || v != true {
				panic("error")
			}
			return rt.Reply(true)
		},
		nil,
		20000000,
		b,
		Array{"hello", "world", true},
	)
}

func BenchmarkRPC_map(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, rtMap RTMap) Return {
			if rtMap.Size() != 2 {
				panic("error")
			}
			if v, err := rtMap.Get("name").ToString(); err != nil ||
				v != "kitty" {
				panic("error")
			}
			if v, err := rtMap.Get("age").ToInt64(); err != nil || v != 12 {
				panic("error")
			}
			return rt.Reply(true)
		},
		nil,
		20000000,
		b,
		Map{"name": "kitty", "age": int64(12)},
	)
}

func BenchmarkRPC_call_reply_rtArray(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, v int64) Return {
			if v == 0 {
				rtArray := rt.NewRTArray(64)
				for i := 0; i < 64; i++ {
					_ = rtArray.Append("hello")
				}
				return rt.Reply(rtArray)
			}
			return rt.Reply(rt.Call("#.test:bench", v-1))
		},
		nil,
		10000000,
		b,
		int64(6),
	)
}

func BenchmarkRPC_call_reply_rtMap(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, v int64) Return {
			if v == 0 {
				rtMap := rt.NewRTMap(32)
				for i := 0; i < 32; i++ {
					_ = rtMap.Set(strconv.Itoa(i), "hello")
				}
				return rt.Reply(rtMap)
			}
			return rt.Reply(rt.Call("#.test:bench", v-1))
		},
		nil,
		10000000,
		b,
		int64(6),
	)
}

func BenchmarkRPC_call_reply_rtValue(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime, v int64, rtValue RTValue) Return {
			if v == 0 {
				return rt.Reply(v)
			}
			return rt.Reply(rt.Call("#.test:bench", v-1, rtValue))
		},
		nil,
		10000000,
		b,
		int64(6),
		Array{
			Map{"name": "kitty", "age": int64(12)},
			Map{"name": "doggy", "age": int64(14)},
			Map{"name": "ducky", "age": int64(9)},
		},
	)
}

func BenchmarkRPC_getServiceData(b *testing.B) {
	testWithRPCBenchmark(
		8192*8,
		1024,
		&testFuncCache{},
		func(rt Runtime) Return {
			if v, ok := rt.GetServiceConfig("dbName"); !ok || v != "mysql" {
				panic("error")
			}
			if v, ok := rt.GetServiceConfig("port"); !ok || v != uint64(3366) {
				panic("error")
			}
			return rt.Reply(true)
		},
		Map{"dbName": "mysql", "port": uint64(3366)},
		20000000,
		b,
	)
}

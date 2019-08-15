package core

import (
	"strconv"
	"testing"
)

func Test_RPCMap_newRPCArray(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(newRPCMap(validCtx).ok()).IsTrue()
	assert(newRPCMap(invalidCtx)).Equals(nilRPCMap)
}

func Test_RPCMap_getStream(t *testing.T) {
	assert := NewAssert(t)
	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validMap := newRPCMap(validCtx)
	invalidMap := newRPCMap(invalidCtx)
	assert(validMap.ctx.getCacheStream()).IsNotNil()
	assert(invalidMap.ctx.getCacheStream()).IsNil()
	assert(invalidMap.Size()).Equals(0)
	assert(len(invalidMap.Keys())).Equals(0)
}

func Test_RPCMap_Get(t *testing.T) {
	assert := NewAssert(t)
	testSmallMap := make(map[string]interface{})
	testLargeMap := make(map[string]interface{})

	testCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	testSmallMap["0"] = nil
	testSmallMap["1"] = false
	testSmallMap["2"] = float64(3.14)
	testSmallMap["3"] = int64(30000)
	testSmallMap["4"] = uint64(30000)
	testSmallMap["5"] = ""
	testSmallMap["6"] = "hello"
	testSmallMap["7"] = []byte{}
	testSmallMap["8"] = []byte{0x53}
	testSmallMap["9"] = newRPCArray(testCtx)
	testSmallMap10 := newRPCArray(testCtx)
	testSmallMap10.Append("world")
	testSmallMap["10"] = testSmallMap10
	testSmallMap["11"] = newRPCMap(testCtx)
	testSmallMap12 := newRPCMap(testCtx)
	testSmallMap12.Set("hello", "world")
	testSmallMap["12"] = testSmallMap12
	testSmallMap["13"] = nil
	testSmallMap["14"] = nil
	testSmallMap["15"] = nil

	testLargeMap["0"] = nil
	testLargeMap["1"] = false
	testLargeMap["2"] = float64(3.14)
	testLargeMap["3"] = int64(30000)
	testLargeMap["4"] = uint64(30000)
	testLargeMap["5"] = ""
	testLargeMap["6"] = "hello"
	testLargeMap["7"] = []byte{}
	testLargeMap["8"] = []byte{0x53}
	testLargeMap["9"] = newRPCArray(testCtx)
	testLargeMap10 := newRPCArray(testCtx)
	testLargeMap10.Append("world")
	testLargeMap["10"] = testLargeMap10
	testLargeMap["11"] = newRPCMap(testCtx)
	testLargeMap12 := newRPCMap(testCtx)
	testLargeMap12.Set("hello", "world")
	testLargeMap["12"] = testLargeMap12
	testLargeMap["13"] = nil
	testLargeMap["14"] = nil
	testLargeMap["15"] = nil
	testLargeMap["16"] = nil

	fnTestMap := func(mp map[string]interface{}, name string, tp string) {
		inner := &rpcInnerContext{
			stream: NewRPCStream(),
		}
		ctx := &rpcContext{
			inner: inner,
		}
		rpcMap := newRPCMap(ctx)
		for k, v := range mp {
			rpcMap.Set(k, v)
		}

		stream := NewRPCStream()
		stream.Write(rpcMap)
		sm, _ := stream.ReadRPCMap(ctx)

		switch tp {
		case "nil":
			assert(sm.GetNil(name)).Equals(true)
			assert(sm.GetNil("no")).Equals(false)
			ctx.close()
			assert(sm.GetNil(name)).Equals(false)
		case "bool":
			assert(sm.GetBool(name)).Equals(mp[name], true)
			assert(sm.GetBool("no")).Equals(false, false)
			ctx.close()
			assert(sm.GetBool(name)).Equals(false, false)
		case "float64":
			assert(sm.GetFloat64(name)).Equals(mp[name], true)
			assert(sm.GetFloat64("no")).Equals(float64(0), false)
			ctx.close()
			assert(sm.GetFloat64(name)).Equals(float64(0), false)
		case "int64":
			assert(sm.GetInt64(name)).Equals(mp[name], true)
			assert(sm.GetInt64("no")).Equals(int64(0), false)
			ctx.close()
			assert(sm.GetInt64(name)).Equals(int64(0), false)
		case "uint64":
			assert(sm.GetUint64(name)).Equals(mp[name], true)
			assert(sm.GetUint64("no")).Equals(uint64(0), false)
			ctx.close()
			assert(sm.GetUint64(name)).Equals(uint64(0), false)
		case "string":
			assert(sm.GetString(name)).Equals(mp[name], true)
			assert(sm.GetString("no")).Equals("", false)
			ctx.close()
			assert(sm.GetString(name)).Equals("", false)
		case "bytes":
			assert(sm.GetBytes(name)).Equals(mp[name], true)
			assert(sm.GetBytes("no")).Equals(emptyBytes, false)
			ctx.close()
			assert(sm.GetBytes(name)).Equals(emptyBytes, false)
		case "rpcArray":
			target1, ok := sm.GetRPCArray(name)
			assert(ok).Equals(true)
			_, ok = sm.GetRPCArray("no")
			assert(ok).Equals(false)
			ctx.close()
			assert(sm.GetRPCArray(name)).Equals(nilRPCArray, false)
			assert(target1.ctx).Equals(ctx)
		case "rpcMap":
			target1, ok := sm.GetRPCMap(name)
			assert(ok).Equals(true)
			_, ok = sm.GetRPCMap("no")
			assert(ok).Equals(false)
			ctx.close()
			assert(sm.GetRPCMap(name)).Equals(nilRPCMap, false)
			assert(target1.ctx).Equals(ctx)
		}

		ctx.inner = inner
		assert(sm.Get(name)).Equals(mp[name], true)
		assert(sm.Get("no")).Equals(nil, false)
		ctx.close()
		assert(sm.Get(name)).Equals(nil, false)
		assert(ctx.close()).IsFalse()
	}

	fnTestMap(testSmallMap, "0", "nil")
	fnTestMap(testLargeMap, "0", "nil")
	fnTestMap(testSmallMap, "1", "bool")
	fnTestMap(testLargeMap, "1", "bool")
	fnTestMap(testSmallMap, "2", "float64")
	fnTestMap(testLargeMap, "2", "float64")
	fnTestMap(testSmallMap, "3", "int64")
	fnTestMap(testLargeMap, "3", "int64")
	fnTestMap(testSmallMap, "4", "uint64")
	fnTestMap(testLargeMap, "4", "uint64")
	fnTestMap(testSmallMap, "5", "string")
	fnTestMap(testLargeMap, "5", "string")
	fnTestMap(testSmallMap, "6", "string")
	fnTestMap(testLargeMap, "6", "string")
	fnTestMap(testSmallMap, "7", "bytes")
	fnTestMap(testLargeMap, "7", "bytes")
	fnTestMap(testSmallMap, "8", "bytes")
	fnTestMap(testLargeMap, "8", "bytes")
	fnTestMap(testSmallMap, "9", "rpcArray")
	fnTestMap(testLargeMap, "9", "rpcArray")
	fnTestMap(testSmallMap, "10", "rpcArray")
	fnTestMap(testLargeMap, "10", "rpcArray")
	fnTestMap(testSmallMap, "11", "rpcMap")
	fnTestMap(testLargeMap, "11", "rpcMap")
	fnTestMap(testSmallMap, "12", "rpcMap")
	fnTestMap(testLargeMap, "12", "rpcMap")
}

func Test_RPCMap_Set(t *testing.T) {
	assert := NewAssert(t)
	ctx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}

	fnTest := func(tp string, name string, value interface{}) {
		validCtx := &rpcContext{
			inner: &rpcInnerContext{
				stream: NewRPCStream(),
			},
		}
		invalidCtx := &rpcContext{
			inner: nil,
		}

		map0 := newRPCMap(validCtx)
		map16 := newRPCMap(validCtx)
		map100 := newRPCMap(validCtx)

		for i := 0; i < 16; i++ {
			map16.Set(strconv.Itoa(i), i)
		}
		for i := 0; i < 100; i++ {
			map100.Set(strconv.Itoa(i), i)
		}
		invalidMap := newRPCMap(invalidCtx)

		switch tp {
		case "nil":
			assert(map0.SetNil(name)).IsTrue()
			assert(map16.SetNil(name)).IsTrue()
			assert(map100.SetNil(name)).IsTrue()
			assert(invalidMap.SetNil(name)).IsFalse()
		case "bool":
			assert(map0.SetBool(name, value.(bool))).IsTrue()
			assert(map16.SetBool(name, value.(bool))).IsTrue()
			assert(map100.SetBool(name, value.(bool))).IsTrue()
			assert(invalidMap.SetBool(name, value.(bool))).IsFalse()
		case "int64":
			assert(map0.SetInt64(name, value.(int64))).IsTrue()
			assert(map16.SetInt64(name, value.(int64))).IsTrue()
			assert(map100.SetInt64(name, value.(int64))).IsTrue()
			assert(invalidMap.SetInt64(name, value.(int64))).IsFalse()
		case "uint64":
			assert(map0.SetUint64(name, value.(uint64))).IsTrue()
			assert(map16.SetUint64(name, value.(uint64))).IsTrue()
			assert(map100.SetUint64(name, value.(uint64))).IsTrue()
			assert(invalidMap.SetUint64(name, value.(uint64))).IsFalse()
		case "float64":
			assert(map0.SetFloat64(name, value.(float64))).IsTrue()
			assert(map16.SetFloat64(name, value.(float64))).IsTrue()
			assert(map100.SetFloat64(name, value.(float64))).IsTrue()
			assert(invalidMap.SetFloat64(name, value.(float64))).IsFalse()
		case "string":
			assert(map0.SetString(name, value.(string))).IsTrue()
			assert(map16.SetString(name, value.(string))).IsTrue()
			assert(map100.SetString(name, value.(string))).IsTrue()
			assert(invalidMap.SetString(name, value.(string))).IsFalse()
		case "bytes":
			assert(map0.SetBytes(name, value.([]byte))).IsTrue()
			assert(map16.SetBytes(name, value.([]byte))).IsTrue()
			assert(map100.SetBytes(name, value.([]byte))).IsTrue()
			assert(invalidMap.SetBytes(name, value.([]byte))).IsFalse()
		case "rpcArray":
			assert(map0.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(map16.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(map100.SetRPCArray(name, value.(rpcArray))).IsTrue()
			assert(invalidMap.SetRPCArray(name, value.(rpcArray))).IsFalse()
		case "rpcMap":
			assert(map0.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(map16.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(map100.SetRPCMap(name, value.(rpcMap))).IsTrue()
			assert(invalidMap.SetRPCMap(name, value.(rpcMap))).IsFalse()
		}

		assert(map0.Set(name, value)).IsTrue()
		assert(map16.Set(name, value)).IsTrue()
		assert(map100.Set(name, value)).IsTrue()
		assert(invalidMap.Set(name, value)).IsFalse()
	}

	fnTest("nil", "1", nil)
	fnTest("bool", "2", false)
	fnTest("float64", "3", float64(3.14))
	fnTest("int64", "4", int64(23))
	fnTest("uint64", "5", uint64(324))
	fnTest("string", "6", "hello")
	fnTest("bytes", "7", []byte{123, 1})
	fnTest("rpcArray", "8", newRPCArray(ctx))
	fnTest("rpcMap", "9", newRPCMap(ctx))

	fnTest("nil", "t1", nil)
	fnTest("bool", "t2", false)
	fnTest("float64", "t3", float64(3.14))
	fnTest("int64", "t4", int64(23))
	fnTest("uint64", "t5", uint64(324))
	fnTest("string", "t6", "hello")
	fnTest("bytes", "t7", []byte{123, 1})
	fnTest("rpcArray", "t8", newRPCArray(ctx))
	fnTest("rpcMap", "t9", newRPCMap(ctx))

	mp := newRPCMap(ctx)
	invalidCtx := &rpcContext{
		inner: nil,
	}
	assert(mp.SetRPCArray("error", newRPCArray(invalidCtx))).IsFalse()
	assert(mp.SetRPCMap("error", newRPCMap(invalidCtx))).IsFalse()
	assert(mp.Set("error", make(chan bool))).IsFalse()
}

func Test_RPCMap_Delete(t *testing.T) {
	assert := NewAssert(t)

	rpcMap := newRPCMap(nil)

	validCtx := &rpcContext{
		inner: &rpcInnerContext{
			stream: NewRPCStream(),
		},
	}
	invalidCtx := &rpcContext{
		inner: nil,
	}
	validMap := newRPCMap(validCtx)
	invalidMap := newRPCMap(invalidCtx)

	assert(rpcMap.Delete("")).Equals(false)
	assert(validMap.Delete("")).Equals(false)
	assert(invalidMap.Delete("")).Equals(false)
	assert(nilRPCMap.Delete("")).Equals(false)

	assert(rpcMap.Delete("hi")).Equals(false)
	assert(validMap.Delete("hi")).Equals(false)
	assert(invalidMap.Delete("hi")).Equals(false)
	assert(nilRPCMap.Delete("hi")).Equals(false)
}
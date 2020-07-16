package internal

import (
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) RPCReplyCacheFunc {
	switch fnString {
	case "S":
		return func(c *RPCContext, s *Stream, f interface{}) bool {
			h, ok := s.ReadString()
			if !ok || s.CanRead() {
				return false
			}
			f.(func(*RPCContext, string) *RPCReturn)(c, h)
			return true
		}
	default:
		return nil
	}
}

func readStringFromFile(filePath string) (string, error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func TestFnCache_basic(t *testing.T) {
	assert := NewAssert(t)
	_, file, _, _ := runtime.Caller(0)

	processor0 := NewProcessor(true, 8192, 16, 32, nil)
	assert(processor0.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-0.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-0.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-0.go")))

	processor1 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor1.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor1.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-1.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-1.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-1.go")))

	processor2 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor2.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCBool) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor2.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-2.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-2.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-2.go")))

	processor3 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor3.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCInt) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor3.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-3.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-3.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-3.go")))

	processor4 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor4.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCUint) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor4.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-4.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-4.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-4.go")))

	processor5 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor5.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCFloat) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor5.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-5.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-5.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-5.go")))

	processor6 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor6.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCString) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor6.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-6.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-6.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-6.go")))

	processor7 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor7.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCBytes) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor7.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-7.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-7.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-7.go")))

	processor8 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor8.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCArray) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor8.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-8.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-8.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-8.go")))

	processor9 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor9.AddService("abc", NewService().
		Reply("sayHello", func(ctx *RPCContext, _ RPCMap) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor9.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-9.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-9.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-9.go")))

	processor10 := NewProcessor(true, 8192, 16, 32, nil)
	_ = processor10.AddService("abc", NewService().
		Reply("sayHello", func(
			ctx *RPCContext, _ RPCBool, _ RPCInt, _ RPCUint, _ RPCFloat, _ RPCString,
			_ RPCBytes, _ RPCArray, _ RPCMap,
		) *RPCReturn {
			return ctx.OK(true)
		}), "")
	assert(processor10.BuildCache(
		"pkgName",
		path.Join(path.Dir(file), "_tmp_/fncache-basic-10.go"),
	)).IsNil()
	assert(readStringFromFile(
		path.Join(path.Dir(file), "_snapshot_/fncache-basic-10.snapshot"),
	)).Equals(readStringFromFile(
		path.Join(path.Dir(file), "_tmp_/fncache-basic-10.go")))

	_ = os.RemoveAll(path.Join(path.Dir(file), "_tmp_"))
}

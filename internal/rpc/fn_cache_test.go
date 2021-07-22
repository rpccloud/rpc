package rpc

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestBuildFuncCache(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)
	defer func() {
		_ = os.RemoveAll(path.Join(path.Dir(curFile), "_tmp_"))
	}()

	t.Run("duplicate kind", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "_tmp_/duplicate-kind.go")
		assert(buildFuncCache("pkgName", filePath, []string{"A", "A"})).Equals(
			base.ErrFnCacheDuplicateKindString.AddDebug("duplicate kind A"),
		)
	})

	t.Run("illegal kind", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "_tmp_/illegal-kind.go")
		assert(buildFuncCache("pkgName", filePath, []string{"T", "A"})).
			Equals(base.ErrFnCacheIllegalKindString.AddDebug("illegal kind T"))
	})

	t.Run("mkdir error", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "fn_cache_test.go", "error.go")
		assert(buildFuncCache("pkgName", filePath, []string{"A"})).
			Equals(base.ErrCacheMkdirAll)
	})

	t.Run("write to file error", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "_snapshot_")
		assert(buildFuncCache("pkgName", filePath, []string{"A"})).
			Equals(base.ErrCacheWriteFile)
	})

	t.Run("kinds is empty", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "_tmp_/test-cache-01.go")
		snapPath := path.Join(curDir, "_snapshot_/test-cache-01.snapshot")
		assert(buildFuncCache("pkgName", filePath, []string{})).IsNil()
		assert(base.ReadFromFile(filePath)).Equals(base.ReadFromFile(snapPath))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		filePath := path.Join(curDir, "_tmp_/test-cache-02.go")
		snapPath := path.Join(curDir, "_snapshot_/test-cache-02.snapshot")
		assert(buildFuncCache("pkgName", filePath, []string{
			"",
			"BIUFSXAMVYZ",
			"B", "I", "U", "F", "S", "X", "A", "M", "V", "Y", "Z",
			"BY", "IY", "UY", "FY", "SY", "XY", "AY", "MY", "VY", "YY", "ZY",
			"BUF", "ABM", "UFS", "XAA", "MMS", "MMM", "AAA", "MFF",
			"BIUFSXAM", "AAAAAAAA", "MAXSFUIB",
			"BIUFSXAMYZ", "BIUSXAMVYZ", "BIUFXAMVYZ",
		})).IsNil()
		assert(base.ReadFromFile(filePath)).Equals(base.ReadFromFile(snapPath))
	})
}

type testFuncCache struct{}

func (p *testFuncCache) Get(fnString string) ActionCacheFunc {
	switch fnString {
	case "":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if !stream.IsReadFinish() {
				return -1
			}

			stream.SetWritePosToBodyStart()
			fn.(func(Runtime) Return)(rt)
			return 0
		}
	case "S":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadString(); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, String) Return)(rt, arg0)
				return 0
			}
		}
	case "I":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadInt64(); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64) Return)(rt, arg0)
				return 0
			}
		}
	case "M":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadMap(); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Map) Return)(rt, arg0)
				return 0
			}
		}
	case "V":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadRTValue(rt); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTValue) Return)(rt, arg0)
				return 0
			}
		}
	case "Y":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadRTArray(rt); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTArray) Return)(rt, arg0)
				return 0
			}
		}
	case "Z":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadRTMap(rt); err != nil {
				return 1
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, RTMap) Return)(rt, arg0)
				return 0
			}
		}
	case "IV":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadInt64(); err != nil {
				return 1
			} else if arg1, err := stream.ReadRTValue(rt); err != nil {
				return 2
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(Runtime, Int64, RTValue) Return)(rt, arg0, arg1)
				return 0
			}
		}
	case "BIUFSXAMVYZ":
		return func(rt Runtime, stream *Stream, fn interface{}) int {
			if arg0, err := stream.ReadBool(); err != nil {
				return 1
			} else if arg1, err := stream.ReadInt64(); err != nil {
				return 2
			} else if arg2, err := stream.ReadUint64(); err != nil {
				return 3
			} else if arg3, err := stream.ReadFloat64(); err != nil {
				return 4
			} else if arg4, err := stream.ReadString(); err != nil {
				return 5
			} else if arg5, err := stream.ReadBytes(); err != nil {
				return 6
			} else if arg6, err := stream.ReadArray(); err != nil {
				return 7
			} else if arg7, err := stream.ReadMap(); err != nil {
				return 8
			} else if arg8, err := stream.ReadRTValue(rt); err != nil {
				return 9
			} else if arg9, err := stream.ReadRTArray(rt); err != nil {
				return 10
			} else if arg10, err := stream.ReadRTMap(rt); err != nil {
				return 11
			} else if !stream.IsReadFinish() {
				return -1
			} else {
				stream.SetWritePosToBodyStart()
				fn.(func(
					Runtime, Bool, Int64, Uint64, Float64,
					String, Bytes, Array, Map, RTValue, RTArray, RTMap,
				) Return)(
					rt, arg0, arg1, arg2, arg3, arg4, arg5,
					arg6, arg7, arg8, arg9, arg10,
				)
				return 0
			}
		}
	default:
		return nil
	}
}

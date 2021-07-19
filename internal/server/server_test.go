package server

import (
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

type testActionCache struct{}

func (p *testActionCache) Get(_ string) rpc.ActionCacheFunc {
	return nil
}

func TestServerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(defaultMaxNumOfThreads).Equal(1024 * 1024)
		assert(defaultThreadsPerCPU).Equal(16384)
		assert(defaultThreadBufferSize).Equal(2048)
		assert(defaultCloseTimeout).Equal(5 * time.Second)
		assert(defaultMaxNodeDepth).Equal(128)
		assert(defaultMaxCallDepth).Equal(128)
		assert(fnNumCPU()).Equal(runtime.NumCPU())
	})
}

func TestNewServer(t *testing.T) {
	t.Run("numOfThreads <= defaultMaxNumOfThreads", func(t *testing.T) {
		fnNumCPU = func() int {
			return 4
		}
		defer func() {
			fnNumCPU = runtime.NumCPU
		}()

		assert := base.NewAssert(t)
		v := NewServer(rpc.NewTestStreamReceiver())
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
		assert(v.sessionServer).IsNotNil()
		assert(v.numOfThreads).Equal(4 * 16384)
		assert(v.maxNodeDepth).Equal(int16(128))
		assert(v.maxCallDepth).Equal(int16(128))
		assert(v.threadBufferSize).Equal(uint32(2048))
		assert(v.actionCache).IsNil()
		assert(v.closeTimeout).Equal(5 * time.Second)
		assert(v.logReceiver).IsNotNil()
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
	})

	t.Run("numOfThreads > defaultMaxNumOfThreads", func(t *testing.T) {
		fnNumCPU = func() int {
			return 256
		}
		defer func() {
			fnNumCPU = runtime.NumCPU
		}()

		assert := base.NewAssert(t)
		v := NewServer(rpc.NewTestStreamReceiver())
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
		assert(v.sessionServer).IsNotNil()
		assert(v.numOfThreads).Equal(1024 * 1024)
		assert(v.maxNodeDepth).Equal(int16(128))
		assert(v.maxCallDepth).Equal(int16(128))
		assert(v.threadBufferSize).Equal(uint32(2048))
		assert(v.actionCache).IsNil()
		assert(v.closeTimeout).Equal(5 * time.Second)
		assert(v.logReceiver).IsNotNil()
		assert(len(v.mountServices)).Equal(0)
		assert(cap(v.mountServices)).Equal(0)
	})
}

func TestServer_Listen(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.Listen("tcp", "127.0.0.1:1234", nil)
		assert(logReceiver.GetStream()).IsNil()
	})
}

func TestServer_ListenWithDebug(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.ListenWithDebug("tcp", "127.0.0.1:1234", nil)
		assert(logReceiver.GetStream()).IsNil()
	})
}
func TestServer_SetNumOfThreads(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.isRunning = true
		_, source := v.SetNumOfThreads(1024), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("numOfThreads == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		_, source := v.SetNumOfThreads(0), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrNumOfThreadsIsWrong.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.SetNumOfThreads(1024)
		assert(v.numOfThreads).Equal(1024)
	})
}

func TestServer_SetThreadBufferSize(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.isRunning = true
		_, source := v.SetThreadBufferSize(1024), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("threadBufferSize == 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		_, source := v.SetThreadBufferSize(0), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrThreadBufferSizeIsWrong.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.SetThreadBufferSize(1023)
		assert(v.threadBufferSize).Equal(uint32(1023))
	})
}

func TestServer_SetActionCache(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.isRunning = true
		_, source := v.SetActionCache(&testActionCache{}), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
		assert(v.actionCache).IsNil()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		ac := &testActionCache{}
		v := NewServer(nil)
		v.SetActionCache(ac)
		assert(v.actionCache).Equal(ac)
	})
}

func TestServer_AddService(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := rpc.NewService()
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.isRunning = true
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := rpc.NewService()
		v := NewServer(nil)
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(v.mountServices[0]).Equal(rpc.NewServiceMeta(
			"t",
			service,
			source,
			nil,
		))
	})
}

func TestServer_BuildReplyCache(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("test ok", func(t *testing.T) {
		defer func() {
			_ = os.RemoveAll(path.Join(curDir, "cache"))
		}()
		assert := base.NewAssert(t)
		v := NewServer(nil)
		assert(v.BuildReplyCache()).Equal(v)
	})

	t.Run("output file exists", func(t *testing.T) {
		defer func() {
			_ = os.RemoveAll(path.Join(curDir, "cache"))
		}()

		_ = os.MkdirAll(path.Join(curDir, "cache"), 0555)
		_ = os.MkdirAll(path.Join(curDir, "cache", "rpc_action_cache.go"), 0555)
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		assert(v.BuildReplyCache()).Equal(v)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrCacheWriteFile.Standardize(),
		)
	})
}

func TestServer_OnReceiveStream(t *testing.T) {
	t.Run("StreamKindRPCInternalRequest", func(t *testing.T) {
		assert := base.NewAssert(t)

		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver).
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		go func() {
			v.Open()
		}()

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetDepth(0)
		stream.WriteString("#.test.Eval")
		stream.WriteString("@")

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		v.OnReceiveStream(stream)

		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
			Equal(nil, base.ErrServerSessionNotFound)

		for !base.IsTCPPortOccupied(8888) {
			time.Sleep(10 * time.Millisecond)
		}
		v.Close()
	})

	t.Run("StreamKindRPCExternalRequest", func(t *testing.T) {
		assert := base.NewAssert(t)

		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver).
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		go func() {
			v.Open()
		}()

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetDepth(0)
		stream.WriteString("#.test.Eval")
		stream.WriteString("@")

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		v.OnReceiveStream(stream)

		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
			Equal(nil, base.ErrServerSessionNotFound)

		for !base.IsTCPPortOccupied(8888) {
			time.Sleep(10 * time.Millisecond)
		}
		v.Close()
	})

	t.Run("StreamKindRPCResponseOK", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver).
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)

		go func() {
			v.Open()
		}()

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		stream.Write(true)

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		v.OnReceiveStream(stream)
		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
			Equal(nil, base.ErrServerSessionNotFound)

		for !base.IsTCPPortOccupied(8888) {
			time.Sleep(10 * time.Millisecond)
		}
		v.Close()
	})

	t.Run("StreamKindRPCResponseError", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver).
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)
		go func() {
			v.Open()
		}()

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseError)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		v.OnReceiveStream(stream)
		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
			Equal(nil, base.ErrServerSessionNotFound)

		for !base.IsTCPPortOccupied(8888) {
			time.Sleep(10 * time.Millisecond)
		}
		v.Close()
	})

	t.Run("StreamKindRPCBoardCast", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver).
			SetNumOfThreads(1024).
			Listen("tcp", "127.0.0.1:8888", nil)
		go func() {
			v.Open()
		}()

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCBoardCast)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		v.OnReceiveStream(stream)
		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
			Equal(nil, base.ErrServerSessionNotFound)

		for !base.IsTCPPortOccupied(8888) {
			time.Sleep(10 * time.Millisecond)
		}
		v.Close()
	})

	t.Run("StreamKindSystemErrorReport log to screen", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindSystemErrorReport)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())
		v.OnReceiveStream(stream)
		assert(stream.GetKind() != rpc.StreamKindSystemErrorReport).IsTrue()
	})

	t.Run("StreamKindConnectResponse", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
		stream.WriteString(base.ErrStream.GetMessage())
		v.OnReceiveStream(stream)
		assert(stream.GetKind() != rpc.StreamKindConnectResponse).IsTrue()
	})

}

func TestServer_Open(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.isRunning = true
		isOpen, source := v.Open(), base.GetFileLine(0)
		assert(isOpen).Equal(false)
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerAlreadyRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("processor create error", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		v.numOfThreads = 0
		assert(v.Open()).IsFalse()
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrNumOfThreadsIsWrong.Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			for !v.IsRunning() {
				time.Sleep(10 * time.Millisecond)
			}

			time.Sleep(200 * time.Millisecond)
			v.Close()
		}()

		assert(v.Open()).IsTrue()
	})
}

func TestServer_IsRunning(t *testing.T) {
	t.Run("not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		assert(v.isRunning).IsFalse()
	})

	t.Run("running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.isRunning = true
		assert(v.isRunning).IsTrue()
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("server is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		logReceiver := rpc.NewTestStreamReceiver()
		v := NewServer(logReceiver)
		isSuccess, source := v.Close(), base.GetFileLine(0)
		assert(isSuccess).IsFalse()
		assert(rpc.ParseResponseStream(logReceiver.GetStream())).Equal(
			nil, base.ErrServerNotRunning.AddDebug(source).Standardize(),
		)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			v.Open()
		}()

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)
		assert(v.Close()).IsTrue()
		assert(v.isRunning).IsFalse()
		assert(v.processor).IsNil()
	})
}

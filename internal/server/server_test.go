package server

import (
	"bytes"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

func captureStdout(fn func()) string {
	oldStdout := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	func() {
		defer func() {
			_ = recover()
		}()
		fn()
	}()

	outCH := make(chan string)
	// copy the output in a separate goroutine so print can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCH <- buf.String()
	}()

	os.Stdout = oldStdout
	_ = w.Close()
	ret := <-outCH
	_ = r.Close()
	return ret
}

type testActionCache struct{}

func (p *testActionCache) Get(_ string) rpc.ActionCacheFunc {
	return nil
}

func TestNewServer(t *testing.T) {
	t.Run("config is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		assert(v.config).IsNotNil()
		assert(len(v.listeners), cap(v.listeners)).Equals(0, 0)
		assert(v.processor).IsNil()
		assert(v.sessionServer).IsNil()
		assert(v.streamHub).IsNil()
		assert(len(v.mountServices), cap(v.mountServices)).Equals(0, 0)
	})

	t.Run("config is not nil", func(t *testing.T) {
		config := GetDefaultServerConfig()
		assert := base.NewAssert(t)
		v := NewServer(config)
		assert(v.config).Equals(config)
		assert(len(v.listeners), cap(v.listeners)).Equals(0, 0)
		assert(v.processor).IsNil()
		assert(v.sessionServer).IsNil()
		assert(v.streamHub).IsNil()
		assert(len(v.mountServices), cap(v.mountServices)).Equals(0, 0)
	})
}

func TestServer_Listen(t *testing.T) {
	t.Run("p.streamHub != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1)
		v := NewServer(nil)
		v.streamHub = rpc.NewStreamHub(
			true, "", base.ErrorLogAll, rpc.StreamHubCallback{
				OnSystemErrorReportStream: func(
					sessionID uint64,
					err *base.Error,
				) {
					errCH <- err
				},
			},
		)

		v.Listen("tcp", "127.0.0.1:1234", nil)
		assert((<-errCH).GetCode()).
			Equals(base.ErrServerAlreadyRunning.GetCode())
		v.streamHub.Close()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.Listen("tcp", "127.0.0.1:1234", nil)
		assert(len(v.listeners)).Equals(1)
		assert(v.listeners[0]).Equals(&listener{
			isDebug:   false,
			network:   "tcp",
			addr:      "127.0.0.1:1234",
			tlsConfig: nil,
		})
	})
}

func TestServer_ListenWithDebug(t *testing.T) {
	t.Run("p.streamHub != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1)
		v := NewServer(nil)
		v.streamHub = rpc.NewStreamHub(
			true, "", base.ErrorLogAll, rpc.StreamHubCallback{
				OnSystemErrorReportStream: func(
					sessionID uint64,
					err *base.Error,
				) {
					errCH <- err
				},
			},
		)

		v.ListenWithDebug("tcp", "127.0.0.1:1234", nil)
		assert((<-errCH).GetCode()).
			Equals(base.ErrServerAlreadyRunning.GetCode())
		v.streamHub.Close()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.ListenWithDebug("tcp", "127.0.0.1:1234", nil)
		assert(len(v.listeners)).Equals(1)
		assert(v.listeners[0]).Equals(&listener{
			isDebug:   true,
			network:   "tcp",
			addr:      "127.0.0.1:1234",
			tlsConfig: nil,
		})
	})
}

func TestServer_AddService(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1)
		service := rpc.NewService()
		v := NewServer(nil)
		v.streamHub = rpc.NewStreamHub(
			true, "", base.ErrorLogAll, rpc.StreamHubCallback{
				OnSystemErrorReportStream: func(
					sessionID uint64,
					err *base.Error,
				) {
					errCH <- err
				},
			},
		)
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(<-errCH).
			Equals(base.ErrServerAlreadyRunning.AddDebug(source).Standardize())
		v.streamHub.Close()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := rpc.NewService()
		v := NewServer(nil)
		_, source := v.AddService("t", service, nil), base.GetFileLine(0)
		assert(v.mountServices[0]).Equals(rpc.NewServiceMeta(
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
		assert(v.BuildReplyCache()).Equals(nil)
	})

	t.Run("output file exists", func(t *testing.T) {
		defer func() {
			_ = os.RemoveAll(path.Join(curDir, "cache"))
		}()

		_ = os.MkdirAll(path.Join(curDir, "cache"), 0555)
		_ = os.MkdirAll(path.Join(curDir, "cache", "rpc_action_cache.go"), 0555)
		assert := base.NewAssert(t)
		v := NewServer(nil)
		assert(v.BuildReplyCache()).Equals(base.ErrCacheWriteFile.Standardize())
	})
}

// func TestServer_OnReceiveStream(t *testing.T) {
// 	t.Run("StreamKindRPCInternalRequest", func(t *testing.T) {
// 		assert := base.NewAssert(t)

// 		logReceiver := rpc.NewTestStreamReceiver()
// 		v := NewServer(logReceiver).
// 			SetNumOfThreads(1024).
// 			Listen("tcp", "127.0.0.1:8888", nil)

// 		go func() {
// 			v.Open()
// 		}()

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindRPCRequest)
// 		stream.SetDepth(0)
// 		stream.WriteString("#.test.Eval")
// 		stream.WriteString("@")

// 		for !v.IsRunning() {
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		v.OnReceiveStream(stream)

// 		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
// 			Equal(nil, base.ErrServerSessionNotFound)

// 		for !base.IsTCPPortOccupied(8888) {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 		v.Close()
// 	})

// 	t.Run("StreamKindRPCExternalRequest", func(t *testing.T) {
// 		assert := base.NewAssert(t)

// 		logReceiver := rpc.NewTestStreamReceiver()
// 		v := NewServer(logReceiver).
// 			SetNumOfThreads(1024).
// 			Listen("tcp", "127.0.0.1:8888", nil)

// 		go func() {
// 			v.Open()
// 		}()

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindRPCRequest)
// 		stream.SetDepth(0)
// 		stream.WriteString("#.test.Eval")
// 		stream.WriteString("@")

// 		for !v.IsRunning() {
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		v.OnReceiveStream(stream)

// 		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
// 			Equal(nil, base.ErrServerSessionNotFound)

// 		for !base.IsTCPPortOccupied(8888) {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 		v.Close()
// 	})

// 	t.Run("StreamKindRPCResponseOK", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		logReceiver := rpc.NewTestStreamReceiver()
// 		v := NewServer(logReceiver).
// 			SetNumOfThreads(1024).
// 			Listen("tcp", "127.0.0.1:8888", nil)

// 		go func() {
// 			v.Open()
// 		}()

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindRPCResponseOK)
// 		stream.Write(true)

// 		for !v.IsRunning() {
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		v.OnReceiveStream(stream)
// 		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
// 			Equal(nil, base.ErrServerSessionNotFound)

// 		for !base.IsTCPPortOccupied(8888) {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 		v.Close()
// 	})

// 	t.Run("StreamKindRPCResponseError", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		logReceiver := rpc.NewTestStreamReceiver()
// 		v := NewServer(logReceiver).
// 			SetNumOfThreads(1024).
// 			Listen("tcp", "127.0.0.1:8888", nil)
// 		go func() {
// 			v.Open()
// 		}()

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindRPCResponseError)
// 		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
// 		stream.WriteString(base.ErrStream.GetMessage())

// 		for !v.IsRunning() {
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		v.OnReceiveStream(stream)
// 		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
// 			Equal(nil, base.ErrServerSessionNotFound)

// 		for !base.IsTCPPortOccupied(8888) {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 		v.Close()
// 	})

// 	t.Run("StreamKindRPCBoardCast", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		logReceiver := rpc.NewTestStreamReceiver()
// 		v := NewServer(logReceiver).
// 			SetNumOfThreads(1024).
// 			Listen("tcp", "127.0.0.1:8888", nil)
// 		go func() {
// 			v.Open()
// 		}()

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindRPCBoardCast)
// 		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
// 		stream.WriteString(base.ErrStream.GetMessage())

// 		for !v.IsRunning() {
// 			time.Sleep(10 * time.Millisecond)
// 		}

// 		v.OnReceiveStream(stream)
// 		assert(rpc.ParseResponseStream(logReceiver.WaitStream())).
// 			Equal(nil, base.ErrServerSessionNotFound)

// 		for !base.IsTCPPortOccupied(8888) {
// 			time.Sleep(10 * time.Millisecond)
// 		}
// 		v.Close()
// 	})

// 	t.Run("StreamKindSystemErrorReport log to screen", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		v := NewServer(nil)

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindSystemErrorReport)
// 		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
// 		stream.WriteString(base.ErrStream.GetMessage())
// 		v.OnReceiveStream(stream)
// 		assert(stream.GetKind() != rpc.StreamKindSystemErrorReport).IsTrue()
// 	})

// 	t.Run("StreamKindConnectResponse", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		v := NewServer(nil)

// 		stream := rpc.NewStream()
// 		stream.SetKind(rpc.StreamKindConnectResponse)
// 		stream.WriteUint64(uint64(base.ErrStream.GetCode()))
// 		stream.WriteString(base.ErrStream.GetMessage())
// 		v.OnReceiveStream(stream)
// 		assert(stream.GetKind() != rpc.StreamKindConnectResponse).IsTrue()
// 	})

// }

func TestServer_Open(t *testing.T) {
	t.Run("server is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1)
		v := NewServer(nil)
		v.streamHub = rpc.NewStreamHub(
			true, "", base.ErrorLogAll, rpc.StreamHubCallback{
				OnSystemErrorReportStream: func(
					sessionID uint64,
					err *base.Error,
				) {
					errCH <- err
				},
			},
		)
		isOpen, source := v.Open(), base.GetFileLine(0)
		assert(isOpen).Equals(false)
		assert(<-errCH).
			Equals(base.ErrServerAlreadyRunning.AddDebug(source).Standardize())
		v.streamHub.Close()
	})

	t.Run("processor create error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.config.numOfThreads = 0

		outStr := captureStdout(func() {
			assert(v.Open()).IsFalse()
		})

		assert(strings.HasSuffix(
			outStr,
			"ConfigFatal[20]: numOfThreads is wrong",
		)).IsTrue()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.config.numOfThreads = 1024
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
		assert(v.IsRunning()).IsFalse()
	})

	t.Run("running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.streamHub = rpc.NewStreamHub(
			false, "", base.ErrorLogAll, rpc.StreamHubCallback{},
		)
		v.streamHub.Close()
		assert(v.IsRunning()).IsTrue()
	})
}

func TestServer_Close(t *testing.T) {
	t.Run("server is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		assert(v.Close()).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewServer(nil)
		v.config.numOfThreads = 1024
		v.Listen("tcp", "0.0.0.0:1234", nil)

		go func() {
			v.Open()
		}()

		for !v.IsRunning() {
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)
		assert(v.Close()).IsTrue()
		assert(v.IsRunning()).IsFalse()
		assert(v.processor).IsNil()
	})
}

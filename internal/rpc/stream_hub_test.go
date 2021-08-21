package rpc

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
)

func TestNewStreamHub(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		callback := StreamHubCallback{}
		v := NewStreamHub(true, "", base.ErrorLogAll, callback)
		assert(v).IsNotNil()
		assert(v.logger).IsNotNil()
		assert(v.logLevel).Equals(base.ErrorLogAll)
	})

	t.Run("err != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		callback := StreamHubCallback{
			OnSystemErrorReportStream: func(sessionID uint64, err *base.Error) {
				errCH <- err
			},
		}
		v := NewStreamHub(true, "//", base.ErrorLogAll, callback)
		assert(v).IsNotNil()
		assert(v.logger).IsNotNil()
		assert(v.logLevel).Equals(base.ErrorLogAll)
		assert((<-errCH).GetCode()).Equals(base.ErrLogOpenFile.GetCode())
	})
}

func TestStreamHub_OnReceiveStream(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for _, kind := range []uint8{
			StreamKindRPCRequest, StreamKindRPCResponseOK,
			StreamKindRPCResponseError, StreamKindRPCBoardCast,
		} {
			streamCH := make(chan *Stream, 1024)
			callback := StreamHubCallback{
				OnRPCRequestStream: func(stream *Stream) {
					streamCH <- stream
				},
				OnRPCResponseOKStream: func(stream *Stream) {
					streamCH <- stream
				},
				OnRPCResponseErrorStream: func(stream *Stream) {
					streamCH <- stream
				},
				OnRPCBoardCastStream: func(stream *Stream) {
					streamCH <- stream
				},
			}
			v := NewStreamHub(true, "", base.ErrorLogAll, callback)
			stream := NewStream()
			stream.SetKind(kind)
			v.OnReceiveStream(stream)
			assert(<-streamCH).Equals(stream)
			assert(len(streamCH)).Equals(0)
		}
	})

	t.Run("stream == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		callback := StreamHubCallback{
			OnSystemErrorReportStream: func(sessionID uint64, err *base.Error) {
				errCH <- err
			},
		}
		v := NewStreamHub(true, "", base.ErrorLogAll, callback)
		v.OnReceiveStream(nil)
		assert(len(errCH)).Equals(0)
	})

	t.Run("case StreamKindSystemErrorReport, ErrorLogAll", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		sessionCH := make(chan uint64, 1024)
		callback := StreamHubCallback{
			OnSystemErrorReportStream: func(sessionID uint64, err *base.Error) {
				errCH <- err
				sessionCH <- sessionID
			},
		}
		v := NewStreamHub(true, "./tmp/test.log", base.ErrorLogAll, callback)
		defer func() {
			os.RemoveAll("./tmp")
		}()
		stream := MakeSystemErrorStream(base.ErrStream)
		stream.SetSessionID(17)
		v.OnReceiveStream(stream)
		logStr, _ := base.ReadFromFile("./tmp/test.log")
		assert(strings.HasSuffix(
			logStr,
			" <sessionID:17> SecurityWarn[1]: stream error",
		)).IsTrue()
		assert(<-errCH).Equals(base.ErrStream)
		assert(len(errCH)).Equals(0)
	})

	t.Run("case StreamKindSystemErrorReport, ErrorLogNone", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		sessionCH := make(chan uint64, 1024)
		callback := StreamHubCallback{
			OnSystemErrorReportStream: func(sessionID uint64, err *base.Error) {
				errCH <- err
				sessionCH <- sessionID
			},
		}
		v := NewStreamHub(true, "./tmp/test.log", 0, callback)
		defer func() {
			os.RemoveAll("./tmp")
		}()
		stream := MakeSystemErrorStream(base.ErrStream)
		stream.SetSessionID(17)
		v.OnReceiveStream(stream)
		assert(base.ReadFromFile("./tmp/test.log")).Equals("", nil)
		assert(len(errCH)).Equals(0)
	})
}

func TestStreamHub_Close(t *testing.T) {
	getFieldPointer := func(ptr interface{}, fieldName string) unsafe.Pointer {
		val := reflect.Indirect(reflect.ValueOf(ptr))
		return unsafe.Pointer(val.FieldByName(fieldName).UnsafeAddr())
	}

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		callback := StreamHubCallback{}
		v := NewStreamHub(true, "", base.ErrorLogAll, callback)
		assert(v.Close()).IsTrue()
	})

	t.Run("err != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		callback := StreamHubCallback{
			OnSystemErrorReportStream: func(sessionID uint64, err *base.Error) {
				errCH <- err
			},
		}
		v := NewStreamHub(true, "./tmp/test.log", base.ErrorLogAll, callback)
		defer func() {
			os.RemoveAll("./tmp")
		}()

		filePtr := (**os.File)(getFieldPointer(v.logger, "file"))
		_ = (*filePtr).Close()
		assert(v.Close()).IsFalse()
		assert((<-errCH).GetCode()).Equals(base.ErrLogCloseFile.GetCode())
	})
}

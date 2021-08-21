package rpc

import (
	"os"
	"reflect"
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

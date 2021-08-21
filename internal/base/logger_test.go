package base

import (
	"bytes"
	"io"
	"os"
	"testing"
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

func TestNewLogger(t *testing.T) {
	t.Run("open file error", func(t *testing.T) {
		assert := NewAssert(t)
		v, e := NewLogger(false, "/")
		assert(v.file).IsNil()
		assert(e.GetCode()).Equals(ErrLogOpenFile.GetCode())
	})

	t.Run("create dir error", func(t *testing.T) {
		assert := NewAssert(t)
		f, _ := os.Create("tmp")
		f.Close()
		defer os.Remove("tmp")
		v, e := NewLogger(false, "./tmp/test.log")
		assert(v.file).IsNil()
		assert(e).Equals(
			ErrLogOpenFile.AddDebug("path tmp is not a directory"),
		)
	})

	t.Run("file is empty", func(t *testing.T) {
		assert := NewAssert(t)
		assert(NewLogger(false, "")).Equals(&Logger{
			isLogToScreen: false,
			file:          nil,
		}, nil)
	})

	t.Run("ok", func(t *testing.T) {
		assert := NewAssert(t)
		v, err := NewLogger(false, "test.log")
		assert(err).IsNil()
		assert(v.isLogToScreen).IsFalse()
		assert(v.file).IsNotNil()
		v.Close()
		_ = os.Remove("test.log")
	})
}

func TestLogger_Log(t *testing.T) {
	getFileContent := func(fileName string) string {
		ret, _ := ReadFromFile(fileName)
		return ret
	}

	t.Run("isLogToScreen == true, file is empty", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(true, "")

		assert(captureStdout(func() {
			v.Log("hello")
		})).Equals("hello")

		v.Close()
	})

	t.Run("isLogToScreen == false, file is not", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(false, "test.log")

		assert(captureStdout(func() {
			v.Log("hello")
		})).Equals("")

		assert(getFileContent("test.log")).Equals("hello")

		v.Close()
		_ = os.Remove("test.log")
	})
}

func TestLogger_Close(t *testing.T) {
	t.Run("ok, file is nil", func(t *testing.T) {
		assert := NewAssert(t)
		v, _ := NewLogger(true, "")
		assert(v.Close()).IsNil()
	})

	t.Run("ok, file is not nil", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")
		assert(v.Close()).IsNil()
		_ = os.Remove("test.log")
	})

	t.Run("file close error", func(t *testing.T) {
		assert := NewAssert(t)

		v, _ := NewLogger(true, "test.log")
		v.file.Close()
		assert(v.Close().GetCode()).Equals(ErrLogCloseFile.GetCode())
		_ = os.Remove("test.log")
	})
}

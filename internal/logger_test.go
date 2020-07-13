package internal

import (
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func runStdWriterLogger(onRun func(logger *Logger)) string {
	if r, w, err := os.Pipe(); err == nil {
		old := os.Stdout // keep backup of the real stdout
		os.Stdout = w
		defer func() {
			os.Stdout = old
		}()

		onRun(NewLogger(NewRPCStdoutLogWriter()))

		retCH := make(chan string)

		go func() {
			buf := make([]byte, 10240)
			pos := 0
			for {
				if n, err := r.Read(buf[pos:]); err == io.EOF {
					break
				} else {
					pos += n
				}
			}
			retCH <- string(buf[:pos])
		}()
		_ = w.Close()

		return <-retCH
	}
	return "<error>"
}

func runCallbackWriterLogger(
	onRun func(logger *Logger),
) (isoTime string, tag string, msg string, extra string) {
	wait := make(chan bool, 1)
	onRun(NewLogger(NewRPCCallbackLogWriter(
		func(_isoTime string, _tag string, _msg string, _extra string) {
			isoTime = _isoTime
			tag = _tag
			msg = _msg
			extra = _extra
			wait <- true
		},
	)))
	<-wait
	return
}

func TestStdLogWriter_Write(t *testing.T) {
	assert := NewRPCAssert(t)

	assert(strings.HasSuffix(
		runStdWriterLogger(func(logger *Logger) {
			logger.Info("message")
		}),
		" Info: message\n",
	)).IsTrue()

	assert(strings.HasSuffix(
		runStdWriterLogger(func(logger *Logger) {
			logger.InfoExtra("message", "extra")
		}),
		"(extra) Info: message\n",
	)).IsTrue()
}

func TestNewLogger(t *testing.T) {
	assert := NewRPCAssert(t)
	logger1 := NewLogger(nil)
	assert(logger1.level).Equals(RPCLogMaskAll)
	assert(logger1.writer).IsNotNil()

	logger2 := NewLogger(NewRPCStdoutLogWriter())
	assert(logger2.level).Equals(RPCLogMaskAll)
	assert(logger2.writer).IsNotNil()
}

func TestLogger_SetLevel(t *testing.T) {
	assert := NewRPCAssert(t)
	logger := NewLogger(nil)

	assert(logger.SetLevel(RPCLogMaskNone - 1)).IsFalse()
	assert(logger.level).Equals(RPCLogMaskAll)

	assert(logger.SetLevel(RPCLogMaskAll + 1)).IsFalse()
	assert(logger.level).Equals(RPCLogMaskAll)

	assert(logger.SetLevel(RPCLogMaskNone)).IsTrue()
	assert(logger.level).Equals(RPCLogMaskNone)

	assert(logger.SetLevel(RPCLogMaskAll)).IsTrue()
	assert(logger.level).Equals(RPCLogMaskAll)

	// test all level and logs
	fnTestLogLevel := func(level int32) int32 {
		ret := int32(0)

		logger := NewLogger(NewRPCCallbackLogWriter(
			func(_ string, tag string, msg string, _ string) {
				if msg == "message" {
					switch tag {
					case "Debug":
						atomic.AddInt32(&ret, RPCLogMaskDebug)
					case "Info":
						atomic.AddInt32(&ret, RPCLogMaskInfo)
					case "Warn":
						atomic.AddInt32(&ret, RPCLogMaskWarn)
					case "Error":
						atomic.AddInt32(&ret, RPCLogMaskError)
					case "Fatal":
						atomic.AddInt32(&ret, RPCLogMaskFatal)
					}
				}
			},
		))
		logger.SetLevel(level)
		logger.Debug("message")
		logger.Info("message")
		logger.Warn("message")
		logger.Error("message")
		logger.Fatal("message")
		time.Sleep(30 * time.Millisecond)
		return atomic.LoadInt32(&ret)
	}

	assert(fnTestLogLevel(RPCLogMaskNone - 1)).Equals(RPCLogMaskAll)
	for i := int32(0); i < 32; i++ {
		assert(fnTestLogLevel(i)).Equals(i)
	}
	assert(fnTestLogLevel(RPCLogMaskAll + 1)).Equals(RPCLogMaskAll)
}

func TestLogger_Debug(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Debug("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Debug")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestLogger_DebugExtra(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.DebugExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Debug")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestLogger_Info(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Info("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Info")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestLogger_InfoExtra(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.InfoExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Info")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestLogger_Warn(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Warn("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Warn")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestLogger_WarnExtra(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.WarnExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Warn")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestLogger_Error(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Error("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Error")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestLogger_ErrorExtra(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.ErrorExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Error")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

func TestLogger_Fatal(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.Fatal("message")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Fatal")
	assert(msg).Equals("message")
	assert(extra).Equals("")
}

func TestLogger_FatalExtra(t *testing.T) {
	assert := NewRPCAssert(t)

	isoTime, tag, msg, extra := runCallbackWriterLogger(func(logger *Logger) {
		logger.FatalExtra("message", "extra")
	})

	assert(len(isoTime) > 0).IsTrue()
	assert(tag).Equals("Fatal")
	assert(msg).Equals("message")
	assert(extra).Equals("extra")
}

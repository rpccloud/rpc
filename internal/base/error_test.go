package base

import (
	"fmt"
	"math"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("check constant", func(t *testing.T) {
		assert := NewAssert(t)

		assert(ErrorTypeConfig).Equals(ErrorType(1))
		assert(ErrorTypeNet).Equals(ErrorType(2))
		assert(ErrorTypeAction).Equals(ErrorType(3))
		assert(ErrorTypeDevelop).Equals(ErrorType(4))
		assert(ErrorTypeKernel).Equals(ErrorType(5))
		assert(ErrorTypeSecurity).Equals(ErrorType(6))

		assert(ErrorLevelInfo).Equals(ErrorLevel(1 << 0))
		assert(ErrorLevelWarn).Equals(ErrorLevel(1 << 1))
		assert(ErrorLevelError).Equals(ErrorLevel(1 << 2))
		assert(ErrorLevelFatal).Equals(ErrorLevel(1 << 3))
	})
}

func TestNewError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(NewError(123, "msg")).Equals(&Error{code: 123, message: "msg"})
	})
}

func TestDefineError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("error redefined", func(t *testing.T) {
		assert := NewAssert(t)
		_ = defineError(ErrorTypeAction, num, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(RunWithCatchPanic(func() {
			_ = defineError(
				ErrorTypeAction, num, ErrorLevelWarn, "msg", "source",
			)
		})).Equals("Error redefined :\n>>> source\n>>> source\n")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := defineError(ErrorTypeAction, num, ErrorLevelWarn, "msg", "source")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeAction, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals("source")
	})
}

func TestDefineConfigError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineConfigError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeConfig, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestDefineNetError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineNetError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeNet, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestDefineActionError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineActionError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeAction, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestDefineDevelopError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineDevelopError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeDevelop, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestDefineKernelError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, s1 := DefineKernelError(num, ErrorLevelWarn, "msg"), GetFileLine(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeKernel, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestDefineSecurityError(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		getFL := GetFileLine
		v1, s1 := DefineSecurityError(num, ErrorLevelWarn, "msg"), getFL(0)
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType(), v1.GetIndex(), v1.GetLevel(), v1.GetMessage()).
			Equals(ErrorTypeSecurity, num, ErrorLevelWarn, "msg")
		assert(errorDefineMap[num]).Equals(s1)
	})
}

func TestError_GetCode(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert((&Error{code: 123}).GetCode()).Equals(uint32(123))
	})
}

func TestError_GetType(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test ErrorTypeConfig", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineConfigError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeConfig)
	})

	t.Run("test ErrorTypeNet", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineNetError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeNet)
	})

	t.Run("test ErrorTypeAction", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineActionError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeAction)
	})

	t.Run("test ErrorTypeDevelop", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineDevelopError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeDevelop)
	})

	t.Run("test ErrorTypeKernel", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineKernelError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeKernel)
	})

	t.Run("test ErrorTypeSecurity", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetType()).Equals(ErrorTypeSecurity)
	})
}

func TestError_GetLevel(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test ErrorLevelWarn", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equals(ErrorLevelWarn)
	})

	t.Run("test ErrorLevelError", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelError, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equals(ErrorLevelError)
	})

	t.Run("test ErrorLevelFatal", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetLevel()).Equals(ErrorLevelFatal)
	})
}

func TestError_GetIndex(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test with uint32 min", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetIndex()).Equals(num)
	})

	t.Run("test with uint32 max", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetIndex()).Equals(num)
	})
}

func TestError_GetMessage(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelFatal, "msg1")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		assert(v1.GetMessage()).Equals("msg1")
	})
}

func TestError_AddDebug(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test from origin error", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineSecurityError(num, ErrorLevelWarn, "")
		v2 := DefineSecurityError(num-1, ErrorLevelFatal, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			errorDefineMutex.Unlock()
		}()
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsFalse()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsFalse()
		assert(v3.GetMessage()).Equals("dbg")
		assert(v4.GetMessage()).Equals("msg\ndbg")
	})

	t.Run("test from derived error", func(t *testing.T) {
		assert := NewAssert(t)
		v := DefineSecurityError(num, ErrorLevelWarn, "")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()

		v1 := v.AddDebug("")
		v2 := v.AddDebug("msg")
		v3 := v1.AddDebug("dbg")
		v4 := v2.AddDebug("dbg")
		assert(fmt.Sprintf("%p", v3) == fmt.Sprintf("%p", v1)).IsTrue()
		assert(fmt.Sprintf("%p", v4) == fmt.Sprintf("%p", v2)).IsTrue()
		assert(v3.GetMessage()).Equals("dbg")
		assert(v4.GetMessage()).Equals("msg\ndbg")
	})
}

func TestError_Standardize(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(ErrStream.Standardize()).Equals(ErrStream)
		assert(ErrAction.AddDebug("").Standardize()).Equals(ErrAction)
	})
}

func TestError_getErrorTypeString(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineConfigError(num, ErrorLevelFatal, "msg")
		v2 := DefineNetError(num-1, ErrorLevelFatal, "msg")
		v3 := DefineActionError(num-2, ErrorLevelFatal, "msg")
		v4 := DefineDevelopError(num-3, ErrorLevelFatal, "msg")
		v5 := DefineKernelError(num-4, ErrorLevelFatal, "msg")
		v6 := DefineSecurityError(num-5, ErrorLevelFatal, "msg")
		v7 := &Error{}
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			delete(errorDefineMap, num-2)
			delete(errorDefineMap, num-3)
			delete(errorDefineMap, num-4)
			delete(errorDefineMap, num-5)
			errorDefineMutex.Unlock()
		}()
		assert(v1.getErrorTypeString()).Equals("Config")
		assert(v2.getErrorTypeString()).Equals("Net")
		assert(v3.getErrorTypeString()).Equals("Action")
		assert(v4.getErrorTypeString()).Equals("Develop")
		assert(v5.getErrorTypeString()).Equals("Kernel")
		assert(v6.getErrorTypeString()).Equals("Security")
		assert(v7.getErrorTypeString()).Equals("")
	})
}

func TestError_getErrorLevelString(t *testing.T) {
	num := ErrorIndex(math.MaxUint16)

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := DefineConfigError(num, ErrorLevelWarn, "msg")
		v2 := DefineConfigError(num-1, ErrorLevelError, "msg")
		v3 := DefineConfigError(num-2, ErrorLevelFatal, "msg")
		v4 := DefineDevelopError(num-3, 0, "msg")
		v5 := DefineDevelopError(num-4, 255, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			delete(errorDefineMap, num-1)
			delete(errorDefineMap, num-2)
			delete(errorDefineMap, num-3)
			delete(errorDefineMap, num-4)
			errorDefineMutex.Unlock()
		}()
		assert(v1.getErrorLevelString()).Equals("Warn")
		assert(v2.getErrorLevelString()).Equals("Error")
		assert(v3.getErrorLevelString()).Equals("Fatal")
		assert(v4.getErrorLevelString()).Equals("")
		assert(v5.getErrorLevelString()).Equals("")
	})
}

func TestError_Error(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		num := ErrorIndex(math.MaxUint16)
		v1 := DefineConfigError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		v2 := v1.AddDebug("dbg")
		assert(v2.Error()).Equals(fmt.Sprintf("ConfigWarn[%d]: msg\ndbg", num))
	})
}

func TestError_ReportString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		num := ErrorIndex(math.MaxUint16)
		v1 := DefineConfigError(num, ErrorLevelWarn, "msg")
		defer func() {
			errorDefineMutex.Lock()
			delete(errorDefineMap, num)
			errorDefineMutex.Unlock()
		}()
		timeHeaderLen := len(TimeNowISOString())
		assert(v1.ReportString(0, 0)[timeHeaderLen:]).Equals(
			" ConfigWarn[65535]: msg",
		)
		assert(v1.ReportString(1, 0)[timeHeaderLen:]).Equals(
			" <machineID:1> ConfigWarn[65535]: msg",
		)
		assert(v1.ReportString(0, 1)[timeHeaderLen:]).Equals(
			" <sessionID:1> ConfigWarn[65535]: msg",
		)
		assert(v1.ReportString(1, 1)[timeHeaderLen:]).Equals(
			" <machineID:1> <sessionID:1> ConfigWarn[65535]: msg",
		)
	})
}

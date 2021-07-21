package base

import (
	"fmt"
	"strconv"
	"sync"
)

// ErrorType ...
type ErrorType uint8

const (
	// ErrorTypeConfig ...
	ErrorTypeConfig = ErrorType(1)
	// ErrorTypeNet ...
	ErrorTypeNet = ErrorType(2)
	// ErrorTypeAction ...
	ErrorTypeAction = ErrorType(3)
	// ErrorTypeDevelop ...
	ErrorTypeDevelop = ErrorType(4)
	// ErrorTypeKernel ...
	ErrorTypeKernel = ErrorType(5)
	// ErrorTypeSecurity ..
	ErrorTypeSecurity = ErrorType(6)
)

// ErrorLevel ...
type ErrorLevel uint8

const (
	// ErrorLevelInfo ...
	ErrorLevelInfo = ErrorLevel(1 << 0)
	// ErrorLevelWarn ...
	ErrorLevelWarn = ErrorLevel(1 << 1)
	// ErrorLevelError ...
	ErrorLevelError = ErrorLevel(1 << 2)
	// ErrorLevelFatal ...
	ErrorLevelFatal = ErrorLevel(1 << 3)
	// ErrorLogAll ...
	ErrorLogAll = ErrorLevelInfo |
		ErrorLevelWarn |
		ErrorLevelError |
		ErrorLevelFatal
)

// ErrorIndex ...
type ErrorIndex uint16

var (
	errorDefineMutex = &sync.Mutex{}
	errorDefineMap   = map[ErrorIndex]string{}
)

// Error ...
type Error struct {
	code    uint32
	message string
}

// NewError ...
func NewError(code uint32, message string) *Error {
	return &Error{
		code:    code,
		message: message,
	}
}

func defineError(
	kind ErrorType,
	index ErrorIndex,
	level ErrorLevel,
	message string,
	source string,
) *Error {
	errorDefineMutex.Lock()
	defer errorDefineMutex.Unlock()

	code := (uint32(kind) << 20) | (uint32(level) << 16) | uint32(index)

	if value, ok := errorDefineMap[index]; ok {
		panic(fmt.Sprintf("Error redefined :\n>>> %s\n>>> %s\n", value, source))
	} else {
		errorDefineMap[index] = source
	}

	return &Error{
		code:    code,
		message: message,
	}
}

// DefineConfigError ...
func DefineConfigError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeConfig, idx, level, msg, GetFileLine(1))
}

// DefineNetError ...
func DefineNetError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeNet, idx, level, msg, GetFileLine(1))
}

// DefineActionError ...
func DefineActionError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeAction, idx, level, msg, GetFileLine(1))
}

// DefineDevelopError ...
func DefineDevelopError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeDevelop, idx, level, msg, GetFileLine(1))
}

// DefineKernelError ...
func DefineKernelError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeKernel, idx, level, msg, GetFileLine(1))
}

// DefineSecurityError ...
func DefineSecurityError(idx ErrorIndex, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeSecurity, idx, level, msg, GetFileLine(1))
}

// GetCode ...
func (p *Error) GetCode() uint32 {
	return p.code & 0xFFFFFF
}

// GetType ...
func (p *Error) GetType() ErrorType {
	return ErrorType((p.code >> 20) & 0x0F)
}

// GetLevel ...
func (p *Error) GetLevel() ErrorLevel {
	return ErrorLevel((p.code >> 16) & 0x0F)
}

// GetIndex ...
func (p *Error) GetIndex() ErrorIndex {
	return ErrorIndex(p.code & 0xFFFF)
}

// GetMessage ...
func (p *Error) GetMessage() string {
	return p.message
}

// AddDebug ...
func (p *Error) AddDebug(debug string) *Error {
	if p.code&0xFF000000 == 0 {
		ret := &Error{code: p.code | 0x01000000}
		if p.message == "" {
			ret.message = ConcatString(debug)
		} else {
			ret.message = ConcatString(p.message, "\n", debug)
		}
		return ret
	}

	if p.message == "" {
		p.message = ConcatString(debug)
	} else {
		p.message = ConcatString(p.message, "\n", debug)
	}
	return p
}

// Standardize ...
func (p *Error) Standardize() *Error {
	p.code &= 0xFFFFFF
	return p
}

func (p *Error) getErrorTypeString() string {
	switch p.GetType() {
	case ErrorTypeConfig:
		return "Config"
	case ErrorTypeNet:
		return "Net"
	case ErrorTypeAction:
		return "Action"
	case ErrorTypeDevelop:
		return "Develop"
	case ErrorTypeKernel:
		return "Kernel"
	case ErrorTypeSecurity:
		return "Security"
	default:
		return ""
	}
}

func (p *Error) getErrorLevelString() string {
	switch p.GetLevel() {
	case ErrorLevelWarn:
		return "Warn"
	case ErrorLevelError:
		return "Error"
	case ErrorLevelFatal:
		return "Fatal"
	default:
		return ""
	}
}

func (p *Error) Error() string {
	return ConcatString(
		p.getErrorTypeString(),
		p.getErrorLevelString(),
		"[",
		strconv.FormatUint(uint64(p.GetIndex()), 10),
		"]: ",
		p.message,
	)
}

func (p *Error) ReportString(machineID uint64, sessionID uint64) string {
	machineString := ""
	if machineID != 0 {
		machineString = ConcatString(
			"<target:",
			strconv.FormatUint(machineID, 10),
			"> ",
		)
	}
	sessionString := ""
	if sessionID != 0 {
		sessionString = ConcatString(
			"<session:",
			strconv.FormatUint(sessionID, 10),
			"> ",
		)
	}

	return ConcatString(
		ConvertToIsoDateString(TimeNow()),
		" ",
		machineString,
		sessionString,
		p.getErrorTypeString(),
		p.getErrorLevelString(),
		"[",
		strconv.FormatUint(uint64(p.GetIndex()), 10),
		"]: ",
		p.message,
	)
}

package base

import (
	"testing"
)

func TestNewError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindRuntimePanic, "message", "fileLine")
	assert(o1.GetKind()).Equal(ErrorKindRuntimePanic)
	assert(o1.GetMessage()).Equal("message")
	assert(o1.GetDebug()).Equal("fileLine")
}

func TestNewReplyError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyError("message")).
		Equal(NewError(ErrorKindReply, "message", ""))
}

func TestNewReplyPanic(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewReplyPanic("message")).
		Equal(NewError(ErrorKindReplyPanic, "message", ""))
}

func TestNewRuntimeError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewRuntimePanic("message")).
		Equal(NewError(ErrorKindRuntimePanic, "message", ""))
}

func TestNewProtocolError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewProtocolError("message")).
		Equal(NewError(ErrorKindProtocol, "message", ""))
}

func TestNewTransportError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewTransportError("message")).
		Equal(NewError(ErrorKindTransport, "message", ""))
}

func TestNewKernelError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewKernelPanic("message")).
		Equal(NewError(ErrorKindKernelPanic, "message", ""))
}

func TestNewSecurityLimitError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewSecurityLimitError("message")).
		Equal(NewError(ErrorKindSecurityLimit, "message", ""))
}

func TestConvertToError(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(ConvertToError(0)).IsNil()

	// Test(2)
	assert(ConvertToError(make(chan bool))).IsNil()

	// Test(3)
	assert(ConvertToError(nil)).IsNil()

	// Test(4)
	assert(ConvertToError(NewKernelPanic("test"))).IsNotNil()
}

func TestRpcError_GetKind(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetKind()).Equal(ErrorKindKernelPanic)
}

func TestRpcError_GetMessage(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetMessage()).Equal("message")
}

func TestRpcError_GetDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o1.GetDebug()).Equal("fileLine")
}

func TestRpcError_AddDebug(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	o1 := NewError(ErrorKindKernelPanic, "message", "")
	assert(o1.AddDebug("fileLine").GetDebug()).Equal("fileLine")

	// Test(2)
	o2 := NewError(ErrorKindKernelPanic, "message", "fileLine")
	assert(o2.AddDebug("fileLine").GetDebug()).Equal("fileLine\nfileLine")
}

func TestRpcError_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(NewError(ErrorKindKernelPanic, "", "").Error()).Equal("")

	// Test(2)
	assert(NewError(ErrorKindKernelPanic, "message", "").Error()).Equal("message")

	// Test(3)
	assert(NewError(ErrorKindKernelPanic, "", "fileLine").Error()).Equal("fileLine")

	// Test(4)
	assert(NewError(ErrorKindKernelPanic, "message", "fileLine").Error()).
		Equal("message\nfileLine")
}

package internal

import "testing"

func TestContext_getThread(t *testing.T) {
	//assert := NewAssert(t)
	//
	//ctx1 := ContextObject{}
	//assert(ctx1.rpcThread).IsNil()
	//
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//
	//ctx2 := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx2.rpcThread).Equals(unsafe.Pointer(rpcThread))

}

func TestContext_stop(t *testing.T) {
	//assert := NewAssert(t)
	//
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx.rpcThread).IsNotNil()
	//ctx.stop()
	//assert(ctx.rpcThread).IsNil()
}

func TestContext_OK(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//assert(rpcThread.execSuccessful).IsFalse()
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//ctx.OK(uint(215))
	//assert(rpcThread.execSuccessful).IsTrue()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(true, true)
	//assert(rpcThread.outStream.ReadUint64()).Equals(Uint64(215), true)
	//assert(rpcThread.outStream.CanRead()).IsFalse()
	//
	//// ctx is stop
	//thread1 := newThread(nil)
	//thread1.stop()
	//assert(thread1.execSuccessful).IsFalse()
	//ctx1 := ContextObject{rpcThread: unsafe.Pointer(thread1)}
	//ctx1.stop()
	//ctx1.OK(uint(215))
	//thread1.outStream.SetReadPos(streamBodyPos)
	//assert(thread1.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(thread1.outStream.GetReadPos()).Equals(streamBodyPos)
	//
	//// value is illegal
	//thread2 := newThread(nil)
	//thread2.stop()
	//assert(thread2.execSuccessful).IsFalse()
	//ctx2 := ContextObject{rpcThread: unsafe.Pointer(thread2)}
	//ctx2.OK(make(chan bool))
	//assert(thread2.execSuccessful).IsFalse()
	//thread2.outStream.SetReadPos(streamBodyPos)
	//assert(thread2.outStream.ReadBool()).Equals(false, true)
	//assert(thread2.outStream.ReadString()).Equals("return type is error", true)
	//dbgMessage, ok := thread2.outStream.ReadString()
	//assert(ok).IsTrue()
	//assert(strings.Contains(dbgMessage, "TestRpcContext_OK")).IsTrue()
	//assert(thread2.outStream.CanRead()).IsFalse()
}

func TestContext_writeError(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//processor := NewProcessor(16, 16, nil, nil)
	//rpcThread := newThread(newThreadPool(processor))
	//rpcThread.stop()
	//rpcThread.execSuccessful = true
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//ctx.writeError("errorMessage", "errorDebug")
	//assert(rpcThread.execSuccessful).IsFalse()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorDebug", true)
	//assert(rpcThread.outStream.CanRead()).IsFalse()
	//
	//// ctx is stop
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execSuccessful = true
	//ctx1 := ContextObject{rpcThread: unsafe.Pointer(thread1)}
	//ctx1.stop()
	//ctx1.writeError("errorMessage", "errorDebug")
	//thread1.outStream.SetReadPos(streamBodyPos)
	//assert(thread1.execSuccessful).IsTrue()
	//assert(thread1.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(thread1.outStream.GetReadPos()).Equals(streamBodyPos)
}

func TestContext_Error(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx.Error(nil)).IsNil()
	//assert(rpcThread.outStream.GetWritePos()).Equals(streamBodyPos)
	//assert(rpcThread.outStream.GetReadPos()).Equals(streamBodyPos)
	//
	//assert(
	//	ctx.Error(NewErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorDebug", true)
	//
	//// ctx have execReplyNode
	//thread1 := newThread(nil)
	//thread1.stop()
	//thread1.execReplyNode = &rpcReplyNode{debugString: "nodeDebug"}
	//ctx1 := ContextObject{rpcThread: unsafe.Pointer(thread1)}
	//assert(
	//	ctx1.Error(NewErrorByDebug("errorMessage", "errorDebug")),
	//).IsNil()
	//thread1.outStream.SetReadPos(streamBodyPos)
	//assert(thread1.outStream.ReadBool()).Equals(false, true)
	//assert(thread1.outStream.ReadString()).Equals("errorMessage", true)
	//assert(thread1.outStream.ReadString()).Equals("errorDebug\nnodeDebug", true)
}

func TestContext_Errorf(t *testing.T) {
	//assert := NewAssert(t)
	//
	//// ctx is ok
	//rpcThread := newThread(nil)
	//rpcThread.stop()
	//ctx := ContextObject{rpcThread: unsafe.Pointer(rpcThread)}
	//assert(ctx.Errorf("error%s", "Message")).IsNil()
	//rpcThread.outStream.SetReadPos(streamBodyPos)
	//assert(rpcThread.outStream.ReadBool()).Equals(false, true)
	//assert(rpcThread.outStream.ReadString()).Equals("errorMessage", true)
	//dbgMessage, ok := rpcThread.outStream.ReadString()
	//assert(ok).IsTrue()
	//
	//assert(strings.Contains(dbgMessage, "TestRpcContext_Errorf")).IsTrue()
}

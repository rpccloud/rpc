package internal

import (
	"sync/atomic"
	"unsafe"
)

type Context = *ContextObject

type ContextObject struct {
	thread unsafe.Pointer
}

func (p *ContextObject) stop() {
	atomic.StorePointer(&p.thread, nil)
}

func (p *ContextObject) getThread() *rpcThread {
	return (*rpcThread)(atomic.LoadPointer(&p.thread))
}

func (p *ContextObject) writeError(message string, debug string) Return {
	if thread := p.getThread(); thread != nil {
		execStream := thread.outStream
		execStream.SetWritePos(streamBodyPos)
		execStream.WriteBool(false)
		execStream.WriteString(message)
		execStream.WriteString(debug)
		thread.execSuccessful = false
	}
	return nilReturn
}

// OK get success ReturnObject  by value
func (p *ContextObject) OK(value interface{}) Return {
	if thread := p.getThread(); thread != nil {
		stream := thread.outStream
		stream.SetWritePos(streamBodyPos)
		stream.WriteBool(true)

		if stream.Write(value) != StreamWriteOK {
			return p.writeError(
				"return type is error",
				GetStackString(1),
			)
		}

		thread.execSuccessful = true
	}
	return nilReturn
}

func (p *ContextObject) Error(err Error) Return {
	if err == nil {
		return nilReturn
	}

	if thread := (*rpcThread)(p.thread); thread != nil &&
		thread.execReplyNode != nil &&
		thread.execReplyNode.debugString != "" {
		err.AddDebug(thread.execReplyNode.debugString)
	}

	return p.writeError(err.GetMessage(), err.GetDebug())
}

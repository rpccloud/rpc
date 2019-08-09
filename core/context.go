package core

import (
	"fmt"
)

type rpcInnerContext struct {
	stream       *RPCStream
	serverThread *rpcThread
	clientThread *uint64
}

type rpcContext struct {
	inner *rpcInnerContext
}

type Context = *rpcContext

func (p *rpcContext) getCacheStream() *RPCStream {
	if p != nil && p.inner != nil {
		return p.inner.stream
	}
	return nil
}

func (p *rpcContext) close() bool {
	if p.inner != nil {
		p.inner = nil
		return true
	}
	return false
}

func (p *rpcContext) writeError(message string, debug string) Return {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			logger := serverThread.processor.logger
			logger.Error(NewRPCErrorWithDebug(message, debug))
			stream := serverThread.execStream
			stream.SetWritePos(13)
			stream.WriteBool(false)
			stream.WriteString(message)
			stream.WriteString(debug)
			serverThread.execSuccessful = false
		}
	}
	return nilReturn
}

// OK get success Return  by value
func (p *rpcContext) OK(value interface{}) Return {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			stream := serverThread.execStream
			stream.SetWritePos(13)
			stream.WriteBool(true)
			if stream.Write(value) == RPCStreamWriteOK {
				serverThread.execSuccessful = true
				return nilReturn
			} else {
				return p.writeError(
					"return type is error",
					GetStackString(1),
				)
			}
		}
	}
	return nilReturn
}

// Failed get failed Return by code and message
func (p *rpcContext) Error(a ...interface{}) Return {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			message := fmt.Sprint(a...)
			err := NewRPCErrorWithDebug(message, GetStackString(1))
			if serverThread.execEchoNode != nil &&
				serverThread.execEchoNode.debugString != "" {
				err.AddDebug(serverThread.execEchoNode.debugString)
			}
			return p.writeError(err.GetMessage(), err.GetDebug())
		}
	}
	return nilReturn
}

func (p *rpcContext) Errorf(format string, a ...interface{}) Return {
	if p.inner != nil {
		serverThread := p.inner.serverThread
		if serverThread != nil {
			message := fmt.Sprintf(format, a...)
			err := NewRPCErrorWithDebug(message, GetStackString(1))
			if serverThread.execEchoNode != nil &&
				serverThread.execEchoNode.debugString != "" {
				err.AddDebug(serverThread.execEchoNode.debugString)
			}
			return p.writeError(err.GetMessage(), err.GetDebug())
		}
	}
	return nilReturn
}

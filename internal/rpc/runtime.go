package rpc

import (
	"math"

	"github.com/rpccloud/rpc/internal/base"
)

// Runtime ...
type Runtime struct {
	id     uint64
	thread *rpcThread
}

func (p Runtime) lock() *rpcThread {
	if thread := p.thread; thread != nil {
		return thread.lock(p.id)
	}
	return nil
}

func (p Runtime) unlock() bool {
	if thread := p.thread; thread != nil {
		return thread.unlock(p.id)
	}

	return false
}

// Reply ...
func (p Runtime) Reply(value interface{}) Return {
	if thread := p.lock(); thread != nil {
		defer p.unlock()
		return thread.Write(value, 1, true)
	}

	base.PublishPanic(
		base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(base.GetFileLine(1)),
	)
	return emptyReturn
}

// Post ...
func (p Runtime) Post(endpoint string, message string, value Any) error {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

		gatewayID, sessionID, ok := base.DecryptSessionEndpoint(endpoint)
		if !ok {
			return base.ErrRuntimePostEndpoint
		}

		stream := NewStream()
		stream.SetKind(StreamKindRPCBoardCast)
		stream.SetGatewayID(gatewayID)
		stream.SetSessionID(sessionID)
		stream.WriteString(thread.GetExecServicePath() + "%" + message)
		if reason := stream.Write(value); reason != StreamWriteOK {
			stream.Release()
			return base.ErrUnsupportedValue.AddDebug(base.ConcatString(
				reason,
			))
		}

		thread.processor.streamReceiver.OnReceiveStream(stream)
		return nil
	}

	return base.ErrRuntimeIllegalInCurrentGoroutine.
		AddDebug(base.GetFileLine(1))
}

// Call ...
func (p Runtime) Call(target string, args ...interface{}) RTValue {
	if thread := p.lock(); thread != nil {
		defer p.unlock()
		frame := thread.top

		// make stream
		stream, err := MakeInternalRequestStream(
			frame.stream.HasStatusBitDebug(),
			frame.depth+1,
			target,
			frame.from,
			args...,
		)
		if err != nil {
			return RTValue{
				err: err.AddDebug(
					base.AddFileLine(thread.GetExecActionNodePath(), 1),
				),
			}
		}
		defer stream.Release()

		// switch thread frame and eval
		func() {
			thread.pushFrame()
			defer thread.popFrame()
			thread.Eval(stream, false)
		}()

		// return
		ret := p.parseResponseStream(stream)
		if ret.err != nil {
			ret.err = ret.err.AddDebug(
				base.AddFileLine(thread.GetExecActionNodePath(), 1),
			)
		}

		return ret
	}

	return RTValue{
		err: base.ErrRuntimeIllegalInCurrentGoroutine.
			AddDebug(base.GetFileLine(1)),
	}

}

// NewRTArray ...
func (p Runtime) NewRTArray(size int) RTArray {
	if p.lock() != nil {
		defer p.unlock()
		return newRTArray(p, size)
	}

	return RTArray{}
}

// NewRTMap ...
func (p Runtime) NewRTMap(size int) RTMap {
	if p.lock() != nil {
		defer p.unlock()
		return newRTMap(p, size)
	}

	return RTMap{}
}

// GetPostEndPoint ...
func (p Runtime) GetPostEndPoint() string {
	if thread := p.lock(); thread != nil {
		defer p.unlock()
		stream := thread.top.stream

		if ret, ok := base.EncryptSessionEndpoint(
			stream.GetGatewayID(),
			stream.GetSessionID(),
		); ok {
			return ret
		}
	}

	return ""
}

// GetServiceConfig ...
func (p Runtime) GetServiceConfig(key string) (Any, bool) {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

		if actionNode := thread.GetActionNode(); actionNode == nil {
			return nil, false
		} else if serviceNode := actionNode.service; serviceNode == nil {
			return nil, false
		} else {
			return serviceNode.GetConfig(key)
		}
	}

	return nil, false
}

// SetServiceConfig ...
func (p Runtime) SetServiceConfig(key string, value Any) bool {
	if thread := p.lock(); thread != nil {
		defer p.unlock()

		if actionNode := thread.GetActionNode(); actionNode == nil {
			return false
		} else if serviceNode := actionNode.service; serviceNode == nil {
			return false
		} else {
			return serviceNode.SetConfig(key, value)
		}
	}

	return false
}

func (p Runtime) parseResponseStream(stream *Stream) RTValue {
	stream.SetReadPos(streamPosBody)

	switch stream.GetKind() {
	case StreamKindRPCResponseOK:
		ret, _ := stream.ReadRTValue(p)
		return ret
	case StreamKindSystemErrorReport:
		fallthrough
	case StreamKindRPCResponseError:
		if errCode, err := stream.ReadUint64(); err != nil {
			return RTValue{err: err}
		} else if errCode == 0 {
			return RTValue{err: base.ErrStream}
		} else if errCode > math.MaxUint32 {
			return RTValue{err: base.ErrStream}
		} else if message, err := stream.ReadString(); err != nil {
			return RTValue{err: err}
		} else if !stream.IsReadFinish() {
			return RTValue{err: base.ErrStream}
		} else {
			return RTValue{err: base.NewError(uint32(errCode), message)}
		}
	default:
		return RTValue{err: base.ErrStream}
	}
}

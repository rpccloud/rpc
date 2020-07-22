package internal

import (
	"sync/atomic"
	"unsafe"
)

type Context = *ContextObject

type ContextObject struct {
	thread unsafe.Pointer
}

func (p *ContextObject) getThread() *rpcThread {
	if thread := (*rpcThread)(atomic.LoadPointer(&p.thread)); thread == nil {
		ReportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if node := thread.execReplyNode; node == nil {
		ReportPanic(
			NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(GetFileLine(2)),
		)
		return nil
	} else if !thread.IsDebug() {
		return thread
	} else if meta := node.replyMeta; meta == nil {
		ReportPanic(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(GetFileLine(0)),
		)
		return nil
	} else {
		codeSource := GetFileLine(2)
		switch meta.GetCheckStatus(codeSource) {
		case rpcReplyCheckStatusOK:
			return thread
		case rpcReplyCheckStatusError:
			ReportPanic(
				NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(codeSource),
			)
			return nil
		default:
			if thread.GetGoId() != CurrentGoroutineID() {
				meta.SetCheckError(codeSource)
				ReportPanic(
					NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(codeSource),
				)
				return nil
			} else {
				meta.SetCheckOK(codeSource)
				return thread
			}
		}
	}
}

func (p *ContextObject) stop() {
	if p == nil {
		ReportPanic(
			NewKernelError(ErrStringUnexpectedNil).AddDebug(GetFileLine(0)),
		)
	} else {
		atomic.StorePointer(&p.thread, nil)
	}
}

func (p *ContextObject) OK(value interface{}) Return {
	if p == nil {
		ReportPanic(
			NewReplyPanic(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// ReportPanic has already run
		return nilReturn
	} else {
		return thread.WriteOK(value, 2)
	}
}

func (p *ContextObject) Error(value error) Return {
	if p == nil {
		ReportPanic(
			NewReplyPanic(ErrStringUnexpectedNil).AddDebug(GetFileLine(1)),
		)
		return nilReturn
	} else if thread := p.getThread(); thread == nil {
		// ReportPanic has already run
		return nilReturn
	} else if err, ok := value.(Error); ok && err != nil {
		return thread.WriteError(
			err.AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else if value != nil {
		return thread.WriteError(
			NewReplyError(value.Error()).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	} else {
		return thread.WriteError(
			NewReplyError(ErrStringUnexpectedNil).
				AddDebug(AddFileLine(thread.GetExecReplyNodePath(), 1)),
		)
	}
}

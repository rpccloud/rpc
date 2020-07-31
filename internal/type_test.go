package internal

import (
	"errors"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestReturnObject_nilReturn(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(nilReturn).Equals(Return(nil))
}

func TestContextObject_nilContext(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	assert(nilContext).Equals(Context(nil))
}

func TestContextObject_getThread(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		fnInner := func(source string) {
			source1 = source
			o1 := ContextObject{thread: nil}
			assert(o1.getThread()).IsNil()
		}
		fnInner(GetFileLine(0))
		return ctx.OK(false)
	})).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source1),
	)

	// Test(2)
	source2 := ""
	assert(testRunOnContext(true, func(processor *Processor, ctx Context) Return {
		fnInner := func(source string) {
			source2 = source
			thread := getFakeThread(false)
			thread.processor = processor
			thread.execReplyNode = nil
			o1 := ContextObject{thread: unsafe.Pointer(thread)}
			assert(o1.getThread()).IsNil()
		}
		fnInner(GetFileLine(0))
		return ctx.OK(false)
	})).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source2),
	)

	// Test(3)
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		assert(ctx.getThread()).IsNotNil()
		return ctx.OK(true)
	})).Equals(true, nil, nil)

	// Test(4)
	source4 := ""
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		go func() {
			func(source string) {
				source4 = source
				assert(ctx.getThread()).IsNil()
			}(GetFileLine(0))
		}()
		time.Sleep(200 * time.Millisecond)
		return ctx.OK(false)
	})).Equals(
		false,
		nil,
		NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source4),
	)

	// Test(5)
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		assert(ctx.getThread()).IsNotNil()
		return ctx.OK(true)
	})).Equals(true, nil, nil)
}

func TestContextObject_stop(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	panic1 := testRunWithCatchPanic(func() {
		ret := Context(nil).stop()
		assert(ret).IsFalse()
	})
	assert(panic1).IsNotNil()
	assert(panic1.GetKind()).Equals(ErrorKindKernelPanic)
	assert(panic1.GetMessage()).Equals("rpc: object is nil")
	assert(strings.Contains(panic1.GetDebug(), "goroutine")).IsTrue()
	assert(strings.Contains(panic1.GetDebug(), "[running]")).IsTrue()

	// Test(2)
	ctx2 := getFakeContext(true)
	assert(ctx2.thread).IsNotNil()
	assert(ctx2.stop()).IsTrue()
	assert(ctx2.thread).IsNil()
}

func TestContextObject_OK(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithCatchPanic(func() {
		ret, source := Context(nil).OK(true), GetFileLine(0)
		source1 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic("rpc: context is nil").AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(testRunWithCatchPanic(func() {
		ret, source := (&ContextObject{thread: nil}).OK(true), GetFileLine(0)
		source2 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source2))

	// Test(3)
	assert(testRunOnContext(true, func(_ *Processor, ctx Context) Return {
		ret := ctx.OK(true)
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(true, nil, nil)
}

func TestContextObject_Error(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	source1 := ""
	assert(testRunWithCatchPanic(func() {
		ret, source := Context(nil).Error(errors.New("error")), GetFileLine(0)
		source1 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic("rpc: context is nil").AddDebug(source1))

	// Test(2)
	source2 := ""
	assert(testRunWithCatchPanic(func() {
		err := NewReplyError("error")
		ret, source := (&ContextObject{thread: nil}).Error(err), GetFileLine(0)
		source2 = source
		assert(ret).Equals(nilReturn)
	})).Equals(NewReplyPanic(ErrStringRunOutOfReplyScope).AddDebug(source2))

	// Test(3)
	source3 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(NewReplyError("error")), GetFileLine(0)
		source3 = ctx.getThread().GetReplyNode().path + " " + source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source3),
		nil,
	)

	// Test(4)
	source4 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(errors.New("error")), GetFileLine(0)
		source4 = ctx.getThread().GetReplyNode().path + " " + source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("error").AddDebug(source4),
		nil,
	)

	// Test(5)
	source5 := ""
	assert(testRunOnContext(false, func(_ *Processor, ctx Context) Return {
		ret, source := ctx.Error(nil), GetFileLine(0)
		source5 = source
		assert(ret).Equals(nilReturn)
		return ret
	})).Equals(
		nil,
		NewReplyError("rpc: Context.Error() argument should not nil").
			AddDebug("#.test:Eval "+source5),
		nil,
	)
}
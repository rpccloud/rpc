package core

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"testing"
)

func TestRuntime_lock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(Runtime{}.lock()).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := getFakeThread(false)
		assert(Runtime{id: thread.top.lockStatus, thread: thread}.lock()).
			Equal(thread)
	})
}

func TestRuntime_unlock(t *testing.T) {
	t.Run("thread is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			Runtime{}.unlock()
		})).IsNil()
	})

	t.Run("thread is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		thread := getFakeThread(false)
		rt := Runtime{id: thread.top.lockStatus, thread: thread}
		assert(base.RunWithCatchPanic(func() {
			rt.lock()
			rt.unlock()
		})).IsNil()
		// if unlock is success, the thread can lock again
		assert(rt.lock()).Equal(thread)
	})
}

func TestContextObject_OK(t *testing.T) {
	t.Run("thread lock error", func(t *testing.T) {
		assert := base.NewAssert(t)
		source := ""
		assert(base.RunWithSubscribePanic(func() {
			_, source = Runtime{}.OK(true), base.GetFileLine(0)
		})).Equal(
			errors.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(source),
		)
	})

	//
	//// Test(2) return value type error
	//source2 := ""
	//assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
	//	// return value type error
	//	ret, source := rt.OK(make(chan bool)), base.GetFileLine(0)
	//	source2 = source
	//	return ret
	//})).Equal(
	//	nil,
	//	nil,
	//	base.NewReplyPanic("value type is not supported").
	//		AddDebug("#.test:Eval "+source2),
	//)
	//
	//// Test(3) OK
	//assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
	//	return rt.OK(true)
	//})).Equal(true, nil, nil)
	//
	//// Test(4) rpcThreadReturnStatusAlreadyCalled
	//source4 := ""
	//assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
	//	// OK called twice
	//	rt.OK(true)
	//	ret, source := rt.OK(make(chan bool)), base.GetFileLine(0)
	//	source4 = source
	//	return ret
	//})).Equal(
	//	nil,
	//	nil,
	//	base.NewReplyPanic("Runtime.OK has been called before").
	//		AddDebug("#.test:Eval "+source4),
	//)
}

//func TestContextObject_Error(t *testing.T) {
//	assert := base.NewAssert(t)
//
//	// Test(1)
//	source1 := ""
//	assert(base.RunWithSubscribePanic(func() {
//		ret, source := Runtime{}.Error(errors.New("error")), base.GetFileLine(0)
//		source1 = source
//		assert(ret).Equal(emptyReturn)
//	})).Equal(base.NewReplyPanic(
//		"Runtime is illegal in current goroutine",
//	).AddDebug(source1))
//
//	// Test(2)
//	source2 := ""
//	assert(base.RunWithSubscribePanic(func() {
//		err := base.NewReplyError("error")
//		ret, source := (&Runtime{thread: nil}).Error(err), base.GetFileLine(0)
//		source2 = source
//		assert(ret).Equal(emptyReturn)
//	})).Equal(
//		base.NewReplyPanic("Runtime is illegal in current goroutine").AddDebug(source2),
//	)
//
//	// Test(3)
//	source3 := ""
//	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
//		ret, source := rt.Error(base.NewReplyError("error")), base.GetFileLine(0)
//		source3 = rt.thread.GetReplyNode().path + " " + source
//		assert(ret).Equal(emptyReturn)
//		return ret
//	})).Equal(
//		nil,
//		base.NewReplyError("error").AddDebug(source3),
//		nil,
//	)
//
//	// Test(4)
//	source4 := ""
//	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
//		ret, source := rt.Error(errors.New("error")), base.GetFileLine(0)
//		source4 = rt.thread.GetReplyNode().path + " " + source
//		assert(ret).Equal(emptyReturn)
//		return ret
//	})).Equal(
//		nil,
//		base.NewReplyError("error").AddDebug(source4),
//		nil,
//	)
//
//	// Test(5)
//	source5 := ""
//	assert(testRunOnContext(false, func(_ *Processor, rt Runtime) Return {
//		ret, source := rt.Error(nil), base.GetFileLine(0)
//		source5 = source
//		assert(ret).Equal(emptyReturn)
//		return ret
//	})).Equal(
//		nil,
//		base.NewReplyError("argument should not nil").
//			AddDebug("#.test:Eval "+source5),
//		nil,
//	)
//
//	// Test(6) rpcThreadReturnStatusAlreadyCalled
//	source6 := ""
//	assert(testRunOnContext(true, func(_ *Processor, rt Runtime) Return {
//		// OK called twice
//		rt.Error(errors.New("error"))
//		ret, source := rt.Error(errors.New("error")), base.GetFileLine(0)
//		source6 = source
//		return ret
//	})).Equal(
//		nil,
//		nil,
//		base.NewReplyPanic("Runtime.Error has been called before").
//			AddDebug("#.test:Eval "+source6),
//	)
//}

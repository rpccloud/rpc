package rpc

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
)

var (
	rpcThreadFrameCache = &sync.Pool{
		New: func() interface{} {
			return &rpcThreadFrame{
				stream:             nil,
				actionNode:         nil,
				from:               "",
				depth:              0,
				cacheArrayItemsPos: 0,
				cacheMapItemsPos:   0,
				cacheArrayEntryPos: 0,
				cacheMapEntryPos:   0,
				retStatus:          0,
				lockStatus:         0,
				parentRTWritePos:   streamPosBody,
				next:               nil,
			}
		},
	}
	zeroCacheEntry = [8]cacheEntry{}
)

type rpcThreadFrame struct {
	stream             *Stream
	actionNode         unsafe.Pointer
	from               string
	depth              uint16
	cacheArrayItemsPos uint32
	cacheMapItemsPos   uint32
	cacheArrayEntryPos uint32
	cacheMapEntryPos   uint32
	retStatus          uint32
	lockStatus         uint64
	parentRTWritePos   int
	next               *rpcThreadFrame
}

func newRPCThreadFrame() *rpcThreadFrame {
	return rpcThreadFrameCache.Get().(*rpcThreadFrame)
}

func (p *rpcThreadFrame) Reset() {
	p.stream = nil
	atomic.StorePointer(&p.actionNode, nil)
	p.from = ""
	p.cacheArrayItemsPos = 0
	p.cacheMapItemsPos = 0
	p.cacheArrayEntryPos = 0
	p.cacheMapEntryPos = 0
	p.parentRTWritePos = streamPosBody
	p.next = nil
}

func (p *rpcThreadFrame) Release() {
	p.Reset()
	rpcThreadFrameCache.Put(p)
}

type cacheEntry struct {
	arrayItems []posRecord
	mapItems   []mapItem
	mapLength  uint32
}

type rpcThread struct {
	processor       *Processor
	inputCH         chan *Stream
	closeCH         unsafe.Pointer
	closeTimeout    time.Duration
	top             *rpcThreadFrame
	rootFrame       rpcThreadFrame
	sequence        uint64
	rtStream        *Stream
	cacheEntry      [8]cacheEntry
	cacheArrayItems []posRecord
	cacheMapItems   []mapItem
	mu              sync.Mutex
}

func newRTArray(rt Runtime, size int) (ret RTArray) {
	if thread := rt.thread; thread != nil && size >= 0 {
		ret.rt = rt
		frame := thread.top

		if frame.cacheArrayEntryPos < 8 {
			ret.items = &thread.cacheEntry[frame.cacheArrayEntryPos].arrayItems
			frame.cacheArrayEntryPos++
		} else {
			ret.items = new([]posRecord)
		}

		end := int(frame.cacheArrayItemsPos) + size
		if end <= len(thread.cacheArrayItems) {
			*ret.items = thread.cacheArrayItems[frame.cacheArrayItemsPos:]
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			frame.cacheArrayItemsPos = uint32(end)
		} else {
			*ret.items = make([]posRecord, 0, size)
		}
	}

	return
}

func newRTMap(rt Runtime, size int) (ret RTMap) {
	if thread := rt.thread; thread != nil && size >= 0 {
		ret.rt = rt
		frame := thread.top

		if frame.cacheMapEntryPos < 8 {
			ret.items = &thread.cacheEntry[frame.cacheMapEntryPos].mapItems
			ret.length = &thread.cacheEntry[frame.cacheMapEntryPos].mapLength
			*ret.length = 0
			frame.cacheMapEntryPos++
		} else {
			ret.items = new([]mapItem)
			ret.length = new(uint32)
		}

		end := int(frame.cacheMapItemsPos) + size
		if end <= len(thread.cacheMapItems) {
			*ret.items = thread.cacheMapItems[frame.cacheMapItemsPos:]
			itemsHeader := (*reflect.SliceHeader)(unsafe.Pointer(ret.items))
			itemsHeader.Len = 0
			itemsHeader.Cap = size
			frame.cacheMapItemsPos = uint32(end)
		} else {
			*ret.items = make([]mapItem, 0, size)
		}
	}

	return
}

func newThread(
	processor *Processor,
	closeTimeout time.Duration,
	bufferSize uint32,
	onEvalFinish func(*rpcThread),
) *rpcThread {
	if processor == nil || onEvalFinish == nil {
		return nil
	}

	retCH := make(chan *rpcThread)

	timeout := closeTimeout
	if timeout < time.Second {
		timeout = time.Second
	}

	inputCH := make(chan *Stream)
	closeCH := make(chan bool, 1)
	go func() {
		numOfArray := int(bufferSize) / (2 * sizeOfPosRecord)
		numOfMap := int(bufferSize) / (2 * sizeOfMapItem)

		thread := &rpcThread{
			processor:       processor,
			inputCH:         inputCH,
			closeCH:         unsafe.Pointer(&closeCH),
			closeTimeout:    timeout,
			sequence:        1<<48 + (rand.Uint64()%(1<<56)/2)*2,
			rtStream:        NewStream(),
			cacheArrayItems: make([]posRecord, numOfArray),
			cacheMapItems:   make([]mapItem, numOfMap),
		}

		thread.top = &thread.rootFrame
		thread.top.parentRTWritePos = streamPosBody
		retCH <- thread

		for stream := <-inputCH; stream != nil; stream = <-inputCH {
			thread.Eval(stream, true)
			onEvalFinish(thread)
			thread.Reset()
		}

		closeCH <- true
	}()

	return <-retCH
}

func (p *rpcThread) Reset() {
	p.rtStream.Reset()
	p.rootFrame.Reset()
	p.cacheEntry = zeroCacheEntry
}

func (p *rpcThread) Close() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if chPtr := (*chan bool)(atomic.LoadPointer(&p.closeCH)); chPtr != nil {
		atomic.StorePointer(&p.closeCH, nil)
		close(p.inputCH)
		select {
		case <-*chPtr:
			p.rtStream.Release()
			p.rtStream = nil
			return true
		case <-time.After(p.closeTimeout):
			return false
		}
	}

	return false
}

func (p *rpcThread) lock(rtID uint64) *rpcThread {
	lockPtr := &p.top.lockStatus
	for !atomic.CompareAndSwapUint64(lockPtr, rtID, rtID+1) {
		curStatus := atomic.LoadUint64(lockPtr)

		if curStatus != rtID && curStatus != rtID+1 {
			return nil
		}

		if atomic.CompareAndSwapUint64(lockPtr, rtID, rtID+1) {
			return p
		}

		runtime.Gosched()
	}

	return p
}

func (p *rpcThread) unlock(rtID uint64) bool {
	return atomic.CompareAndSwapUint64(&p.top.lockStatus, rtID+1, rtID)
}

func (p *rpcThread) pushFrame() {
	frame := newRPCThreadFrame()
	frame.cacheArrayItemsPos = p.top.cacheArrayItemsPos
	frame.cacheMapItemsPos = p.top.cacheMapItemsPos
	frame.cacheArrayEntryPos = p.top.cacheArrayEntryPos
	frame.cacheMapEntryPos = p.top.cacheMapEntryPos
	frame.parentRTWritePos = p.rtStream.GetWritePos()
	frame.next = p.top
	p.top = frame
}

func (p *rpcThread) popFrame() {
	if frame := p.top.next; frame != nil {
		p.rtStream.SetWritePos(p.top.parentRTWritePos)
		p.top.Release()
		p.top = frame
	}
}

func (p *rpcThread) GetActionNode() *rpcActionNode {
	return (*rpcActionNode)(atomic.LoadPointer(&p.top.actionNode))
}

func (p *rpcThread) GetExecActionNodePath() string {
	if node := p.GetActionNode(); node != nil {
		return node.path
	}
	return ""
}

func (p *rpcThread) GetExecServicePath() string {
	if node := p.GetActionNode(); node != nil && node.service != nil {
		return node.service.path
	}
	return ""
}

func (p *rpcThread) GetExecActionDebug() string {
	if node := p.GetActionNode(); node != nil && node.meta != nil {
		return base.ConcatString(node.path, " ", node.meta.fileLine)
	}
	return ""
}

func (p *rpcThread) Write(value interface{}, skip uint, debug bool) Return {
	frame := p.top

	if frame.retStatus != 0 {
		value = base.ErrRuntimeReplyHasBeenCalled
	}

	stream := frame.stream
	stream.SetWritePosToBodyStart()
	writeErr := (*base.Error)(nil)

	if reason := stream.Write(value); reason == StreamWriteOK {
		stream.SetKind(StreamKindRPCResponseOK)
		frame.retStatus = 1
		return emptyReturn
	} else if err, ok := value.(*base.Error); ok {
		if err == nil {
			writeErr = base.ErrUnsupportedValue.AddDebug("value is nil")
		} else {
			writeErr = err
		}
	} else if e, ok := value.(error); ok && !base.IsNil(e) {
		writeErr = base.ErrAction.AddDebug(e.Error())
	} else if v, ok := value.(RTValue); ok && v.err != nil {
		writeErr = v.err
	} else {
		writeErr = base.ErrUnsupportedValue.AddDebug(reason)
	}

	frame.retStatus = 2
	stream.SetKind(StreamKindRPCResponseError)

	if debug {
		msg := writeErr.GetMessage()
		stream.WriteUint64(uint64(writeErr.GetCode()))
		if msg == "" {
			stream.WriteString(base.ConcatString(
				writeErr.GetMessage(),
				base.AddFileLine(p.GetExecActionNodePath(), skip+1),
			))
		} else {
			stream.WriteString(base.ConcatString(
				writeErr.GetMessage(),
				"\n",
				base.AddFileLine(p.GetExecActionNodePath(), skip+1),
			))
		}
	} else {
		stream.WriteUint64(uint64(writeErr.GetCode()))
		stream.WriteString(writeErr.GetMessage())
	}
	return emptyReturn
}

func (p *rpcThread) PutStream(stream *Stream) (ret bool) {
	if stream == nil {
		return false
	}

	defer func() {
		if v := recover(); v != nil {
			ret = false
		}
	}()

	if atomic.LoadPointer(&p.closeCH) != nil {
		p.inputCH <- stream
		return true
	}
	return false
}

func (p *rpcThread) Eval(inStream *Stream, needCallback bool) Return {
	timeStart := base.TimeNow()
	frame := p.top
	frame.stream = inStream
	p.sequence += 2
	rtID := p.sequence
	frame.lockStatus = rtID
	frame.retStatus = 0
	frame.depth = inStream.GetDepth()
	execActionNode := (*rpcActionNode)(nil)
	argErrorIndex := 0

	defer func() {
		if v := recover(); v != nil {
			// write runtime error
			p.Write(
				base.ErrActionPanic.
					AddDebug(fmt.Sprintf("runtime error: %v", v)).
					AddDebug(p.GetExecActionDebug()).
					AddDebug(string(debug.Stack())),
				0,
				false,
			)
		}

		defer func() {
			if v := recover(); v != nil {
				// kernel error
				base.PublishPanic(
					base.ErrThreadEvalFatal.AddDebug(fmt.Sprintf("%v", v)).
						AddDebug(string(debug.Stack())),
				)
			}
		}()

		for !atomic.CompareAndSwapUint64(&frame.lockStatus, rtID, rtID+2) {
			time.Sleep(10 * time.Millisecond)
		}

		if frame.retStatus == 0 {
			p.Write(
				base.ErrRuntimeExternalReturn.AddDebug(p.GetExecActionDebug()),
				0,
				false,
			)
		} else {
			// count
			if execActionNode != nil {
				execActionNode.indicator.Count(
					base.TimeNow().Sub(timeStart),
					frame.retStatus == 1,
				)
			}
		}

		// callback
		inStream.SetReadPosToBodyStart()

		if needCallback {
			p.processor.streamReceiver.OnReceiveStream(inStream)
		}
	}()

	// set exec action node
	actionPath, _, err := inStream.readUnsafeString()
	if err != nil {
		return p.Write(err, 0, false)
	} else if actionNode, ok := p.processor.actionsMap[actionPath]; !ok {
		return p.Write(
			base.ErrTargetNotExist.AddDebug(base.ConcatString(
				"rpc-call: ",
				actionPath,
				" does not exist",
			)),
			0,
			false,
		)
	} else {
		execActionNode = actionNode
		atomic.StorePointer(&frame.actionNode, unsafe.Pointer(execActionNode))
	}

	if frame.depth >= p.processor.maxCallDepth {
		return p.Write(
			base.ErrCallOverflow.
				AddDebug(base.ConcatString(
					"call ",
					actionPath,
					" level(",
					strconv.FormatUint(uint64(frame.depth), 10),
					") overflows",
				)),
			0,
			false,
		)
	} else if frame.from, _, err = inStream.readUnsafeString(); err != nil {
		return p.Write(err, 0, false)
	} else {
		// create context
		rt := Runtime{id: rtID, thread: p}

		if fnCache := execActionNode.cacheFN; fnCache != nil {
			argErrorIndex = fnCache(rt, inStream, execActionNode.meta.handler)
			if argErrorIndex == 0 {
				return emptyReturn
			}
		} else {
			args := make([]reflect.Value, 0, 8)
			args = append(args, reflect.ValueOf(rt))
			for i := 1; i < len(execActionNode.argTypes); i++ {
				var rv reflect.Value
				switch execActionNode.argTypes[i] {
				case int64Type:
					if v, err := inStream.ReadInt64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case uint64Type:
					if v, err := inStream.ReadUint64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case float64Type:
					if v, err := inStream.ReadFloat64(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case boolType:
					if v, err := inStream.ReadBool(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case stringType:
					if v, err := inStream.ReadString(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case bytesType:
					if v, err := inStream.ReadBytes(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case arrayType:
					if v, err := inStream.ReadArray(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case rtArrayType:
					if v, err := inStream.ReadRTArray(rt); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case mapType:
					if v, err := inStream.ReadMap(); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case rtMapType:
					if v, err := inStream.ReadRTMap(rt); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				case rtValueType:
					if v, err := inStream.ReadRTValue(rt); err == nil {
						rv = reflect.ValueOf(v)
					} else {
						argErrorIndex = i
					}
				default:
					argErrorIndex = i
				}

				if argErrorIndex != 0 {
					break
				} else {
					args = append(args, rv)
				}
			}

			if argErrorIndex == 0 {
				if inStream.IsReadFinish() {
					inStream.SetWritePosToBodyStart()
					execActionNode.reflectFn.Call(args)
					return emptyReturn
				}

				argErrorIndex = -1
			}
		}

		if val, err := inStream.Read(); err != nil {
			return p.Write(err, 0, false)
		} else if argErrorIndex < 0 {
			return p.Write(base.ErrStream, 0, false)
		} else if !inStream.HasStatusBitDebug() {
			return p.Write(
				base.ErrArgumentsNotMatch.
					AddDebug(base.ConcatString(
						"rpc-call: ",
						actionPath,
						" arguments does not match",
					)),
				0,
				false,
			)
		} else {
			var typeName string
			if val == nil {
				typeName = "<nil>"
			} else {
				typeName = convertTypeToString(reflect.ValueOf(val).Type())
			}

			return p.Write(
				base.ErrArgumentsNotMatch.AddDebug(
					fmt.Sprintf(
						"rpc-call: %s %s argument does not match. "+
							"want: %s got: %s\n%s",
						actionPath,
						base.ConvertOrdinalToString(uint(argErrorIndex)),
						convertTypeToString(
							execActionNode.argTypes[argErrorIndex],
						),
						typeName,
						p.GetExecActionDebug(),
					)),
				0,
				false,
			)
		}
	}
}

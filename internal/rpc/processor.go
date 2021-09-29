package rpc

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

const (
	rootName               = "#"
	freeGroups             = 256
	processorStatusClosed  = 0
	processorStatusRunning = 1
)

var (
	nodeNameRegex   = regexp.MustCompile(`^[_0-9a-zA-Z]+$`)
	actionNameRegex = regexp.MustCompile(
		`^([_a-zA-Z][_0-9a-zA-Z]*)` +
			`|` +
			`(\$onMount)|(\$onUnmount)|(\$onTimer)$`,
	)
	emptyEvalBack   = func(*Stream) {}
	emptyEvalFinish = func(*rpcThread) {}
)

type rpcActionNode struct {
	path       string
	meta       *ActionMeta
	service    *rpcServiceNode
	cacheFN    ActionCacheFunc
	reflectFn  reflect.Value
	callString string
	argTypes   []reflect.Type
	indicator  *base.PerformanceIndicator
}

type rpcServiceNode struct {
	path    string
	addMeta *ServiceMeta
	depth   uint16
	isMount bool
	data    Map
}

func (p *rpcServiceNode) GetConfig(key string) (Any, bool) {
	v, ok := p.data[key]
	return v, ok
}

func (p *rpcServiceNode) SetConfig(key string, value Any) bool {
	if p.data != nil {
		p.data[key] = value
		return true
	}

	return false
}

// Processor ...
type Processor struct {
	status            int32
	actionsMap        map[string]*rpcActionNode
	servicesMap       map[string]*rpcServiceNode
	maxNodeDepth      uint16
	maxCallDepth      uint16
	threads           []*rpcThread
	systemThread      *rpcThread
	freeCHArray       []chan *rpcThread
	readThreadPos     uint64
	writeThreadPos    uint64
	panicSubscription *base.PanicSubscription
	streamReceiver    IStreamReceiver
	closeCH           chan string
	mu                sync.Mutex
	muSystemInvoke    sync.Mutex
}

// NewProcessor ...
func NewProcessor(
	numOfThreads int,
	maxNodeDepth int16,
	maxCallDepth int16,
	threadBufferSize uint32,
	fnCache ActionCache,
	closeTimeout time.Duration,
	mountServices []*ServiceMeta,
	streamReceiver IStreamReceiver,
) *Processor {
	if streamReceiver == nil {
		panic("streamReceiver is nil")
	}

	if numOfThreads <= 0 {
		streamReceiver.OnReceiveStream(
			MakeSystemErrorStream(base.ErrNumOfThreadsIsWrong),
		)
		return nil
	} else if maxNodeDepth <= 0 {
		streamReceiver.OnReceiveStream(
			MakeSystemErrorStream(base.ErrMaxNodeDepthIsWrong),
		)
		return nil
	} else if maxCallDepth <= 0 {
		streamReceiver.OnReceiveStream(
			MakeSystemErrorStream(base.ErrProcessorMaxCallDepthIsWrong),
		)
		return nil
	} else {
		size := ((numOfThreads + freeGroups - 1) / freeGroups) * freeGroups
		ret := &Processor{
			status:         processorStatusRunning,
			actionsMap:     make(map[string]*rpcActionNode),
			servicesMap:    make(map[string]*rpcServiceNode),
			maxNodeDepth:   uint16(maxNodeDepth),
			maxCallDepth:   uint16(maxCallDepth),
			threads:        make([]*rpcThread, size),
			freeCHArray:    nil,
			readThreadPos:  0,
			writeThreadPos: 0,
			streamReceiver: streamReceiver,
			closeCH:        make(chan string),
		}

		// subscribe panic
		ret.panicSubscription = base.SubscribePanic(func(err *base.Error) {
			defer func() {
				_ = recover()
			}()
			streamReceiver.OnReceiveStream(MakeSystemErrorStream(err))
		})

		// init system thread
		ret.systemThread = newThread(
			ret,
			closeTimeout,
			threadBufferSize,
			emptyEvalFinish,
		)

		// mount nodes
		ret.servicesMap[rootName] = &rpcServiceNode{
			path:    rootName,
			addMeta: nil,
			depth:   0,
			data:    Map{},
		}

		for _, meta := range mountServices {
			if err := ret.mountNode(rootName, meta, fnCache); err != nil {
				streamReceiver.OnReceiveStream(MakeSystemErrorStream(err))
				return nil
			}
		}

		// start config update
		go func() {
			counter := uint64(0)
			evalTime := base.TimeNow()
			for atomic.LoadInt32(&ret.status) == processorStatusRunning {
				time.Sleep(50 * time.Millisecond)
				timeNow := base.TimeNow()
				if timeNow.Sub(evalTime) > time.Second {
					evalTime = timeNow
					counter++
					ret.onTimer(counter)
				}
			}
			ret.closeCH <- ""
		}()

		// start threads
		freeCHArray := make([]chan *rpcThread, freeGroups)
		for i := 0; i < freeGroups; i++ {
			freeCHArray[i] = make(chan *rpcThread, size/freeGroups)
		}
		ret.freeCHArray = freeCHArray

		for i := 0; i < size; i++ {
			thread := newThread(
				ret,
				closeTimeout,
				threadBufferSize,
				func(thread *rpcThread) {
					defer func() {
						_ = recover()
					}()
					freeCHArray[atomic.AddUint64(
						&ret.writeThreadPos,
						1,
					)%freeGroups] <- thread
				},
			)
			ret.threads[i] = thread
			ret.freeCHArray[i%freeGroups] <- thread
		}

		return ret
	}
}

// Close ...
func (p *Processor) Close() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.CompareAndSwapInt32(
		&p.status,
		processorStatusRunning,
		processorStatusClosed,
	) {
		// wait for config update thread finish
		<-p.closeCH

		// close worker threads
		for i := 0; i < len(p.threads); i++ {
			go func(idx int) {
				if p.threads[idx].Close() {
					p.closeCH <- ""
				} else {
					p.closeCH <- p.threads[idx].GetExecActionDebug()
				}
			}(i)
		}

		// wait all rpcThread close
		errMap := make(map[string]int)
		for i := 0; i < len(p.threads); i++ {
			if errString := <-p.closeCH; errString != "" {
				if v, ok := errMap[errString]; ok {
					errMap[errString] = v + 1
				} else {
					errMap[errString] = 1
				}
			}
		}

		errList := make([]string, 0)
		for k, v := range errMap {
			if v > 1 {
				errList = append(
					errList,
					fmt.Sprintf("%s (%d goroutines)", k, v),
				)
			} else {
				errList = append(
					errList,
					fmt.Sprintf("%s (%d goroutine)", k, v),
				)
			}
		}

		if len(errList) > 0 {
			p.streamReceiver.OnReceiveStream(MakeSystemErrorStream(
				base.ErrActionCloseTimeout.AddDebug(base.ConcatString(
					"the following actions can not close: \n\t",
					strings.Join(errList, "\n\t"),
				)),
			))
		}

		p.unmount("#")
		p.systemThread.Close()

		for _, freeCH := range p.freeCHArray {
			close(freeCH)
		}

		if p.panicSubscription != nil {
			p.panicSubscription.Close()
			p.panicSubscription = nil
		}

		return len(errList) == 0
	}

	return false
}

// PutStream ...
func (p *Processor) PutStream(stream *Stream) (ret bool) {
	defer func() {
		if v := recover(); v != nil {
			ret = false
		}
	}()

	if thread := <-p.freeCHArray[atomic.AddUint64(
		&p.readThreadPos,
		1,
	)%freeGroups]; thread != nil {
		success := thread.PutStream(stream)
		if !success {
			p.freeCHArray[atomic.AddUint64(
				&p.writeThreadPos,
				1,
			)%freeGroups] <- thread
		}
		return success
	}

	return false
}

// BuildCache ...
func (p *Processor) BuildCache(pkgName string, path string) *base.Error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt32(&p.status) == processorStatusRunning {
		retMap := make(map[string]bool)
		for _, action := range p.actionsMap {
			if fnTypeString, err := getFuncKind(action.reflectFn); err == nil {
				retMap[fnTypeString] = true
			}
		}

		fnKinds := make([]string, 0)
		for key := range retMap {
			fnKinds = append(fnKinds, key)
		}

		return buildFuncCache(pkgName, path, fnKinds)
	}

	return base.ErrProcessorIsNotRunning
}

func (p *Processor) onTimer(sequence uint64) {
	for key := range p.servicesMap {
		p.invokeSystemAction(key, "$onTimer", sequence)
	}
}

func (p *Processor) invokeSystemAction(
	path string,
	name string,
	args ...interface{},
) bool {
	p.muSystemInvoke.Lock()
	defer p.muSystemInvoke.Unlock()

	actionPath := path + ":" + name
	if _, ok := p.actionsMap[actionPath]; ok {
		stream, _ := MakeInternalRequestStream(true, 0, actionPath, "", args...)
		defer func() {
			stream.Release()
		}()
		p.systemThread.Eval(stream, false)
		return true
	}

	return false
}

func (p *Processor) mountNode(
	parentServiceNodePath string,
	nodeMeta *ServiceMeta,
	fnCache ActionCache,
) *base.Error {
	if nodeMeta == nil {
		return base.ErrProcessorNodeMetaIsNil
	} else if !nodeNameRegex.MatchString(nodeMeta.name) {
		return base.ErrServiceName.
			AddDebug(fmt.Sprintf("service name %s is illegal", nodeMeta.name)).
			AddDebug(nodeMeta.fileLine)
	} else if nodeMeta.service == nil {
		return base.ErrServiceIsNil.AddDebug(nodeMeta.fileLine)
	} else {
		parentNode := p.servicesMap[parentServiceNodePath]
		servicePath := parentServiceNodePath + "." + nodeMeta.name
		if parentNode.depth+1 > p.maxNodeDepth { // depth overflows
			return base.ErrServiceOverflow.
				AddDebug(fmt.Sprintf(
					"service path %s overflows "+
						"(max depth: %d, current depth:%d)",
					servicePath,
					p.maxNodeDepth, parentNode.depth+1,
				)).
				AddDebug(nodeMeta.fileLine)
		} else if item, ok := p.servicesMap[servicePath]; ok { // path exists
			return base.ErrServiceName.
				AddDebug(fmt.Sprintf(
					"duplicated service name %s",
					nodeMeta.name,
				)).
				AddDebug(fmt.Sprintf(
					"current:\n%s\nconflict:\n%s",
					base.AddPrefixPerLine(nodeMeta.fileLine, "\t"),
					base.AddPrefixPerLine(item.addMeta.fileLine, "\t"),
				))
		} else {
			node := &rpcServiceNode{
				path:    servicePath,
				addMeta: nodeMeta,
				depth:   parentNode.depth + 1,
				data:    Map{},
				isMount: false,
			}

			for k, v := range nodeMeta.data {
				node.data[k] = v
			}

			// mount the node
			p.servicesMap[servicePath] = node

			// mount the actions
			for _, actionMeta := range nodeMeta.service.actions {
				if err := p.mountAction(node, actionMeta, fnCache); err != nil {
					p.unmount(servicePath)
					return err
				}
			}

			// invoke onMount
			p.invokeSystemAction(servicePath, "$onMount")
			node.isMount = true

			// mount children
			for _, v := range nodeMeta.service.children {
				err := p.mountNode(servicePath, v, fnCache)
				if err != nil {
					p.unmount(servicePath)
					return err
				}
			}

			return nil
		}
	}
}

func (p *Processor) mountAction(
	serviceNode *rpcServiceNode,
	meta *ActionMeta,
	fnCache ActionCache,
) *base.Error {
	if meta == nil {
		return base.ErrProcessorActionMetaIsNil
	} else if !actionNameRegex.MatchString(meta.name) {
		return base.ErrActionName.
			AddDebug(fmt.Sprintf("action name %s is illegal", meta.name)).
			AddDebug(meta.fileLine)
	} else if meta.handler == nil {
		return base.ErrActionHandler.
			AddDebug("handler is nil").
			AddDebug(meta.fileLine)
	} else if fn := reflect.ValueOf(meta.handler); fn.Kind() != reflect.Func {
		return base.ErrActionHandler.AddDebug(fmt.Sprintf(
			"handler must be func(rt %s, ...) %s",
			convertTypeToString(runtimeType),
			convertTypeToString(returnType),
		)).AddDebug(meta.fileLine)
	} else if fnTypeString, err := getFuncKind(fn); err != nil {
		// Check action handler is right
		return err.AddDebug(meta.fileLine)
	} else {
		actionPath := serviceNode.path + ":" + meta.name
		if item, ok := p.actionsMap[actionPath]; ok {
			// check the action path is not occupied
			return base.ErrActionName.
				AddDebug(base.ConcatString(
					"duplicated action name ",
					meta.name,
				)).
				AddDebug(fmt.Sprintf(
					"current:\n%s\nconflict:\n%s",
					base.AddPrefixPerLine(meta.fileLine, "\t"),
					base.AddPrefixPerLine(item.meta.fileLine, "\t"),
				))
		}

		// mount the action
		numOfArgs := fn.Type().NumIn()
		argTypes := make([]reflect.Type, numOfArgs)
		argStrings := make([]string, numOfArgs)
		for i := 0; i < numOfArgs; i++ {
			argTypes[i] = fn.Type().In(i)
			argStrings[i] = convertTypeToString(argTypes[i])
		}

		actionNode := &rpcActionNode{
			path:      actionPath,
			meta:      meta,
			service:   serviceNode,
			cacheFN:   nil,
			reflectFn: fn,
			callString: fmt.Sprintf(
				"%s(%s) %s",
				actionPath,
				strings.Join(argStrings, ", "),
				convertTypeToString(returnType),
			),
			argTypes:  argTypes,
			indicator: base.NewPerformanceIndicator(),
		}

		if fnCache != nil {
			actionNode.cacheFN = fnCache.Get(fnTypeString)
		}

		p.actionsMap[actionPath] = actionNode
		return nil
	}
}

func (p *Processor) unmount(path string) {
	// invoke onUnmount
	for key, v := range p.servicesMap {
		if strings.HasPrefix(key, path) {
			if v.isMount {
				p.invokeSystemAction(key, "$onUnmount")
				v.isMount = false
			}
		}
	}

	// clean node actions or sibling actions
	for key := range p.actionsMap {
		if strings.HasPrefix(key, path) {
			delete(p.actionsMap, key)
		}
	}

	// clean node or sibling nodes
	for key := range p.servicesMap {
		if strings.HasPrefix(key, path) {
			delete(p.servicesMap, key)
		}
	}
}

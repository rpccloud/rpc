package server

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

// Channel ...
type Channel struct {
	sequence   uint64
	backTimeNS int64
	backStream *rpc.Stream
}

// In ...
func (p *Channel) In(id uint64) (canIn bool, backStream *rpc.Stream) {
	if id > p.sequence {
		p.Clean()
		p.sequence = id
		return true, nil
	} else if id == p.sequence {
		return false, p.backStream
	} else {
		return false, nil
	}
}

// Out ...
func (p *Channel) Out(stream *rpc.Stream) (canOut bool) {
	id := stream.GetCallbackID()

	if id == p.sequence {
		if p.backTimeNS == 0 {
			p.backTimeNS = base.TimeNow().UnixNano()
			p.backStream = stream
			return true
		}
		return false
	} else if id == 0 {
		return true
	} else {
		return false
	}
}

// IsTimeout ...
func (p *Channel) IsTimeout(nowNS int64, timeout int64) bool {
	return p.backTimeNS > 0 && nowNS-p.backTimeNS > timeout
}

// Clean ...
func (p *Channel) Clean() {
	p.backTimeNS = 0
	if p.backStream != nil {
		p.backStream.Release()
		p.backStream = nil
	}
}

// Session ...
type Session struct {
	id            uint64
	sessionServer *SessionServer
	security      string
	conn          *adapter.StreamConn
	channels      []Channel
	activeTimeNS  int64
	prev          *Session
	next          *Session
	sync.Mutex
}

// InitSession ...
func InitSession(
	sessionServer *SessionServer,
	streamConn *adapter.StreamConn,
	stream *rpc.Stream,
) {
	if stream.GetCallbackID() != 0 {
		stream.Release()
		sessionServer.OnConnError(streamConn, base.ErrStream)
	} else if kind := stream.GetKind(); kind != rpc.StreamKindConnectRequest {
		stream.Release()
		sessionServer.OnConnError(streamConn, base.ErrStream)
	} else if sessionString, err := stream.ReadString(); err != nil {
		stream.Release()
		sessionServer.OnConnError(streamConn, base.ErrStream)
	} else if !stream.IsReadFinish() {
		stream.Release()
		sessionServer.OnConnError(streamConn, base.ErrStream)
	} else {
		session := (*Session)(nil)
		config := sessionServer.config

		// try to find session by session string
		strArray := strings.Split(sessionString, "-")
		if len(strArray) == 2 && len(strArray[1]) == 32 {
			if id, err := strconv.ParseUint(strArray[0], 10, 64); err == nil {
				if s, ok := sessionServer.GetSession(id); ok && s.security == strArray[1] {
					session = s
				}
			}
		}

		// if session not find by session string, create a new session
		if session == nil {
			if sessionServer.TotalSessions() >= int64(config.serverMaxSessions) {
				stream.Release()
				sessionServer.OnConnError(streamConn, base.ErrServerSessionSeedOverflows)
				return
			}

			session = &Session{
				id:            sessionServer.CreateSessionID(),
				sessionServer: sessionServer,
				security:      base.GetRandString(32),
				conn:          nil,
				channels:      make([]Channel, config.numOfChannels),
				activeTimeNS:  base.TimeNow().UnixNano(),
				prev:          nil,
				next:          nil,
			}

			sessionServer.AddSession(session)
		}

		streamConn.SetReceiver(session)

		stream.SetWritePosToBodyStart()
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString(fmt.Sprintf("%d-%s", session.id, session.security))
		stream.WriteInt64(int64(config.numOfChannels))
		stream.WriteInt64(int64(config.transLimit))
		stream.WriteInt64(int64(config.heartbeatInterval / time.Millisecond))
		stream.WriteInt64(int64(config.heartbeatTimeout / time.Millisecond))
		streamConn.WriteStreamAndRelease(stream)

		session.OnConnOpen(streamConn)
	}
}

// TimeCheck ...
func (p *Session) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	if sessionServer := p.sessionServer; sessionServer != nil {
		config := sessionServer.config

		if p.conn != nil {
			// conn timeout
			if !p.conn.IsActive(nowNS, config.heartbeatTimeout) {
				p.conn.Close()
			}
		} else {
			// session timeout
			if nowNS-p.activeTimeNS > int64(config.serverSessionTimeout) {
				p.activeTimeNS = 0
			}
		}

		// channel timeout
		timeoutNS := int64(config.serverCacheTimeout)
		for i := 0; i < len(p.channels); i++ {
			if channel := &p.channels[i]; channel.IsTimeout(nowNS, timeoutNS) {
				channel.Clean()
			}
		}
	}
}

// OutStream ...
func (p *Session) OutStream(stream *rpc.Stream) {
	p.Lock()
	defer p.Unlock()

	if stream != nil {
		switch stream.GetKind() {
		case rpc.StreamKindRPCResponseOK:
			fallthrough
		case rpc.StreamKindRPCResponseError:
			// record stream
			callbackID := stream.GetCallbackID()
			channel := &p.channels[callbackID%uint64(len(p.channels))]
			if channel.Out(stream) && p.conn != nil {
				p.conn.WriteStreamAndRelease(stream.Clone())
			} else {
				stream.Release()
			}
		case rpc.StreamKindRPCBoardCast:
			p.conn.WriteStreamAndRelease(stream)
		default:
			stream.Release()
		}
	}
}

// OnConnOpen ...
func (p *Session) OnConnOpen(streamConn *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = streamConn
}

// OnConnReadStream ...
func (p *Session) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *rpc.Stream,
) {
	p.Lock()
	defer p.Unlock()

	switch stream.GetKind() {
	case rpc.StreamKindPing:
		if stream.IsReadFinish() {
			p.activeTimeNS = base.TimeNow().UnixNano()
			stream.SetKind(rpc.StreamKindPong)
			streamConn.WriteStreamAndRelease(stream)
		} else {
			p.OnConnError(streamConn, base.ErrStream)
			stream.Release()
		}
	case rpc.StreamKindRPCRequest:
		if cbID := stream.GetCallbackID(); cbID > 0 {
			channel := &p.channels[cbID%uint64(len(p.channels))]
			if accepted, backStream := channel.In(cbID); accepted {
				stream.SetSessionID(p.id)
				// who receives the stream is responsible for releasing it
				p.sessionServer.streamReceiver.OnReceiveStream(stream)
			} else if backStream != nil {
				// do not release the backStream, so we need to clone it
				streamConn.WriteStreamAndRelease(backStream.Clone())
				stream.Release()
			} else {
				// ignore the stream
				stream.Release()
			}
		} else {
			p.OnConnError(streamConn, base.ErrStream)
			stream.Release()
		}
	default:
		p.OnConnError(streamConn, base.ErrStream)
		stream.Release()
	}
}

// OnConnError ...
func (p *Session) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	errStream := rpc.MakeSystemErrorStream(err)
	errStream.SetSessionID(p.id)
	p.sessionServer.streamReceiver.OnReceiveStream(errStream)

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *Session) OnConnClose(_ *adapter.StreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = nil
}

// SessionPool ...
type SessionPool struct {
	sessionServer *SessionServer
	idMap         map[uint64]*Session
	head          *Session
	sync.Mutex
}

// NewSessionPool ...
func NewSessionPool(sessionServer *SessionServer) *SessionPool {
	return &SessionPool{
		sessionServer: sessionServer,
		idMap:         map[uint64]*Session{},
		head:          nil,
	}
}

// Get ...
func (p *SessionPool) Get(id uint64) (*Session, bool) {
	p.Lock()
	defer p.Unlock()

	ret, ok := p.idMap[id]
	return ret, ok
}

// Add ...
func (p *SessionPool) Add(session *Session) bool {
	p.Lock()
	defer p.Unlock()

	if session == nil {
		return false
	}

	if _, exist := p.idMap[session.id]; !exist {
		p.idMap[session.id] = session

		if p.head != nil {
			p.head.prev = session
		}

		session.prev = nil
		session.next = p.head
		p.head = session

		atomic.AddInt64(&p.sessionServer.totalSessions, 1)
		return true
	}

	return false
}

// TimeCheck ...
func (p *SessionPool) TimeCheck(nowNS int64) {
	p.Lock()
	defer p.Unlock()

	node := p.head
	for node != nil {
		node.TimeCheck(nowNS)

		// remove it from the list
		if node.activeTimeNS == 0 {
			delete(p.idMap, node.id)

			if node.prev != nil {
				node.prev.next = node.next
			}

			if node.next != nil {
				node.next.prev = node.prev
			}

			if node == p.head {
				p.head = node.next
			}

			atomic.AddInt64(&p.sessionServer.totalSessions, -1)
		}

		node = node.next
	}
}

// SessionServer ...
type SessionServer struct {
	isRunning      bool
	sessionSeed    uint64
	totalSessions  int64
	sessionMapList []*SessionPool
	streamReceiver rpc.IStreamReceiver
	closeCH        chan bool
	config         *SessionConfig
	adapters       []*adapter.Adapter
	orcManager     *base.ORCManager
	sync.Mutex
}

// SessionServer ...
func NewSessionServer(
	listeners []*listener,
	config *SessionConfig,
	streamReceiver rpc.IStreamReceiver,
) *SessionServer {
	if streamReceiver == nil {
		panic("streamReceiver is nil")
	}

	ret := &SessionServer{
		isRunning:      false,
		sessionSeed:    0,
		totalSessions:  0,
		sessionMapList: make([]*SessionPool, 1024),
		streamReceiver: streamReceiver,
		closeCH:        make(chan bool, 1),
		config:         config,
		adapters:       make([]*adapter.Adapter, len(listeners)),
		orcManager:     base.NewORCManager(),
	}

	for i := 0; i < 1024; i++ {
		ret.sessionMapList[i] = NewSessionPool(ret)
	}

	for i := 0; i < len(listeners); i++ {
		ret.adapters[i] = adapter.NewServerAdapter(
			listeners[i].isDebug,
			listeners[i].network,
			listeners[i].addr,
			listeners[i].tlsConfig,
			config.serverReadBufferSize,
			config.serverWriteBufferSize,
			ret,
		)
	}

	return ret
}

// TotalSessions ...
func (p *SessionServer) TotalSessions() int64 {
	return atomic.LoadInt64(&p.totalSessions)
}

// AddSession ...
func (p *SessionServer) AddSession(session *Session) bool {
	return p.sessionMapList[session.id%1024].Add(session)
}

// GetSession ...
func (p *SessionServer) GetSession(id uint64) (*Session, bool) {
	return p.sessionMapList[id%1024].Get(id)
}

// CreateSessionID ...
func (p *SessionServer) CreateSessionID() uint64 {
	return atomic.AddUint64(&p.sessionSeed, 1)
}

// TimeCheck ...
func (p *SessionServer) TimeCheck(nowNS int64) {
	for i := 0; i < 1024; i++ {
		p.sessionMapList[i].TimeCheck(nowNS)
	}
}

// Open ...
func (p *SessionServer) Open() {
	p.orcManager.Open(func() bool {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			p.streamReceiver.OnReceiveStream(
				rpc.MakeSystemErrorStream(base.ErrServerAlreadyRunning),
			)
			return false
		} else if len(p.adapters) <= 0 {
			p.streamReceiver.OnReceiveStream(
				rpc.MakeSystemErrorStream(base.ErrServerNoListenersAvailable),
			)
			return false
		} else {
			p.isRunning = true
			return true
		}
	})

	// -------------------------------------------------------------------------
	// Notice:
	//      if p.orcManager.Close() is called between Open and Run. Run will not
	// execute at all.
	// -------------------------------------------------------------------------
	p.orcManager.Run(func(isRunning func() bool) bool {
		waitCH := make(chan bool)
		waitCount := 0

		for _, item := range p.adapters {
			waitCount++
			item.Open()
			go func(adapter *adapter.Adapter) {
				adapter.Run()
				waitCH <- true
			}(item)
		}

		for isRunning() {
			startNS := base.TimeNow().UnixNano()
			p.TimeCheck(startNS)
			base.WaitWhileRunning(startNS, isRunning, time.Second)
		}

		for waitCount > 0 {
			<-waitCH
			waitCount--
		}

		return true
	})
}

// Close ...
func (p *SessionServer) Close() {
	p.orcManager.Close(func() bool {
		for _, item := range p.adapters {
			item.Close()
		}
		return true
	}, func() {
		p.Lock()
		defer p.Unlock()
		p.isRunning = false
	})
}

// OutStream ...
func (p *SessionServer) OutStream(stream *rpc.Stream) {
	if session, ok := p.GetSession(stream.GetSessionID()); ok {
		session.OutStream(stream)
	} else {
		errStream := rpc.MakeSystemErrorStream(base.ErrServerSessionNotFound)
		errStream.SetSessionID(stream.GetSessionID())
		p.streamReceiver.OnReceiveStream(errStream)
		stream.Release()
	}
}

// OnConnOpen ...
func (p *SessionServer) OnConnOpen(_ *adapter.StreamConn) {
	// ignore
	// we will add some security checks here
}

// OnConnReadStream ...
func (p *SessionServer) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *rpc.Stream,
) {
	InitSession(p, streamConn, stream)
}

// OnConnError ...
func (p *SessionServer) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(err))

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *SessionServer) OnConnClose(_ *adapter.StreamConn) {
	// ignore
	// streamConn is not attached to a session
	// If it happens multiple times on one ip, it may be an attack
}

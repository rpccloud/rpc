package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

func newSession(id uint64, sessionServer *SessionServer) *Session {
	return &Session{
		id:            id,
		sessionServer: sessionServer,
		security:      base.GetRandString(32),
		conn:          nil,
		channels:      make([]Channel, sessionServer.sessionConfig.numOfChannels),
		activeTimeNS:  base.TimeNow().UnixNano(),
		prev:          nil,
		next:          nil,
	}
}

type testNetConn struct {
	writeBuffer []byte
	isRunning   bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		isRunning:   true,
		writeBuffer: make([]byte, 0),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(b []byte) (n int, err error) {
	p.writeBuffer = append(p.writeBuffer, b...)
	return len(b), nil
}

func (p *testNetConn) Close() error {
	p.isRunning = false
	return nil
}

func (p *testNetConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (p *testNetConn) SetDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetReadDeadline(_ time.Time) error {
	panic("not implemented")
}

func (p *testNetConn) SetWriteDeadline(_ time.Time) error {
	panic("not implemented")
}

func prepareTestSession() (*Session, adapter.IConn, *testNetConn) {
	sessionServer := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
	session := newSession(11, sessionServer)
	netConn := newTestNetConn()
	syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
	streamConn := adapter.NewStreamConn(false, syncConn, session)
	syncConn.SetNext(streamConn)
	sessionServer.AddSession(session)
	return session, syncConn, netConn
}

func checkSessionList(head *Session) bool {
	if head == nil {
		return true
	}

	if head.prev != nil {
		return false
	}

	list := make([]*Session, 0)
	item := head
	for item != nil {
		list = append(list, item)
		item = item.next
	}

	idx := len(list) - 1
	item = list[idx]

	for idx >= 0 {
		if item == list[idx] {
			item = item.prev
			idx--
		} else {
			break
		}
	}

	return idx == -1 && item == nil
}

func testTimeCheck(pos int) bool {
	nowNS := base.TimeNow().UnixNano()
	v := NewSessionPool(&SessionServer{sessionConfig: GetDefaultSessionConfig()})

	s1 := &Session{id: 1, activeTimeNS: nowNS}
	s2 := &Session{id: 2, activeTimeNS: nowNS}
	s3 := &Session{id: 3, activeTimeNS: nowNS}
	v.Add(s3)
	v.Add(s2)
	v.Add(s1)

	var firstSession *Session

	switch pos {
	case 1:
		s1.activeTimeNS = 0
		firstSession = s2
	case 2:
		s2.activeTimeNS = 0
		firstSession = s1
	case 3:
		s3.activeTimeNS = 0
		firstSession = s1
	default:
		v.TimeCheck(nowNS)
		return v.sessionServer.totalSessions == 3 &&
			v.head == s1 &&
			checkSessionList(v.head)
	}

	v.TimeCheck(nowNS)
	return v.sessionServer.totalSessions == 2 &&
		v.head == firstSession &&
		checkSessionList(v.head)
}

func TestGetDefaultSessionConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		cfg := GetDefaultSessionConfig()
		assert(cfg.numOfChannels).Equal(32)
		assert(cfg.transLimit).Equal(4 * 1024 * 1024)
		assert(cfg.heartbeat).Equal(4 * time.Second)
		assert(cfg.heartbeatTimeout).Equal(8 * time.Second)
		assert(cfg.serverMaxSessions).Equal(10240000)
		assert(cfg.serverSessionTimeout).Equal(120 * time.Second)
		assert(cfg.serverReadBufferSize).Equal(1200)
		assert(cfg.serverWriteBufferSize).Equal(1200)
		assert(cfg.serverCacheTimeout).Equal(10 * time.Second)
	})
}

func TestChannel_In(t *testing.T) {
	t.Run("old id without back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream()}
		assert(v.In(0)).Equal(false, nil)
		assert(v.In(9)).Equal(false, nil)
	})

	t.Run("old id with back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream()}
		assert(v.In(10)).Equal(false, rpc.NewStream())
	})

	t.Run("new id", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream(), backTimeNS: 1}
		assert(v.In(11)).Equal(true, nil)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})
}

func TestChannel_Out(t *testing.T) {
	t.Run("id is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		assert(v.Out(stream)).Equal(true)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})

	t.Run("id equals sequence", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equal(true)
		assert(v.backTimeNS > 0, v.backStream != nil).Equal(true, true)
	})

	t.Run("id equals sequence, but not in", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(10), nil)
	})

	t.Run("id is wrong 01", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(9)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})

	t.Run("id is wrong 02", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(11)
		assert(v.Out(stream)).Equal(false)
		assert(v.backTimeNS, v.backStream).Equal(int64(0), nil)
	})
}

func TestChannel_IsTimeout(t *testing.T) {
	t.Run("backTimeNS is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 0, backStream: rpc.NewStream()}
		assert(v.IsTimeout(base.TimeNow().UnixNano(), int64(time.Second))).
			IsFalse()
	})

	t.Run("not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		nowNS := base.TimeNow().UnixNano()
		v := &Channel{
			sequence:   10,
			backTimeNS: nowNS,
			backStream: rpc.NewStream(),
		}
		assert(v.IsTimeout(nowNS, int64(100*time.Second))).IsFalse()
	})

	t.Run("timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		nowNS := base.TimeNow().UnixNano()
		v := &Channel{
			sequence:   10,
			backTimeNS: nowNS - 101,
			backStream: rpc.NewStream(),
		}
		assert(v.IsTimeout(nowNS, 100)).IsTrue()
	})
}

func TestChannel_Clean(t *testing.T) {
	t.Run("backStream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 1, backStream: nil}
		v.Clean()
		assert(v.sequence, v.backTimeNS, v.backStream).
			Equal(uint64(10), int64(0), nil)
	})

	t.Run("backStream is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 1, backStream: rpc.NewStream()}
		v.Clean()
		assert(v.sequence, v.backTimeNS, v.backStream).
			Equal(uint64(10), int64(0), nil)
	})
}

func TestInitSession(t *testing.T) {
	t.Run("stream callbackID != 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			sessionServer,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.SetCallbackID(1)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("kind is not StreamKindConnectRequest", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			sessionServer,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("read session string error", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			sessionServer,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteBool(true)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("read stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)

		streamConn := adapter.NewStreamConn(
			false,
			adapter.NewServerSyncConn(netConn, 1200, 1200),
			sessionServer,
		)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.WriteBool(false)
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("max sessions limit", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		sessionServer.sessionConfig.serverMaxSessions = 1
		sessionServer.AddSession(&Session{
			id:       234,
			security: "12345678123456781234567812345678",
		})

		syncConn := adapter.NewServerSyncConn(newTestNetConn(), 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, sessionServer)
		syncConn.SetNext(streamConn)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())

		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGateWaySeedOverflows)
	})

	t.Run("stream is ok, create new session", func(t *testing.T) {
		assert := base.NewAssert(t)
		id := uint64(234)
		security := "12345678123456781234567812345678"
		testCollection := map[string]bool{
			"234-12345678123456781234567812345678":   true,
			"0234-12345678123456781234567812345678":  true,
			"":                                       false,
			"-":                                      false,
			"-S":                                     false,
			"-SecurityPasswordSecurityPass":          false,
			"-SecurityPasswordSecurityPassword":      false,
			"-SecurityPasswordSecurityPasswordEx":    false,
			"*-S":                                    false,
			"*-SecurityPasswordSecurityPassword":     false,
			"*-SecurityPasswordSecurityPasswordEx":   false,
			"ABC-S":                                  false,
			"ABC-SecurityPasswordSecurityPassword":   false,
			"ABC-SecurityPasswordSecurityPasswordEx": false,
			"1-S":                                    false,
			"1-SecurityPasswordSecurityPassword":     false,
			"1-SecurityPasswordSecurityPasswordEx":   false,
			"-234-SecurityPasswordSecurityPassword":  false,
			"234-":                                   false,
			"234-S":                                  false,
			"234-SecurityPasswordSecurityPassword":   false,
			"234-SecurityPasswordSecurityPasswordEx": false,
			"-234-":                                  false,
			"-234-234-":                              false,
			"------":                                 false,
		}

		for connStr, exist := range testCollection {
			sessionServer := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
			sessionServer.AddSession(&Session{id: id, security: security, sessionServer: sessionServer})
			netConn := newTestNetConn()
			syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
			streamConn := adapter.NewStreamConn(false, syncConn, sessionServer)
			syncConn.SetNext(streamConn)

			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindConnectRequest)
			stream.WriteString(connStr)
			stream.BuildStreamCheck()
			streamConn.OnReadBytes(stream.GetBuffer())

			if exist {
				assert(sessionServer.totalSessions).Equal(int64(1))
			} else {
				assert(sessionServer.totalSessions).Equal(int64(2))
				v, _ := sessionServer.GetSession(1)
				assert(v.id).Equal(uint64(1))
				assert(v.sessionServer).Equal(sessionServer)
				assert(len(v.security)).Equal(32)
				assert(v.conn).IsNotNil()
				assert(len(v.channels)).Equal(GetDefaultSessionConfig().numOfChannels)
				assert(cap(v.channels)).Equal(GetDefaultSessionConfig().numOfChannels)
				nowNS := base.TimeNow().UnixNano()
				assert(nowNS-v.activeTimeNS < int64(time.Second)).IsTrue()
				assert(nowNS-v.activeTimeNS >= 0).IsTrue()
				assert(v.prev).IsNil()
				assert(v.next).IsNil()

				rs := rpc.NewStream()
				rs.PutBytesTo(netConn.writeBuffer, 0)

				sessionConfig := sessionServer.sessionConfig

				assert(rs.GetKind()).
					Equal(uint8(rpc.StreamKindConnectResponse))
				assert(rs.ReadString()).
					Equal(fmt.Sprintf("%d-%s", v.id, v.security), nil)
				assert(rs.ReadInt64()).Equal(int64(sessionConfig.numOfChannels), nil)
				assert(rs.ReadInt64()).Equal(int64(sessionConfig.transLimit), nil)
				assert(rs.ReadInt64()).
					Equal(int64(sessionConfig.heartbeat/time.Millisecond), nil)
				assert(rs.ReadInt64()).
					Equal(int64(sessionConfig.heartbeatTimeout/time.Millisecond), nil)
				assert(rs.IsReadFinish()).IsTrue()
				assert(rs.CheckStream()).IsTrue()
			}
		}
	})
}

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		sessionServer := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v := newSession(3, sessionServer)
		assert(v.id).Equal(uint64(3))
		assert(v.sessionServer).Equal(sessionServer)
		assert(len(v.security)).Equal(32)
		assert(v.conn).IsNil()
		assert(len(v.channels)).Equal(GetDefaultSessionConfig().numOfChannels)
		assert(cap(v.channels)).Equal(GetDefaultSessionConfig().numOfChannels)
		assert(base.TimeNow().UnixNano()-v.activeTimeNS < int64(time.Second)).
			IsTrue()
		assert(v.prev).IsNil()
		assert(v.next).IsNil()
	})
}

func TestSession_TimeCheck(t *testing.T) {
	t.Run("p.conn is active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		session.sessionServer.sessionConfig.heartbeatTimeout = 100 * time.Millisecond
		syncConn.OnOpen()
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsTrue()
		assert(session.conn).IsNotNil()
	})

	t.Run("p.conn is not active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		session.sessionServer.sessionConfig.heartbeatTimeout = 1 * time.Millisecond
		syncConn.OnOpen()
		time.Sleep(30 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("p.conn is nil, session is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()
		session.sessionServer.sessionConfig.serverSessionTimeout = time.Second
		session.sessionServer.TimeCheck(base.TimeNow().UnixNano())
		assert(session.sessionServer.TotalSessions()).Equal(int64(1))
	})

	t.Run("p.conn is nil, session is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()
		session.sessionServer.sessionConfig.serverSessionTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
		session.sessionServer.TimeCheck(base.TimeNow().UnixNano())
		assert(session.sessionServer.TotalSessions()).Equal(int64(0))
	})

	t.Run("p.channels is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()

		// fill the channels
		for i := 0; i < session.sessionServer.sessionConfig.numOfChannels; i++ {
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}

		session.sessionServer.sessionConfig.serverCacheTimeout = 10 * time.Millisecond
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.sessionServer.sessionConfig.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}
	})

	t.Run("p.channels is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession()

		// fill the channels
		for i := 0; i < session.sessionServer.sessionConfig.numOfChannels; i++ {
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
		}

		session.sessionServer.sessionConfig.serverCacheTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.sessionServer.sessionConfig.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS).Equal(int64(0))
			assert(session.channels[i].backStream).IsNil()
		}
	})
}

func TestSession_OutStream(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession()
		session.OutStream(nil)
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("p.conn is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession()
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		session.OutStream(stream)
		assert(len(netConn.writeBuffer)).Equal(0)
		assert(session.channels[0].backStream).Equal(stream)
	})

	t.Run("stream kind error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := 1; i <= len(session.channels); i++ {
			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindConnectResponse)
			stream.SetCallbackID(uint64(i))
			session.OutStream(stream)
		}

		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("stream can not out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := len(session.channels) + 1; i <= 2*len(session.channels); i++ {
			stream := rpc.NewStream()
			stream.SetKind(rpc.StreamKindRPCResponseOK)
			stream.SetCallbackID(uint64(i))
			session.OutStream(stream)
		}

		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("stream can out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		exceptBuffer := make([]byte, 0)
		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := 1; i <= len(session.channels); i++ {
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i))
			stream.SetKind(rpc.StreamKindRPCResponseOK)
			stream.BuildStreamCheck()
			exceptBuffer = append(exceptBuffer, stream.GetBuffer()...)
			session.OutStream(stream)
		}

		assert(netConn.writeBuffer).Equal(exceptBuffer)
	})

	t.Run("stream is StreamKindRPCBoardCast", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		// ignore the init stream
		netConn.writeBuffer = make([]byte, 0)

		exceptBuffer := make([]byte, 0)
		for i := 1; i <= len(session.channels); i++ {
			(&session.channels[i%len(session.channels)]).In(uint64(i))
		}

		for i := 1; i <= len(session.channels); i++ {
			stream := rpc.NewStream()
			stream.SetSessionID(5678)
			stream.SetKind(rpc.StreamKindRPCBoardCast)
			stream.WriteString("#.test%Msg")
			stream.WriteString("HI")
			stream.BuildStreamCheck()
			exceptBuffer = append(exceptBuffer, stream.GetBuffer()...)
			session.OutStream(stream)
		}

		assert(netConn.writeBuffer).Equal(exceptBuffer)
	})
}

func TestSession_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		session.OnConnOpen(streamConn)
		assert(session.conn).Equal(streamConn)
	})
}

func TestSession_OnConnReadStream(t *testing.T) {
	t.Run("cbID == 0, kind == StreamKindPing ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		sendStream := rpc.NewStream()
		sendStream.SetKind(rpc.StreamKindPing)
		session.OnConnReadStream(streamConn, sendStream)
		backStream := rpc.NewStream()
		backStream.PutBytesTo(netConn.writeBuffer, 0)
		assert(backStream.GetKind()).Equal(uint8(rpc.StreamKindPong))
		assert(backStream.IsReadFinish()).IsTrue()
	})

	t.Run("cbID == 0, kind == StreamKindPing error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindPing)
		stream.WriteBool(true)

		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver
		session.OnConnReadStream(streamConn, stream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("cbID > 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)

		streamReceiver := rpc.NewTestStreamReceiver()
		session, syncConn, _ := prepareTestSession()
		session.sessionServer.streamReceiver = streamReceiver

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		stream.SetKind(rpc.StreamKindRPCRequest)
		session.OnConnReadStream(streamConn, stream)

		backStream := streamReceiver.GetStream()
		assert(backStream.GetSessionID()).Equal(uint64(11))
	})

	t.Run("cbID > 0, accept = false, backStream != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		cacheStream := rpc.NewStream()
		cacheStream.SetCallbackID(10)
		cacheStream.BuildStreamCheck()

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)

		(&session.channels[10%len(session.channels)]).In(10)
		(&session.channels[10%len(session.channels)]).Out(cacheStream)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(netConn.writeBuffer).Equal(cacheStream.GetBuffer())
	})

	t.Run("cbID > 0, accept = false, backStream == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession()

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		(&session.channels[10%len(session.channels)]).In(10)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeBuffer)).Equal(0)
	})

	t.Run("cbID == 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindRPCRequest)

		session.OnConnReadStream(streamConn, stream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})

	t.Run("cbID == 0, kind err", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver
		session.OnConnReadStream(streamConn, rpc.NewStream())
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})
}

func TestSession_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver
		session.OnConnError(streamConn, base.ErrStream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})
}

func TestSession_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession()
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		session.OnConnOpen(streamConn)
		assert(session.conn).IsNotNil()
		session.OnConnClose(streamConn)
		assert(session.conn).IsNil()
	})
}

func TestNewSessionPool(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		sessionServer := &SessionServer{}
		v := NewSessionPool(sessionServer)
		assert(v.sessionServer).Equal(sessionServer)
		assert(len(v.idMap)).Equal(0)
		assert(v.head).IsNil()
	})
}

func TestSessionPool_Add(t *testing.T) {
	t.Run("session is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		assert(v.Add(nil)).IsFalse()
	})

	t.Run("session has already exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session := &Session{id: 10}
		assert(v.Add(session)).IsTrue()
		assert(v.Add(session)).IsFalse()
		assert(v.sessionServer.totalSessions).Equal(int64(1))
	})

	t.Run("add one session", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session := &Session{id: 10}
		assert(v.Add(session)).IsTrue()
		assert(v.sessionServer.totalSessions).Equal(int64(1))
		assert(v.head).Equal(session)
		assert(session.prev).Equal(nil)
		assert(session.next).Equal(nil)
	})

	t.Run("add two sessions", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session1 := &Session{id: 11}
		session2 := &Session{id: 12}
		assert(v.Add(session1)).IsTrue()
		assert(v.Add(session2)).IsTrue()
		assert(v.sessionServer.totalSessions).Equal(int64(2))
		assert(v.head).Equal(session2)
		assert(session2.prev).Equal(nil)
		assert(session2.next).Equal(session1)
		assert(session1.prev).Equal(session2)
		assert(session1.next).Equal(nil)
	})
}

func TestSessionPool_Get(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session1 := &Session{id: 11}
		session2 := &Session{id: 12}
		assert(v.Add(session1)).IsTrue()
		assert(v.Add(session2)).IsTrue()
		assert(v.sessionServer.totalSessions).Equal(int64(2))
		assert(v.Get(11)).Equal(session1, true)
		assert(v.Get(12)).Equal(session2, true)
		assert(v.Get(10)).Equal(nil, false)
	})
}

func TestSessionPool_TimeCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(testTimeCheck(0)).IsTrue()
		assert(testTimeCheck(1)).IsTrue()
		assert(testTimeCheck(2)).IsTrue()
		assert(testTimeCheck(3)).IsTrue()
	})
}

func TestSessionServerBasic(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(sessionManagerVectorSize).Equal(1024)
	})
}

func TestNewSessionServer(t *testing.T) {
	t.Run("streamReceiver is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewSessionServer(GetDefaultSessionConfig(), nil)
		})).Equal("streamReceiver is nil")
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		assert(v.isRunning).Equal(false)
		assert(v.sessionSeed).Equal(uint64(0))
		assert(v.totalSessions).Equal(int64(0))
		assert(len(v.sessionMapList)).Equal(sessionManagerVectorSize)
		assert(cap(v.sessionMapList)).Equal(sessionManagerVectorSize)
		assert(v.streamReceiver).Equal(streamReceiver)
		assert(len(v.closeCH)).Equal(0)
		assert(cap(v.closeCH)).Equal(1)
		assert(v.sessionConfig).Equal(GetDefaultSessionConfig())
		assert(len(v.adapters)).Equal(0)
		assert(cap(v.adapters)).Equal(0)
		assert(v.orcManager).IsNotNil()

		for i := 0; i < sessionManagerVectorSize; i++ {
			assert(v.sessionMapList[i]).IsNotNil()
		}
	})
}

func TestSessionServer_TotalSessions(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.totalSessions = 54321
		assert(v.TotalSessions()).Equal(int64(54321))
	})
}

func TestSessionServer_AddSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsFalse()
		}
	})
}

func TestSessionServer_GetSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())

		for i := uint64(1); i < 100; i++ {
			session := newSession(i, v)
			assert(v.AddSession(session)).IsTrue()
		}

		for i := uint64(1); i < 100; i++ {
			s, ok := v.GetSession(i)
			assert(s).IsNotNil()
			assert(ok).IsTrue()
		}

		for i := uint64(100); i < 200; i++ {
			s, ok := v.GetSession(i)
			assert(s).IsNil()
			assert(ok).IsFalse()
		}
	})
}

func TestSessionServer_CreateSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		assert(v.CreateSessionID()).Equal(uint64(1))
		assert(v.CreateSessionID()).Equal(uint64(2))
	})
}

func TestSessionServer_TimeCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		for i := uint64(1); i <= sessionManagerVectorSize; i++ {
			session := newSession(i, v)
			session.activeTimeNS = 0
			assert(v.AddSession(session)).IsTrue()
		}

		assert(v.TotalSessions()).Equal(int64(sessionManagerVectorSize))
		v.TimeCheck(base.TimeNow().UnixNano())
		assert(v.TotalSessions()).Equal(int64(0))
	})
}

func TestSessionServer_Listen(t *testing.T) {
	t.Run("SessionServer is running", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.isRunning = true
		assert(v.Listen("tcp", "0.0.0.0:8080", nil)).Equal(v)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGatewayAlreadyRunning)
	})

	t.Run("SessionServer is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		tlsConfig := &tls.Config{}
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		assert(v.Listen("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
		assert(len(v.adapters)).Equal(1)
		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
			false,
			"tcp",
			"0.0.0.0:8080",
			tlsConfig,
			v.sessionConfig.serverReadBufferSize,
			v.sessionConfig.serverWriteBufferSize,
			v,
		))
	})
}

func TestSessionServer_ListenWithDebug(t *testing.T) {
	t.Run("SessionServer is running", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.isRunning = true
		assert(v.ListenWithDebug("tcp", "0.0.0.0:8080", nil)).Equal(v)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGatewayAlreadyRunning)
	})

	t.Run("SessionServer is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		tlsConfig := &tls.Config{}
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		assert(v.ListenWithDebug("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
		assert(len(v.adapters)).Equal(1)
		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
			true,
			"tcp",
			"0.0.0.0:8080",
			tlsConfig,
			v.sessionConfig.serverReadBufferSize,
			v.sessionConfig.serverWriteBufferSize,
			v,
		))
	})
}

func TestSessionServer_Open(t *testing.T) {
	t.Run("it is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.isRunning = true
		v.Open()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGatewayAlreadyRunning)
	})

	t.Run("no valid adapter", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.Open()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGatewayNoAvailableAdapter)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(&Session{id: 10})
		v.Listen("tcp", "127.0.0.1:8000", nil)
		v.Listen("tcp", "127.0.0.1:8001", nil)

		go func() {
			for v.TotalSessions() == 1 {
				time.Sleep(10 * time.Millisecond)
			}
			assert(v.isRunning).IsTrue()
			_, err1 := net.Listen("tcp", "127.0.0.1:8000")
			_, err2 := net.Listen("tcp", "127.0.0.1:8001")
			assert(err1).IsNotNil()
			assert(err2).IsNotNil()
			v.Close()
			waitCH <- true
		}()
		assert(v.TotalSessions()).Equal(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestSessionServer_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(&Session{id: 10})
		v.Listen("tcp", "127.0.0.1:8000", nil)

		go func() {
			for v.TotalSessions() == 1 {
				time.Sleep(10 * time.Millisecond)
			}
			assert(v.isRunning).IsTrue()
			v.Close()
			assert(v.isRunning).IsFalse()
			ln1, err1 := net.Listen("tcp", "127.0.0.1:8000")
			ln2, err2 := net.Listen("tcp", "127.0.0.1:8001")
			assert(err1).IsNil()
			assert(err2).IsNil()
			_ = ln1.Close()
			_ = ln2.Close()
			waitCH <- true
		}()

		assert(v.TotalSessions()).Equal(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestSessionServer_ReceiveStreamFromRouter(t *testing.T) {
	t.Run("session is exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		stream := rpc.NewStream()
		stream.SetSessionID(10)
		v.OutStream(stream)
		assert(streamReceiver.GetStream()).IsNil()
	})

	t.Run("session is not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		stream := rpc.NewStream()
		stream.SetSessionID(11)
		v.OutStream(stream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrGateWaySessionNotFound)
	})
}

func TestSessionServer_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnOpen(nil)
		})).IsNil()
	})
}

func TestSessionServer_OnConnReadStream(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		syncConn := adapter.NewServerSyncConn(newTestNetConn(), 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(v.totalSessions).Equal(int64(1))
	})
}

func TestSessionServer_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		netConn := newTestNetConn()
		syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		assert(netConn.isRunning).IsTrue()
		v.OnConnError(streamConn, base.ErrStream)
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equal(nil, base.ErrStream)
	})
}

func TestSessionServer_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnClose(nil)
		})).IsNil()
	})
}

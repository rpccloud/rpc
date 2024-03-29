package server

import (
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
		channels:      make([]Channel, sessionServer.config.numOfChannels),
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

func prepareTestSession(
	listeners []*listener,
) (*Session, adapter.IConn, *testNetConn) {
	sessionServer := NewSessionServer(
		listeners,
		GetDefaultSessionConfig(),
		rpc.NewTestStreamReceiver(),
	)
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
	v := NewSessionPool(&SessionServer{config: GetDefaultSessionConfig()})

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

func TestChannel_In(t *testing.T) {
	t.Run("old id without back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream()}
		assert(v.In(0)).Equals(false, nil)
		assert(v.In(9)).Equals(false, nil)
	})

	t.Run("old id with back stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream()}
		assert(v.In(10)).Equals(false, rpc.NewStream())
	})

	t.Run("new id", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backStream: rpc.NewStream(), backTimeNS: 1}
		assert(v.In(11)).Equals(true, nil)
		assert(v.backTimeNS, v.backStream).Equals(int64(0), nil)
	})
}

func TestChannel_Out(t *testing.T) {
	t.Run("id is zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		assert(v.Out(stream)).Equals(true)
		assert(v.backTimeNS, v.backStream).Equals(int64(0), nil)
	})

	t.Run("id equals sequence", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equals(true)
		assert(v.backTimeNS > 0, v.backStream != nil).Equals(true, true)
	})

	t.Run("id equals sequence, but not in", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		assert(v.Out(stream)).Equals(false)
		assert(v.backTimeNS, v.backStream).Equals(int64(10), nil)
	})

	t.Run("id is wrong 01", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(9)
		assert(v.Out(stream)).Equals(false)
		assert(v.backTimeNS, v.backStream).Equals(int64(0), nil)
	})

	t.Run("id is wrong 02", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10}
		stream := rpc.NewStream()
		stream.SetCallbackID(11)
		assert(v.Out(stream)).Equals(false)
		assert(v.backTimeNS, v.backStream).Equals(int64(0), nil)
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
			Equals(uint64(10), int64(0), nil)
	})

	t.Run("backStream is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 10, backTimeNS: 1, backStream: rpc.NewStream()}
		v.Clean()
		assert(v.sequence, v.backTimeNS, v.backStream).
			Equals(uint64(10), int64(0), nil)
	})
}

func TestInitSession(t *testing.T) {
	t.Run("stream callbackID != 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), streamReceiver,
		)

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
			Equals(nil, base.ErrStream)
	})

	t.Run("kind is not StreamKindConnectRequest", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), streamReceiver,
		)

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
			Equals(nil, base.ErrStream)
	})

	t.Run("read session string error", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), streamReceiver,
		)

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
			Equals(nil, base.ErrStream)
	})

	t.Run("read stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		netConn := newTestNetConn()
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), streamReceiver,
		)

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
			Equals(nil, base.ErrStream)
	})

	t.Run("max sessions limit", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), streamReceiver,
		)
		sessionServer.config.serverMaxSessions = 1
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
			Equals(nil, base.ErrServerSessionSeedOverflows)
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
			sessionServer := NewSessionServer(
				nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
			)
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
				assert(sessionServer.totalSessions).Equals(int64(1))
			} else {
				assert(sessionServer.totalSessions).Equals(int64(2))
				v, _ := sessionServer.GetSession(1)
				assert(v.id).Equals(uint64(1))
				assert(v.sessionServer).Equals(sessionServer)
				assert(len(v.security)).Equals(32)
				assert(v.conn).IsNotNil()
				assert(len(v.channels)).Equals(GetDefaultSessionConfig().numOfChannels)
				assert(cap(v.channels)).Equals(GetDefaultSessionConfig().numOfChannels)
				nowNS := base.TimeNow().UnixNano()
				assert(nowNS-v.activeTimeNS < int64(time.Second)).IsTrue()
				assert(nowNS-v.activeTimeNS >= 0).IsTrue()
				assert(v.prev).IsNil()
				assert(v.next).IsNil()

				rs := rpc.NewStream()
				rs.PutBytesTo(netConn.writeBuffer, 0)

				config := sessionServer.config

				assert(rs.GetKind()).
					Equals(uint8(rpc.StreamKindConnectResponse))
				assert(rs.ReadString()).
					Equals(fmt.Sprintf("%d-%s", v.id, v.security), nil)
				assert(rs.ReadInt64()).Equals(int64(config.numOfChannels), nil)
				assert(rs.ReadInt64()).Equals(int64(config.transLimit), nil)
				assert(rs.ReadInt64()).
					Equals(int64(config.heartbeatInterval/time.Millisecond), nil)
				assert(rs.ReadInt64()).
					Equals(int64(config.heartbeatTimeout/time.Millisecond), nil)
				assert(rs.IsReadFinish()).IsTrue()
				assert(rs.CheckStream()).IsTrue()
			}
		}
	})
}

func TestNewSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		sessionServer := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)
		v := newSession(3, sessionServer)
		assert(v.id).Equals(uint64(3))
		assert(v.sessionServer).Equals(sessionServer)
		assert(len(v.security)).Equals(32)
		assert(v.conn).IsNil()
		assert(len(v.channels)).Equals(GetDefaultSessionConfig().numOfChannels)
		assert(cap(v.channels)).Equals(GetDefaultSessionConfig().numOfChannels)
		assert(base.TimeNow().UnixNano()-v.activeTimeNS < int64(time.Second)).
			IsTrue()
		assert(v.prev).IsNil()
		assert(v.next).IsNil()
	})
}

func TestSession_TimeCheck(t *testing.T) {
	t.Run("p.conn is active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
		session.sessionServer.config.heartbeatTimeout = 100 * time.Millisecond
		syncConn.OnOpen()
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsTrue()
		assert(session.conn).IsNotNil()
	})

	t.Run("p.conn is not active", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
		session.sessionServer.config.heartbeatTimeout = 1 * time.Millisecond
		syncConn.OnOpen()
		time.Sleep(30 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("p.conn is nil, session is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession(nil)
		session.sessionServer.config.serverSessionTimeout = time.Second
		session.sessionServer.TimeCheck(base.TimeNow().UnixNano())
		assert(session.sessionServer.TotalSessions()).Equals(int64(1))
	})

	t.Run("p.conn is nil, session is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession(nil)
		session.sessionServer.config.serverSessionTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
		session.sessionServer.TimeCheck(base.TimeNow().UnixNano())
		assert(session.sessionServer.TotalSessions()).Equals(int64(0))
	})

	t.Run("p.channels is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession(nil)

		// fill the channels
		for i := 0; i < session.sessionServer.config.numOfChannels; i++ {
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}

		session.sessionServer.config.serverCacheTimeout = 10 * time.Millisecond
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.sessionServer.config.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS > 0).IsTrue()
			assert(session.channels[i].backStream).IsNotNil()
		}
	})

	t.Run("p.channels is timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, _ := prepareTestSession(nil)

		// fill the channels
		for i := 0; i < session.sessionServer.config.numOfChannels; i++ {
			stream := rpc.NewStream()
			stream.SetCallbackID(uint64(i) + 1)
			session.channels[i].In(stream.GetCallbackID())
			session.channels[i].Out(stream)
		}

		session.sessionServer.config.serverCacheTimeout = 1 * time.Millisecond
		time.Sleep(30 * time.Millisecond)
		session.TimeCheck(base.TimeNow().UnixNano())

		for i := 0; i < session.sessionServer.config.numOfChannels; i++ {
			assert(session.channels[i].backTimeNS).Equals(int64(0))
			assert(session.channels[i].backStream).IsNil()
		}
	})
}

func TestSession_OutStream(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession(nil)
		session.OutStream(nil)
		assert(len(netConn.writeBuffer)).Equals(0)
	})

	t.Run("p.conn is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, _, netConn := prepareTestSession(nil)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		session.OutStream(stream)
		assert(len(netConn.writeBuffer)).Equals(0)
		assert(session.channels[0].backStream).Equals(stream)
	})

	t.Run("stream kind error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
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

		assert(len(netConn.writeBuffer)).Equals(0)
	})

	t.Run("stream can not out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
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

		assert(len(netConn.writeBuffer)).Equals(0)
	})

	t.Run("stream can out", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
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

		assert(netConn.writeBuffer).Equals(exceptBuffer)
	})

	t.Run("stream is StreamKindRPCBoardCast", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
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

		assert(netConn.writeBuffer).Equals(exceptBuffer)
	})
}

func TestSession_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession(nil)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		session.OnConnOpen(streamConn)
		assert(session.conn).Equals(streamConn)
	})
}

func TestSession_OnConnReadStream(t *testing.T) {
	t.Run("cbID == 0, kind == StreamKindPing ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
		syncConn.OnOpen()
		netConn.writeBuffer = make([]byte, 0)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		sendStream := rpc.NewStream()
		sendStream.SetKind(rpc.StreamKindPing)
		session.OnConnReadStream(streamConn, sendStream)
		backStream := rpc.NewStream()
		backStream.PutBytesTo(netConn.writeBuffer, 0)
		assert(backStream.GetKind()).Equals(uint8(rpc.StreamKindPong))
		assert(backStream.IsReadFinish()).IsTrue()
	})

	t.Run("cbID == 0, kind == StreamKindPing error", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)
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
			Equals(nil, base.ErrStream)
	})

	t.Run("cbID > 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)

		streamReceiver := rpc.NewTestStreamReceiver()
		session, syncConn, _ := prepareTestSession(nil)
		session.sessionServer.streamReceiver = streamReceiver

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(10)
		stream.SetKind(rpc.StreamKindRPCRequest)
		session.OnConnReadStream(streamConn, stream)

		backStream := streamReceiver.GetStream()
		assert(backStream.GetSessionID()).Equals(uint64(11))
	})

	t.Run("cbID > 0, accept = false, backStream != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)

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
		assert(netConn.writeBuffer).Equals(cacheStream.GetBuffer())
	})

	t.Run("cbID > 0, accept = false, backStream == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, netConn := prepareTestSession(nil)

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		syncConn.SetNext(streamConn)
		(&session.channels[10%len(session.channels)]).In(10)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCRequest)
		stream.SetCallbackID(10)
		session.OnConnOpen(streamConn)
		netConn.writeBuffer = make([]byte, 0)
		session.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeBuffer)).Equals(0)
	})

	t.Run("cbID == 0, accept = true, backStream = nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession(nil)
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver

		streamConn := adapter.NewStreamConn(false, syncConn, session)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindRPCRequest)

		session.OnConnReadStream(streamConn, stream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrStream)
	})

	t.Run("cbID == 0, kind err", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession(nil)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver
		session.OnConnReadStream(streamConn, rpc.NewStream())
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrStream)
	})
}

func TestSession_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession(nil)
		streamConn := adapter.NewStreamConn(false, syncConn, session)
		streamReceiver := rpc.NewTestStreamReceiver()
		session.sessionServer.streamReceiver = streamReceiver
		session.OnConnError(streamConn, base.ErrStream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrStream)
	})
}

func TestSession_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		session, syncConn, _ := prepareTestSession(nil)
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
		assert(v.sessionServer).Equals(sessionServer)
		assert(len(v.idMap)).Equals(0)
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
		assert(v.sessionServer.totalSessions).Equals(int64(1))
	})

	t.Run("add one session", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session := &Session{id: 10}
		assert(v.Add(session)).IsTrue()
		assert(v.sessionServer.totalSessions).Equals(int64(1))
		assert(v.head).Equals(session)
		assert(session.prev).Equals(nil)
		assert(session.next).Equals(nil)
	})

	t.Run("add two sessions", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionPool(&SessionServer{})
		session1 := &Session{id: 11}
		session2 := &Session{id: 12}
		assert(v.Add(session1)).IsTrue()
		assert(v.Add(session2)).IsTrue()
		assert(v.sessionServer.totalSessions).Equals(int64(2))
		assert(v.head).Equals(session2)
		assert(session2.prev).Equals(nil)
		assert(session2.next).Equals(session1)
		assert(session1.prev).Equals(session2)
		assert(session1.next).Equals(nil)
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
		assert(v.sessionServer.totalSessions).Equals(int64(2))
		assert(v.Get(11)).Equals(session1, true)
		assert(v.Get(12)).Equals(session2, true)
		assert(v.Get(10)).Equals(nil, false)
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
		assert(1024).Equals(1024)
	})
}

func TestNewSessionServer(t *testing.T) {
	t.Run("streamReceiver is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(base.RunWithCatchPanic(func() {
			NewSessionServer(nil, GetDefaultSessionConfig(), nil)
		})).Equals("streamReceiver is nil")
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		assert(v.isRunning).Equals(false)
		assert(v.sessionSeed).Equals(uint64(0))
		assert(v.totalSessions).Equals(int64(0))
		assert(len(v.sessionMapList)).Equals(1024)
		assert(cap(v.sessionMapList)).Equals(1024)
		assert(v.streamReceiver).Equals(streamReceiver)
		assert(len(v.closeCH)).Equals(0)
		assert(cap(v.closeCH)).Equals(1)
		assert(v.config).Equals(GetDefaultSessionConfig())
		assert(len(v.adapters)).Equals(0)
		assert(cap(v.adapters)).Equals(0)
		assert(v.orcManager).IsNotNil()

		for i := 0; i < 1024; i++ {
			assert(v.sessionMapList[i]).IsNotNil()
		}
	})
}

func TestSessionServer_TotalSessions(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)
		v.totalSessions = 54321
		assert(v.TotalSessions()).Equals(int64(54321))
	})
}

func TestSessionServer_AddSession(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)

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
		v := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)

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
		v := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)
		assert(v.CreateSessionID()).Equals(uint64(1))
		assert(v.CreateSessionID()).Equals(uint64(2))
	})
}

func TestSessionServer_TimeCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(
			nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver(),
		)
		for i := uint64(1); i <= 1024; i++ {
			session := newSession(i, v)
			session.activeTimeNS = 0
			assert(v.AddSession(session)).IsTrue()
		}

		assert(v.TotalSessions()).Equals(int64(1024))
		v.TimeCheck(base.TimeNow().UnixNano())
		assert(v.TotalSessions()).Equals(int64(0))
	})
}

// func TestSessionServer_Listen(t *testing.T) {
// 	t.Run("SessionServer is not running", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		tlsConfig := &tls.Config{}
// 		v := NewSessionServer(nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
// 		assert(v.Listen("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
// 		assert(len(v.adapters)).Equal(1)
// 		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
// 			false,
// 			"tcp",
// 			"0.0.0.0:8080",
// 			tlsConfig,
// 			v.sessionConfig.serverReadBufferSize,
// 			v.sessionConfig.serverWriteBufferSize,
// 			v,
// 		))
// 	})
// }

// func TestSessionServer_ListenWithDebug(t *testing.T) {
// 	t.Run("SessionServer is not running", func(t *testing.T) {
// 		assert := base.NewAssert(t)
// 		tlsConfig := &tls.Config{}
// 		v := NewSessionServer(nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
// 		assert(v.ListenWithDebug("tcp", "0.0.0.0:8080", tlsConfig)).Equal(v)
// 		assert(len(v.adapters)).Equal(1)
// 		assert(v.adapters[0]).Equal(adapter.NewServerAdapter(
// 			true,
// 			"tcp",
// 			"0.0.0.0:8080",
// 			tlsConfig,
// 			v.sessionConfig.serverReadBufferSize,
// 			v.sessionConfig.serverWriteBufferSize,
// 			v,
// 		))
// 	})
// }

func TestSessionServer_Open(t *testing.T) {
	t.Run("it is already running", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		v.isRunning = true
		v.Open()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrServerAlreadyRunning)
	})

	t.Run("no valid adapter", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		v.Open()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrServerNoListenersAvailable)
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		v := NewSessionServer(
			[]*listener{
				{
					isDebug:   false,
					network:   "tcp",
					addr:      "127.0.0.1:8000",
					tlsConfig: nil,
				},
				{
					isDebug:   false,
					network:   "tcp",
					addr:      "127.0.0.1:8001",
					tlsConfig: nil,
				},
			}, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(&Session{id: 10})

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
		assert(v.TotalSessions()).Equals(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestSessionServer_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		waitCH := make(chan bool)
		v := NewSessionServer([]*listener{
			{
				isDebug:   false,
				network:   "tcp",
				addr:      "127.0.0.1:8000",
				tlsConfig: nil,
			},
		}, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(&Session{id: 10})

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

		assert(v.TotalSessions()).Equals(int64(1))
		v.Open()
		<-waitCH
	})
}

func TestSessionServer_ReceiveStreamFromRouter(t *testing.T) {
	t.Run("session is exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		stream := rpc.NewStream()
		stream.SetSessionID(10)
		v.OutStream(stream)
		assert(streamReceiver.GetStream()).IsNil()
	})

	t.Run("session is not exist", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		stream := rpc.NewStream()
		stream.SetSessionID(11)
		v.OutStream(stream)
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrServerSessionNotFound)
	})
}

func TestSessionServer_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnOpen(nil)
		})).IsNil()
	})
}

func TestSessionServer_OnConnReadStream(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		syncConn := adapter.NewServerSyncConn(newTestNetConn(), 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)

		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectRequest)
		stream.WriteString("")
		stream.BuildStreamCheck()
		streamConn.OnReadBytes(stream.GetBuffer())
		assert(v.totalSessions).Equals(int64(1))
	})
}

func TestSessionServer_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		streamReceiver := rpc.NewTestStreamReceiver()
		v := NewSessionServer(nil, GetDefaultSessionConfig(), streamReceiver)
		v.AddSession(newSession(10, v))
		netConn := newTestNetConn()
		syncConn := adapter.NewServerSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		assert(netConn.isRunning).IsTrue()
		v.OnConnError(streamConn, base.ErrStream)
		assert(netConn.isRunning).IsFalse()
		assert(rpc.ParseResponseStream(streamReceiver.GetStream())).
			Equals(nil, base.ErrStream)
	})
}

func TestSessionServer_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSessionServer(nil, GetDefaultSessionConfig(), rpc.NewTestStreamReceiver())
		v.AddSession(newSession(10, v))
		assert(base.RunWithCatchPanic(func() {
			v.OnConnClose(nil)
		})).IsNil()
	})
}

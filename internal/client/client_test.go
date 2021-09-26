package client

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"github.com/rpccloud/rpc/internal/server"
)

type testNetConn struct {
	writeCH   chan []byte
	isRunning bool
}

func newTestNetConn() *testNetConn {
	return &testNetConn{
		isRunning: true,
		writeCH:   make(chan []byte, 1024),
	}
}

func (p *testNetConn) Read(_ []byte) (n int, err error) {
	panic("not implemented")
}

func (p *testNetConn) Write(b []byte) (n int, err error) {
	buf := make([]byte, len(b))
	copy(buf, b)
	p.writeCH <- buf
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

func getTestServer() *server.Server {
	userService := rpc.NewService().
		On("SayHello", func(rt rpc.Runtime, name rpc.String) rpc.Return {
			return rt.Reply("hello " + name)
		}).
		On("Sleep", func(rt rpc.Runtime, timeNS int64) rpc.Return {
			time.Sleep(time.Duration(timeNS))
			return rt.Reply(nil)
		}).
		On("PostMessage", func(rt rpc.Runtime, timeNS int64) rpc.Return {
			return rt.Reply(
				rt.Post(rt.GetPostEndPoint(), "@Post", rpc.Array{true, timeNS}),
			)
		})

	rpcServer := server.NewServer(
		server.GetDefaultServerConfig().SetNumOfThreads(256),
	).ListenWithDebug("ws", "0.0.0.0:8765", nil)
	rpcServer.AddService("user", userService, nil)

	go func() {
		rpcServer.Open()
	}()

	time.Sleep(100 * time.Millisecond)

	return rpcServer
}

func checkClientPreSendList(c *Client, arr []*SendItem) bool {
	if len(arr) == 0 {
		return c.preSendHead == nil && c.preSendTail == nil
	}

	if c.preSendHead != arr[0] || c.preSendTail != arr[len(arr)-1] {
		return false
	}

	if c.preSendTail.next != nil {
		return false
	}

	for i := 0; i < len(arr)-1; i++ {
		if arr[i].next != arr[i+1] {
			return false
		}
	}

	return true
}

type TestAdapter struct {
	isDebug    bool
	isClient   bool
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   adapter.IReceiver
	service    base.IORCService
	orcManager *base.ORCManager
}

func TestSubscription_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		client := &Client{
			subscriptionMap: make(map[string][]*Subscription),
		}
		sub := &Subscription{
			id:        13,
			client:    client,
			onMessage: func(value rpc.Any) {},
		}
		client.subscriptionMap["#.test%Message"] = []*Subscription{sub}

		sub.Close()
		assert(sub.id).Equals(int64(0))
		assert(sub.client).Equals(nil)
		assert(sub.onMessage).Equals(nil)
		assert(client.subscriptionMap).Equals(map[string][]*Subscription{})
	})
}

func TestSendItem_NewSendItem(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(5432)
		assert(v.isRunning).IsTrue()
		assert(base.TimeNow().UnixNano()-v.startTimeNS < int64(time.Second)).
			IsTrue()
		assert(base.TimeNow().UnixNano()-v.startTimeNS >= 0).IsTrue()
		assert(v.sendTimeNS).Equals(int64(0))
		assert(v.timeoutNS).Equals(int64(5432))
		assert(len(v.returnCH)).Equals(0)
		assert(cap(v.returnCH)).Equals(1)
		assert(v.sendStream).IsNotNil()
		assert(v.next).IsNil()
	})
}

func TestSendItem_Back(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		assert(v.Back(nil)).IsFalse()
		assert(len(v.returnCH)).Equals(0)
	})

	t.Run("item is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		v.isRunning = false
		assert(v.Back(rpc.NewStream())).IsFalse()
		assert(len(v.returnCH)).Equals(0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		stream := rpc.NewStream()
		assert(v.Back(stream)).IsTrue()
		assert(len(v.returnCH)).Equals(1)
		assert(<-v.returnCH).Equals(stream)
	})
}

func TestSendItem_CheckTime(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Millisecond))
		v.sendStream.SetCallbackID(15)
		time.Sleep(100 * time.Millisecond)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsTrue()
		assert(v.isRunning).IsFalse()
		assert(len(v.returnCH)).Equals(1)
		stream := <-v.returnCH
		assert(stream.GetCallbackID()).Equals(uint64(15))
		assert(stream.ReadUint64()).
			Equals(uint64(base.ErrClientTimeout.GetCode()), nil)
		assert(stream.ReadString()).
			Equals(base.ErrClientTimeout.GetMessage(), nil)
		assert(stream.IsReadFinish()).IsTrue()
	})

	t.Run("it is not timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Second))
		v.sendStream.SetCallbackID(15)
		time.Sleep(10 * time.Millisecond)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
		assert(v.isRunning).IsTrue()
		assert(len(v.returnCH)).Equals(0)
	})

	t.Run("it is not running", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(int64(time.Millisecond))
		v.sendStream.SetCallbackID(15)
		time.Sleep(10 * time.Millisecond)
		v.isRunning = false
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
		assert(v.isRunning).IsFalse()
		assert(len(v.returnCH)).Equals(0)
	})
}

func TestSendItem_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewSendItem(0)
		for i := 0; i < 10000; i++ {
			v.sendStream.PutBytes([]byte{1})
		}

		v.Release()
		assert(v.sendStream.GetWritePos()).Equals(rpc.StreamHeadSize)
	})

	t.Run("test put back", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := NewSendItem(0)
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}

func TestChannel_Use(t *testing.T) {
	t.Run("p.item != nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: &SendItem{}}
		assert(v.Use(&SendItem{}, 32)).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642}
		item := NewSendItem(0)
		assert(v.Use(item, 32)).IsTrue()
		assert(v.sequence).Equals(uint64(674))
		assert(v.item).Equals(item)
		assert(item.sendStream.GetCallbackID()).Equals(uint64(674))
		nowNS := base.TimeNow().UnixNano()
		assert(nowNS-v.item.sendTimeNS < int64(time.Second)).IsTrue()
		assert(nowNS-v.item.sendTimeNS > -int64(time.Second)).IsTrue()
	})
}

func TestChannel_Free(t *testing.T) {
	t.Run("p.item == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: nil}
		assert(v.Free(rpc.NewStream())).IsFalse()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		item := NewSendItem(0)
		v := &Channel{sequence: 642, item: item}
		stream := rpc.NewStream()
		assert(len(item.returnCH)).Equals(0)
		assert(v.Free(stream)).IsTrue()
		assert(v.item).IsNil()
		assert(len(item.returnCH)).Equals(1)
		assert(<-item.returnCH).Equals(stream)
	})
}

func TestChannel_CheckTime(t *testing.T) {
	t.Run("p.item == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: nil}
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
	})

	t.Run("CheckTime return false", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Channel{sequence: 642, item: NewSendItem(int64(time.Second))}
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsFalse()
	})

	t.Run("CheckTime return true", func(t *testing.T) {
		assert := base.NewAssert(t)
		item := NewSendItem(int64(time.Millisecond))
		v := &Channel{sequence: 642, item: item}

		time.Sleep(10 * time.Millisecond)
		assert(len(item.returnCH)).Equals(0)
		assert(v.CheckTime(base.TimeNow().UnixNano())).IsTrue()
		assert(v.item).IsNil()
		assert(len(item.returnCH)).Equals(1)
	})
}

func TestNewClient(t *testing.T) {
	type TestORCManager struct {
		status      uint32
		statusCond  sync.Cond
		runningCond sync.Cond
		mu          sync.Mutex
	}

	t.Run("test", func(t *testing.T) {
		testServer := getTestServer()
		defer testServer.Close()

		assert := base.NewAssert(t)
		v := NewClient(
			"ws", "127.0.0.1:8765", nil, 1024, 2048, func(_ *base.Error) {},
		)

		for {
			v.mu.Lock()
			conn := v.conn
			v.mu.Unlock()

			if conn == nil {
				time.Sleep(10 * time.Millisecond)
			} else {
				break
			}
		}

		assert(v.config).Equals(&Config{
			numOfChannels:    32,
			transLimit:       4 * 1024 * 1024,
			heartbeat:        4 * time.Second,
			heartbeatTimeout: 8 * time.Second,
		})
		assert(len(v.sessionString) >= 34).IsTrue()
		testAdapter := (*TestAdapter)(unsafe.Pointer(v.adapter))
		assert(testAdapter.isDebug).IsFalse()
		assert(testAdapter.isClient).IsTrue()
		assert(testAdapter.network).Equals("ws")
		assert(testAdapter.addr).Equals("127.0.0.1:8765")
		assert(testAdapter.tlsConfig).Equals(nil)
		assert(testAdapter.rBufSize).Equals(1024)
		assert(testAdapter.wBufSize).Equals(2048)
		assert(testAdapter.receiver).Equals(v)
		assert(testAdapter.service).IsNotNil()
		adapterOrcManager := (*TestORCManager)(
			unsafe.Pointer(testAdapter.orcManager),
		)
		// orcStatusReady | orcLockBit = 1 | 1 << 2 = 5
		assert(atomic.LoadUint32(&adapterOrcManager.status)).
			Equals(uint32(2))
		assert(&adapterOrcManager.statusCond).IsNotNil()
		assert(&adapterOrcManager.runningCond).IsNotNil()
		assert(&adapterOrcManager.mu).IsNotNil()
		assert(v.preSendHead).IsNil()
		assert(v.preSendTail).IsNil()
		assert(len(v.channels)).Equals(32)
		assert(v.lastPingTimeNS > 0).IsTrue()
		// orcStatusReady | orcLockBit = 1 | 1 << 2 = 5
		assert(atomic.LoadUint32(
			&(*TestORCManager)(unsafe.Pointer(v.orcManager)).status,
		) % 8).Equals(uint32(2))

		// check tryLoop
		_, err := v.Send(
			500*time.Millisecond,
			"#.user:Sleep",
			int64(2*time.Second),
		)
		assert(err).Equals(base.ErrClientTimeout)
		v.Close()
	})
}

func TestClient_tryToSendPing(t *testing.T) {
	t.Run("p.conn == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{}

		v.tryToSendPing(1)
		assert(v.lastPingTimeNS).Equals(int64(0))
	})

	t.Run("do not need to ping", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: base.TimeNow().UnixNano(),
			config:         &Config{heartbeat: time.Second},
		}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.tryToSendPing(base.TimeNow().UnixNano())
		assert(len(netConn.writeCH)).Equals(0)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeat: time.Second},
		}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.tryToSendPing(base.TimeNow().UnixNano())

		stream := rpc.NewStream()
		stream.PutBytesTo(<-netConn.writeCH, 0)
		assert(stream.GetKind()).Equals(uint8(rpc.StreamKindPing))
		assert(stream.IsReadFinish()).IsTrue()
		assert(stream.CheckStream()).IsTrue()
	})
}

func TestClient_tryToTimeout(t *testing.T) {
	fnTest := func(totalItems int, timeoutItems int) bool {
		v := &Client{
			config: &Config{heartbeatTimeout: 9 * time.Millisecond},
		}

		if totalItems < 0 || totalItems < timeoutItems {
			panic("error")
		}

		beforeData := make([]*SendItem, totalItems)
		afterData := make([]*SendItem, totalItems)
		nowNS := base.TimeNow().UnixNano()

		for i := 0; i < totalItems; i++ {
			item := NewSendItem(int64(time.Second))

			if v.preSendTail == nil {
				v.preSendHead = item
				v.preSendTail = item
			} else {
				v.preSendTail.next = item
				v.preSendTail = item
			}

			beforeData[i] = item
			afterData[i] = item
		}

		for i := 0; i < timeoutItems; i++ {
			rand.Seed(time.Now().UnixNano())
			idx := rand.Int() % len(afterData)
			afterData[idx].timeoutNS = 0
			afterData = append(afterData[:idx], afterData[idx+1:]...)
		}

		if !checkClientPreSendList(v, beforeData) {
			return false
		}

		v.tryToTimeout(nowNS + int64(500*time.Millisecond))
		return checkClientPreSendList(v, afterData)
	}

	t.Run("check if the channels has been swept", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
		}
		item := NewSendItem(int64(5 * time.Millisecond))
		v.channels[0].Use(item, 1)

		v.tryToTimeout(item.sendTimeNS + int64(4*time.Millisecond))
		assert(v.channels[0].sequence).Equals(uint64(1))
		assert(v.channels[0].item).IsNotNil()

		v.tryToTimeout(item.sendTimeNS + int64(10*time.Millisecond))
		assert(v.channels[0].sequence).Equals(uint64(1))
		assert(v.channels[0].item).IsNil()
	})

	t.Run("check if the conn has been swept", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
		}

		// conn is nil
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(4*time.Millisecond))
		assert(v.conn).IsNil()

		// set conn
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		// conn is active
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(4*time.Millisecond))
		assert(netConn.isRunning).IsTrue()

		// conn is not active
		v.tryToTimeout(base.TimeNow().UnixNano() + int64(20*time.Millisecond))
		assert(netConn.isRunning).IsFalse()
	})

	t.Run("item timeout", func(t *testing.T) {
		assert := base.NewAssert(t)
		for n := 0; n < 10; n++ {
			for i := 0; i < 10; i++ {
				for j := 0; j <= i; j++ {
					assert(fnTest(i, j)).IsTrue()
				}
			}
		}
	})
}

func TestClient_tryToDeliverPreSendMessages(t *testing.T) {
	fnTest := func(totalPreItems int, chSize int, chFree int) bool {
		if chSize < chFree {
			panic("error")
		}

		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, chSize),
		}

		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		chFreeArr := make([]int, chSize)

		for i := 0; i < len(v.channels); i++ {
			(&v.channels[i]).sequence = uint64(i)
			(&v.channels[i]).item = nil
			chFreeArr[i] = i
		}

		itemsArray := make([]*SendItem, totalPreItems)
		for i := 0; i < totalPreItems; i++ {
			itemsArray[i] = NewSendItem(int64(time.Second))
			if v.preSendHead == nil {
				v.preSendHead = itemsArray[i]
				v.preSendTail = itemsArray[i]
			} else {
				v.preSendTail.next = itemsArray[i]
				v.preSendTail = itemsArray[i]
			}
		}

		if !checkClientPreSendList(v, itemsArray) {
			panic("error")
		}

		for len(chFreeArr) > chFree {
			rand.Seed(base.TimeNow().UnixNano())
			idx := rand.Int() % len(chFreeArr)
			(&v.channels[chFreeArr[idx]]).item = NewSendItem(int64(time.Second))
			chFreeArr = append(chFreeArr[:idx], chFreeArr[idx+1:]...)
		}

		v.tryToDeliverPreSendMessages()

		return checkClientPreSendList(
			v,
			itemsArray[base.MinInt(len(itemsArray), chFree):],
		)
	}

	t.Run("p.conn == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			channels:       make([]Channel, 1),
			preSendHead:    NewSendItem(0),
		}
		v.tryToDeliverPreSendMessages()
		assert(v.preSendHead).IsNotNil()
	})

	t.Run("p.channel == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			lastPingTimeNS: 10000,
			config:         &Config{heartbeatTimeout: 9 * time.Millisecond},
			conn:           adapter.NewStreamConn(false, nil, nil),
			preSendHead:    NewSendItem(0),
		}
		v.tryToDeliverPreSendMessages()
		assert(v.preSendHead).IsNotNil()
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for n := 16; n <= 32; n++ {
			for i := 0; i < 10; i++ {
				for j := 0; j <= n; j++ {
					assert(fnTest(i, n, j)).IsTrue()
				}
			}
		}
	})
}

func TestClient_Subscribe(t *testing.T) {
	t.Run("test basic", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClient(
			"ws", "127.0.0.1:8080", nil, 1200, 1200, func(b *base.Error) {},
		)
		defer v.Close()

		sub1 := v.Subscribe("#.test", "Message01", func(value rpc.Any) {})
		sub2 := v.Subscribe("#.test", "Message01", func(value rpc.Any) {})
		sub3 := v.Subscribe("#.test", "Message02", func(value rpc.Any) {})

		assert(sub1, sub2, sub3).IsNotNil()
		assert(v.subscriptionMap).Equals(map[string][]*Subscription{
			"#.test%Message01": {sub1, sub2},
			"#.test%Message02": {sub3},
		})
	})

	t.Run("test message", func(t *testing.T) {
		assert := base.NewAssert(t)

		rpcServer := getTestServer()
		defer rpcServer.Close()

		rpcClient := NewClient(
			"ws", "0.0.0.0:8765", nil, 1200, 1200, func(b *base.Error) {},
		)
		defer rpcClient.Close()

		waitCH := make(chan rpc.Any, 1)
		rpcClient.Subscribe("#.user", "@Post", func(value rpc.Any) {
			waitCH <- value
		})
		assert(rpcClient.Send(5*time.Second, "#.user:PostMessage", 2345)).
			Equals(nil, nil)
		assert(<-waitCH).Equals(rpc.Array{true, int64(2345)})
	})
}

func TestClient_unsubscribe(t *testing.T) {
	assert := base.NewAssert(t)
	v := NewClient(
		"ws", "127.0.0.1:8080", nil, 1200, 1200, func(b *base.Error) {},
	)
	defer v.Close()

	sub1 := v.Subscribe("#.test", "Message01", func(value rpc.Any) {})
	sub2 := v.Subscribe("#.test", "Message01", func(value rpc.Any) {})
	sub3 := v.Subscribe("#.test", "Message02", func(value rpc.Any) {})

	assert(v.subscriptionMap).Equals(map[string][]*Subscription{
		"#.test%Message01": {sub1, sub2},
		"#.test%Message02": {sub3},
	})
	v.unsubscribe(sub1.id)
	assert(v.subscriptionMap).Equals(map[string][]*Subscription{
		"#.test%Message01": {sub2},
		"#.test%Message02": {sub3},
	})
	v.unsubscribe(sub2.id)
	assert(v.subscriptionMap).Equals(map[string][]*Subscription{
		"#.test%Message02": {sub3},
	})
	v.unsubscribe(sub3.id)
	assert(v.subscriptionMap).Equals(map[string][]*Subscription{})
}

func TestClient_Send(t *testing.T) {
	t.Run("args error", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{
			channels: make([]Channel, 0),
		}

		assert(v.Send(time.Second, "#.user:SayHello", make(chan bool))).
			Equals(nil, base.ErrUnsupportedValue.AddDebug(
				"value type(chan bool) is not supported",
			))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		rpcServer := getTestServer()
		defer rpcServer.Close()

		rpcClient := NewClient(
			"ws", "0.0.0.0:8765", nil, 1200, 1200, func(b *base.Error) {},
		)
		defer rpcClient.Close()

		waitCH := make(chan []interface{})
		for i := 0; i < 100; i++ {
			go func() {
				v, err := rpcClient.Send(
					6*time.Second,
					"#.user:SayHello",
					"kitty",
				)
				waitCH <- []interface{}{v, err}
			}()
		}

		for i := 0; i < 100; i++ {
			assert(<-waitCH...).Equals("hello kitty", nil)
		}
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClient(
			"ws", "127.0.0.1:1234", nil, 1200, 1200, func(b *base.Error) {},
		)
		assert(v.adapter).IsNotNil()
		assert(v.Close()).IsTrue()
		assert(v.adapter).IsNil()
	})
}

func TestClient_OnConnOpen(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{sessionString: "123456"}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn

		v.OnConnOpen(streamConn)

		stream := rpc.NewStream()
		stream.PutBytesTo(<-netConn.writeCH, 0)
		assert(stream.GetKind()).
			Equals(uint8(rpc.StreamKindConnectRequest))
		assert(stream.ReadString()).Equals("123456", nil)
	})
}

func TestClient_OnConnReadStream(t *testing.T) {
	fnTestClient := func() (
		*Client,
		*adapter.StreamConn,
		*testNetConn,
		chan *base.Error) {

		errCH := make(chan *base.Error, 1024)
		v := &Client{
			config:          &Config{},
			subscriptionMap: map[string][]*Subscription{},
			onError: func(err *base.Error) {
				errCH <- err
			},
		}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		return v, streamConn, netConn, errCH
	}

	t.Run("conn == nil, stream.callbackID != 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(12)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, kind != StreamKindConnectResponse", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, read sessionString error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, read numOfChannels error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, numOfChannels config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.SetCallbackID(0)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(0)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrClientConfig)
	})

	t.Run("conn == nil, read transLimit error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, transLimit config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(0)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrClientConfig)
	})

	t.Run("conn == nil, read heartbeat error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, heartbeat config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(0)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrClientConfig)
	})

	t.Run("conn == nil, read heartbeatTimeout error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, heartbeatTimeout config error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(0)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrClientConfig)
	})

	t.Run("conn == nil, stream is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second))
		stream.WriteInt64(int64(2 * time.Second))
		stream.WriteBool(false)
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("conn == nil, sessionString != p.sessionString", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second / time.Millisecond))
		stream.WriteInt64(int64(2 * time.Second / time.Millisecond))
		v, streamConn, _, errCH := fnTestClient()
		v.OnConnReadStream(streamConn, stream)
		assert(len(errCH)).Equals(0)
		assert(v.sessionString).Equals("12-87654321876543218765432187654321")

		assert(v.config.numOfChannels).Equals(32)
		assert(v.config.transLimit).Equals(4 * 1024 * 1024)
		assert(v.config.heartbeat).Equals(1 * time.Second)
		assert(v.config.heartbeatTimeout).Equals(2 * time.Second)
		for i := 0; i < 32; i++ {
			assert(v.channels[i].sequence).Equals(uint64(i))
			assert(v.channels[i].item).IsNil()
		}
		assert(v.lastPingTimeNS > 0).IsTrue()
	})

	t.Run("conn == nil, sessionString == p.sessionString", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(0)
		stream.SetKind(rpc.StreamKindConnectResponse)
		stream.WriteString("12-87654321876543218765432187654321")
		stream.WriteInt64(32)
		stream.WriteInt64(4 * 1024 * 1024)
		stream.WriteInt64(int64(time.Second / time.Millisecond))
		stream.WriteInt64(int64(2 * time.Second / time.Millisecond))
		v, streamConn, netConn, _ := fnTestClient()
		v.channels = make([]Channel, 32)
		for i := 0; i < 32; i++ {
			(&v.channels[i]).sequence = uint64(i)
			(&v.channels[i]).Use(NewSendItem(0), 32)
		}

		v.sessionString = "12-87654321876543218765432187654321"
		v.OnConnReadStream(streamConn, stream)
		assert(len(netConn.writeCH)).Equals(32)
		assert(v.lastPingTimeNS > 0).IsTrue()
	})

	t.Run("p.conn != nil, StreamKindPong error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindPong)
		stream.Write("error")
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("p.conn != nil, StreamKindPong ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindPong)
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(len(errCH)).Equals(0)
	})

	t.Run("p.conn != nil, StreamKindRPCResponseOK ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(17 + 32)
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		v, streamConn, _, _ := fnTestClient()
		v.conn = streamConn
		v.channels = make([]Channel, 32)
		(&v.channels[17]).sequence = 17
		(&v.channels[17]).Use(NewSendItem(0), 32)
		v.OnConnReadStream(streamConn, stream)
		assert(v.channels[17].item).IsNil()
	})

	t.Run("p.conn != nil, StreamKindRPCResponseOK error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(17)
		stream.SetKind(rpc.StreamKindRPCResponseOK)
		v, streamConn, _, _ := fnTestClient()
		v.conn = streamConn
		v.channels = make([]Channel, 32)
		(&v.channels[17]).sequence = 17
		(&v.channels[17]).Use(NewSendItem(0), 32)
		v.OnConnReadStream(streamConn, stream)
		assert(v.channels[17].item).IsNotNil()
	})

	t.Run("p.conn != nil, StreamKindRPCResponseError ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(17 + 32)
		stream.SetKind(rpc.StreamKindRPCResponseError)
		v, streamConn, _, _ := fnTestClient()
		v.conn = streamConn
		v.channels = make([]Channel, 32)
		(&v.channels[17]).sequence = 17
		(&v.channels[17]).Use(NewSendItem(0), 32)
		v.OnConnReadStream(streamConn, stream)
		assert(v.channels[17].item).IsNil()
	})

	t.Run(
		"p.conn != nil, StreamKindRPCResponseError error",
		func(t *testing.T) {
			assert := base.NewAssert(t)
			stream := rpc.NewStream()
			stream.SetCallbackID(17)
			stream.SetKind(rpc.StreamKindRPCResponseError)
			v, streamConn, _, _ := fnTestClient()
			v.conn = streamConn
			v.channels = make([]Channel, 32)
			(&v.channels[17]).sequence = 17
			(&v.channels[17]).Use(NewSendItem(0), 32)
			v.OnConnReadStream(streamConn, stream)
			assert(v.channels[17].item).IsNotNil()
		},
	)

	t.Run("p.conn != nil, getKind() error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetCallbackID(17 + 32)
		stream.SetKind(rpc.StreamKindConnectResponse)
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("StreamKindRPCBoardCast Read path error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCBoardCast)
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("StreamKindRPCBoardCast Read value error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCBoardCast)
		stream.WriteString("#.test%Message")
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("StreamKindRPCBoardCast Read is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCBoardCast)
		stream.WriteString("#.test%Message")
		stream.WriteBool(true)
		stream.WriteString("error")
		v, streamConn, _, errCH := fnTestClient()
		v.conn = streamConn
		v.OnConnReadStream(streamConn, stream)
		assert(<-errCH).Equals(base.ErrStream)
	})

	t.Run("StreamKindRPCBoardCast ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCBoardCast)
		stream.WriteString("#.test%Message")
		stream.WriteString("Hello")
		v, streamConn, _, _ := fnTestClient()
		v.conn = streamConn
		var ret rpc.Any
		v.Subscribe("#.test", "Message", func(value rpc.Any) {
			ret = value
		})
		v.OnConnReadStream(streamConn, stream)
		assert(ret).Equals("Hello")
	})
}

func TestClient_OnConnError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		errCH := make(chan *base.Error, 1024)
		v := &Client{sessionString: "123456", onError: func(err *base.Error) {
			errCH <- err
		}}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		v.OnConnError(streamConn, base.ErrStream)
		assert(<-errCH).Equals(base.ErrStream)
		assert(netConn.isRunning).IsFalse()
	})
}

func TestClient_OnConnClose(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := &Client{}
		netConn := newTestNetConn()
		syncConn := adapter.NewClientSyncConn(netConn, 1200, 1200)
		streamConn := adapter.NewStreamConn(false, syncConn, v)
		syncConn.SetNext(streamConn)
		v.conn = streamConn
		v.OnConnClose(streamConn)
		assert(v.conn).IsNil()
	})
}

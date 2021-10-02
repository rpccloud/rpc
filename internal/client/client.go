// Package client ...
package client

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

// Config ...
type Config struct {
	numOfChannels    int
	transLimit       int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration
}

// Subscription ...
type Subscription struct {
	id        int64
	client    *Client
	onMessage func(value rpc.Any)
}

// Close ...
func (p *Subscription) Close() {
	if p.client != nil {
		p.client.unsubscribe(p.id)
		p.id = 0
		p.onMessage = nil
		p.client = nil
	}
}

var sendItemCache = &sync.Pool{
	New: func() interface{} {
		return &SendItem{
			returnCH:   make(chan *rpc.Stream, 1),
			sendStream: rpc.NewStream(),
		}
	},
}

// SendItem ...
type SendItem struct {
	isRunning   bool
	startTimeNS int64
	sendTimeNS  int64
	timeoutNS   int64
	returnCH    chan *rpc.Stream
	sendStream  *rpc.Stream
	next        *SendItem
}

// NewSendItem ...
func NewSendItem(timeoutNS int64) *SendItem {
	ret := sendItemCache.Get().(*SendItem)
	ret.isRunning = true
	ret.startTimeNS = base.TimeNow().UnixNano()
	ret.sendTimeNS = 0
	ret.timeoutNS = timeoutNS
	ret.next = nil
	return ret
}

// Back ...
func (p *SendItem) Back(stream *rpc.Stream) bool {
	if stream == nil || !p.isRunning {
		return false
	}

	p.returnCH <- stream
	return true
}

// CheckTime ...
func (p *SendItem) CheckTime(nowNS int64) bool {
	if nowNS-p.startTimeNS > p.timeoutNS && p.isRunning {
		p.isRunning = false

		// return timeout stream
		stream := rpc.NewStream()
		stream.SetKind(rpc.StreamKindRPCResponseError)
		stream.SetCallbackID(p.sendStream.GetCallbackID())
		stream.WriteUint64(uint64(base.ErrClientTimeout.GetCode()))
		stream.WriteString(base.ErrClientTimeout.GetMessage())
		p.returnCH <- stream
		return true
	}

	return false
}

// Release ...
func (p *SendItem) Release() {
	p.sendStream.Reset()
	sendItemCache.Put(p)
}

// Channel ...
type Channel struct {
	sequence uint64
	item     *SendItem
}

// Use ...
func (p *Channel) Use(item *SendItem, channelSize int) bool {
	if p.item == nil {
		p.sequence += uint64(channelSize)
		item.sendStream.SetCallbackID(p.sequence)
		p.item = item
		p.item.sendTimeNS = base.TimeNow().UnixNano()
		return true
	}

	return false
}

// Free ...
func (p *Channel) Free(stream *rpc.Stream) bool {
	if item := p.item; item != nil {
		p.item = nil
		return item.Back(stream)
	}

	return false
}

// CheckTime ...
func (p *Channel) CheckTime(nowNS int64) bool {
	if p.item != nil && p.item.CheckTime(nowNS) {
		p.item = nil
		return true
	}

	return false
}

// Client ...
type Client struct {
	config          *Config
	sessionString   string
	adapter         *adapter.Adapter
	conn            *adapter.StreamConn
	preSendHead     *SendItem
	preSendTail     *SendItem
	channels        []Channel
	lastPingTimeNS  int64
	orcManager      *base.ORCManager
	onError         func(err *base.Error)
	subscriptionMap map[string][]*Subscription
	mu              sync.Mutex
}

// NewClient ...
func NewClient(
	network string,
	addr string,
	path string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	onError func(err *base.Error),
) *Client {
	ret := &Client{
		config:          &Config{},
		sessionString:   "",
		adapter:         nil,
		conn:            nil,
		preSendHead:     nil,
		preSendTail:     nil,
		channels:        nil,
		lastPingTimeNS:  0,
		orcManager:      base.NewORCManager(),
		subscriptionMap: make(map[string][]*Subscription),
		onError:         onError,
	}

	// init adapter
	clientAdapter := adapter.NewClientAdapter(
		network, addr, path, tlsConfig, rBufSize, wBufSize, ret,
	)
	clientAdapter.Open()
	go func() {
		clientAdapter.Run()
	}()
	ret.adapter = clientAdapter

	ret.orcManager.Open(func() bool {
		return true
	})

	go func() {
		ret.orcManager.Run(func(isRunning func() bool) {
			for isRunning() {
				time.Sleep(time.Second)
				nowNS := base.TimeNow().UnixNano()
				ret.mu.Lock()
				ret.tryToTimeout(nowNS)
				ret.tryToDeliverPreSendMessages()
				ret.tryToSendPing(nowNS)
				ret.mu.Unlock()
			}
		})
	}()

	return ret
}

func (p *Client) initChannel(size int) {
	p.channels = make([]Channel, size)
	for i := 0; i < len(p.channels); i++ {
		(&p.channels[i]).sequence = uint64(i)
		(&p.channels[i]).item = nil
	}
}

func (p *Client) initConn(stream *rpc.Stream) {
	if kind := stream.GetKind(); kind != rpc.StreamKindConnectResponse {
		p.OnConnError(p.conn, base.ErrStream)
	} else if sessionString, err := stream.ReadString(); err != nil {
		p.OnConnError(p.conn, err)
	} else if numOfChannels, err := stream.ReadInt64(); err != nil {
		p.OnConnError(p.conn, err)
	} else if numOfChannels <= 0 {
		p.OnConnError(p.conn, base.ErrClientConfig)
	} else if transLimit, err := stream.ReadInt64(); err != nil {
		p.OnConnError(p.conn, err)
	} else if transLimit <= 0 {
		p.OnConnError(p.conn, base.ErrClientConfig)
	} else if heartbeat, err := stream.ReadInt64(); err != nil {
		p.OnConnError(p.conn, err)
	} else if heartbeat <= 0 {
		p.OnConnError(p.conn, base.ErrClientConfig)
	} else if heartbeatTimeout, err := stream.ReadInt64(); err != nil {
		p.OnConnError(p.conn, err)
	} else if heartbeatTimeout <= 0 {
		p.OnConnError(p.conn, base.ErrClientConfig)
	} else if !stream.IsReadFinish() {
		p.OnConnError(p.conn, base.ErrStream)
	} else if sessionString != p.sessionString {
		// new session
		p.sessionString = sessionString

		// update config
		p.config.numOfChannels = int(numOfChannels)
		p.config.transLimit = int(transLimit)
		p.config.heartbeat = time.Duration(heartbeat) * time.Millisecond
		p.config.heartbeatTimeout =
			time.Duration(heartbeatTimeout) * time.Millisecond

		// init channel
		p.initChannel(p.config.numOfChannels)
	} else {
		// try to resend channel message
		for i := 0; i < len(p.channels); i++ {
			if item := (&p.channels[i]).item; item != nil {
				p.conn.WriteStreamAndRelease(item.sendStream.Clone())
			}
		}
	}

	p.lastPingTimeNS = base.TimeNow().UnixNano()
}

func (p *Client) tryToSendPing(nowNS int64) {
	if p.conn == nil || nowNS-p.lastPingTimeNS < int64(p.config.heartbeat) {
		return
	}

	// Send Ping
	p.lastPingTimeNS = nowNS
	stream := rpc.NewStream()
	stream.SetKind(rpc.StreamKindPing)
	stream.SetCallbackID(0)
	p.conn.WriteStreamAndRelease(stream)
}

func (p *Client) tryToTimeout(nowNS int64) {
	// sweep pre send list
	preValidItem := (*SendItem)(nil)
	item := p.preSendHead
	for item != nil {
		if item.CheckTime(nowNS) {
			nextItem := item.next

			if preValidItem == nil {
				p.preSendHead = nextItem
			} else {
				preValidItem.next = nextItem
			}

			if item == p.preSendTail {
				p.preSendTail = preValidItem
			}

			item.next = nil
			item = nextItem
		} else {
			preValidItem = item
			item = item.next
		}
	}

	// sweep the channels
	for i := 0; i < len(p.channels); i++ {
		(&p.channels[i]).CheckTime(nowNS)
	}

	// check conn timeout
	if p.conn != nil {
		if !p.conn.IsActive(nowNS, p.config.heartbeatTimeout) {
			p.conn.Close()
		}
	}
}

func (p *Client) tryToDeliverPreSendMessages() {
	if p.conn == nil || p.channels == nil {
		return
	}

	findFree := 0
	channelSize := len(p.channels)

	for findFree < channelSize && p.preSendHead != nil {
		// find a free channel
		for findFree < channelSize && p.channels[findFree].item != nil {
			findFree++
		}

		if findFree < channelSize {
			// remove sendItem from linked list
			item := p.preSendHead
			if item == p.preSendTail {
				p.preSendHead = nil
				p.preSendTail = nil
			} else {
				p.preSendHead = p.preSendHead.next
			}
			item.next = nil

			(&p.channels[findFree]).Use(item, channelSize)
			p.conn.WriteStreamAndRelease(item.sendStream.Clone())
		}
	}
}

// Subscribe ...
func (p *Client) Subscribe(
	nodePath string,
	message string,
	fn func(value rpc.Any),
) *Subscription {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret := &Subscription{
		id:        base.GetSeed(),
		client:    p,
		onMessage: fn,
	}
	path := nodePath + "%" + message
	list, ok := p.subscriptionMap[path]
	if !ok {
		list = make([]*Subscription, 0)
	}
	list = append(list, ret)

	p.subscriptionMap[path] = list
	return ret
}

func (p *Client) unsubscribe(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, list := range p.subscriptionMap {
		pos := -1
		for i := 0; i < len(list); i++ {
			if list[i].id == id {
				pos = i
			}
		}

		// remove if id exists
		if pos >= 0 {
			list = append(list[:pos], list[pos+1:]...)
		}

		if len(list) > 0 {
			p.subscriptionMap[key] = list
		} else {
			delete(p.subscriptionMap, key)
		}
	}
}

// Send ...
func (p *Client) Send(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, *base.Error) {
	item := NewSendItem(int64(timeout))
	defer item.Release()

	item.sendStream.SetKind(rpc.StreamKindRPCRequest)
	// set depth
	item.sendStream.SetDepth(0)
	// write target
	item.sendStream.WriteString(target)
	// write from
	item.sendStream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if eStr := item.sendStream.Write(args[i]); eStr != rpc.StreamWriteOK {
			return nil, base.ErrUnsupportedValue.AddDebug(eStr)
		}
	}

	// add item to the list tail
	p.mu.Lock()
	if p.preSendTail == nil {
		p.preSendHead = item
		p.preSendTail = item
	} else {
		p.preSendTail.next = item
		p.preSendTail = item
	}
	p.tryToDeliverPreSendMessages()
	p.mu.Unlock()

	// wait for response
	backStream := <-item.returnCH
	defer backStream.Release()

	return rpc.ParseResponseStream(backStream)
}

// Close ...
func (p *Client) Close() bool {
	return p.orcManager.Close(func() {
		p.adapter.Close()
	}, func() {
		p.adapter = nil
	})
}

// OnConnOpen ...
func (p *Client) OnConnOpen(streamConn *adapter.StreamConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	stream := rpc.NewStream()
	stream.SetKind(rpc.StreamKindConnectRequest)
	stream.SetCallbackID(0)
	stream.WriteString(p.sessionString)
	streamConn.WriteStreamAndRelease(stream)
}

// OnConnReadStream ...
func (p *Client) OnConnReadStream(
	streamConn *adapter.StreamConn,
	stream *rpc.Stream,
) {
	p.mu.Lock()
	defer p.mu.Unlock()

	callbackID := stream.GetCallbackID()

	if p.conn == nil {
		p.conn = streamConn

		if callbackID != 0 {
			p.OnConnError(streamConn, base.ErrStream)
		} else {
			p.initConn(stream)
		}

		stream.Release()
	} else {
		switch stream.GetKind() {
		case rpc.StreamKindRPCResponseOK:
			fallthrough
		case rpc.StreamKindRPCResponseError:
			channel := &p.channels[callbackID%uint64(len(p.channels))]
			if channel.sequence == callbackID {
				channel.Free(stream)
				p.tryToDeliverPreSendMessages()
			} else {
				stream.Release()
			}
		case rpc.StreamKindRPCBoardCast:
			if actionPath, err := stream.ReadString(); err != nil {
				p.OnConnError(streamConn, err)
			} else if value, err := stream.Read(); err != nil {
				p.OnConnError(streamConn, err)
			} else if !stream.IsReadFinish() {
				p.OnConnError(streamConn, base.ErrStream)
			} else {
				if list, ok := p.subscriptionMap[actionPath]; ok {
					for i := 0; i < len(list); i++ {
						list[i].onMessage(value)
					}
				}
			}
			stream.Release()
		case rpc.StreamKindPong:
			if !stream.IsReadFinish() {
				p.OnConnError(streamConn, base.ErrStream)
			}
			stream.Release()
		default:
			p.OnConnError(streamConn, base.ErrStream)
			stream.Release()
		}
	}
}

// OnConnError ...
func (p *Client) OnConnError(streamConn *adapter.StreamConn, err *base.Error) {
	if p.onError != nil {
		p.onError(err)
	}

	if streamConn != nil {
		streamConn.Close()
	}
}

// OnConnClose ...
func (p *Client) OnConnClose(_ *adapter.StreamConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = nil
}

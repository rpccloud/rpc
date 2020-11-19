package client

import (
	"fmt"
	"github.com/rpccloud/rpc/internal"
	"github.com/rpccloud/rpc/internal/adapter/websocket"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
	"github.com/rpccloud/rpc/internal/errors"
	"github.com/rpccloud/rpc/internal/gateway"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	clientStatusClosed  = int32(0)
	clientStatusClosing = int32(0)
	clientStatusRunning = int32(1)
)

// Client ...
type Client struct {
	status               int32
	closeCH              chan bool
	connectString        string
	sessionString        string
	conn                 internal.IStreamConn
	preSendHead          *SendItem
	preSendTail          *SendItem
	channels             []Channel
	freeChannels         *FreeChannelStack
	config               gateway.SessionConfig
	lastTimeoutCheckTime time.Time
	lastControlSendTime  time.Time
	sync.Mutex
}

func newClient(connectString string) (*Client, *base.Error) {
	adapter := (internal.IClientAdapter)(nil)

	if urlInfo, e := url.Parse(connectString); e != nil {
		return nil, errors.ErrClientConnectString.AddDebug(e.Error())
	} else if urlInfo.Scheme == "ws" || urlInfo.Scheme == "wss" {
		adapter = websocket.NewWebsocketClientAdapter(connectString)
	} else {
		return nil, errors.ErrClientConnectString.AddDebug(
			fmt.Sprintf("unsupported scheme %s", urlInfo.Scheme),
		)
	}

	ret := &Client{
		status:               clientStatusRunning,
		closeCH:              make(chan bool, 1),
		connectString:        connectString,
		sessionString:        "",
		conn:                 nil,
		preSendHead:          nil,
		preSendTail:          nil,
		config:               gateway.SessionConfig{},
		channels:             nil,
		freeChannels:         nil,
		lastTimeoutCheckTime: base.TimeNow(),
		lastControlSendTime:  base.TimeNow(),
	}

	go func() {
		for atomic.LoadInt32(&ret.status) == clientStatusRunning {
			adapter.Open(ret.onConnRun, ret.onError)

			if atomic.LoadInt32(&ret.status) == clientStatusRunning {
				time.Sleep(time.Second)
			}
		}
	}()

	go func() {
		for atomic.LoadInt32(&ret.status) == clientStatusRunning {
			now := base.TimeNow()
			ret.tryToTimeout(now)
			ret.tryToDeliverPreSendMessages()
			ret.tryToSendPing(now)
			time.Sleep(100 * time.Millisecond)
		}

		atomic.StoreInt32(&ret.status, clientStatusClosed)
		ret.closeCH <- true
	}()

	return ret, nil
}

func (p *Client) onError(err *base.Error) {
	fmt.Println("client onError: ", err)
}

// Close ...
func (p *Client) Close() bool {
	if !atomic.CompareAndSwapInt32(
		&p.status,
		clientStatusRunning,
		clientStatusClosing,
	) {
		p.onError(errors.ErrClientNotRunning)
		return false
	}

	err := func() *base.Error {
		p.Lock()
		defer p.Unlock()

		if p.conn != nil {
			return p.conn.Close()
		} else {
			return nil
		}
	}()

	if err != nil {
		p.onError(errors.ErrClientNotRunning)
		return false
	}

	<-p.closeCH
	return true
}

func (p *Client) initConn(conn internal.IStreamConn) *base.Error {
	sendStream := core.NewStream()
	backStream := (*core.Stream)(nil)

	defer func() {
		sendStream.Release()
		if backStream != nil {
			backStream.Release()
		}
	}()

	sendStream.SetCallbackID(0)
	sendStream.WriteInt64(core.ControlStreamConnectRequest)
	sendStream.WriteString(p.sessionString)

	if err := conn.WriteStream(sendStream, 3*time.Second); err != nil {
		return err
	} else if backStream, err = conn.ReadStream(3*time.Second, 0); err != nil {
		return err
	} else if backStream.GetCallbackID() != 0 {
		return errors.ErrStream
	} else if kind, err := backStream.ReadInt64(); err != nil {
		return err
	} else if kind != core.ControlStreamConnectResponse {
		return errors.ErrStream
	} else if sessionString, err := backStream.ReadString(); err != nil {
		return err
	} else if config, err := gateway.ReadSessionConfig(backStream); err != nil {
		return err
	} else if !backStream.IsReadFinish() {
		return errors.ErrStream
	} else {
		p.Lock()
		defer p.Unlock()

		if sessionString != p.sessionString {
			// new session
			p.sessionString = sessionString
			p.config = config
			p.channels = make([]Channel, config.NumOfChannels())
			p.freeChannels = NewFreeChannelStack(int(config.NumOfChannels()))
			for i := 0; i < len(p.channels); i++ {
				p.channels[i].id = i
				p.channels[i].client = p
				p.channels[i].seq = uint64(config.NumOfChannels()) + uint64(i)
				p.channels[i].item = nil
				p.freeChannels.Push(i)
			}
		} else if !p.config.Equals(&config) {
			// old session, but config changes
			p.sessionString = ""
			return errors.ErrClientConfigChanges
		} else {
			// old session
		}

		return nil
	}
}

func (p *Client) setConn(conn internal.IStreamConn) {
	p.Lock()
	defer p.Unlock()
	p.conn = conn
}

func (p *Client) onCallbackStream(stream *core.Stream, callbackID uint64) {
	p.Lock()
	defer p.Unlock()

	if p.channels != nil {
		if chSize := uint64(len(p.channels)); chSize > 0 {
			p.channels[callbackID%chSize].OnCallbackStream(stream)
		}
	}
}

func (p *Client) onConnRun(conn internal.IStreamConn) {
	// init conn
	if err := p.initConn(conn); err != nil {
		p.onError(err)
		return
	}

	err := (*base.Error)(nil)
	p.setConn(conn)

	defer func() {
		p.setConn(nil)

		if err != nil {
			p.onError(err)
		}

		if err := conn.Close(); err != nil {
			p.onError(err)
		}
	}()

	// receive messages
	for atomic.LoadInt32(&p.status) == clientStatusRunning {
		if stream, e := conn.ReadStream(
			p.config.ReadTimeout(),
			p.config.TransLimit(),
		); e != nil {
			if e != errors.ErrStreamConnIsClosed {
				err = e
			}
			return
		} else if callbackID := stream.GetCallbackID(); callbackID > 0 {
			p.onCallbackStream(stream, callbackID)
		} else if kind, e := stream.ReadInt64(); e != nil {
			err = e
			return
		} else if kind == core.ControlStreamPong {
			// ignore
		} else {
			// broadcast message is not supported now
			err = errors.ErrStream
			return
		}
	}
}

func (p *Client) tryToSendPing(now time.Time) {
	p.Lock()
	defer p.Unlock()

	deltaTime := now.Sub(p.lastControlSendTime)

	if p.conn == nil {
		return
	} else if deltaTime < p.config.Heartbeat() {
		return
	} else {
		// Send Ping
		p.lastControlSendTime = now
		sendStream := core.NewStream()
		defer sendStream.Release()
		sendStream.SetCallbackID(0)
		sendStream.WriteInt64(core.ControlStreamPing)
		if err := p.conn.WriteStream(
			sendStream,
			p.config.WriteTimeout(),
		); err != nil {
			p.onError(err)
		}
	}
}

func (p *Client) tryToTimeout(now time.Time) {
	p.Lock()
	defer p.Unlock()

	if now.Sub(p.lastTimeoutCheckTime) > 800*time.Millisecond {
		p.lastTimeoutCheckTime = now

		// sweep pre send list
		preValidItem := (*SendItem)(nil)
		item := p.preSendHead
		for item != nil {
			if item.CheckAndTimeout(now) {
				nextItem := item.next

				if preValidItem == nil {
					p.preSendHead = nextItem
				} else {
					preValidItem.next = nextItem
				}

				if item == p.preSendTail {
					p.preSendTail = preValidItem
					if p.preSendTail != nil {
						p.preSendTail.next = nil
					}
				}
				item = nextItem
			} else {
				preValidItem = item
				item = item.next
			}
		}

		// sweep the channels
		for i := 0; i < len(p.channels); i++ {
			p.channels[i].OnTimeout(now)
		}
	}
}

func (p *Client) tryToDeliverPreSendMessages() {
	p.Lock()
	defer p.Unlock()

	for p.tryToDeliverPreSendOneMessage() {
	}
}

func (p *Client) tryToDeliverPreSendOneMessage() bool {
	if atomic.LoadInt32(&p.status) != clientStatusRunning { // not running
		return false
	} else if p.conn == nil { // not connected
		return false
	} else if p.preSendHead == nil { // preSend queue is empty
		return false
	} else if p.channels == nil {
		return false
	} else if channelID, ok := p.freeChannels.Pop(); !ok {
		return false
	} else {
		channel := &p.channels[channelID]

		// get and set the send item
		item := p.preSendHead
		if item == p.preSendTail {
			p.preSendHead = nil
			p.preSendTail = nil
		} else {
			p.preSendHead = p.preSendHead.next
		}

		item.id = atomic.LoadUint64(&channel.seq)
		item.next = nil
		item.sendStream.SetCallbackID(item.id)
		channel.item = item

		// try to send
		if err := p.conn.WriteStream(
			item.sendStream,
			p.config.WriteTimeout(),
		); err != nil {
			p.onError(err)
			return false
		}

		item.sendTime = base.TimeNow()
		return true
	}
}

// SendMessage ...
func (p *Client) SendMessage(
	timeout time.Duration,
	target string,
	args ...interface{},
) (interface{}, *base.Error) {
	item := newSendItem()
	defer item.Release()

	item.timeout = timeout

	// set depth
	item.sendStream.SetDepth(0)
	// write target
	item.sendStream.WriteString(target)
	// write from
	item.sendStream.WriteString("@")
	// write args
	for i := 0; i < len(args); i++ {
		if reason := item.sendStream.Write(args[i]); reason != core.StreamWriteOK {
			return nil, errors.ErrUnsupportedValue.AddDebug(reason)
		}
	}

	// add item to the list tail
	p.Lock()
	if p.preSendTail == nil {
		p.preSendHead = item
		p.preSendTail = item
	} else {
		p.preSendTail.next = item
		p.preSendTail = item
	}
	p.Unlock()

	p.tryToDeliverPreSendMessages()

	// wait for response
	retStream := <-item.returnCH
	defer retStream.Release()
	return core.ParseResponseStream(retStream)
}

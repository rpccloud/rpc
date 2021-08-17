package router

import (
	"crypto/tls"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

const (
	channelActionInit     = 1
	channelActionContinue = 2
	channelActionReset    = 3
	channelActionData     = 4
	channelActionTimer    = 5
)

// ConnectMeta ...
type ConnectMeta struct {
	addr      string
	tlsConfig *tls.Config
	id        *base.GlobalID
}

func connReadBytes(
	conn net.Conn,
	timeout time.Duration,
	b []byte,
) (int, *base.Error) {
	pos := 0
	length := 2

	if len(b) < 2 {
		return -1, base.ErrRouterConnProtocol
	}

	if e := conn.SetReadDeadline(base.TimeNow().Add(timeout)); e != nil {
		return -1, base.ErrRouterConnRead.AddDebug(e.Error())
	}

	for pos < length {
		n, e := conn.Read(b[pos:])

		if e != nil {
			return -1, base.ErrRouterConnRead.AddDebug(e.Error())
		}

		pos += n

		if length == 2 && pos >= 2 {
			length = int(binary.LittleEndian.Uint16(b))
			if length > len(b) {
				return -1, base.ErrRouterConnProtocol
			}
		}
	}

	return length, nil
}

func connWriteBytes(
	conn net.Conn,
	timeout time.Duration,
	b []byte,
) *base.Error {
	pos := 0

	if len(b) < 2 || int(binary.LittleEndian.Uint16(b)) != len(b) {
		return base.ErrRouterConnProtocol
	}

	if e := conn.SetWriteDeadline(base.TimeNow().Add(timeout)); e != nil {
		return base.ErrRouterConnWrite.AddDebug(e.Error())
	}

	for pos < len(b) {
		n, e := conn.Write(b[pos:])
		if e != nil {
			return base.ErrRouterConnWrite.AddDebug(e.Error())
		}
		pos += n
	}

	return nil
}

// Channel ...
type Channel struct {
	needReset              uint8
	conn                   net.Conn
	streamCH               chan *rpc.Stream
	streamReceiver         rpc.IStreamReceiver
	sendPrepareSequence    uint64
	sendConfirmSequence    uint64
	sendBuffers            [numOfCacheBuffer][bufferSize]byte
	receiveSequence        uint64
	receiveBuffer          [bufferSize]byte
	receiveStreamGenerator *rpc.StreamGenerator
	closeCH                chan bool
	orcManager             *base.ORCManager
	sync.Mutex
}

// NewChannel ...
func NewChannel(
	index uint16,
	connMeta *ConnectMeta,
	streamCH chan *rpc.Stream,
	streamReceiver rpc.IStreamReceiver,
) *Channel {
	ret := &Channel{
		needReset:              0,
		conn:                   nil,
		streamCH:               streamCH,
		streamReceiver:         streamReceiver,
		sendPrepareSequence:    0,
		sendConfirmSequence:    0,
		receiveSequence:        0,
		receiveStreamGenerator: rpc.NewStreamGenerator(streamReceiver),
		closeCH:                make(chan bool),
		orcManager:             base.NewORCManager(),
	}

	ret.orcManager.Open(func() bool {
		return true
	})

	if connMeta != nil {
		go func() {
			ret.orcManager.Run(func(isRunning func() bool) {
				for isRunning() {
					startNS := base.TimeNow().UnixNano()
					if err := ret.runMasterThread(index, connMeta); err != nil {
						streamReceiver.OnReceiveStream(
							rpc.MakeSystemErrorStream(err),
						)
					}
					base.WaitWhileRunning(
						startNS,
						isRunning,
						time.Second,
					)
				}
			})
		}()
	}

	return ret
}

func (p *Channel) runMasterThread(
	index uint16,
	connMeta *ConnectMeta,
) *base.Error {
	var conn net.Conn
	var e error
	buffer := make([]byte, 32)

	// dial
	if connMeta.tlsConfig == nil {
		conn, e = net.Dial("tcp", connMeta.addr)
	} else {
		conn, e = tls.Dial("tcp", connMeta.addr, connMeta.tlsConfig)
	}

	// deal dial error
	if e != nil {
		return base.ErrRouterConnDial.AddDebug(e.Error())
	}

	// make init frame
	binary.LittleEndian.PutUint16(buffer, 32)
	binary.LittleEndian.PutUint16(buffer[2:], channelActionInit)
	binary.LittleEndian.PutUint16(buffer[4:], index)
	binary.LittleEndian.PutUint64(buffer[6:], connMeta.id.GetID())
	buffer[14] = p.needReset
	binary.LittleEndian.PutUint64(buffer[16:], p.sendConfirmSequence)
	binary.LittleEndian.PutUint64(buffer[24:], p.receiveSequence)

	// send init frame
	if err := connWriteBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	// read response frame
	if _, err := connReadBytes(conn, time.Second, buffer); err != nil {
		return err
	}

	sendSequence := uint64(0)
	switch binary.LittleEndian.Uint16(buffer[2:]) {
	case channelActionContinue:
		rReceiveSequence := binary.LittleEndian.Uint64(buffer[24:])
		if rReceiveSequence < p.sendConfirmSequence ||
			rReceiveSequence > p.sendPrepareSequence {
			p.needReset = 1
			return base.ErrRouterConnProtocol
		}
		sendSequence = rReceiveSequence
	case channelActionReset:
		p.needReset = 0
		p.sendPrepareSequence = 0
		p.sendConfirmSequence = 0
		p.receiveSequence = 0
		p.receiveStreamGenerator.Reset()
		sendSequence = 0
	default:
		p.needReset = 1
		return base.ErrRouterConnProtocol
	}

	p.RunWithConn(sendSequence, conn)

	return nil
}

func (p *Channel) runSlaveThread(
	conn net.Conn,
	initBuffer [32]byte,
) *base.Error {
	rNeedReset := initBuffer[14]
	rSendConfirmSequence := binary.LittleEndian.Uint64(initBuffer[16:])
	rReceiveSequence := binary.LittleEndian.Uint64(initBuffer[24:])

	sendSequence := uint64(0)
	if p.needReset == 0 && rNeedReset == 0 &&
		rSendConfirmSequence <= p.receiveSequence &&
		rReceiveSequence >= p.sendConfirmSequence &&
		rReceiveSequence <= p.sendPrepareSequence {
		binary.LittleEndian.PutUint16(initBuffer[2:], channelActionContinue)
		binary.LittleEndian.PutUint64(initBuffer[24:], p.receiveSequence)
		sendSequence = rReceiveSequence
	} else {
		p.needReset = 0
		p.sendPrepareSequence = 0
		p.sendConfirmSequence = 0
		p.receiveSequence = 0
		p.receiveStreamGenerator.Reset()
		binary.LittleEndian.PutUint16(initBuffer[2:], channelActionReset)
		sendSequence = 0
	}

	// send response frame
	if err := connWriteBytes(conn, time.Second, initBuffer[:]); err != nil {
		return err
	}

	// run async
	go func() {
		p.RunWithConn(sendSequence, conn)
	}()

	return nil
}

// RunWithConn ...
func (p *Channel) RunWithConn(sendSequence uint64, conn net.Conn) {
	running := uint32(1)

	isRunning := func() bool {
		return p.orcManager.IsRunning() && atomic.LoadUint32(&running) == 1
	}

	p.conn = conn

	sendCH := make(chan uint64, numOfCacheBuffer)
	for id := sendSequence + 1; id <= p.sendPrepareSequence; id++ {
		sendCH <- id
	}

	waitCH := make(chan bool)

	go func() {
		p.runMakeFrame(isRunning, sendCH)
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		if err := p.runRead(conn, isRunning); err != nil {
			p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(err))
		}
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	go func() {
		if err := p.runWrite(conn, isRunning, sendCH); err != nil {
			p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(err))
		}
		atomic.StoreUint32(&running, 0)
		_ = conn.Close()
		waitCH <- true
	}()

	<-waitCH
	<-waitCH
	<-waitCH
	p.conn = nil
}

func (p *Channel) canPrepare() bool {
	sendPrepare := atomic.LoadUint64(&p.sendPrepareSequence)
	sendConfirm := atomic.LoadUint64(&p.sendConfirmSequence)
	return sendPrepare-sendConfirm < numOfCacheBuffer
}

func (p *Channel) runMakeFrame(isRunning func() bool, sendCH chan uint64) {
	stream := (*rpc.Stream)(nil)

	for {
		// wait for valid frame id
		for isRunning() && !p.canPrepare() {
			time.Sleep(30 * time.Millisecond)
		}

		if !isRunning() {
			return
		}

		frameID := atomic.AddUint64(&p.sendPrepareSequence, 1)

		// init data frame
		frameBuffer := p.sendBuffers[frameID%numOfCacheBuffer]
		binary.LittleEndian.PutUint16(frameBuffer[2:], channelActionData)
		binary.LittleEndian.PutUint64(frameBuffer[4:], frameID)

		// gat first stream
		if stream == nil {
			stream = <-p.streamCH
		}

		// write stream to buffer
		bufferPos := 20
		streamPos := 0
		for stream != nil && bufferSize-bufferPos >= rpc.StreamBlockSize {
			peekBuf, finish := stream.PeekBufferSlice(
				streamPos, bufferSize-bufferPos,
			)
			copyLen := copy(frameBuffer[bufferPos:], peekBuf)
			streamPos += copyLen
			bufferPos += copyLen

			if finish {
				streamPos = 0
				select {
				case stream = <-p.streamCH:
				default:
					stream = nil
				}
			}
		}
		binary.LittleEndian.PutUint16(frameBuffer[:], uint16(bufferPos))
		sendCH <- frameID
	}
}

func (p *Channel) runRead(conn net.Conn, isRunning func() bool) *base.Error {
	for isRunning() {
		n, err := connReadBytes(conn, 3*time.Second, p.receiveBuffer[:])
		channelAction := binary.LittleEndian.Uint16(p.receiveBuffer[2:])

		if !isRunning() {
			return nil
		}

		rReceiveSequence := uint64(0)

		if err != nil {
			return err
		} else if n < 20 {
			return base.ErrRouterConnProtocol
		} else if channelAction == channelActionData {
			rSendSequence := binary.LittleEndian.Uint64(p.receiveBuffer[4:])
			rReceiveSequence = binary.LittleEndian.Uint64(p.receiveBuffer[12:])

			// check rSendSequence
			if rSendSequence != p.receiveSequence+1 {
				p.needReset = 1
				return base.ErrRouterConnProtocol
			}

			// process receiveBuffer
			err = p.receiveStreamGenerator.OnBytes(p.receiveBuffer[20:])
			if err != nil {
				p.needReset = 1
				return err
			}

			// update receiveSequence
			atomic.StoreUint64(&p.receiveSequence, rSendSequence)
		} else if channelAction == channelActionTimer {
			rReceiveSequence = binary.LittleEndian.Uint64(p.receiveBuffer[4:])
		} else {
			return base.ErrRouterConnProtocol
		}

		// check rReceiveSequence
		if rReceiveSequence <= atomic.LoadUint64(&p.sendPrepareSequence) {
			p.needReset = 1
			return base.ErrRouterConnProtocol
		}

		// update sendConfirmSequence
		if rReceiveSequence > atomic.LoadUint64(&p.sendConfirmSequence) {
			atomic.StoreUint64(&p.sendConfirmSequence, rReceiveSequence)
		}
	}

	return nil
}

func (p *Channel) runWrite(
	conn net.Conn,
	isRunning func() bool,
	sendCH chan uint64,
) *base.Error {
	var timerBuffer [12]byte
	binary.LittleEndian.PutUint16(timerBuffer[:], 12)
	binary.LittleEndian.PutUint16(timerBuffer[2:], channelActionTimer)
	receiveSequence := uint64(0)

	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case id := <-sendCH:
			buffer := p.sendBuffers[id%numOfCacheBuffer][:]
			receiveSequence = atomic.LoadUint64(&p.receiveSequence)
			binary.LittleEndian.PutUint64(buffer[12:], receiveSequence)
			if err := connWriteBytes(conn, time.Second, buffer); err != nil {
				return err
			}
		case <-timer.C:
			if !isRunning() {
				return nil
			}

			if v := atomic.LoadUint64(&p.receiveSequence); v > receiveSequence {
				binary.LittleEndian.PutUint64(timerBuffer[4:], v)
				if err := connWriteBytes(
					conn, time.Second, timerBuffer[:],
				); err != nil {
					return err
				}
			}
		}
	}
}

// Close ...
func (p *Channel) Close() {
	p.orcManager.Close(nil, nil)
}

package adapter

import (
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

// SyncConn ...
type SyncConn struct {
	isServer  bool
	isRunning bool
	conn      net.Conn
	next      IConn
	rBuf      []byte
	wBuf      []byte
	mu        sync.Mutex
}

// NewServerSyncConn ...
func NewServerSyncConn(netConn net.Conn, rBufSize int, wBufSize int) *SyncConn {
	return &SyncConn{
		isServer:  true,
		isRunning: true,
		conn:      netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// NewClientSyncConn ...
func NewClientSyncConn(netConn net.Conn, rBufSize int, wBufSize int) *SyncConn {
	return &SyncConn{
		isServer:  false,
		isRunning: true,
		conn:      netConn,
		next:      nil,
		rBuf:      make([]byte, rBufSize),
		wBuf:      make([]byte, wBufSize),
	}
}

// SetNext ...
func (p *SyncConn) SetNext(next IConn) {
	p.next = next
}

// OnOpen ...
func (p *SyncConn) OnOpen() {
	p.next.OnOpen()
}

// OnClose ...
func (p *SyncConn) OnClose() {
	p.next.OnClose()
}

// OnError ...
func (p *SyncConn) OnError(err *base.Error) {
	p.next.OnError(err)
}

// Close ...
func (p *SyncConn) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		p.isRunning = false
		if e := p.conn.Close(); e != nil {
			p.OnError(base.ErrConnClose.AddDebug(e.Error()))
		}
	}
}

// LocalAddr ...
func (p *SyncConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

// RemoteAddr ...
func (p *SyncConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

// OnReadReady ...
func (p *SyncConn) OnReadReady() bool {
	n, e := p.conn.Read(p.rBuf)
	if e != nil {
		if e != io.EOF {
			if p.isServer {
				p.OnError(base.ErrConnRead.AddDebug(e.Error()))
			} else {
				p.mu.Lock()
				ignoreReport := (!p.isRunning) &&
					strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix)
				p.mu.Unlock()

				if !ignoreReport {
					p.OnError(base.ErrConnRead.AddDebug(e.Error()))
				}
			}
		}

		return false
	}

	p.next.OnReadBytes(p.rBuf[:n])
	return true
}

// OnWriteReady ...
func (p *SyncConn) OnWriteReady() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	isTriggerFinish := false

	for !isTriggerFinish {
		bufLen := 0

		for !isTriggerFinish && bufLen < len(p.wBuf) {
			if n := p.next.OnFillWrite(p.wBuf[bufLen:]); n > 0 {
				bufLen += n
			} else {
				isTriggerFinish = true
			}
		}

		start := 0
		for start < bufLen {
			if n, e := p.conn.Write(p.wBuf[start:bufLen]); e != nil {
				p.OnError(base.ErrConnWrite.AddDebug(e.Error()))
				return false
			} else if n == 0 {
				return false
			} else {
				start += n
			}
		}
	}

	return true
}

// OnReadBytes ...
func (p *SyncConn) OnReadBytes(_ []byte) {
	panic("kernel error: it should not be called")
}

// OnFillWrite ...
func (p *SyncConn) OnFillWrite(_ []byte) int {
	panic("kernel error: it should not be called")
}

const streamConnStatusRunning = int32(1)
const streamConnStatusClosed = int32(0)

// StreamConn ...
type StreamConn struct {
	isDebug             bool
	status              int32
	prev                IConn
	receiver            IReceiver
	writeCH             chan *rpc.Stream
	writeStream         *rpc.Stream
	readStreamGenerator *rpc.StreamGenerator
	writePos            int
	activeTimeNS        int64
}

// NewStreamConn ...
func NewStreamConn(isDebug bool, prev IConn, receiver IReceiver) *StreamConn {
	ret := &StreamConn{
		isDebug:             isDebug,
		status:              streamConnStatusRunning,
		prev:                prev,
		receiver:            receiver,
		writeCH:             make(chan *rpc.Stream, 16),
		readStreamGenerator: nil,
		writeStream:         nil,
		writePos:            0,
		activeTimeNS:        base.TimeNow().UnixNano(),
	}
	ret.readStreamGenerator = rpc.NewStreamGenerator(ret)
	return ret
}

// SetReceiver ...
func (p *StreamConn) SetReceiver(receiver IReceiver) {
	p.receiver = receiver
}

// OnOpen ...
func (p *StreamConn) OnOpen() {
	p.receiver.OnConnOpen(p)
}

// OnClose ...
func (p *StreamConn) OnClose() {
	p.receiver.OnConnClose(p)
}

// OnError ...
func (p *StreamConn) OnError(err *base.Error) {
	p.receiver.OnConnError(p, err)
}

// OnReceiveStream ...
func (p *StreamConn) OnReceiveStream(stream *rpc.Stream) {
	atomic.StoreInt64(&p.activeTimeNS, base.TimeNow().UnixNano())
	if p.isDebug {
		stream.SetStatusBitDebug()
		stream.BuildStreamCheck()
	}
	p.receiver.OnConnReadStream(p, stream)
}

// OnReadBytes ...
func (p *StreamConn) OnReadBytes(b []byte) {
	if err := p.readStreamGenerator.OnBytes(b); err != nil {
		p.receiver.OnConnError(p, base.ErrStream)
	}
}

// OnFillWrite ...
func (p *StreamConn) OnFillWrite(b []byte) int {
	if p.writeStream == nil {
		select {
		case stream := <-p.writeCH:
			p.writeStream = stream
			p.writePos = 0
		default:
			return 0
		}
	}

	if p.writeStream == nil {
		return 0
	}

	peekBuf, finish := p.writeStream.PeekBufferSlice(p.writePos, len(b))

	if len(peekBuf) <= 0 {
		p.OnError(
			base.ErrOnFillWriteFatal,
		)
		return 0
	}

	copyBytes := copy(b, peekBuf)
	p.writePos += copyBytes

	if finish {
		p.writeStream.Release()
		p.writeStream = nil
		p.writePos = 0
	}

	return copyBytes
}

// Close ...
func (p *StreamConn) Close() {
	if atomic.CompareAndSwapInt32(
		&p.status,
		streamConnStatusRunning,
		streamConnStatusClosed,
	) {
		close(p.writeCH)
		p.prev.Close()
	}
}

// LocalAddr ...
func (p *StreamConn) LocalAddr() net.Addr {
	return p.prev.LocalAddr()
}

// RemoteAddr ...
func (p *StreamConn) RemoteAddr() net.Addr {
	return p.prev.RemoteAddr()
}

// WriteStreamAndRelease ...
func (p *StreamConn) WriteStreamAndRelease(stream *rpc.Stream) {
	func() {
		defer func() {
			_ = recover()
		}()

		stream.BuildStreamCheck()
		p.writeCH <- stream
	}()

	p.prev.OnWriteReady()
}

// IsActive ...
func (p *StreamConn) IsActive(nowNS int64, timeout time.Duration) bool {
	return nowNS-atomic.LoadInt64(&p.activeTimeNS) < int64(timeout)
}

// SetNext ...
func (p *StreamConn) SetNext(_ IConn) {
	panic("kernel error: it should not be called")
}

// OnReadReady ...
func (p *StreamConn) OnReadReady() bool {
	panic("kernel error: it should not be called")
}

// OnWriteReady ...
func (p *StreamConn) OnWriteReady() bool {
	panic("kernel error: it should not be called")
}

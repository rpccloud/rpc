package async

import (
	"github.com/rpccloud/rpc/internal/adapter"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/errors"
	"golang.org/x/sys/unix"
	"net"
	"sync"
)

type Conn struct {
	channel       *Channel
	fd            int
	next          adapter.XConn
	lAddr         net.Addr
	rAddr         net.Addr
	rBuf          []byte
	wBuf          []byte
	wStartPos     int
	wEndPos       int
	canWriteReady bool
	sync.Mutex
}

func NewConn(
	channel *Channel,
	fd int,
	lAddr net.Addr,
	rAddr net.Addr,
	rBufSize int,
	wBufSize int,
) *Conn {
	return &Conn{
		channel:       channel,
		fd:            fd,
		next:          nil,
		lAddr:         lAddr,
		rAddr:         rAddr,
		rBuf:          make([]byte, rBufSize),
		wBuf:          make([]byte, wBufSize),
		wStartPos:     0,
		wEndPos:       0,
		canWriteReady: false,
	}
}

func (p *Conn) SetNext(next adapter.XConn) {
	p.next = next
}

func (p *Conn) OnReadReady() {
	if n, e := readFD(p.fd, p.rBuf); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	} else {
		p.OnReadBytes(p.rBuf[:n])
	}
}

func (p *Conn) DoWrite() bool {
	isFillFinish := false

	// fill buffer
	for !isFillFinish && p.wEndPos < len(p.wBuf) {
		n := 0
		if n, isFillFinish = p.OnFillWrite(p.wBuf[p.wEndPos:]); n > 0 {
			p.wEndPos += n
		}
	}

	// write buffer
	if n, e := writeFD(p.fd, p.wBuf[p.wStartPos:p.wEndPos]); e != nil {
		if e == unix.EINTR {
			return false
		} else {
			p.OnError(errors.ErrTemp.AddDebug(e.Error()))
		}
	} else {
		p.wStartPos += n
	}

	if p.wStartPos == p.wEndPos {
		p.wStartPos = 0
		p.wEndPos = 0
		return isFillFinish
	}

	return false
}

func (p *Conn) OnWriteReady() {
	p.Lock()
	defer p.Unlock()

	if p.canWriteReady {
		p.canWriteReady = !p.DoWrite()
	}
}

func (p *Conn) OnOpen() {
	p.next.OnOpen()
}

func (p *Conn) OnClose() {
	p.next.OnClose()
}

func (p *Conn) OnError(err *base.Error) {
	p.next.OnError(err)
}

func (p *Conn) OnReadBytes(b []byte) {
	p.next.OnReadBytes(b)
}

func (p *Conn) OnFillWrite(b []byte) (int, bool) {
	return p.next.OnFillWrite(b)
}

func (p *Conn) TriggerWrite() {
	p.Lock()
	defer p.Unlock()

	p.canWriteReady = !p.DoWrite()
}

func (p *Conn) Close() {
	if e := closeFD(p.fd); e != nil {
		p.OnError(errors.ErrTemp.AddDebug(e.Error()))
	}
}

func (p *Conn) LocalAddr() net.Addr {
	return p.lAddr
}

func (p *Conn) RemoteAddr() net.Addr {
	return p.rAddr
}

func (p *Conn) GetFD() int {
	return p.fd
}

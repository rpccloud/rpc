package router

import (
	"encoding/binary"
	"net"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

const (
	numOfChannelPerSlot = 8
	numOfCacheBuffer    = 512
	bufferSize          = 65536
)

// Slot ...
type Slot struct {
	dataCH       chan *rpc.Stream
	dataChannels []*Channel
}

// NewSlot ...
func NewSlot(
	connectMeta *ConnectMeta,
	streamReceiver rpc.IStreamReceiver,
) *Slot {
	ret := &Slot{
		dataCH:       make(chan *rpc.Stream, 8192),
		dataChannels: make([]*Channel, numOfChannelPerSlot),
	}

	for i := 0; i < numOfChannelPerSlot; i++ {
		ret.dataChannels[i] = NewChannel(
			uint16(i), connectMeta, ret.dataCH, streamReceiver,
		)
	}

	return ret
}

func (p *Slot) AddSlaveConn(conn net.Conn, initBuffer [32]byte) *base.Error {
	index := binary.LittleEndian.Uint16(initBuffer[4:])
	if index < numOfChannelPerSlot && conn != nil {
		channel := p.dataChannels[index]
		if err := channel.runSlaveThread(conn, initBuffer); err != nil {
			return err
		}

		return nil
	}

	return base.ErrRouterConnProtocol
}

// SendStream ...
func (p *Slot) SendStream(s *rpc.Stream) (ret bool) {
	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	if s != nil {
		p.dataCH <- s
		return true
	}

	return false
}

// Close ...
func (p *Slot) Close() {
	for i := 0; i < numOfChannelPerSlot; i++ {
		p.dataChannels[i].Close()
	}

	close(p.dataCH)
}

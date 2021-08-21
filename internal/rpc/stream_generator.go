package rpc

import "github.com/rpccloud/rpc/internal/base"

// StreamGenerator ...
type StreamGenerator struct {
	streamReceiver IStreamReceiver
	streamPos      int
	streamBuffer   []byte
	stream         *Stream
}

// NewStreamGenerator ...
func NewStreamGenerator(streamReceiver IStreamReceiver) *StreamGenerator {
	return &StreamGenerator{
		streamReceiver: streamReceiver,
		streamPos:      0,
		streamBuffer:   make([]byte, StreamHeadSize),
		stream:         nil,
	}
}

// Reset ...
func (p *StreamGenerator) Reset() {
	p.streamPos = 0
	if p.stream != nil {
		p.stream.Release()
		p.stream = nil
	}
}

// OnBytes ...
func (p *StreamGenerator) OnBytes(b []byte) *base.Error {
	// fill header
	if p.stream == nil {
		if p.streamPos < StreamHeadSize {
			copyBytes := copy(p.streamBuffer[p.streamPos:], b)
			p.streamPos += copyBytes
			b = b[copyBytes:]
		}

		if p.streamPos < StreamHeadSize {
			// not error
			return nil
		}

		p.stream = NewStream()
		p.stream.PutBytesTo(p.streamBuffer, 0)
		p.streamPos = 0
	}

	// fill body
	byteLen := len(b)
	streamLength := int(p.stream.GetLength())
	remains := streamLength - p.stream.GetWritePos()

	if remains < 0 {
		return base.ErrStream
	}

	writeBuf := b[:base.MinInt(byteLen, remains)]
	p.stream.PutBytes(writeBuf)
	streamPos := p.stream.GetWritePos()

	if streamPos == streamLength {
		if p.stream.CheckStream() {
			p.streamReceiver.OnReceiveStream(p.stream)
			p.stream = nil
		} else {
			return base.ErrStream
		}
	}

	if byteLen > len(writeBuf) {
		return p.OnBytes(b[len(writeBuf):])
	}

	return nil
}

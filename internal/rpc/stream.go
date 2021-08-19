package rpc

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
)

const (
	streamVersion            = 1
	streamBlockSize          = 512
	streamFrameArrayInitSize = 8

	streamPosVersion    = 0
	streamPosStatusBit  = 1
	streamPosKind       = 2
	streamPosPriority   = 3
	streamPosLength     = 4
	streamPosCheckSum   = 8
	streamPosZoneID     = 16
	streamPosTargetID   = 18
	streamPosSourceID   = 26
	streamPosGatewayID  = 34
	streamPosSessionID  = 42
	streamPosCallbackID = 50
	streamPosDepth      = 58
	streamPosBody       = 60

	streamStatusBitDebug = 0

	// StreamBlockSize ...
	StreamBlockSize = streamBlockSize
	// StreamHeadSize ...
	StreamHeadSize = streamPosBody
	// StreamWriteOK ...
	StreamWriteOK = ""
	// StreamWriteOverflow ...
	StreamWriteOverflow = " overflows"
	// StreamWriteIsNotAvailable ...
	StreamWriteIsNotAvailable = " is not available"

	// StreamKindConnectRequest ...
	StreamKindConnectRequest = 1
	// StreamKindConnectResponse ...
	StreamKindConnectResponse = 2
	// StreamKindPing ...
	StreamKindPing = 3
	// StreamKindPong ...
	StreamKindPong = 4
	// StreamKindRPCRequest ...
	StreamKindRPCRequest = 5
	// StreamKindRPCResponseOK ...
	StreamKindRPCResponseOK = 6
	// StreamKindRPCResponseError ...
	StreamKindRPCResponseError = 7
	// StreamKindRPCBoardCast ...
	StreamKindRPCBoardCast = 8
	// StreamKindSystemErrorReport ...
	StreamKindSystemErrorReport = 9
)

var (
	constStreamFrame0 = [streamBlockSize]byte{streamVersion}
	constStreamFrameN = [streamBlockSize]byte{0}
	initStreamFrame0  = constStreamFrame0[0:]
	initStreamFrameN  = constStreamFrameN[0:]
	streamCache       = &base.SyncPool{
		New: func() interface{} {
			ret := &Stream{
				readSeg:    0,
				readIndex:  streamPosBody,
				writeSeg:   0,
				writeIndex: streamPosBody,
			}

			zeroFrame := ret.bufferBytes[0:]
			copy(zeroFrame, initStreamFrame0)
			ret.frames = ret.bufferFrames[0:1]
			ret.frames[0] = &zeroFrame
			ret.readFrame = zeroFrame
			ret.writeFrame = zeroFrame
			return ret
		},
	}
	frameCache = &base.SyncPool{
		New: func() interface{} {
			ret := make([]byte, streamBlockSize)
			return &ret
		},
	}
	readSkipArray = [256]int{
		-0x01, +0x01, +0x01, +0x01, +0x01, +0x09, +0x03, +0x05,
		+0x09, +0x03, +0x05, +0x09, -0x01, -0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
		+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
		+0x01, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08, +0x09,
		+0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10, +0x11,
		+0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18, +0x19,
		+0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20, +0x21,
		+0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28, +0x29,
		+0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30, +0x31,
		+0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38, +0x39,
		+0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x40, +0x00,
		+0x01, +0x02, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08,
		+0x09, +0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10,
		+0x11, +0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18,
		+0x19, +0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20,
		+0x21, +0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28,
		+0x29, +0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30,
		+0x31, +0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38,
		+0x39, +0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x00,
	}
)

// Stream ...
type Stream struct {
	frames []*[]byte

	readSeg   int
	readIndex int
	readFrame []byte

	writeSeg   int
	writeIndex int
	writeFrame []byte

	bufferBytes  [streamBlockSize]byte
	bufferFrames [streamFrameArrayInitSize]*[]byte
}

// NewStream ...
func NewStream() *Stream {
	return streamCache.Get().(*Stream)
}

// Reset ...
func (p *Stream) Reset() {
	copy(*p.frames[0], initStreamFrame0)

	// reset frames
	if len(p.frames) != 1 {
		for i := 1; i < len(p.frames); i++ {
			copy(*(p.frames[i]), initStreamFrameN)
			frameCache.Put(p.frames[i])
			p.frames[i] = nil
		}

		p.frames = p.bufferFrames[0:1]

		if p.readSeg != 0 {
			p.readSeg = 0
			p.readFrame = *(p.frames[0])
		}

		if p.writeSeg != 0 {
			p.writeSeg = 0
			p.writeFrame = *(p.frames[0])
		}
	}

	p.readIndex = streamPosBody
	p.writeIndex = streamPosBody
}

// Release clean the Stream
func (p *Stream) Release() {
	p.Reset()
	streamCache.Put(p)
}

// Clone ...
func (p *Stream) Clone() *Stream {
	ret := NewStream()
	copy(*(ret.frames[0]), *(p.frames[0]))
	for i := 1; i <= p.writeSeg; i++ {
		// use gotoNextWriteFrame do not broke the bufferFrames
		ret.gotoNextWriteFrame()
		copy(*(ret.frames[i]), *(p.frames[i]))
	}

	ret.readSeg = p.readSeg
	ret.readIndex = p.readIndex
	ret.readFrame = *(ret.frames[ret.readSeg])

	ret.writeSeg = p.writeSeg
	ret.writeIndex = p.writeIndex

	ret.writeFrame = *(ret.frames[ret.writeSeg])

	return ret
}

// GetVersion ...
func (p *Stream) GetVersion() uint8 {
	return (*p.frames[0])[streamPosVersion]
}

// SetVersion ...
func (p *Stream) SetVersion(version uint8) {
	(*p.frames[0])[streamPosVersion] = version
}

// HasStatusBitDebug ...
func (p *Stream) HasStatusBitDebug() bool {
	return (*p.frames[0])[streamPosStatusBit]&(1<<streamStatusBitDebug) != 0
}

// SetStatusBitDebug ...
func (p *Stream) SetStatusBitDebug() {
	(*p.frames[0])[streamPosStatusBit] |= 1 << streamStatusBitDebug
}

// ClearStatusBitDebug ...
func (p *Stream) ClearStatusBitDebug() {
	(*p.frames[0])[streamPosStatusBit] &= (1 << streamStatusBitDebug) ^ 0xFF
}

// GetKind ...
func (p *Stream) GetKind() uint8 {
	return (*p.frames[0])[streamPosKind]
}

// SetKind ...
func (p *Stream) SetKind(v uint8) {
	(*p.frames[0])[streamPosKind] = v
}

// GetPriority ...
func (p *Stream) GetPriority() uint8 {
	return (*p.frames[0])[streamPosPriority]
}

// SetPriority ...
func (p *Stream) SetPriority(v uint8) {
	(*p.frames[0])[streamPosPriority] = v
}

// GetLength ...
func (p *Stream) GetLength() uint32 {
	return binary.LittleEndian.Uint32((*p.frames[0])[streamPosLength:])
}

func (p *Stream) getCheckSum() uint64 {
	checksum := uint64(0)
	arr := []uint64(nil)
	arrHeader := (*reflect.SliceHeader)(unsafe.Pointer(&arr))

	if p.writeSeg > 0 {
		for i := 0; i < p.writeSeg; i++ {
			frameHeader := (*reflect.SliceHeader)(unsafe.Pointer(p.frames[i]))
			arrHeader.Len = streamBlockSize / 8
			arrHeader.Cap = streamBlockSize / 8
			arrHeader.Data = frameHeader.Data

			for j := streamBlockSize/8 - 1; j >= 0; j-- {
				checksum ^= arr[j]
			}
		}
	}

	u64End := (p.writeIndex + 7) / 8
	for i := p.writeIndex; i < u64End*8; i++ {
		p.writeFrame[i] = 0
	}

	frameHeader := (*reflect.SliceHeader)(unsafe.Pointer(&p.writeFrame))
	arrHeader.Len = u64End
	arrHeader.Cap = u64End
	arrHeader.Data = frameHeader.Data
	for j := u64End - 1; j >= 0; j-- {
		checksum ^= arr[j]
	}
	return checksum
}

// BuildStreamCheck ...
func (p *Stream) BuildStreamCheck() {
	binary.LittleEndian.PutUint32(
		(*p.frames[0])[streamPosLength:],
		uint32(p.GetWritePos()),
	)

	buf := (*p.frames[0])[streamPosCheckSum:]
	binary.LittleEndian.PutUint64(buf, 0)
	binary.LittleEndian.PutUint64(buf, p.getCheckSum())
}

// CheckStream ...
func (p *Stream) CheckStream() bool {
	return p.getCheckSum() == 0 && uint32(p.GetWritePos()) == p.GetLength()
}

// GetZoneID ...
func (p *Stream) GetZoneID() uint16 {
	return binary.LittleEndian.Uint16((*p.frames[0])[streamPosZoneID:])
}

// SetZoneID ...
func (p *Stream) SetZoneID(v uint16) {
	binary.LittleEndian.PutUint16((*p.frames[0])[streamPosZoneID:], v)
}

// GetTargetID ...
func (p *Stream) GetTargetID() uint64 {
	return binary.LittleEndian.Uint64((*p.frames[0])[streamPosTargetID:])
}

// SetTargetID ...
func (p *Stream) SetTargetID(v uint64) {
	binary.LittleEndian.PutUint64((*p.frames[0])[streamPosTargetID:], v)
}

// GetSourceID ...
func (p *Stream) GetSourceID() uint64 {
	return binary.LittleEndian.Uint64((*p.frames[0])[streamPosSourceID:])
}

// SetSourceID ...
func (p *Stream) SetSourceID(v uint64) {
	binary.LittleEndian.PutUint64((*p.frames[0])[streamPosSourceID:], v)
}

// GetGatewayID ...
func (p *Stream) GetGatewayID() uint64 {
	return binary.LittleEndian.Uint64((*p.frames[0])[streamPosGatewayID:])
}

// SetGatewayID ...
func (p *Stream) SetGatewayID(v uint64) {
	binary.LittleEndian.PutUint64((*p.frames[0])[streamPosGatewayID:], v)
}

// GetSessionID ...
func (p *Stream) GetSessionID() uint64 {
	return binary.LittleEndian.Uint64((*p.frames[0])[streamPosSessionID:])
}

// SetSessionID ...
func (p *Stream) SetSessionID(v uint64) {
	binary.LittleEndian.PutUint64((*p.frames[0])[streamPosSessionID:], v)
}

// GetCallbackID ...
func (p *Stream) GetCallbackID() uint64 {
	return binary.LittleEndian.Uint64((*p.frames[0])[streamPosCallbackID:])
}

// SetCallbackID ...
func (p *Stream) SetCallbackID(v uint64) {
	binary.LittleEndian.PutUint64((*p.frames[0])[streamPosCallbackID:], v)
}

// GetDepth ...
func (p *Stream) GetDepth() uint16 {
	return binary.LittleEndian.Uint16((*p.frames[0])[streamPosDepth:])
}

// SetDepth ...
func (p *Stream) SetDepth(v uint16) {
	binary.LittleEndian.PutUint16((*p.frames[0])[streamPosDepth:], v)
}

// GetReadPos get the current read pos of the stream
func (p *Stream) GetReadPos() int {
	return p.readSeg*streamBlockSize + p.readIndex
}

// SetReadPos set the current read pos of the stream
func (p *Stream) SetReadPos(pos int) bool {
	if pos >= streamPosBody {
		readSeg := pos / streamBlockSize
		readIndex := pos % streamBlockSize
		if (readSeg == p.writeSeg && readIndex <= p.writeIndex) ||
			(readSeg < p.writeSeg) {
			p.readIndex = readIndex
			if p.readSeg != readSeg {
				p.readSeg = readSeg
				p.readFrame = *(p.frames[readSeg])
			}
			return true
		}
	}
	return false
}

// SetReadPosToBodyStart ...
func (p *Stream) SetReadPosToBodyStart() {
	p.readIndex = streamPosBody
	if p.readSeg != 0 {
		p.readSeg = 0
		p.readFrame = *(p.frames[0])
	}
}

// GetWritePos ...
func (p *Stream) GetWritePos() int {
	return p.writeSeg*streamBlockSize + p.writeIndex
}

// SetWritePos ...
func (p *Stream) SetWritePos(pos int) bool {
	if pos >= streamPosBody {
		numToCreate := pos/streamBlockSize - len(p.frames) + 1
		for numToCreate > 0 {
			p.frames = append(p.frames, frameCache.Get().(*[]byte))
			numToCreate--
		}

		p.writeIndex = pos % streamBlockSize
		writeSeg := pos / streamBlockSize
		if p.writeSeg != writeSeg {
			p.writeSeg = writeSeg
			p.writeFrame = *(p.frames[writeSeg])
		}
		return true
	}
	return false
}

// SetWritePosToBodyStart ...
func (p *Stream) SetWritePosToBodyStart() {
	p.writeIndex = streamPosBody
	if p.writeSeg != 0 {
		p.writeSeg = 0
		p.writeFrame = *(p.frames[0])
	}
}

// CanRead return true if the stream is not finish
func (p *Stream) CanRead() bool {
	return p.readIndex < p.writeIndex || p.readSeg < p.writeSeg
}

// IsReadFinish return true if the stream is finish
func (p *Stream) IsReadFinish() bool {
	return p.readIndex == p.writeIndex && p.readSeg == p.writeSeg
}

func (p *Stream) gotoNextWriteFrame() {
	p.writeSeg++
	p.writeIndex = 0
	if p.writeSeg == len(p.frames) {
		p.frames = append(p.frames, frameCache.Get().(*[]byte))
	}
	p.writeFrame = *(p.frames[p.writeSeg])
}

func (p *Stream) gotoNextReadFrameUnsafe() {
	p.readSeg++
	p.readIndex = 0
	p.readFrame = *(p.frames[p.readSeg])
}

func (p *Stream) gotoNextReadByteUnsafe() {
	p.readIndex++
	if p.readIndex == streamBlockSize {
		p.gotoNextReadFrameUnsafe()
	}
}

func (p *Stream) hasNBytesToRead(n int) bool {
	return p.GetReadPos()+n <= p.GetWritePos()
}

func (p *Stream) isSafetyReadNBytesInCurrentFrame(n int) bool {
	lineEnd := p.readIndex + n
	return (lineEnd < streamBlockSize) &&
		(lineEnd <= p.writeIndex || p.readSeg < p.writeSeg)
}

func (p *Stream) peekNBytesCrossFrameUnsafe(n int) []byte {
	v := make([]byte, n)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	copy(v[copyBytes:], *(p.frames[p.readSeg+1]))
	return v
}

func (p *Stream) readNBytesCrossFrameUnsafe(n int) []byte {
	v := make([]byte, n)
	copyBytes := copy(v, p.readFrame[p.readIndex:])
	p.readSeg++
	p.readFrame = *(p.frames[p.readSeg])
	p.readIndex = copy(v[copyBytes:], p.readFrame)
	return v
}

// return how many bytes to skip, 0 means error
func (p *Stream) peekSkip() (int, byte) {
	op := p.readFrame[p.readIndex]
	if skip := readSkipArray[op]; skip > 0 && p.CanRead() {
		return skip, op
	} else if skip < 0 {
		return 0, op
	} else if p.isSafetyReadNBytesInCurrentFrame(5) {
		b := p.readFrame[p.readIndex:]
		return int(uint32(b[1]) |
			(uint32(b[2]) << 8) |
			(uint32(b[3]) << 16) |
			(uint32(b[4]) << 24)), op
	} else if p.hasNBytesToRead(5) {
		b := p.peekNBytesCrossFrameUnsafe(5)
		return int(uint32(b[1]) |
			(uint32(b[2]) << 8) |
			(uint32(b[3]) << 16) |
			(uint32(b[4]) << 24)), op
	} else {
		return 0, op
	}
}

func (p *Stream) writeStreamNext(s *Stream) bool {
	if skip, _ := s.peekSkip(); skip <= 0 {
		return false
	} else if s.isSafetyReadNBytesInCurrentFrame(skip) &&
		p.writeIndex+skip < streamBlockSize {
		copy(
			p.writeFrame[p.writeIndex:],
			s.readFrame[s.readIndex:s.readIndex+skip],
		)
		p.writeIndex += skip
		s.readIndex += skip
		return true
	} else if s.hasNBytesToRead(skip) {
		for skip > 0 {
			copyEnd := base.MinInt(s.readIndex+skip, streamBlockSize)
			n := copy(
				p.writeFrame[p.writeIndex:],
				s.readFrame[s.readIndex:copyEnd],
			)
			skip -= n
			p.writeIndex += n
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			s.readIndex += n
			if s.readIndex == streamBlockSize {
				s.gotoNextReadFrameUnsafe()
			}
		}
		return true
	} else {
		return false
	}
}

// GetBuffer ...
func (p *Stream) GetBuffer() []byte {
	length := p.GetWritePos()
	ret := make([]byte, length)
	for i := 0; i <= p.writeSeg; i++ {
		copy(ret[i*streamBlockSize:], *(p.frames[i]))
	}
	return ret
}

// GetBufferUnsafe ...
func (p *Stream) GetBufferUnsafe() []byte {
	if p.writeSeg == 0 {
		return p.writeFrame[0:p.writeIndex]
	}
	return p.GetBuffer()
}

// PeekBufferSlice ...
func (p *Stream) PeekBufferSlice(pos int, max int) (ret []byte, isFinish bool) {
	if pos < 0 || max <= 0 {
		return nil, true
	}

	peekIndex := pos % streamBlockSize
	peekSeg := pos / streamBlockSize

	if peekSeg < p.writeSeg {
		peekEnd := base.MinInt(peekIndex+max, streamBlockSize)
		return (*p.frames[peekSeg])[peekIndex:peekEnd],
			peekSeg*streamBlockSize+peekEnd >= p.GetWritePos()
	} else if peekSeg == p.writeSeg {
		if peekIndex < p.writeIndex {
			peekEnd := base.MinInt(peekIndex+max, p.writeIndex)
			return (*p.frames[peekSeg])[peekIndex:peekEnd],
				peekEnd == p.writeIndex
		}
		return nil, true
	} else {
		return nil, true
	}
}

// PutBytes ...
func (p *Stream) PutBytes(v []byte) {
	if p.writeIndex+len(v) < streamBlockSize {
		p.writeIndex += copy(p.writeFrame[p.writeIndex:], v)
	} else {
		for len(v) > 0 {
			n := copy(p.writeFrame[p.writeIndex:], v)
			p.writeIndex += n
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			v = v[n:]
		}
	}
}

// PutBytesTo ...
func (p *Stream) PutBytesTo(v []byte, pos int) bool {
	if pos+len(v) < streamPosBody {
		return false
	}

	if pos >= streamPosBody {
		p.SetWritePos(pos)
	} else {
		p.writeIndex = pos
		p.writeSeg = 0
		p.writeFrame = *(p.frames[0])
	}

	p.PutBytes(v)
	return true
}

// WriteNil ...
func (p *Stream) WriteNil() {
	p.writeFrame[p.writeIndex] = 1
	p.writeIndex++
	if p.writeIndex == streamBlockSize {
		p.gotoNextWriteFrame()
	}
}

// WriteBool write bool value to stream
func (p *Stream) WriteBool(v bool) {
	switch v {
	case true:
		p.writeFrame[p.writeIndex] = 2
	case false:
		p.writeFrame[p.writeIndex] = 3
	}
	p.writeIndex++
	if p.writeIndex == streamBlockSize {
		p.gotoNextWriteFrame()
	}
}

// WriteFloat64 write float64 value to stream
func (p *Stream) WriteFloat64(value float64) {
	if value == 0 {
		p.writeFrame[p.writeIndex] = 4
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else {
		v := math.Float64bits(value)
		if p.writeIndex < streamBlockSize-9 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 5
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			b[5] = byte(v >> 32)
			b[6] = byte(v >> 40)
			b[7] = byte(v >> 48)
			b[8] = byte(v >> 56)
			p.writeIndex += 9
		} else {
			p.PutBytes([]byte{
				5,
				byte(v),
				byte(v >> 8),
				byte(v >> 16),
				byte(v >> 24),
				byte(v >> 32),
				byte(v >> 40),
				byte(v >> 48),
				byte(v >> 56),
			})
		}
	}
}

// WriteInt64 write int64 value to stream
func (p *Stream) WriteInt64(v int64) {
	if v > -8 && v < 33 {
		p.writeFrame[p.writeIndex] = byte(v + 21)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return
	}

	if v >= -32768 && v < 32768 {
		v += 32768
		if p.writeIndex < streamBlockSize-3 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 6
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.PutBytes([]byte{
			6,
			byte(v),
			byte(v >> 8),
		})
	} else if v >= -2147483648 && v < 2147483648 {
		v += 2147483648
		if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 7
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.PutBytes([]byte{
			7,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		var uv uint64
		if v < 0 {
			uv = 9223372036854775808 - uint64(-v)
		} else {
			uv = 9223372036854775808 + uint64(v)
		}
		if p.writeIndex < streamBlockSize-9 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 8
			b[1] = byte(uv)
			b[2] = byte(uv >> 8)
			b[3] = byte(uv >> 16)
			b[4] = byte(uv >> 24)
			b[5] = byte(uv >> 32)
			b[6] = byte(uv >> 40)
			b[7] = byte(uv >> 48)
			b[8] = byte(uv >> 56)
			p.writeIndex += 9
			return
		}
		p.PutBytes([]byte{
			8,
			byte(uv),
			byte(uv >> 8),
			byte(uv >> 16),
			byte(uv >> 24),
			byte(uv >> 32),
			byte(uv >> 40),
			byte(uv >> 48),
			byte(uv >> 56),
		})
	}
}

// WriteUint64 write uint64 value to stream
func (p *Stream) WriteUint64(v uint64) {
	if v < 10 {
		p.writeFrame[p.writeIndex] = byte(v + 54)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if v < 65536 {
		if p.writeIndex < streamBlockSize-3 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 9
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			p.writeIndex += 3
			return
		}
		p.PutBytes([]byte{
			9,
			byte(v),
			byte(v >> 8),
		})
	} else if v < 4294967296 {
		if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 10
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			p.writeIndex += 5
			return
		}
		p.PutBytes([]byte{
			10,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
	} else {
		if p.writeIndex < streamBlockSize-9 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 11
			b[1] = byte(v)
			b[2] = byte(v >> 8)
			b[3] = byte(v >> 16)
			b[4] = byte(v >> 24)
			b[5] = byte(v >> 32)
			b[6] = byte(v >> 40)
			b[7] = byte(v >> 48)
			b[8] = byte(v >> 56)
			p.writeIndex += 9
			return
		}
		p.PutBytes([]byte{
			11,
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
		})
	}
}

// WriteString ...
func (p *Stream) WriteString(v string) {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 128
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < streamBlockSize-2 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 128)
			b[length+1] = 0
			p.writeIndex += copy(b[1:], v) + 2
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 128)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		// write body
		p.PutBytes([]byte(v))
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else {
		totalLength := length + 6
		if p.writeIndex+totalLength < streamBlockSize {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			b[totalLength-1] = 0
			p.writeIndex += copy(b[5:], v) + 6
			return
		} else if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 191
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += 5
		} else {
			p.PutBytes([]byte{
				191,
				byte(uint32(totalLength)),
				byte(uint32(totalLength) >> 8),
				byte(uint32(totalLength) >> 16),
				byte(uint32(totalLength) >> 24),
			})
		}
		// write body
		p.PutBytes(base.StringToBytesUnsafe(v))
		// write zero tail
		p.writeFrame[p.writeIndex] = 0
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	}
}

// WriteBytes ...
func (p *Stream) WriteBytes(v Bytes) {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 192
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
	} else if length < 63 {
		if p.writeIndex+length < streamBlockSize-1 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = byte(length + 192)
			p.writeIndex += copy(b[1:], v) + 1
			return
		}
		// write header
		p.writeFrame[p.writeIndex] = byte(length + 192)
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		// write body
		p.PutBytes(v)
	} else {
		totalLength := length + 5
		if p.writeIndex+totalLength < streamBlockSize {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += copy(b[5:], v) + 5
			return
		} else if p.writeIndex < streamBlockSize-5 {
			b := p.writeFrame[p.writeIndex:]
			b[0] = 255
			b[1] = byte(uint32(totalLength))
			b[2] = byte(uint32(totalLength) >> 8)
			b[3] = byte(uint32(totalLength) >> 16)
			b[4] = byte(uint32(totalLength) >> 24)
			p.writeIndex += 5
		} else {
			p.PutBytes([]byte{
				255,
				byte(uint32(totalLength)),
				byte(uint32(totalLength) >> 8),
				byte(uint32(totalLength) >> 16),
				byte(uint32(totalLength) >> 24),
			})
		}
		// write body
		p.PutBytes(v)
	}
}

func (p *Stream) writeArray(v Array, depth int) string {
	length := len(v)
	if length == 0 {
		p.writeFrame[p.writeIndex] = 64
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < streamBlockSize-5 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(64 + length)
	} else {
		b[0] = 95
	}

	if length > 30 {
		if p.writeIndex < streamBlockSize-4 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.PutBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for i := 0; i < length; i++ {
		if reason := p.write(v[i], depth-1); reason != StreamWriteOK {
			p.SetWritePos(startPos)
			return base.ConcatString("[", strconv.Itoa(i), "]", reason)
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.SetWritePos(startPos + 1)
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.SetWritePos(endPos)
	}

	return StreamWriteOK
}

func (p *Stream) writeMap(v Map, depth int) string {
	length := len(v)

	if length == 0 {
		p.writeFrame[p.writeIndex] = 96
		p.writeIndex++
		if p.writeIndex == streamBlockSize {
			p.gotoNextWriteFrame()
		}
		return StreamWriteOK
	}

	startPos := p.GetWritePos()

	b := p.writeFrame[p.writeIndex:]
	if p.writeIndex < streamBlockSize-5 {
		p.writeIndex += 5
	} else {
		b = b[0:1]
		p.SetWritePos(startPos + 5)
	}

	if length < 31 {
		b[0] = byte(96 + length)
	} else {
		b[0] = 127
	}

	if length > 30 {
		if p.writeIndex < streamBlockSize-4 {
			l := p.writeFrame[p.writeIndex:]
			l[0] = byte(uint32(length))
			l[1] = byte(uint32(length) >> 8)
			l[2] = byte(uint32(length) >> 16)
			l[3] = byte(uint32(length) >> 24)
			p.writeIndex += 4
		} else {
			p.PutBytes([]byte{
				byte(uint32(length)),
				byte(uint32(length) >> 8),
				byte(uint32(length) >> 16),
				byte(uint32(length) >> 24),
			})
		}
	}

	for name, value := range v {
		p.WriteString(name)
		if reason := p.write(value, depth-1); reason != StreamWriteOK {
			p.SetWritePos(startPos)
			return base.ConcatString("[\"", name, "\"]", reason)
		}
	}

	totalLength := uint32(p.GetWritePos() - startPos)
	if len(b) > 1 {
		b[1] = byte(totalLength)
		b[2] = byte(totalLength >> 8)
		b[3] = byte(totalLength >> 16)
		b[4] = byte(totalLength >> 24)
	} else {
		endPos := p.GetWritePos()
		p.SetWritePos(startPos + 1)
		p.PutBytes([]byte{
			byte(totalLength),
			byte(totalLength >> 8),
			byte(totalLength >> 16),
			byte(totalLength >> 24),
		})
		p.SetWritePos(endPos)
	}

	return StreamWriteOK
}

func (p *Stream) writeRTArray(v RTArray) string {
	if thread := v.rt.thread; thread != nil {
		readStream := thread.rtStream
		length := len(*v.items)

		if length == 0 {
			p.writeFrame[p.writeIndex] = 64
			p.writeIndex++
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			return StreamWriteOK
		}

		startPos := p.GetWritePos()

		b := p.writeFrame[p.writeIndex:]
		if p.writeIndex < streamBlockSize-5 {
			p.writeIndex += 5
		} else {
			b = b[0:1]
			p.SetWritePos(startPos + 5)
		}

		if length < 31 {
			b[0] = byte(64 + length)
		} else {
			b[0] = 95
		}

		if length > 30 {
			if p.writeIndex < streamBlockSize-4 {
				l := p.writeFrame[p.writeIndex:]
				l[0] = byte(uint32(length))
				l[1] = byte(uint32(length) >> 8)
				l[2] = byte(uint32(length) >> 16)
				l[3] = byte(uint32(length) >> 24)
				p.writeIndex += 4
			} else {
				p.PutBytes([]byte{
					byte(uint32(length)),
					byte(uint32(length) >> 8),
					byte(uint32(length) >> 16),
					byte(uint32(length) >> 24),
				})
			}
		}

		for i := 0; i < length; i++ {
			items := *v.items
			readStream.SetReadPos(int(items[i].getPos()))
			if !p.writeStreamNext(readStream) {
				p.SetWritePos(startPos)
				return StreamWriteIsNotAvailable
			}
		}

		totalLength := uint32(p.GetWritePos() - startPos)
		if len(b) > 1 {
			b[1] = byte(totalLength)
			b[2] = byte(totalLength >> 8)
			b[3] = byte(totalLength >> 16)
			b[4] = byte(totalLength >> 24)
		} else {
			endPos := p.GetWritePos()
			p.SetWritePos(startPos + 1)
			p.PutBytes([]byte{
				byte(totalLength),
				byte(totalLength >> 8),
				byte(totalLength >> 16),
				byte(totalLength >> 24),
			})
			p.SetWritePos(endPos)
		}

		return StreamWriteOK
	}

	return StreamWriteIsNotAvailable
}

// WriteRPCMap write RPCMap value to stream
func (p *Stream) writeRTMap(v RTMap) string {
	if thread := v.rt.thread; thread != nil {
		readStream := thread.rtStream
		length := int(*v.length)

		if length == 0 {
			p.writeFrame[p.writeIndex] = 96
			p.writeIndex++
			if p.writeIndex == streamBlockSize {
				p.gotoNextWriteFrame()
			}
			return StreamWriteOK
		}

		startPos := p.GetWritePos()

		b := p.writeFrame[p.writeIndex:]
		if p.writeIndex < streamBlockSize-5 {
			p.writeIndex += 5
		} else {
			b = b[0:1]
			p.SetWritePos(startPos + 5)
		}

		if length < 31 {
			b[0] = byte(96 + length)
		} else {
			b[0] = 127
		}

		if length > 30 {
			if p.writeIndex < streamBlockSize-4 {
				l := p.writeFrame[p.writeIndex:]
				l[0] = byte(uint32(length))
				l[1] = byte(uint32(length) >> 8)
				l[2] = byte(uint32(length) >> 16)
				l[3] = byte(uint32(length) >> 24)
				p.writeIndex += 4
			} else {
				p.PutBytes([]byte{
					byte(uint32(length)),
					byte(uint32(length) >> 8),
					byte(uint32(length) >> 16),
					byte(uint32(length) >> 24),
				})
			}
		}

		if v.items != nil {
			items := *v.items
			for i := 0; i < length; i++ {
				p.WriteString(items[i].key)
				readStream.SetReadPos(int(items[i].pos.getPos()))
				if !p.writeStreamNext(readStream) {
					p.SetWritePos(startPos)
					return StreamWriteIsNotAvailable
				}
			}
		}

		totalLength := uint32(p.GetWritePos() - startPos)
		if len(b) > 1 {
			b[1] = byte(totalLength)
			b[2] = byte(totalLength >> 8)
			b[3] = byte(totalLength >> 16)
			b[4] = byte(totalLength >> 24)
		} else {
			endPos := p.GetWritePos()
			p.SetWritePos(startPos + 1)
			p.PutBytes([]byte{
				byte(totalLength),
				byte(totalLength >> 8),
				byte(totalLength >> 16),
				byte(totalLength >> 24),
			})
			p.SetWritePos(endPos)
		}

		return StreamWriteOK
	}

	return StreamWriteIsNotAvailable
}

func (p *Stream) writeRTValue(v RTValue) string {
	if thread := v.rt.thread; thread != nil {
		readStream := thread.rtStream
		if v.err != nil {
			return StreamWriteIsNotAvailable
		}
		if !readStream.SetReadPos(int(v.pos)) {
			return StreamWriteIsNotAvailable
		}
		if !p.writeStreamNext(readStream) {
			return StreamWriteIsNotAvailable
		}
		return StreamWriteOK
	}

	return StreamWriteIsNotAvailable
}

func (p *Stream) Write(v interface{}) string {
	if reason := p.write(v, 64); reason != StreamWriteOK {
		return "value" + reason
	}

	return StreamWriteOK
}

func (p *Stream) write(v interface{}, depth int) string {
	if depth <= 0 {
		return StreamWriteOverflow
	}

	if v == nil {
		p.WriteNil()
		return StreamWriteOK
	}

	switch v := v.(type) {
	case bool:
		p.WriteBool(v)
		return StreamWriteOK
	case int:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int8:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int16:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int32:
		p.WriteInt64(int64(v))
		return StreamWriteOK
	case int64:
		p.WriteInt64(v)
		return StreamWriteOK
	case uint:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint8:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint16:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint32:
		p.WriteUint64(uint64(v))
		return StreamWriteOK
	case uint64:
		p.WriteUint64(v)
		return StreamWriteOK
	case float32:
		p.WriteFloat64(float64(v))
		return StreamWriteOK
	case float64:
		p.WriteFloat64(v)
		return StreamWriteOK
	case string:
		p.WriteString(v)
		return StreamWriteOK
	case Bytes:
		p.WriteBytes(v)
		return StreamWriteOK
	case Array:
		return p.writeArray(v, depth)
	case Map:
		return p.writeMap(v, depth)
	case RTArray:
		return p.writeRTArray(v)
	case RTMap:
		return p.writeRTMap(v)
	case RTValue:
		return p.writeRTValue(v)
	default:
		return fmt.Sprintf(" type(%T) is not supported", v)
	}
}

// ReadNil read nil
func (p *Stream) ReadNil() (interface{}, *base.Error) {
	if p.CanRead() && p.readFrame[p.readIndex] == 1 {
		p.gotoNextReadByteUnsafe()
		return nil, nil
	}

	return nil, base.ErrStream
}

// ReadBool read bool
func (p *Stream) ReadBool() (bool, *base.Error) {
	if p.CanRead() {
		switch p.readFrame[p.readIndex] {
		case 2:
			p.gotoNextReadByteUnsafe()
			return true, nil
		case 3:
			p.gotoNextReadByteUnsafe()
			return false, nil
		}
	}
	return false, base.ErrStream
}

// ReadFloat64 read a float64
func (p *Stream) ReadFloat64() (float64, *base.Error) {
	v := p.readFrame[p.readIndex]
	if v == 4 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return 0, nil
		}
	} else if v == 5 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return math.Float64frombits(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56),
			), nil
		}
		if p.hasNBytesToRead(9) {
			b := p.readNBytesCrossFrameUnsafe(9)
			return math.Float64frombits(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56),
			), nil
		}
	}
	return 0, base.ErrStream
}

// ReadInt64 read a int64
func (p *Stream) ReadInt64() (int64, *base.Error) {
	v := p.readFrame[p.readIndex]
	if v > 13 && v < 54 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return int64(v) - 21, nil
		}
	} else if v == 6 {
		if p.isSafetyReadNBytesInCurrentFrame(3) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, nil
		}
		if p.hasNBytesToRead(3) {
			b := p.readNBytesCrossFrameUnsafe(3)
			return int64(uint16(b[1])|
				(uint16(b[2])<<8),
			) - 32768, nil
		}
	} else if v == 7 {
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, nil
		}
		if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			return int64(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24),
			) - 2147483648, nil
		}
	} else if v == 8 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return int64(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56) -
					9223372036854775808), nil
		}
		if p.hasNBytesToRead(9) {
			b := p.readNBytesCrossFrameUnsafe(9)
			return int64(
				uint64(b[1]) |
					(uint64(b[2]) << 8) |
					(uint64(b[3]) << 16) |
					(uint64(b[4]) << 24) |
					(uint64(b[5]) << 32) |
					(uint64(b[6]) << 40) |
					(uint64(b[7]) << 48) |
					(uint64(b[8]) << 56) -
					9223372036854775808), nil
		}
	}
	return 0, base.ErrStream
}

// ReadUint64 read a uint64
func (p *Stream) ReadUint64() (uint64, *base.Error) {
	v := p.readFrame[p.readIndex]
	if v > 53 && v < 64 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return uint64(v) - 54, nil
		}
	} else if v == 9 {
		if p.isSafetyReadNBytesInCurrentFrame(3) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 3
			return uint64(b[1]) |
				(uint64(b[2]) << 8), nil
		}
		if p.hasNBytesToRead(3) {
			b := p.readNBytesCrossFrameUnsafe(3)
			return uint64(b[1]) |
				(uint64(b[2]) << 8), nil
		}
	} else if v == 10 {
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 5
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), nil
		}
		if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24), nil
		}
	} else if v == 11 {
		if p.isSafetyReadNBytesInCurrentFrame(9) {
			b := p.readFrame[p.readIndex:]
			p.readIndex += 9
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24) |
				(uint64(b[5]) << 32) |
				(uint64(b[6]) << 40) |
				(uint64(b[7]) << 48) |
				(uint64(b[8]) << 56), nil
		}
		if p.hasNBytesToRead(9) {
			b := p.readNBytesCrossFrameUnsafe(9)
			return uint64(b[1]) |
				(uint64(b[2]) << 8) |
				(uint64(b[3]) << 16) |
				(uint64(b[4]) << 24) |
				(uint64(b[5]) << 32) |
				(uint64(b[6]) << 40) |
				(uint64(b[7]) << 48) |
				(uint64(b[8]) << 56), nil
		}
	}
	return 0, base.ErrStream
}

// ReadString read a string value
func (p *Stream) ReadString() (string, *base.Error) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return "", nil
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if isUTF8Bytes(b) {
					p.readIndex += strLen + 2
					return string(b), nil
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return string(b), nil
			}
			p.SetReadPos(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex : p.readIndex+strLen]
					if isUTF8Bytes(b) {
						p.readIndex += strLen + 1
						return string(b), nil
					}
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return string(b), nil
				}
			}
		}
		p.SetReadPos(readStart)
	}

	return "", base.ErrStream
}

func (p *Stream) GetNodePathUnsafe() (ret string, err *base.Error) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.CanRead() {
			return "", nil
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			return bytesToNodePathUnsafe(
				p.readFrame[p.readIndex+1 : p.readIndex+strLen+1],
			), nil
		} else if p.hasNBytesToRead(strLen + 2) {
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			return bytesToNodePathUnsafe(b), nil
		}
	} else if v == 191 {
		strLen := -1

		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				return bytesToNodePathUnsafe(
					p.readFrame[p.readIndex : p.readIndex+strLen],
				), nil
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				return bytesToNodePathUnsafe(b), nil
			}
		}
	}

	return "", base.ErrStream
}

func (p *Stream) readUnsafeString() (ret string, safe bool, err *base.Error) {
	// empty string
	v := p.readFrame[p.readIndex]
	if v == 128 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return "", true, nil
		}
	} else if v > 128 && v < 191 {
		strLen := int(v - 128)
		if p.isSafetyReadNBytesInCurrentFrame(strLen + 2) {
			if p.readFrame[p.readIndex+strLen+1] == 0 {
				b := p.readFrame[p.readIndex+1 : p.readIndex+strLen+1]
				if isUTF8Bytes(b) {
					p.readIndex += strLen + 2
					return base.BytesToStringUnsafe(b), false, nil
				}
			}
		} else if p.hasNBytesToRead(strLen + 2) {
			readStart := p.GetReadPos()
			b := make([]byte, strLen)
			copyBytes := copy(b, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(b[copyBytes:], p.readFrame)
			}
			if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
				p.gotoNextReadByteUnsafe()
				return base.BytesToStringUnsafe(b), true, nil
			}
			p.SetReadPos(readStart)
		}
	} else if v == 191 {
		readStart := p.GetReadPos()
		strLen := -1

		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			strLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 6
		}

		if strLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(strLen + 1) {
				if p.readFrame[p.readIndex+strLen] == 0 {
					b := p.readFrame[p.readIndex : p.readIndex+strLen]
					if isUTF8Bytes(b) {
						p.readIndex += strLen + 1
						return base.BytesToStringUnsafe(b), false, nil
					}
				}
			} else if p.hasNBytesToRead(strLen + 1) {
				b := make([]byte, strLen)
				reads := 0
				for reads < strLen {
					readLen := copy(b[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				if p.readFrame[p.readIndex] == 0 && isUTF8Bytes(b) {
					p.gotoNextReadByteUnsafe()
					return base.BytesToStringUnsafe(b), true, nil
				}
			}
		}

		p.SetReadPos(readStart)
	}
	return "", true, base.ErrStream
}

// ReadBytes read a Bytes value
func (p *Stream) ReadBytes() (Bytes, *base.Error) {
	// empty bytes
	v := p.readFrame[p.readIndex]

	if v == 192 {
		if p.CanRead() {
			p.gotoNextReadByteUnsafe()
			return Bytes{}, nil
		}
	} else if v > 192 && v < 255 {
		bytesLen := int(v - 192)
		ret := make(Bytes, bytesLen)
		if p.isSafetyReadNBytesInCurrentFrame(bytesLen + 1) {
			copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += bytesLen + 1
			return ret, nil
		} else if p.hasNBytesToRead(bytesLen + 1) {
			copyBytes := copy(ret, p.readFrame[p.readIndex+1:])
			p.readIndex += copyBytes + 1
			if p.readIndex == streamBlockSize {
				p.readSeg++
				p.readFrame = *(p.frames[p.readSeg])
				p.readIndex = copy(ret[copyBytes:], p.readFrame)
			}
			return ret, nil
		}
	} else if v == 255 {
		readStart := p.GetReadPos()
		bytesLen := -1
		if p.isSafetyReadNBytesInCurrentFrame(5) {
			b := p.readFrame[p.readIndex:]
			bytesLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 5
			p.readIndex += 5
		} else if p.hasNBytesToRead(5) {
			b := p.readNBytesCrossFrameUnsafe(5)
			bytesLen = int(uint32(b[1])|
				(uint32(b[2])<<8)|
				(uint32(b[3])<<16)|
				(uint32(b[4])<<24)) - 5
		}

		if bytesLen > 62 {
			if p.isSafetyReadNBytesInCurrentFrame(bytesLen) {
				ret := make(Bytes, bytesLen)
				copy(ret, p.readFrame[p.readIndex:])
				p.readIndex += bytesLen
				return ret, nil
			} else if p.hasNBytesToRead(bytesLen) {
				ret := make(Bytes, bytesLen)
				reads := 0
				for reads < bytesLen {
					readLen := copy(ret[reads:], p.readFrame[p.readIndex:])
					reads += readLen
					p.readIndex += readLen
					if p.readIndex == streamBlockSize {
						p.gotoNextReadFrameUnsafe()
					}
				}
				return ret, nil
			}
		}
		p.SetReadPos(readStart)
	}
	return Bytes{}, base.ErrStream
}

// ReadArray ...
func (p *Stream) ReadArray() (Array, *base.Error) {
	v := p.readFrame[p.readIndex]
	if v >= 64 && v < 96 {
		arrLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 64 {
			if p.CanRead() {
				p.gotoNextReadByteUnsafe()
				return Array{}, nil
			}
		} else if v < 95 {
			arrLen = int(v - 64)
			if p.isSafetyReadNBytesInCurrentFrame(5) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyReadNBytesInCurrentFrame(9) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				arrLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				p.readIndex += 9
			} else if p.hasNBytesToRead(9) {
				bytes8 := p.readNBytesCrossFrameUnsafe(9)
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				arrLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		if arrLen > 0 && totalLen > 4 {
			ret := make(Array, arrLen)
			for i := 0; i < arrLen; i++ {
				if rv, err := p.Read(); err == nil {
					ret[i] = rv
				} else {
					p.SetReadPos(readStart)
					return Array{}, err
				}
			}
			if p.GetReadPos() == readStart+totalLen {
				return ret, nil
			}
		}
		p.SetReadPos(readStart)
	}

	return Array{}, base.ErrStream
}

// ReadMap read a Map value
func (p *Stream) ReadMap() (Map, *base.Error) {
	v := p.readFrame[p.readIndex]
	if v >= 96 && v < 128 {
		mapLen := 0
		totalLen := 0
		readStart := p.GetReadPos()

		if v == 96 {
			if p.CanRead() {
				p.gotoNextReadByteUnsafe()
				return Map{}, nil
			}
		} else if v < 127 {
			mapLen = int(v - 96)
			if p.isSafetyReadNBytesInCurrentFrame(5) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				p.readIndex += 5
			} else if p.hasNBytesToRead(5) {
				bytes4 := p.readNBytesCrossFrameUnsafe(5)
				totalLen = int(uint32(bytes4[1]) |
					(uint32(bytes4[2]) << 8) |
					(uint32(bytes4[3]) << 16) |
					(uint32(bytes4[4]) << 24))
			}
		} else {
			if p.isSafetyReadNBytesInCurrentFrame(9) {
				b := p.readFrame[p.readIndex:]
				totalLen = int(uint32(b[1]) |
					(uint32(b[2]) << 8) |
					(uint32(b[3]) << 16) |
					(uint32(b[4]) << 24))
				mapLen = int(uint32(b[5]) |
					(uint32(b[6]) << 8) |
					(uint32(b[7]) << 16) |
					(uint32(b[8]) << 24))
				p.readIndex += 9
			} else if p.hasNBytesToRead(9) {
				bytes8 := p.readNBytesCrossFrameUnsafe(9)
				totalLen = int(uint32(bytes8[1]) |
					(uint32(bytes8[2]) << 8) |
					(uint32(bytes8[3]) << 16) |
					(uint32(bytes8[4]) << 24))
				mapLen = int(uint32(bytes8[5]) |
					(uint32(bytes8[6]) << 8) |
					(uint32(bytes8[7]) << 16) |
					(uint32(bytes8[8]) << 24))
			}
		}

		end := readStart + totalLen
		if mapLen > 0 && totalLen > 4 {
			ret := Map{}
			for i := 0; i < mapLen; i++ {
				if name, err := p.ReadString(); err != nil {
					p.SetReadPos(readStart)
					return Map{}, err
				} else if rv, err := p.Read(); err != nil {
					p.SetReadPos(readStart)
					return Map{}, err
				} else {
					ret[name] = rv
				}
			}
			if p.GetReadPos() == end {
				return ret, nil
			}
		}
		p.SetReadPos(readStart)
	}

	return Map{}, base.ErrStream
}

// Read read a generic value
func (p *Stream) Read() (ret Any, err *base.Error) {
	defer func() {
		if err != nil {
			ret = nil
		}
	}()

	op := p.readFrame[p.readIndex]
	switch op {
	case byte(1):
		return p.ReadNil()
	case byte(2):
		fallthrough
	case byte(3):
		return p.ReadBool()
	case byte(4):
		fallthrough
	case byte(5):
		return p.ReadFloat64()
	case byte(6):
		fallthrough
	case byte(7):
		fallthrough
	case byte(8):
		return p.ReadInt64()
	case byte(9):
		fallthrough
	case byte(10):
		fallthrough
	case byte(11):
		return p.ReadUint64()
	case byte(12):
		return nil, base.ErrStream
	case byte(13):
		return nil, base.ErrStream
	}

	switch op >> 6 {
	case 0:
		if op < 54 {
			return p.ReadInt64()
		}
		return p.ReadUint64()
	case 1:
		if op < 96 {
			return p.ReadArray()
		}
		return p.ReadMap()
	case 2:
		return p.ReadString()
	default:
		return p.ReadBytes()
	}
}

// ReadRTArray read a RPCArray value
func (p *Stream) ReadRTArray(rt Runtime) (RTArray, *base.Error) {
	if thread := rt.thread; thread != nil {
		v := p.readFrame[p.readIndex]
		if v >= 64 && v < 96 {
			cs := thread.rtStream
			arrLen := 0
			totalLen := 0
			readStart := p.GetReadPos()

			if p != cs {
				cs.SetReadPos(cs.GetWritePos())
				if !cs.writeStreamNext(p) {
					return RTArray{}, base.ErrStream
				}
			}

			start := cs.GetReadPos()

			if v == 64 {
				if cs.CanRead() {
					cs.gotoNextReadByteUnsafe()
					return newRTArray(rt, 0), nil
				}
			} else if v < 95 {
				arrLen = int(v - 64)
				if cs.isSafetyReadNBytesInCurrentFrame(5) {
					b := cs.readFrame[cs.readIndex:]
					totalLen = int(uint32(b[1]) |
						(uint32(b[2]) << 8) |
						(uint32(b[3]) << 16) |
						(uint32(b[4]) << 24))
					cs.readIndex += 5
				} else if cs.hasNBytesToRead(5) {
					bytes4 := cs.readNBytesCrossFrameUnsafe(5)
					totalLen = int(uint32(bytes4[1]) |
						(uint32(bytes4[2]) << 8) |
						(uint32(bytes4[3]) << 16) |
						(uint32(bytes4[4]) << 24))
				}
			} else {
				if cs.isSafetyReadNBytesInCurrentFrame(9) {
					b := cs.readFrame[cs.readIndex:]
					totalLen = int(uint32(b[1]) |
						(uint32(b[2]) << 8) |
						(uint32(b[3]) << 16) |
						(uint32(b[4]) << 24))
					arrLen = int(uint32(b[5]) |
						(uint32(b[6]) << 8) |
						(uint32(b[7]) << 16) |
						(uint32(b[8]) << 24))
					cs.readIndex += 9
				} else if cs.hasNBytesToRead(9) {
					bytes8 := cs.readNBytesCrossFrameUnsafe(9)
					totalLen = int(uint32(bytes8[1]) |
						(uint32(bytes8[2]) << 8) |
						(uint32(bytes8[3]) << 16) |
						(uint32(bytes8[4]) << 24))
					arrLen = int(uint32(bytes8[5]) |
						(uint32(bytes8[6]) << 8) |
						(uint32(bytes8[7]) << 16) |
						(uint32(bytes8[8]) << 24))
				}
			}

			end := start + totalLen
			if arrLen > 0 && totalLen > 4 {
				ret := newRTArray(rt, arrLen)
				for i := 0; i < arrLen; i++ {
					op := cs.readFrame[cs.readIndex]
					skip, _ := cs.peekSkip()
					itemPos := cs.GetReadPos()

					if skip <= 0 {
						p.SetReadPos(readStart)
						return RTArray{}, base.ErrStream
					} else if itemPos+skip > end {
						p.SetReadPos(readStart)
						return RTArray{}, base.ErrStream
					} else if cs.isSafetyReadNBytesInCurrentFrame(skip) {
						cs.readIndex += skip
					} else if !cs.SetReadPos(itemPos + skip) {
						p.SetReadPos(readStart)
						return RTArray{}, base.ErrStream
					}

					if op>>6 == 2 {
						*ret.items = append(
							*ret.items,
							makePosRecord(int64(itemPos), true),
						)
					} else {
						*ret.items = append(
							*ret.items,
							makePosRecord(int64(itemPos), false),
						)
					}
				}
				if cs.GetReadPos() == end {
					return ret, nil
				}
			}
			p.SetReadPos(readStart)
		}
		return RTArray{}, base.ErrStream
	}

	return RTArray{}, base.ErrRuntimeIllegalInCurrentGoroutine.
		AddDebug(base.GetFileLine(1))
}

// ReadRTMap read a RTMap value
func (p *Stream) ReadRTMap(rt Runtime) (RTMap, *base.Error) {
	if thread := rt.thread; thread != nil {
		v := p.readFrame[p.readIndex]
		if v >= 96 && v < 128 {
			cs := thread.rtStream
			mapLen := 0
			totalLen := 0
			readStart := p.GetReadPos()
			if p != cs {
				cs.SetReadPos(cs.GetWritePos())
				if !cs.writeStreamNext(p) {
					return RTMap{}, base.ErrStream
				}
			}
			start := cs.GetReadPos()
			if v == 96 {
				if cs.CanRead() {
					cs.gotoNextReadByteUnsafe()
					return newRTMap(rt, 0), nil
				}
			} else if v < 127 {
				mapLen = int(v - 96)
				if cs.isSafetyReadNBytesInCurrentFrame(5) {
					b := cs.readFrame[cs.readIndex:]
					totalLen = int(uint32(b[1]) |
						(uint32(b[2]) << 8) |
						(uint32(b[3]) << 16) |
						(uint32(b[4]) << 24))
					cs.readIndex += 5
				} else if cs.hasNBytesToRead(5) {
					bytes4 := cs.readNBytesCrossFrameUnsafe(5)
					totalLen = int(uint32(bytes4[1]) |
						(uint32(bytes4[2]) << 8) |
						(uint32(bytes4[3]) << 16) |
						(uint32(bytes4[4]) << 24))
				}
			} else {
				if cs.isSafetyReadNBytesInCurrentFrame(9) {
					b := cs.readFrame[cs.readIndex:]
					totalLen = int(uint32(b[1]) |
						(uint32(b[2]) << 8) |
						(uint32(b[3]) << 16) |
						(uint32(b[4]) << 24))
					mapLen = int(uint32(b[5]) |
						(uint32(b[6]) << 8) |
						(uint32(b[7]) << 16) |
						(uint32(b[8]) << 24))
					cs.readIndex += 9
				} else if cs.hasNBytesToRead(9) {
					bytes8 := cs.readNBytesCrossFrameUnsafe(9)
					totalLen = int(uint32(bytes8[1]) |
						(uint32(bytes8[2]) << 8) |
						(uint32(bytes8[3]) << 16) |
						(uint32(bytes8[4]) << 24))
					mapLen = int(uint32(bytes8[5]) |
						(uint32(bytes8[6]) << 8) |
						(uint32(bytes8[7]) << 16) |
						(uint32(bytes8[8]) << 24))
				}
			}

			end := start + totalLen
			if mapLen > 0 && totalLen > 4 {
				ret := newRTMap(rt, mapLen)
				for i := 0; i < mapLen; i++ {
					key, _, err := cs.readUnsafeString()
					if err != nil {
						p.SetReadPos(readStart)
						return RTMap{}, err
					}
					op := cs.readFrame[cs.readIndex]
					skip, _ := cs.peekSkip()
					itemPos := cs.GetReadPos()

					if skip <= 0 {
						p.SetReadPos(readStart)
						return RTMap{}, base.ErrStream
					} else if itemPos+skip > end {
						p.SetReadPos(readStart)
						return RTMap{}, base.ErrStream
					} else if cs.isSafetyReadNBytesInCurrentFrame(skip) {
						cs.readIndex += skip
					} else if !cs.SetReadPos(itemPos + skip) {
						p.SetReadPos(readStart)
						return RTMap{}, base.ErrStream
					}

					if op>>6 == 2 {
						ret.appendValue(key, makePosRecord(int64(itemPos), true))
					} else {
						ret.appendValue(key, makePosRecord(int64(itemPos), false))
					}
				}
				if cs.GetReadPos() == end {
					return ret, nil
				}
			}
			p.SetReadPos(readStart)
		}
		return RTMap{}, base.ErrStream
	}

	return RTMap{}, base.ErrRuntimeIllegalInCurrentGoroutine.
		AddDebug(base.GetFileLine(1))
}

// ReadRTValue read a RTValue value
func (p *Stream) ReadRTValue(rt Runtime) (RTValue, *base.Error) {
	if thread := rt.thread; thread != nil {
		cs := thread.rtStream
		if p != cs {
			cs.SetReadPos(cs.GetWritePos())
			if !cs.writeStreamNext(p) {
				return RTValue{err: base.ErrStream}, base.ErrStream
			}
		}

		if op := cs.readFrame[cs.readIndex]; op>>6 == 2 {
			pos := int64(cs.GetReadPos())
			cacheString, cacheSafe, cacheError := cs.readUnsafeString()
			return RTValue{
				err:        nil,
				rt:         rt,
				pos:        pos,
				cacheBytes: base.StringToBytesUnsafe(cacheString),
				cacheSafe:  cacheSafe,
				cacheError: cacheError,
			}, nil
		}

		startPos := cs.GetReadPos()
		skip, _ := cs.peekSkip()

		if skip <= 0 {
			return RTValue{err: base.ErrStream}, base.ErrStream
		} else if cs.isSafetyReadNBytesInCurrentFrame(skip) {
			cs.readIndex += skip
		} else if !cs.SetReadPos(startPos + skip) {
			return RTValue{err: base.ErrStream}, base.ErrStream
		}

		return RTValue{
			err:        nil,
			rt:         rt,
			pos:        int64(startPos),
			cacheBytes: nil,
			cacheError: base.ErrStream,
			cacheSafe:  false,
		}, nil
	}

	return RTValue{err: base.ErrRuntimeIllegalInCurrentGoroutine},
		base.ErrRuntimeIllegalInCurrentGoroutine
}

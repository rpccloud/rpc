// Package rpc ...
package rpc

import (
	"math"
	"reflect"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
)

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

// IStreamReceiver ...
type IStreamReceiver interface {
	OnReceiveStream(stream *Stream)
}

// TestStreamReceiver ...
type TestStreamReceiver struct {
	streamCH chan *Stream
}

// NewTestStreamReceiver ...
func NewTestStreamReceiver() *TestStreamReceiver {
	return &TestStreamReceiver{
		streamCH: make(chan *Stream, 10240),
	}
}

// OnReceiveStream ...
func (p *TestStreamReceiver) OnReceiveStream(stream *Stream) {
	p.streamCH <- stream
}

// GetStream ...
func (p *TestStreamReceiver) GetStream() *Stream {
	select {
	case stream := <-p.streamCH:
		return stream
	default:
		return nil
	}
}

// WaitStream ...
func (p *TestStreamReceiver) WaitStream() *Stream {
	return <-p.streamCH
}

// TotalStreams ...
func (p *TestStreamReceiver) TotalStreams() int {
	return len(p.streamCH)
}

func bytesToNodePathUnsafe(bytes []byte) (ret string) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	if bytesHeader.Len <= 0 {
		return ""
	}

	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&ret))
	stringHeader.Data = bytesHeader.Data

	length := bytesHeader.Len

	for idx := 0; idx < length; idx++ {
		c := bytes[idx]
		if c >= 128 || c == 58 {
			stringHeader.Len = idx
			return
		}
	}

	stringHeader.Len = length
	return
}

func isUTF8Bytes(bytes []byte) bool {
	idx := 0
	length := len(bytes)

	for idx < length {
		c := bytes[idx]
		if c < 128 {
			idx++
		} else if c < 224 {
			if (idx+2 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) {
				return false
			}
			idx += 2
		} else if c < 240 {
			if (idx+3 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) {
				return false
			}
			idx += 3
		} else if c < 248 {
			if (idx+4 > length) ||
				(bytes[idx+1]&0xC0 != 0x80) ||
				(bytes[idx+2]&0xC0 != 0x80) ||
				(bytes[idx+3]&0xC0 != 0x80) {
				return false
			}
			idx += 4
		} else {
			return false
		}
	}

	return idx == length
}

func getFuncKind(fn reflect.Value) (string, *base.Error) {
	if fn.Kind() != reflect.Func {
		return "", base.ErrActionHandler.
			AddDebug("handler must be a function")
	} else if fn.Type().NumIn() < 1 ||
		fn.Type().In(0) != reflect.ValueOf(Runtime{}).Type() {
		return "", base.ErrActionHandler.AddDebug(base.ConcatString(
			"handler 1st argument type must be ",
			convertTypeToString(runtimeType)),
		)
	} else if fn.Type().NumOut() != 1 ||
		fn.Type().Out(0) != reflect.ValueOf(emptyReturn).Type() {
		return "", base.ErrActionHandler.AddDebug(base.ConcatString(
			"handler return type must be ",
			convertTypeToString(returnType),
		))
	} else {
		sb := base.NewStringBuilder()
		defer sb.Release()

		for i := 1; i < fn.Type().NumIn(); i++ {
			switch fn.Type().In(i) {
			case boolType:
				sb.AppendByte(vkBool)
			case int64Type:
				sb.AppendByte(vkInt64)
			case uint64Type:
				sb.AppendByte(vkUint64)
			case float64Type:
				sb.AppendByte(vkFloat64)
			case stringType:
				sb.AppendByte(vkString)
			case bytesType:
				sb.AppendByte(vkBytes)
			case arrayType:
				sb.AppendByte(vkArray)
			case mapType:
				sb.AppendByte(vkMap)
			case rtValueType:
				sb.AppendByte(vkRTValue)
			case rtArrayType:
				sb.AppendByte(vkRTArray)
			case rtMapType:
				sb.AppendByte(vkRTMap)
			default:
				return "", base.ErrActionHandler.AddDebug(
					base.ConcatString(
						"handler ",
						base.ConvertOrdinalToString(1+uint(i)),
						" argument type ",
						fn.Type().In(i).String(),
						" is not supported",
					))
			}
		}

		return sb.String(), nil
	}
}

func convertTypeToString(reflectType reflect.Type) string {
	switch reflectType {
	case nil:
		return "<nil>"
	case runtimeType:
		return "rpc.Runtime"
	case returnType:
		return "rpc.Return"
	case boolType:
		return "rpc.Bool"
	case int64Type:
		return "rpc.Int64"
	case uint64Type:
		return "rpc.Uint64"
	case float64Type:
		return "rpc.Float64"
	case stringType:
		return "rpc.String"
	case bytesType:
		return "rpc.Bytes"
	case arrayType:
		return "rpc.Array"
	case mapType:
		return "rpc.Map"
	case rtValueType:
		return "rpc.RTValue"
	case rtArrayType:
		return "rpc.RTArray"
	case rtMapType:
		return "rpc.RTMap"
	default:
		return reflectType.String()
	}
}

func getFastKey(s string) uint32 {
	bytes := base.StringToBytesUnsafe(s)

	if size := len(bytes); size > 0 {
		return uint32(bytes[0])<<16 |
			uint32(bytes[size>>1])<<8 |
			uint32(bytes[size-1])
	}

	return 0
}

// MakeSystemErrorStream ...
func MakeSystemErrorStream(err *base.Error) *Stream {
	if err != nil {
		stream := NewStream()
		stream.SetKind(StreamKindSystemErrorReport)
		stream.WriteUint64(uint64(err.GetCode()))
		stream.WriteString(err.GetMessage())
		return stream
	}

	return nil
}

// MakeInternalRequestStream ...
func MakeInternalRequestStream(
	debug bool,
	depth uint16,
	target string,
	from string,
	args ...interface{},
) (*Stream, *base.Error) {
	stream := NewStream()
	stream.SetKind(StreamKindRPCRequest)

	// set debug bit
	if debug {
		stream.SetStatusBitDebug()
	} else {
		stream.ClearStatusBitDebug()
	}

	// set depth
	stream.SetDepth(depth)
	// write target
	stream.WriteString(target)
	// write from
	stream.WriteString(from)
	// write args
	for i := 0; i < len(args); i++ {
		if reason := stream.Write(args[i]); reason != StreamWriteOK {
			stream.Release()
			return nil, base.ErrUnsupportedValue.AddDebug(
				base.ConcatString(
					base.ConvertOrdinalToString(uint(i)+2),
					" argument: ",
					reason,
				),
			)
		}
	}

	return stream, nil
}

// ParseResponseStream ...
func ParseResponseStream(stream *Stream) (Any, *base.Error) {
	switch stream.GetKind() {
	case StreamKindRPCResponseOK:
		return stream.Read()
	case StreamKindSystemErrorReport:
		fallthrough
	case StreamKindRPCResponseError:
		if errCode, err := stream.ReadUint64(); err != nil {
			return nil, err
		} else if errCode == 0 {
			return nil, base.ErrStream
		} else if errCode > math.MaxUint32 {
			return nil, base.ErrStream
		} else if message, err := stream.ReadString(); err != nil {
			return nil, err
		} else if !stream.IsReadFinish() {
			return nil, base.ErrStream
		} else {
			return nil, base.NewError(uint32(errCode), message)
		}
	default:
		return nil, base.ErrStream
	}
}

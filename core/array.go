package core

import "sync"

////////////////////////////////////////////////////////////////////////////////
// RPCArray
////////////////////////////////////////////////////////////////////////////////
var rpcArrayInnerCache = sync.Pool{
	New: func() interface{} {
		return &rpcArrayInner{
			items: make([]int, 0, 64),
		}
	},
}

type rpcArrayInner struct {
	items []int
}

func (p *rpcArrayInner) free() {
	if cap(p.items) == 64 {
		p.items = p.items[:0]
	} else {
		p.items = make([]int, 0, 64)
	}
	rpcArrayInnerCache.Put(p)
}

// RPCArray ...
type rpcArray struct {
	ctx *rpcContext
	in  *rpcArrayInner
}

func newRPCArray(ctx *rpcContext) rpcArray {
	if ctx != nil && ctx.inner != nil && ctx.inner.stream != nil {
		return rpcArray{
			ctx: ctx,
			in:  rpcArrayInnerCache.Get().(*rpcArrayInner),
		}
	}
	return nilRPCArray
}

func (p rpcArray) ok() bool {
	return p.in != nil &&
		p.ctx != nil &&
		p.ctx.inner != nil &&
		p.ctx.inner.stream != nil
}

// Release ...
func (p rpcArray) release() {
	if p.in != nil {
		p.in.free()
		p.in = nil
	}
}

func (p rpcArray) getIS() (*rpcArrayInner, *rpcStream) {
	if p.in != nil && p.ctx != nil && p.ctx.inner != nil {
		return p.in, p.ctx.inner.stream
	} else {
		return nil, nil
	}
}

// Size ...
func (p rpcArray) Size() int {
	if in, _ := p.getIS(); in != nil {
		return len(in.items)
	} else {
		return 0
	}
}

// GetNil ...
func (p rpcArray) GetNil(index int) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadNil()
	}
	return false
}

// SetNil ...
func (p rpcArray) SetNil(index int) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteNil()
		return true
	}
	return false
}

// AppendNil ...
func (p rpcArray) AppendNil() bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteNil()
		return true
	}
	return false
}

// GetBool ...
func (p rpcArray) GetBool(index int) (bool, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadBool()
	}
	return false, false
}

// SetBool ...
func (p rpcArray) SetBool(index int, value bool) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteBool(value)
		return true
	}
	return false
}

// AppendBool ...
func (p rpcArray) AppendBool(value bool) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteBool(value)
		return true
	}
	return false
}

// GetInt64 ...
func (p rpcArray) GetInt64(index int) (int64, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadInt64()
	}
	return 0, false
}

// SetInt64 ...
func (p rpcArray) SetInt64(index int, value int64) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteInt64(value)
		return true
	}
	return false
}

// AppendInt64 ...
func (p rpcArray) AppendInt64(value int64) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteInt64(value)
		return true
	}
	return false
}

// GetUint64 ...
func (p rpcArray) GetUint64(index int) (uint64, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadUint64()
	}
	return 0, false
}

// SetUint64 ...
func (p rpcArray) SetUint64(index int, value uint64) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteUint64(value)
		return true
	}
	return false
}

// AppendUint64 ...
func (p rpcArray) AppendUint64(value uint64) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteUint64(value)
		return true
	}
	return false
}

// GetFloat64 ...
func (p rpcArray) GetFloat64(index int) (float64, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadFloat64()
	}
	return 0, false
}

// SetFloat64 ...
func (p rpcArray) SetFloat64(index int, value float64) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteFloat64(value)
		return true
	}
	return false
}

// AppendFloat64 ...
func (p rpcArray) AppendFloat64(value float64) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteFloat64(value)
		return true
	}
	return false
}

// GetRPCString ...
func (p rpcArray) GetString(index int) (string, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadString()
	}
	return emptyString, false
}

// SetRPCString ...
func (p rpcArray) SetString(index int, value string) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteString(value)
		return true
	}
	return false
}

// AppendRPCString ...
func (p rpcArray) AppendString(value string) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteString(value)
		return true
	}
	return false
}

// GetRPCBytes ...
func (p rpcArray) GetBytes(index int) ([]byte, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadBytes()
	}
	return emptyBytes, false
}

// SetRPCBytes ...
func (p rpcArray) SetBytes(index int, value []byte) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		in.items[index] = s.GetWritePos()
		s.WriteBytes(value)
		return true
	}
	return false
}

// AppendRPCBytes ...
func (p rpcArray) AppendBytes(value []byte) bool {
	if in, s := p.getIS(); s != nil {
		in.items = append(in.items, s.GetWritePos())
		s.WriteBytes(value)
		return true
	}
	return false
}

// GetRPCArray ...
func (p rpcArray) GetRPCArray(index int) (rpcArray, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadRPCArray(p.ctx)
	}
	return nilRPCArray, false
}

// SetRPCArray ...
func (p rpcArray) SetRPCArray(index int, value rpcArray) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		pos := s.GetWritePos()
		if s.WriteRPCArray(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCArray ...
func (p rpcArray) AppendRPCArray(value rpcArray) bool {
	if in, s := p.getIS(); s != nil {
		pos := s.GetWritePos()
		if s.WriteRPCArray(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// GetRPCMap ...
func (p rpcArray) GetRPCMap(index int) (rpcMap, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.ReadRPCMap(p.ctx)
	}
	return nilRPCMap, false
}

// SetRPCMap ...
func (p rpcArray) SetRPCMap(index int, value rpcMap) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		pos := s.GetWritePos()
		if s.WriteRPCMap(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// AppendRPCMap ...
func (p rpcArray) AppendRPCMap(value rpcMap) bool {
	if in, s := p.getIS(); s != nil {
		pos := s.GetWritePos()
		if s.WriteRPCMap(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}

// Get ...
func (p rpcArray) Get(index int) (interface{}, bool) {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		s.setReadPosUnsafe(in.items[index])
		return s.Read(p.ctx)
	}
	return nil, false
}

// Set ...
func (p rpcArray) Set(index int, value interface{}) bool {
	if in, s := p.getIS(); s != nil && index >= 0 && index < len(in.items) {
		pos := s.GetWritePos()
		if s.Write(value) == RPCStreamWriteOK {
			in.items[index] = pos
			return true
		}
	}
	return false
}

// Append ...
func (p rpcArray) Append(value interface{}) bool {
	if in, s := p.getIS(); s != nil {
		pos := s.GetWritePos()
		if s.Write(value) == RPCStreamWriteOK {
			in.items = append(in.items, pos)
			return true
		}
	}
	return false
}
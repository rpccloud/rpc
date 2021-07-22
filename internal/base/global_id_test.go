package base

import (
	"encoding/binary"
	"errors"
	"net"
	"testing"
)

func TestNewGlobalID(t *testing.T) {
	t.Run("net.Listen error", func(t *testing.T) {
		assert := NewAssert(t)
		originFN := fnNetListen
		fnNetListen = func(network, address string) (net.Listener, error) {
			return nil, errors.New("error")
		}
		defer func() {
			fnNetListen = originFN
		}()

		assert(NewGlobalID()).Equals(nil)
	})

	t.Run("net.Interfaces error", func(t *testing.T) {
		assert := NewAssert(t)
		originFN := fnNetInterfaces
		fnNetInterfaces = func() ([]net.Interface, error) {
			return nil, errors.New("error")
		}
		defer func() {
			fnNetInterfaces = originFN
		}()
		assert(NewGlobalID()).Equals(nil)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := NewAssert(t)
		assert(NewGlobalID()).IsNotNil()
	})
}

func TestGlobalID_GetID(t *testing.T) {
	t.Run("port error", func(t *testing.T) {
		assert := NewAssert(t)
		v := &GlobalID{port: 0, mac: []byte{3, 4, 5, 6, 7, 8}}
		assert(v.GetID()).Equals(uint64(0))
	})

	t.Run("mac error", func(t *testing.T) {
		assert := NewAssert(t)
		v := &GlobalID{port: 513, mac: []byte{3, 4, 5, 6, 7}}
		assert(v.GetID()).Equals(uint64(0))
	})

	t.Run("test ok", func(t *testing.T) {
		assert := NewAssert(t)
		v := &GlobalID{port: 513, mac: []byte{3, 4, 5, 6, 7, 8}}
		buffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(buffer, v.GetID())
		assert(buffer).Equals([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	})
}

func TestGlobalID_Close(t *testing.T) {
	t.Run("p.listener == nil", func(t *testing.T) {
		assert := NewAssert(t)
		v := &GlobalID{port: 0, mac: []byte{3, 4, 5, 6, 7, 8}}
		v.Close()
		assert(v.listener).IsNil()
		assert(v.port).Equals(uint16(0))
		assert(v.mac).IsNil()
	})

	t.Run("test ok", func(t *testing.T) {
		assert := NewAssert(t)
		v := NewGlobalID()
		v.Close()
		assert(v.listener).IsNil()
		assert(v.port).Equals(uint16(0))
		assert(v.mac).IsNil()
	})
}

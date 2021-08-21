package rpc

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestNewStreamGenerator(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := NewTestStreamReceiver()
		v := NewStreamGenerator(receiver)
		assert(v).IsNotNil()
		assert(v.streamReceiver).Equals(receiver)
		assert(v.streamPos).Equals(0)
		assert(len(v.streamBuffer), cap(v.streamBuffer)).
			Equals(StreamHeadSize, StreamHeadSize)
		assert(v.stream).IsNil()
	})
}

func TestStreamGenerator_Reset(t *testing.T) {
	t.Run("stream is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := NewTestStreamReceiver()
		v := NewStreamGenerator(receiver)
		v.streamPos = 20
		v.Reset()
		assert(v.streamPos).Equals(0)
	})

	t.Run("stream is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := NewTestStreamReceiver()
		v := NewStreamGenerator(receiver)
		v.streamPos = 120
		v.stream = NewStream()
		v.Reset()
		assert(v.streamPos).Equals(0)
		assert(v.stream).IsNil()
	})
}

func TestStreamGenerator_OnBytes(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		fnTest := func(
			numOfStream int,
			addBytes int,
			fragment int) bool {
			receiver := NewTestStreamReceiver()
			v := NewStreamGenerator(receiver)
			buffer := make([]byte, 0)
			streamArray := make([]*Stream, numOfStream)
			for i := 0; i < numOfStream; i++ {
				streamArray[i] = NewStream()
				streamArray[i].PutBytes(
					[]byte(base.GetRandString(addBytes)),
				)
				streamArray[i].BuildStreamCheck()
				buffer = append(buffer, streamArray[i].GetBuffer()...)
			}

			for i := 0; i < len(buffer); i += fragment {
				if v.OnBytes(
					buffer[i:base.MinInt(i+fragment, len(buffer))],
				) != nil {
					fmt.Println("A")
					return false
				}
			}

			for i := 0; i < numOfStream; i++ {
				s := receiver.GetStream()
				if s == nil {
					fmt.Println("B")
					return false
				}
				if !reflect.DeepEqual(
					streamArray[i].GetBuffer(),
					s.GetBuffer(),
				) {
					fmt.Println("C")
					fmt.Println(streamArray[i].GetBuffer())
					fmt.Println(s.GetBuffer())
					return false
				}
			}
			return true
		}

		for _, i := range []int{0, 1, 2, 3, 500} {
			for _, j := range []int{1, 2, 3, 57, 63, 452, 512, 800, 1200} {
				assert(fnTest(i, rand.Int()%200, j)).IsTrue()
			}
		}
	})

	t.Run("remains < 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := NewTestStreamReceiver()
		v := NewStreamGenerator(receiver)
		assert(v.OnBytes(NewStream().GetBuffer())).Equals(base.ErrStream)
	})

	t.Run("stream check error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := NewTestStreamReceiver()
		v := NewStreamGenerator(receiver)
		s := NewStream()
		s.PutBytes([]byte{1, 2, 3})
		s.BuildStreamCheck()
		buffer := s.GetBuffer()
		buffer[len(buffer)-1] = 0
		assert(v.OnBytes(buffer)).Equals(base.ErrStream)
	})
}

package rpc

import (
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestPosRecord_getPos(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(posRecord(12).getPos()).Equals(int64(12))
		assert(posRecord(12 | 0x8000000000000000).getPos()).Equals(int64(12))
	})
}

func TestPosRecord_isString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(posRecord(12).isString()).IsFalse()
		assert(posRecord(12 | 0x8000000000000000).isString()).IsTrue()
	})
}

func TestPosRecord_makePosRecord(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(makePosRecord(12, false)).Equals(posRecord(12))
		assert(makePosRecord(12, true)).
			Equals(posRecord(0x8000000000000000 | 12))
	})
}

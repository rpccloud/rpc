package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

func TestCheckEqual(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(checkEqual(0, 1)).IsFalse()
		assert(checkEqual(1, 1)).IsTrue()
		assert(checkEqual(errors.New("error"), nil)).IsFalse()
		assert(checkEqual(error(nil), nil)).IsTrue()
		assert(checkEqual(errors.New("error"), errors.New("error"))).IsTrue()
	})
}

func TestAllCPUTimes(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(allCPUTimes()).IsNotNil()
	})

	t.Run("status fail", func(t *testing.T) {
		// hook checkEqual
		saveCheckEqual := checkEqual
		checkEqual = func(a, b interface{}) bool {
			return !saveCheckEqual(a, b)
		}
		defer func() {
			checkEqual = saveCheckEqual
		}()

		assert := base.NewAssert(t)
		assert(allCPUTimes()).IsNil()
	})
}

func TestGetCPUPercent(t *testing.T) {
	t.Run("interval <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(getCPUPercent(0)).Equals(float64(1), false)
	})

	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		cpu, ok := getCPUPercent(200 * time.Millisecond)
		assert(cpu >= 0 && cpu <= 1).IsTrue()
		assert(ok).IsTrue()
	})
}

func TestCalculateCPUPercent(t *testing.T) {
	t.Run("t1 == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(calculateCPUPercent(nil, &cpuTimesStat{})).
			Equals(float64(1), false)
	})

	t.Run("t2 == nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(calculateCPUPercent(&cpuTimesStat{}, nil)).
			Equals(float64(1), false)
	})

	t.Run("t2Busy < t1Busy", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(calculateCPUPercent(&cpuTimesStat{User: 1}, &cpuTimesStat{})).
			Equals(float64(1), false)
	})

	t.Run("t2All <= t1All", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(calculateCPUPercent(&cpuTimesStat{Idle: 1}, &cpuTimesStat{})).
			Equals(float64(1), true)
	})

	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(calculateCPUPercent(
			&cpuTimesStat{},
			&cpuTimesStat{User: 1, Idle: 1},
		)).Equals(0.5, true)
	})
}

func TestGetMetrics(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetMetrics(200 * time.Millisecond)).IsNotNil()
	})

	t.Run("interval <= 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetMetrics(0)).IsNil()
	})
}

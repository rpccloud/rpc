package base

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(gTimeMaster).IsNotNil()
		assert(string(defaultISODateBuffer)).
			Equals("0000-00-00T00:00:00.000+00:00")
		for i := 0; i < 100; i++ {
			assert(string(intToStringCache2[i])).Equals(fmt.Sprintf("%02d", i))
		}
		for i := 0; i < 1000; i++ {
			assert(string(intToStringCache3[i])).Equals(fmt.Sprintf("%03d", i))
		}
		for i := 0; i < 10000; i++ {
			assert(string(intToStringCache4[i])).Equals(fmt.Sprintf("%04d", i))
		}
		assert(timeMasterStatusNone).Equals(int32(0))
		assert(timeMasterStatusRunning).Equals(int32(1))
		assert(timeMasterStatusClosed).Equals(int32(2))
	})
}

func TestNewTimeMaster(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		assert(v1.timeNowPointer).IsNil()
		assert(v1.timeSpeedCounter).IsNotNil()
	})
}

func TestTimeMaster_Run(t *testing.T) {
	t.Run("status is timeMasterStatusNone low speed", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		for i := 0; i < 100; i++ {
			v1.Count()
		}
		time.Sleep(50 * time.Millisecond)
		go func() {
			v1.Run()
		}()
		time.Sleep(50 * time.Millisecond)
		assert(atomic.LoadPointer(&v1.timeNowPointer)).IsNil()
		v1.Close()
	})

	t.Run("status is timeMasterStatusNone high speed", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		for i := 0; i < 10000; i++ {
			v1.Count()
		}
		time.Sleep(50 * time.Millisecond)
		go func() {
			v1.Run()
		}()
		time.Sleep(50 * time.Millisecond)
		tm := (*timeInfo)(atomic.LoadPointer(&v1.timeNowPointer))
		assert(IsTimeApproximatelyEqual(tm.time, time.Now())).IsTrue()
		assert(tm.timeISOString).Equals(ConvertToIsoDateString(tm.time))
		v1.Close()
	})

	t.Run("status is timeMasterStatusRunning", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &timeMaster{status: timeMasterStatusRunning}
		assert(v1.Run()).IsFalse()
	})

	t.Run("status is timeMasterStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &timeMaster{status: timeMasterStatusClosed}
		assert(v1.Run()).IsFalse()
	})
}

func TestTimeMaster_Close(t *testing.T) {
	t.Run("status is timeMasterStatusNone", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &timeMaster{status: timeMasterStatusNone}
		assert(v1.Close()).IsFalse()
	})

	t.Run("status is timeMasterStatusRunning", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		r1 := make(chan bool)
		go func() {
			r1 <- v1.Run()
		}()
		time.Sleep(50 * time.Millisecond)
		assert(v1.Close()).IsTrue()
		assert(<-r1).IsTrue()
	})

	t.Run("status is timeMasterStatusClosed", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := &timeMaster{status: timeMasterStatusClosed}
		assert(v1.Close()).IsFalse()
	})
}

func TestTimeMaster_Count(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		v1.Count()
		assert(atomic.LoadInt64(&v1.timeSpeedCounter.total)).Equals(int64(1))
	})
}

func TestTimeMaster_TimeNow(t *testing.T) {
	t.Run("timeNow low mode", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		assert(IsTimeApproximatelyEqual(v1.TimeNow(), time.Now())).IsTrue()
	})

	t.Run("timeNow high mode", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		for i := 0; i < 1000; i++ {
			v1.Count()
		}
		time.Sleep(50 * time.Millisecond)
		go func() {
			v1.Run()
		}()
		time.Sleep(50 * time.Millisecond)
		assert(IsTimeApproximatelyEqual(v1.TimeNow(), time.Now())).IsTrue()
	})
}

func TestTimeMaster_TimeNowISOString(t *testing.T) {
	t.Run("timeNow low mode", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		parseTime, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			v1.TimeNowISOString(),
		)
		assert(IsTimeApproximatelyEqual(parseTime, time.Now())).IsTrue()
	})

	t.Run("timeNow high mode", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := newTimeMaster()
		for i := 0; i < 1000; i++ {
			v1.Count()
		}
		time.Sleep(50 * time.Millisecond)
		go func() {
			v1.Run()
		}()
		time.Sleep(50 * time.Millisecond)
		parseTime, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			v1.TimeNowISOString(),
		)
		assert(IsTimeApproximatelyEqual(parseTime, time.Now())).IsTrue()
	})
}

func TestConvertToIsoDateString(t *testing.T) {
	t.Run("year 0000", func(t *testing.T) {
		assert := NewAssert(t)
		v1, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"0000-01-01T00:00:00+00:00",
		)
		assert(ConvertToIsoDateString(v1)).
			Equals("0000-01-01T00:00:00.000+00:00")
	})

	t.Run("year 9999", func(t *testing.T) {
		assert := NewAssert(t)
		v1, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"9998-01-01T00:00:00+00:00",
		)
		v1 = v1.Add(1000000 * time.Hour)
		assert(ConvertToIsoDateString(v1)).
			Equals("9999-01-30T16:00:00.000+00:00")
	})

	t.Run("test timezone", func(t *testing.T) {
		assert := NewAssert(t)

		v1, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"2222-12-22T11:11:11.333-11:59",
		)
		assert(ConvertToIsoDateString(v1)).
			Equals("2222-12-22T11:11:11.333-11:59")

		v2, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"2222-12-22T11:11:11.333+11:59",
		)
		assert(ConvertToIsoDateString(v2)).
			Equals("2222-12-22T11:11:11.333+11:59")

		v3, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"2222-12-22T11:11:11.333+00:00",
		)
		assert(ConvertToIsoDateString(v3)).
			Equals("2222-12-22T11:11:11.333+00:00")

		v4, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"2222-12-22T11:11:11.333-00:00",
		)
		assert(ConvertToIsoDateString(v4)).
			Equals("2222-12-22T11:11:11.333+00:00")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		start, _ := time.Parse(
			"2006-01-02T15:04:05.999Z07:00",
			"0001-01-01T00:00:00+00:00",
		)
		for i := 0; i < 100000; i++ {
			parseTime, err := time.Parse(
				"2006-01-02T15:04:05.999Z07:00",
				ConvertToIsoDateString(start),
			)
			assert(err).IsNil()
			assert(parseTime.UnixNano()).Equals(start.UnixNano())
			start = start.Add(2710991970000000)
		}
	})
}

func TestTimeNow(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsTimeApproximatelyEqual(TimeNow(), time.Now()))
	})
}

func TestTimeNowISOString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1, _ := time.Parse("2006-01-02T15:04:05.999Z07:00", TimeNowISOString())
		assert(IsTimeApproximatelyEqual(v1, time.Now()))
	})
}

func TestIsTimeApproximatelyEqual(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)

		now := time.Now()

		assert(IsTimeApproximatelyEqual(now, now.Add(500*time.Millisecond))).
			IsFalse()
		assert(IsTimeApproximatelyEqual(now.Add(500*time.Millisecond), now)).
			IsFalse()
		assert(IsTimeApproximatelyEqual(now, now.Add(499*time.Millisecond))).
			IsTrue()
		assert(IsTimeApproximatelyEqual(now.Add(499*time.Millisecond), now)).
			IsTrue()
	})
}

package core

import (
	"testing"
)

func Test_NewStringBuilder(t *testing.T) {
	assert := newAssert(t)

	builder := NewStringBuilder()
	assert(len(builder.buffer)).Equals(0)
	assert(cap(builder.buffer)).Equals(4096)
	builder.Release()
}

func Test_StringBuilder_Release(t *testing.T) {
	assert := newAssert(t)

	builder := NewStringBuilder()

	for i := 0; i < 4096; i++ {
		builder.AppendString("S")
	}
	builder.Release()
	builder = NewStringBuilder()
	assert(len(builder.buffer)).Equals(0)
	assert(cap(builder.buffer)).Equals(4096)

	for i := 0; i < 4097; i++ {
		builder.AppendString("S")
	}
	builder.Release()
	builder = NewStringBuilder()
	assert(len(builder.buffer)).Equals(0)
	assert(cap(builder.buffer)).Equals(4096)

	builder.Release()
}

func Test_StringBuilder_AppendBytes(t *testing.T) {
	assert := newAssert(t)

	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "hello"
	}

	builder := NewStringBuilder()
	builder.AppendBytes([]byte(longString))
	assert(builder.String()).Equals(longString)
	builder.Release()
}

func Test_StringBuilder_AppendString(t *testing.T) {
	assert := newAssert(t)

	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "hello"
	}

	var testCollection = [][2]interface{}{
		{[]string{""}, ""},
		{[]string{"a"}, "a"},
		{[]string{"中国"}, "中国"},
		{[]string{"🀄🀄🀄🀄🀄🀄🀄️"}, "🀄🀄🀄🀄🀄🀄🀄️"},
		{[]string{longString}, longString},
		{[]string{"", "🀄🀄🀄🀄🀄🀄🀄️"}, "🀄🀄🀄🀄🀄🀄🀄️"},
		{[]string{"中国", "🀄🀄🀄🀄🀄🀄🀄️"}, "中国🀄🀄🀄🀄🀄🀄🀄️"},
	}

	for _, item := range testCollection {
		builder := NewStringBuilder()
		for i := 0; i < len(item[0].([]string)); i++ {
			builder.AppendString(item[0].([]string)[i])
		}
		assert(builder.String()).Equals(item[1])
		builder.Release()
	}
}

func TestStringBuilder_AppendFormat(t *testing.T) {
	assert := newAssert(t)

	builder := NewStringBuilder()
	builder.AppendFormat("test")
	assert(builder.String()).Equals("test")
	builder.Release()

	builder = NewStringBuilder()
	builder.AppendFormat("test %d", 100)
	assert(builder.String()).Equals("test 100")
	builder.Release()

	builder = NewStringBuilder()
	builder.AppendFormat("test %s %d", "hello", 100)
	assert(builder.String()).Equals("test hello 100")
	builder.Release()
}

func Test_StringBuilder_String(t *testing.T) {
	assert := newAssert(t)

	builder := NewStringBuilder()
	builder.AppendString("a")
	assert(builder.String()).Equals("a")
}

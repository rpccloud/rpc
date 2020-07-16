package internal

import (
	"testing"
)

func TestNewStringBuilder(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	assert(len(builder.buffer)).Equals(0)
	assert(cap(builder.buffer)).Equals(4096)
	builder.Release()
}

func TestStringBuilder_Release(t *testing.T) {
	assert := NewAssert(t)

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

func TestStringBuilder_AppendByte(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	builder.AppendByte('a')
	builder.AppendByte('b')
	builder.AppendByte('c')
	assert(builder.String()).Equals("abc")
	builder.Release()
}

func TestStringBuilder_AppendBytes(t *testing.T) {
	assert := NewAssert(t)

	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "hello"
	}

	builder := NewStringBuilder()
	builder.AppendBytes([]byte(longString))
	assert(builder.String()).Equals(longString)
	builder.Release()
}

func TestStringBuilder_AppendString(t *testing.T) {
	assert := NewAssert(t)

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

func TestStringBuilder_Merge(t *testing.T) {
	assert := NewAssert(t)

	sb1 := NewStringBuilder()
	sb2 := NewStringBuilder()

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("")
	sb2.AppendString("123")

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("123")

	sb1.Merge(sb2)
	assert(sb1.String()).Equals("123123")

	sb1.Merge(nil)
	assert(sb1.String()).Equals("123123")
}

func TestStringBuilder_IsEmpty(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	assert(builder.IsEmpty()).IsTrue()
	builder.AppendString("a")
	assert(builder.IsEmpty()).IsFalse()
}

func TestStringBuilder_String(t *testing.T) {
	assert := NewAssert(t)

	builder := NewStringBuilder()
	builder.AppendString("a")
	assert(builder.String()).Equals("a")
}

package common

import (
	"testing"
	"unsafe"
)

type testObject struct {
	test string
}

func Test_GetStackString(t *testing.T) {
	assert := NewAssert(t)
	assert(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0]).Contains("Test_GetStackString")
	assert(FindLinesByPrefix(
		GetStackString(0),
		"-01",
	)[0]).Contains("utils_test")
}

func Test_FindLinesByPrefix(t *testing.T) {
	assert := NewAssert(t)

	ret := FindLinesByPrefix("", "")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("")

	ret = FindLinesByPrefix("", "hello")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("hello", "dd")
	assert(len(ret)).Equals(0)

	ret = FindLinesByPrefix("  ddhello", "dd")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals("  ddhello")

	ret = FindLinesByPrefix(" \t ddhello", "dd")
	assert(len(ret)).Equals(1)
	assert(ret[0]).Equals(" \t ddhello")

	ret = FindLinesByPrefix(" \t ddhello\ndd\n", "dd")
	assert(len(ret)).Equals(2)
	assert(ret[0]).Equals(" \t ddhello")
	assert(ret[1]).Equals("dd")
}

func Test_GetByteArrayDebugString(t *testing.T) {
	assert := NewAssert(t)
	assert(GetByteArrayDebugString([]byte{})).Equals(
		"",
	)
	assert(GetByteArrayDebugString([]byte{1, 2})).Equals(
		"0000: 0x01 0x02 ",
	)
	assert(GetByteArrayDebugString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})).Equals(
		"0000: 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08 0x09 0x0a 0x0b 0x0c 0x0d 0x0e 0x0f 0x10 ",
	)
	assert(GetByteArrayDebugString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17})).Equals(
		"0000: 0x01 0x02 0x03 0x04 0x05 0x06 0x07 0x08 0x09 0x0a 0x0b 0x0c 0x0d 0x0e 0x0f 0x10 \n0016: 0x11 ",
	)
}

func Test_GetUrlBySchemeHostPortAndPath(t *testing.T) {
	assert := NewAssert(t)

	assert(GetURLBySchemeHostPortAndPath("", "127.0.0.1", 8080, "/world")).
		Equals("")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "")).
		Equals("ws://127.0.0.1:8080/")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "/")).
		Equals("ws://127.0.0.1:8080/")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "world")).
		Equals("ws://127.0.0.1:8080/world")
	assert(GetURLBySchemeHostPortAndPath("ws", "127.0.0.1", 8080, "/world")).
		Equals("ws://127.0.0.1:8080/world")
}

func Test_ConvertOrdinalToString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConvertOrdinalToString(0)).Equals("")
	assert(ConvertOrdinalToString(1)).Equals("1st")
	assert(ConvertOrdinalToString(2)).Equals("2nd")
	assert(ConvertOrdinalToString(3)).Equals("3rd")
	assert(ConvertOrdinalToString(4)).Equals("4th")
	assert(ConvertOrdinalToString(10)).Equals("10th")
	assert(ConvertOrdinalToString(100)).Equals("100th")
}

func Test_GetObjectFieldPointer(t *testing.T) {
	assert := NewAssert(t)
	obj := &testObject{
		test: "hi",
	}
	assert(GetObjectFieldPointer(obj, "test")).Equals(unsafe.Pointer(&obj.test))
}

func Test_AddPrefixPerLine(t *testing.T) {
	assert := NewAssert(t)

	assert(AddPrefixPerLine("", "")).Equals("")
	assert(AddPrefixPerLine("a", "")).Equals("a")
	assert(AddPrefixPerLine("\n", "")).Equals("\n")
	assert(AddPrefixPerLine("a\n", "")).Equals("a\n")
	assert(AddPrefixPerLine("a\nb", "")).Equals("a\nb")
	assert(AddPrefixPerLine("", "-")).Equals("-")
	assert(AddPrefixPerLine("a", "-")).Equals("-a")
	assert(AddPrefixPerLine("\n", "-")).Equals("-\n-")
	assert(AddPrefixPerLine("a\n", "-")).Equals("-a\n-")
	assert(AddPrefixPerLine("a\nb", "-")).Equals("-a\n-b")
}

func Test_isNil(t *testing.T) {
	assert := NewAssert(t)

	assert(isNil(nil)).IsTrue()
	assert(isNil((*RPCStream)(nil))).IsTrue()
	assert(isNil((*RPCArray)(nil))).IsTrue()
	assert(isNil((*RPCMap)(nil))).IsTrue()

	assert(isNil(nilRPCArray)).IsFalse()
	assert(isNil(nilRPCMap)).IsFalse()

	unsafeNil := unsafe.Pointer(nil)
	uintptrNil := uintptr(0)

	assert(isNil(unsafeNil)).IsTrue()
	assert(isNil(uintptrNil)).IsTrue()
}

func Test_equals(t *testing.T) {
	assert := NewAssert(t)
	invalidPub := &PubControl{ctx: nil}

	loggerPtr := NewLogger()
	testCollection := [][3]interface{}{
		{true, true, true},
		{false, false, true},
		{false, true, false},
		{false, 0, false},
		{true, 1, false},
		{true, nil, false},
		{0, 0, true},
		{3, 4, false},
		{3, int(3), true},
		{3, int32(3), false},
		{3, nil, false},
		{3.14, 3.14, true},
		{3.14, 3.15, false},
		{3.14, float32(3.14), false},
		{3.14, float64(3.14), true},
		{3.14, nil, false},
		{"", "", true},
		{"abc", "abc", true},
		{"abc", "ab", false},
		{"", nil, false},
		{"", 6, false},
		{"", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("")}, true},
		{"abc", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("abc")}, true},
		{"abc", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("ab")}, false},
		{"abc", RPCString{pub: invalidPub, status: rpcStatusAllocated, bytes: ([]byte)("abc")}, false},
		{"hi", RPCString{status: rpcStatusAllocated, bytes: nil}, false},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("")}, "", true},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("abc")}, "abc", true},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("abc")}, "ab", false},
		{RPCString{pub: invalidPub, status: rpcStatusAllocated, bytes: ([]byte)("abc")}, "ab", false},
		{RPCString{status: rpcStatusAllocated, bytes: nil}, "hi", false},
		{errorRPCString, errorRPCString, true},

		{[]byte{}, []byte{}, true},
		{[]byte{12}, []byte{12}, true},
		{[]byte{12, 13}, []byte{12, 13}, true},
		{[]byte{12, 13}, 12, false},
		{[]byte{13, 12}, []byte{12, 13}, false},
		{[]byte{12}, []byte{12, 13}, false},
		{[]byte{12, 13}, []byte{12}, false},
		{[]byte{13, 12}, nil, false},
		{[]byte{}, nil, false},
		{[]byte{}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{}}, true},
		{[]byte{12, 13}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{12, 13}}, true},
		{[]byte{12, 13}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{12}}, false},
		{[]byte{12, 13}, RPCBytes{pub: invalidPub, status: rpcStatusAllocated, bytes: []byte{12}}, false},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{}}, []byte{}, true},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{12, 13}}, []byte{12, 13}, true},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{12}}, []byte{12, 13}, false},
		{RPCBytes{pub: invalidPub, status: rpcStatusAllocated, bytes: []byte{12}}, []byte{12, 13}, false},

		{nilRPCMap, nilRPCMap, true},
		{toRPCMap(map[string]interface{}{}), toRPCMap(map[string]interface{}{}), true},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			true,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			toRPCArray([]interface{}{9007199254740991}),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			toRPCMap(map[string]interface{}{"test": 9007199254740991, "3": 9007199254740991}),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991, "3": 9007199254740991}),
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			toRPCMap(map[string]interface{}{"test": 9007199254740990}),
			false,
		},
		{
			toRPCMap(map[string]interface{}{"test": 9007199254740991}),
			toRPCMap(nil),
			false,
		},
		{toRPCMap(map[string]interface{}{}), nil, false},
		{nilRPCArray, nilRPCArray, true},
		{toRPCArray([]interface{}{}), toRPCArray([]interface{}{}), true},
		{toRPCArray([]interface{}{1}), toRPCArray([]interface{}{1}), true},
		{toRPCArray([]interface{}{1, 2}), toRPCArray([]interface{}{1, 2}), true},
		{toRPCArray([]interface{}{1, 2}), 3, false},
		{toRPCArray([]interface{}{1, 2}), toRPCArray([]interface{}{1}), false},
		{toRPCArray([]interface{}{1}), toRPCArray([]interface{}{1, 2}), false},
		{toRPCArray([]interface{}{1, 2}), toRPCArray([]interface{}{2, 1}), false},
		{toRPCArray([]interface{}{1, 2}), toRPCArray(nil), false},
		{toRPCArray([]interface{}{}), toRPCArray(nil), false},

		{nil, nil, true},
		{nil, (*Logger)(nil), true},
		{(*Logger)(nil), nil, true},
		{nil, []interface{}(nil), true},
		{nil, map[string]interface{}(nil), true},
		{nil, []byte(nil), true},
		{nil, []byte{}, false},
		{RPCArray{}, nil, false},
		{RPCMap{}, nil, false},
		{[]byte{}, nil, false},

		{NewRPCErrorWithDebug("m1", "d1"), NewRPCErrorWithDebug("m1", "d1"), true},
		{NewRPCErrorWithDebug("", "d1"), NewRPCErrorWithDebug("m1", "d1"), false},
		{NewRPCErrorWithDebug("m1", ""), NewRPCErrorWithDebug("m1", "d1"), false},
		{NewRPCErrorWithDebug("m1", ""), nil, false},
		{NewRPCErrorWithDebug("m1", ""), 3, false},

		{loggerPtr, loggerPtr, true},
		{NewLogger(), NewLogger(), false},

		{errorRPCBytes, errorRPCBytes, true},
		{nilRPCArray, nilRPCArray, true},
		{nilRPCMap, nilRPCMap, true},
	}

	for _, item := range testCollection {
		assert(equals(item[0], item[1]) == item[2]).IsTrue()
	}
}

func Test_equals_exceptions(t *testing.T) {
	assert := NewAssert(t)

	rightArray := newRPCArray(nil)
	errorArray := newRPCArray(nil)
	rightArray.Append(true)
	errorArray.Append(true)
	(*errorArray.getStream().frames[0])[1] = 13
	assert(equals(rightArray, errorArray)).IsFalse()
	assert(equals(errorArray, rightArray)).IsFalse()

	rightMap := newRPCMap(nil)
	errorMap := newRPCMap(nil)
	rightMap.Set("0", true)
	errorMap.Set("0", true)
	(*errorMap.getStream().frames[0])[1] = 13
	assert(equals(rightMap, errorMap)).IsFalse()
	assert(equals(errorMap, rightMap)).IsFalse()
}

func Test_contains(t *testing.T) {
	assert := NewAssert(t)
	invalidPub := &PubControl{ctx: nil}

	testCollection := [][3]interface{}{
		{"hello world", "world", 1},
		{"hello world", "you", 0},
		{"hello world", 3, -1},
		{"hello world", nil, -1},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, "world", 1},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, "you", 0},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, 3, -1},
		{RPCString{status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, nil, -1},
		{RPCString{pub: invalidPub, status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, "world", -1},
		{"hello world", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("world")}, 1},
		{"hello world", RPCString{status: rpcStatusAllocated, bytes: ([]byte)("you")}, 0},
		{"hello world", RPCString{pub: invalidPub, status: rpcStatusAllocated, bytes: ([]byte)("hello world")}, -1},
		{toRPCArray([]interface{}{1, 2, int64(3)}), int64(3), 1},
		{toRPCArray([]interface{}{1, 2, int64(3)}), int(3), 0},
		{toRPCArray([]interface{}{1, 2, 3}), 0, 0},
		{toRPCArray([]interface{}{1, 2, 3}), nil, 0},
		{toRPCArray([]interface{}{1, 2, 3}), true, 0},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}), "1", -1},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}), "3", -1},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}), true, -1},
		{toRPCMap(map[string]interface{}{"1": 1, "2": 2}), nil, -1},
		{[]byte{}, []byte{}, 1},
		{[]byte{1, 2, 3, 4}, []byte{}, 1},
		{[]byte{1, 2, 3, 4}, []byte{2, 3}, 1},
		{[]byte{1, 2}, []byte{1, 2}, 1},
		{[]byte{1, 2}, []byte{1, 2, 3}, 0},
		{[]byte{1, 2, 3, 4}, []byte{2, 4}, 0},
		{[]byte{1, 2}, 1, -1},
		{[]byte{1, 2}, true, -1},
		{[]byte{1, 2}, nil, -1},

		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{}}, []byte{}, 1},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2, 3, 4}}, []byte{}, 1},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2, 3, 4}}, []byte{2, 3}, 1},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, []byte{1, 3}, 0},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, 1, -1},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, true, -1},
		{RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, nil, -1},
		{RPCBytes{pub: invalidPub, status: rpcStatusAllocated, bytes: []byte{1, 2, 3, 4}}, []byte{}, -1},

		{[]byte{}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{}}, 1},
		{[]byte{1, 2, 3, 4}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{}}, 1},
		{[]byte{1, 2, 3, 4}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{2, 3}}, 1},

		{[]byte{1, 2}, RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 3}}, 0},
		{1, RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, -1},
		{true, RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, -1},
		{nil, RPCBytes{status: rpcStatusAllocated, bytes: []byte{1, 2}}, -1},
		{[]byte{1, 2, 3, 4}, RPCBytes{pub: invalidPub, status: rpcStatusAllocated, bytes: []byte{2, 3}}, -1},

		{nil, "3", -1},
		{nil, nil, -1},
		{true, 3, -1},
		{float64(0), float64(0), -1},
	}

	for _, v := range testCollection {
		assert(contains(v[0], v[1])).Equals(v[2])
	}
}

func Test_contains_exceptions(t *testing.T) {
	assert := NewAssert(t)

	errorArray := newRPCArray(nil)
	errorArray.Append(true)
	(*errorArray.getStream().frames[0])[1] = 13

	assert(contains(errorArray, true)).Equals(0)
}

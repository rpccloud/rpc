package rpc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func getTestDepthArray(depth int) [2]interface{} {
	if depth <= 1 {
		return [2]interface{}{true, ""}
	}

	ret := getTestDepthArray(depth - 1)
	return [2]interface{}{Array{ret[0]}, "[0]" + ret[1].(string)}
}

func getTestDepthMap(depth int) [2]interface{} {
	if depth <= 1 {
		return [2]interface{}{true, ""}
	}

	ret := getTestDepthMap(depth - 1)
	return [2]interface{}{Map{"a": ret[0]}, "[\"a\"]" + ret[1].(string)}
}

var streamTestSuccessCollections = map[string][][2]interface{}{
	"nil": {
		{nil, []byte{0x01}},
	},
	"bool": {
		{true, []byte{0x02}},
		{false, []byte{0x03}},
	},
	"float64": {
		{float64(0), []byte{0x04}},
		{float64(100), []byte{
			0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0x40,
		}},
		{3.1415926, []byte{
			0x05, 0x4A, 0xD8, 0x12, 0x4D, 0xFB, 0x21, 0x09, 0x40,
		}},
		{-3.1415926, []byte{
			0x05, 0x4A, 0xD8, 0x12, 0x4D, 0xFB, 0x21, 0x09, 0xC0,
		}},
		{float64(-100), []byte{
			0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0xC0,
		}},
	},
	"int64": {
		{int64(-9223372036854775808), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}},
		{int64(-9007199254740992), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x7F,
		}},
		{int64(-9007199254740991), []byte{
			0x08, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe0, 0x7F,
		}},
		{int64(-9007199254740990), []byte{
			0x08, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x7F,
		}},
		{int64(-2147483649), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0x7F, 0xFF, 0xFF, 0xFF, 0x7F,
		}},
		{int64(-2147483648), []byte{0x07, 0x00, 0x00, 0x00, 0x00}},
		{int64(-32769), []byte{0x07, 0xFF, 0x7F, 0xFF, 0x7F}},
		{int64(-32768), []byte{0x06, 0x00, 0x00}},
		{int64(-8), []byte{0x06, 0xF8, 0x7F}},
		{int64(-7), []byte{0x0E}},
		{int64(-1), []byte{0x14}},
		{int64(0), []byte{0x15}},
		{int64(1), []byte{0x16}},
		{int64(32), []byte{0x35}},
		{int64(33), []byte{0x06, 0x21, 0x80}},
		{int64(32767), []byte{0x06, 0xFF, 0xFF}},
		{int64(32768), []byte{0x07, 0x00, 0x80, 0x00, 0x80}},
		{int64(2147483647), []byte{0x07, 0xFF, 0xFF, 0xFF, 0xFF}},
		{int64(2147483648), []byte{
			0x08, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x80,
		}},
		{int64(9007199254740990), []byte{
			0x08, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x80,
		}},
		{int64(9007199254740991), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x80,
		}},
		{int64(9007199254740992), []byte{
			0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x80,
		}},
		{int64(9223372036854775807), []byte{
			0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}},
	},
	"uint64": {
		{uint64(0), []byte{0x36}},
		{uint64(9), []byte{0x3F}},
		{uint64(10), []byte{0x09, 0x0A, 0x00}},
		{uint64(65535), []byte{0x09, 0xFF, 0xFF}},
		{uint64(65536), []byte{0x0A, 0x00, 0x00, 0x01, 0x00}},
		{uint64(4294967295), []byte{0x0A, 0xFF, 0xFF, 0xFF, 0xFF}},
		{uint64(4294967296), []byte{
			0x0B, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		}},
		{uint64(9007199254740990), []byte{
			0x0B, 0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x00,
		}},
		{uint64(9007199254740991), []byte{
			0x0B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x1F, 0x00,
		}},
		{uint64(9007199254740992), []byte{
			0x0B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x00,
		}},
		{uint64(18446744073709551615), []byte{
			0x0B, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		}},
	},
	"string": {
		{"", []byte{0x80}},
		{"a", []byte{0x81, 0x61, 0x00}},
		{"😀☘️🀄️©️🌈🎩", []byte{
			0x9E, 0xF0, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
			0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
			0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
			0xA9, 0x00,
		}},
		{"😀中☘️文🀄️©️🌈🎩测试a\n\r\b", []byte{
			0xAE, 0xF0, 0x9F, 0x98, 0x80, 0xE4, 0xB8, 0xAD, 0xE2, 0x98,
			0x98, 0xEF, 0xB8, 0x8F, 0xE6, 0x96, 0x87, 0xF0, 0x9F, 0x80,
			0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0,
			0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xE6, 0xB5, 0x8B,
			0xE8, 0xAF, 0x95, 0x61, 0x0A, 0x0D, 0x08, 0x00,
		}},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", []byte{
			0xBF, 0x45, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		}},
		{"😀☘️🀄️©️🌈🎩😛👩👩👦👨👩👦👦👼🗣👑👚👹👺🌳🍊", []byte{
			0xBF, 0x64, 0x00, 0x00, 0x00, 0xF0, 0x9F, 0x98, 0x80, 0xE2,
			0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
			0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
			0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
			0x9F, 0x91, 0xA9, 0xF0, 0x9F, 0x91, 0xA9, 0xF0, 0x9F, 0x91,
			0xA6, 0xF0, 0x9F, 0x91, 0xA8, 0xF0, 0x9F, 0x91, 0xA9, 0xF0,
			0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
			0xBC, 0xF0, 0x9F, 0x97, 0xA3, 0xF0, 0x9F, 0x91, 0x91, 0xF0,
			0x9F, 0x91, 0x9A, 0xF0, 0x9F, 0x91, 0xB9, 0xF0, 0x9F, 0x91,
			0xBA, 0xF0, 0x9F, 0x8C, 0xB3, 0xF0, 0x9F, 0x8D, 0x8A, 0x00,
		}},
	},
	"bytes": {
		{[]byte{}, []byte{0xC0}},
		{[]byte{0xDA}, []byte{0xC1, 0xDA}},
		{[]byte{
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61,
		}, []byte{
			0xFF, 0x44, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		}},
	},
	"array": {
		{Array{}, []byte{64}},
		{Array{true}, []byte{
			65, 6, 0, 0, 0, 2,
		}},
		{Array{"", true}, []byte{
			66, 7, 0, 0, 0, 0x80, 0x02,
		}},
		{Array{"a", true}, []byte{
			66, 9, 0, 0, 0, 0x81, 0x61, 0x00, 0x02,
		}},
		{Array{
			true, false,
		}, []byte{
			66, 7, 0, 0, 0, 2, 3,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
		}, []byte{
			94, 35, 0, 0, 0, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true,
		}, []byte{
			95, 40, 0, 0, 0, 31, 0, 0, 0, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		}},
		{Array{
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true, true, true,
			true, true,
		}, []byte{
			95, 41, 0, 0, 0, 32, 0, 0, 0, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			2,
		}},
	},
	"map": {
		{Map{}, []byte{0x60}},
		{Map{"1": true}, []byte{
			0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02,
		}},
		{Map{"1": ""}, []byte{
			0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x80,
		}},
		{Map{"1": "a"}, []byte{
			0x61, 0x0B, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x81, 0x61, 0x00,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true,
		}, []byte{
			0x7E, 0x7D, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x81,
			0x32, 0x00, 0x02, 0x81, 0x33, 0x00, 0x02, 0x81, 0x34, 0x00,
			0x02, 0x81, 0x35, 0x00, 0x02, 0x81, 0x36, 0x00, 0x02, 0x81,
			0x37, 0x00, 0x02, 0x81, 0x38, 0x00, 0x02, 0x81, 0x39, 0x00,
			0x02, 0x81, 0x61, 0x00, 0x02, 0x81, 0x62, 0x00, 0x02, 0x81,
			0x63, 0x00, 0x02, 0x81, 0x64, 0x00, 0x02, 0x81, 0x65, 0x00,
			0x02, 0x81, 0x66, 0x00, 0x02, 0x81, 0x67, 0x00, 0x02, 0x81,
			0x68, 0x00, 0x02, 0x81, 0x69, 0x00, 0x02, 0x81, 0x6A, 0x00,
			0x02, 0x81, 0x6B, 0x00, 0x02, 0x81, 0x6C, 0x00, 0x02, 0x81,
			0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00, 0x02, 0x81, 0x6F, 0x00,
			0x02, 0x81, 0x70, 0x00, 0x02, 0x81, 0x71, 0x00, 0x02, 0x81,
			0x72, 0x00, 0x02, 0x81, 0x73, 0x00, 0x02, 0x81, 0x74, 0x00,
			0x02, 0x81, 0x75, 0x00, 0x02,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true,
		}, []byte{
			0x7F, 0x85, 0x00, 0x00, 0x00, 0x1F, 0x00, 0x00, 0x00, 0x81,
			0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
			0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
			0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
			0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
			0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
			0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
			0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
			0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
			0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
			0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
			0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
			0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
			0x76, 0x00, 0x02,
		}},
		{Map{
			"1": true, "2": true, "3": true, "4": true,
			"5": true, "6": true, "7": true, "8": true,
			"9": true, "a": true, "b": true, "c": true,
			"d": true, "e": true, "f": true, "g": true,
			"h": true, "i": true, "j": true, "k": true,
			"l": true, "m": true, "n": true, "o": true,
			"p": true, "q": true, "r": true, "s": true,
			"t": true, "u": true, "v": true, "w": true,
		}, []byte{
			0x7F, 0x89, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x81,
			0x31, 0x00, 0x02, 0x81, 0x32, 0x00, 0x02, 0x81, 0x33, 0x00,
			0x02, 0x81, 0x34, 0x00, 0x02, 0x81, 0x35, 0x00, 0x02, 0x81,
			0x36, 0x00, 0x02, 0x81, 0x37, 0x00, 0x02, 0x81, 0x38, 0x00,
			0x02, 0x81, 0x39, 0x00, 0x02, 0x81, 0x61, 0x00, 0x02, 0x81,
			0x62, 0x00, 0x02, 0x81, 0x63, 0x00, 0x02, 0x81, 0x64, 0x00,
			0x02, 0x81, 0x65, 0x00, 0x02, 0x81, 0x66, 0x00, 0x02, 0x81,
			0x67, 0x00, 0x02, 0x81, 0x68, 0x00, 0x02, 0x81, 0x69, 0x00,
			0x02, 0x81, 0x6A, 0x00, 0x02, 0x81, 0x6B, 0x00, 0x02, 0x81,
			0x6C, 0x00, 0x02, 0x81, 0x6D, 0x00, 0x02, 0x81, 0x6E, 0x00,
			0x02, 0x81, 0x6F, 0x00, 0x02, 0x81, 0x70, 0x00, 0x02, 0x81,
			0x71, 0x00, 0x02, 0x81, 0x72, 0x00, 0x02, 0x81, 0x73, 0x00,
			0x02, 0x81, 0x74, 0x00, 0x02, 0x81, 0x75, 0x00, 0x02, 0x81,
			0x76, 0x00, 0x02, 0x81, 0x77, 0x00, 0x02,
		}},
	},
}

var streamTestWriteCollections = map[string][][2]interface{}{
	"array": {
		{
			Array{true, true, true, make(chan bool), true},
			"[3] type(chan bool) is not supported",
		},
		{
			getTestDepthArray(65)[0],
			getTestDepthArray(65)[1].(string) + " overflows",
		},
		{
			getTestDepthArray(64)[0],
			StreamWriteOK,
		},
	},
	"map": {
		{
			Map{"0": 0, "1": make(chan bool)},
			"[\"1\"] type(chan bool) is not supported",
		},
		{
			getTestDepthMap(65)[0],
			getTestDepthMap(65)[1].(string) + " overflows",
		},
		{
			getTestDepthMap(64)[0],
			StreamWriteOK,
		},
	},
	"rtArray": {
		{
			RTArray{},
			" is not available",
		},
		{
			RTArray{rt: testRuntime, items: &[]posRecord{0}},
			" is not available",
		},
	},
	"rtMap": {
		{
			RTMap{},
			" is not available",
		},
		{
			RTMap{
				rt:    testRuntime,
				items: &[]mapItem{{"key", getFastKey("key"), 0}},
				length: (func() *uint32 {
					v := uint32(1)
					return &v
				})(),
			},
			" is not available",
		},
	},
	"rtValue": {
		{
			RTValue{},
			" is not available",
		},
		{
			RTValue{rt: testRuntime, pos: 0},
			" is not available",
		},
		{
			RTValue{rt: testRuntime, pos: streamPosBody},
			" is not available",
		},
		{
			RTValue{rt: testRuntime, err: base.ErrStream},
			" is not available",
		},
	},
	"value": {
		{nil, StreamWriteOK},
		{true, StreamWriteOK},
		{0, StreamWriteOK},
		{int8(0), StreamWriteOK},
		{int16(0), StreamWriteOK},
		{int32(0), StreamWriteOK},
		{int64(0), StreamWriteOK},
		{uint(0), StreamWriteOK},
		{uint8(0), StreamWriteOK},
		{uint16(0), StreamWriteOK},
		{uint32(0), StreamWriteOK},
		{uint64(0), StreamWriteOK},
		{float32(0), StreamWriteOK},
		{float64(0), StreamWriteOK},
		{"", StreamWriteOK},
		{[]byte{}, StreamWriteOK},
		{Array{}, StreamWriteOK},
		{Map{}, StreamWriteOK},
		{RTArray{rt: testRuntime, items: &[]posRecord{}}, StreamWriteOK},
		{
			RTMap{rt: testRuntime, items: &[]mapItem{}, length: new(uint32)},
			StreamWriteOK,
		},
		{RTValue{}, "value is not available"},
		{RTArray{}, "value is not available"},
		{RTMap{}, "value is not available"},
		{
			Array{true, true, true, make(chan bool), true},
			"value[3] type(chan bool) is not supported",
		},
		{
			getTestDepthArray(65)[0],
			"value" + getTestDepthArray(65)[1].(string) + " overflows",
		},
		{
			Map{"0": 0, "1": make(chan bool)},
			"value[\"1\"] type(chan bool) is not supported",
		},
		{
			getTestDepthMap(65)[0],
			"value" + getTestDepthMap(65)[1].(string) + " overflows",
		},
	},
}

func getTestRange(start uint, end uint, head uint, tail uint, step uint) []int {
	ret := make([]int, 0)
	for i := start; i <= end; i++ {
		pos := i % streamBlockSize
		if i < start+head {
			ret = append(ret, int(i))
		} else if pos < head {
			ret = append(ret, int(i))
		} else if pos > streamBlockSize-tail {
			ret = append(ret, int(i))
		} else {
			if i%step == 0 {
				ret = append(ret, int(i))
			}
		}
	}
	return ret
}

func TestStream(t *testing.T) {
	t.Run("test constant", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(streamVersion).Equals(1)
		assert(streamBlockSize).Equals(512)
		assert(streamBlockSize % 8).Equals(0)
		assert(streamFrameArrayInitSize).Equals(8)
		assert(streamPosVersion).Equals(0)
		assert(streamPosStatusBit).Equals(1)
		assert(streamPosKind).Equals(2)
		assert(streamPosPriority).Equals(3)
		assert(streamPosLength).Equals(4)
		assert(streamPosCheckSum).Equals(8)
		assert(streamPosCheckSum % 8).Equals(0)
		assert(streamPosZoneID).Equals(16)
		assert(streamPosTargetID).Equals(18)
		assert(streamPosSourceID).Equals(26)
		assert(streamPosGatewayID).Equals(34)
		assert(streamPosSessionID).Equals(42)
		assert(streamPosCallbackID).Equals(50)
		assert(streamPosDepth).Equals(58)
		assert(streamPosBody).Equals(60)
		assert(streamStatusBitDebug).Equals(0)
		assert(StreamBlockSize).Equals(512)
		assert(StreamHeadSize).Equals(60)
		assert(StreamWriteOK).Equals("")
		assert(StreamKindConnectRequest).Equals(1)
		assert(StreamKindConnectResponse).Equals(2)
		assert(StreamKindPing).Equals(3)
		assert(StreamKindPong).Equals(4)
		assert(StreamKindRPCRequest).Equals(5)
		assert(StreamKindRPCResponseOK).Equals(6)
		assert(StreamKindRPCResponseError).Equals(7)
		assert(StreamKindRPCBoardCast).Equals(8)
		assert(StreamKindSystemErrorReport).Equals(9)
	})

	t.Run("test initStreamFrame0", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(len(initStreamFrame0)).Equals(streamBlockSize)
		assert(cap(initStreamFrame0)).Equals(streamBlockSize)
		assert(initStreamFrame0[0]).Equals(byte(streamVersion))
		for i := 1; i < streamBlockSize; i++ {
			assert(initStreamFrame0[i]).Equals(uint8(0))
		}
	})

	t.Run("test initStreamFrameN", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(len(initStreamFrameN)).Equals(streamBlockSize)
		assert(cap(initStreamFrameN)).Equals(streamBlockSize)
		for i := 0; i < streamBlockSize; i++ {
			assert(initStreamFrameN[i]).Equals(uint8(0))
		}
	})

	t.Run("test streamCache", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := streamCache.Get().(*Stream)
		assert(len(stream.frames)).Equals(1)
		assert(cap(stream.frames)).Equals(streamFrameArrayInitSize)
		assert(stream.readSeg).Equals(0)
		assert(stream.readIndex).Equals(streamPosBody)
		assert(stream.readFrame).Equals(*(stream.frames[0]))
		assert(stream.writeSeg).Equals(0)
		assert(stream.writeIndex).Equals(streamPosBody)
		assert(stream.writeFrame).Equals(*(stream.frames[0]))
		assert(len(stream.bufferBytes)).Equals(streamBlockSize)
		assert(cap(stream.bufferBytes)).Equals(streamBlockSize)
		assert(len(stream.bufferFrames)).Equals(streamFrameArrayInitSize)
		assert(cap(stream.bufferFrames)).Equals(streamFrameArrayInitSize)
		assert(*stream.frames[0]).Equals(stream.bufferBytes[0:])
		streamCache.Put(stream)
	})

	t.Run("test frameCache", func(t *testing.T) {
		assert := base.NewAssert(t)
		frame := frameCache.Get().(*[]byte)
		assert(len(*frame)).Equals(streamBlockSize)
		assert(cap(*frame)).Equals(streamBlockSize)
		frameCache.Put(frame)
	})

	t.Run("test readSkipArray", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(readSkipArray).Equals([256]int{
			-0x01, +0x01, +0x01, +0x01, +0x01, +0x09, +0x03, +0x05,
			+0x09, +0x03, +0x05, +0x09, -0x01, -0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01, +0x01,
			+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x01, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00, +0x00,
			+0x01, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08, +0x09,
			+0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10, +0x11,
			+0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18, +0x19,
			+0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20, +0x21,
			+0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28, +0x29,
			+0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30, +0x31,
			+0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38, +0x39,
			+0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x40, +0x00,
			+0x01, +0x02, +0x03, +0x04, +0x05, +0x06, +0x07, +0x08,
			+0x09, +0x0a, +0x0b, +0x0c, +0x0d, +0x0e, +0x0f, +0x10,
			+0x11, +0x12, +0x13, +0x14, +0x15, +0x16, +0x17, +0x18,
			+0x19, +0x1a, +0x1b, +0x1c, +0x1d, +0x1e, +0x1f, +0x20,
			+0x21, +0x22, +0x23, +0x24, +0x25, +0x26, +0x27, +0x28,
			+0x29, +0x2a, +0x2b, +0x2c, +0x2d, +0x2e, +0x2f, +0x30,
			+0x31, +0x32, +0x33, +0x34, +0x35, +0x36, +0x37, +0x38,
			+0x39, +0x3a, +0x3b, +0x3c, +0x3d, +0x3e, +0x3f, +0x00,
		})
	})
}

func TestNewStream(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		assert(len(stream.frames)).Equals(1)
		assert(cap(stream.frames)).Equals(streamFrameArrayInitSize)
		assert(stream.readSeg).Equals(0)
		assert(stream.readIndex).Equals(streamPosBody)
		assert(stream.readFrame).Equals(*stream.frames[0])
		assert(stream.writeSeg).Equals(0)
		assert(stream.writeIndex).Equals(streamPosBody)
		assert(stream.writeFrame).Equals(*stream.frames[0])
		assert(*stream.frames[0]).Equals(initStreamFrame0)
		stream.Release()
	})
}

func TestStream_Reset(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()

		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i += 7 {
			assert(len(stream.frames)).Equals(1)
			assert(cap(stream.frames)).Equals(streamFrameArrayInitSize)
			assert(stream.readSeg).Equals(0)
			assert(stream.readIndex).Equals(streamPosBody)
			assert(stream.readFrame).Equals(*stream.frames[0])
			assert(stream.writeSeg).Equals(0)
			assert(stream.writeIndex).Equals(streamPosBody)
			assert(stream.writeFrame).Equals(*stream.frames[0])
			assert(*stream.frames[0]).Equals(initStreamFrame0)

			for j := 0; j < i; j++ {
				stream.PutBytes([]byte{12})
				if stream.writeIndex == 0 {
					assert(stream.writeFrame).Equals(initStreamFrameN)
				}
			}
			stream.SetReadPos(i)
			stream.Reset()
		}

		stream.Release()
	})
}

func TestStream_Release(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		mp := map[string]bool{}
		for i := 0; i < 1000; i++ {
			v := NewStream()
			mp[fmt.Sprintf("%p", v)] = true
			v.Release()
		}
		assert(len(mp) < 1000).IsTrue()
	})
}

func TestStream_Clone(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		totalBytes := 1600
		buffer := make([]byte, totalBytes)
		for i := 0; i < len(buffer); i++ {
			buffer[i] = byte(i)
		}

		for i := 0; i < totalBytes; i++ {
			v := NewStream()
			v.PutBytes(buffer[:i])
			c := v.Clone()
			assert(v.GetBuffer()).Equals(c.GetBuffer())
			assert(v.readSeg).Equals(c.readSeg)
			assert(v.readIndex).Equals(c.readIndex)
			v.Release()
		}
	})
}

func TestStream_GetVersion(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewStream()
		assert(v.GetVersion()).Equals(uint8(1))
		v.Release()
	})
}

func TestStream_SetVersion(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			v.SetVersion(uint8(i))
			assert(v.GetVersion()).Equals(uint8(i))
			v.Release()
		}
	})
}

func TestStream_HasStatusBitDebug(t *testing.T) {
	t.Run("test bit set", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			v.SetStatusBitDebug()
			assert(v.HasStatusBitDebug()).IsTrue()
			v.Release()
		}
	})

	t.Run("test bit unset", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			v.ClearStatusBitDebug()
			assert(v.HasStatusBitDebug()).IsFalse()
			v.Release()
		}
	})
}

func TestStream_SetStatusBitDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			if !v.HasStatusBitDebug() {
				v.SetStatusBitDebug()
				assert(v.HasStatusBitDebug()).IsTrue()
				v.ClearStatusBitDebug()
			}
			assert((*v.frames[0])[streamPosStatusBit]).Equals(byte(i))
			v.Release()
		}
	})
}

func TestStream_ClearStatusBitDebug(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 256; i++ {
			v := NewStream()
			(*v.frames[0])[streamPosStatusBit] = byte(i)
			if v.HasStatusBitDebug() {
				v.ClearStatusBitDebug()
				assert(v.HasStatusBitDebug()).IsFalse()
				v.SetStatusBitDebug()
			}
			assert((*v.frames[0])[streamPosStatusBit]).Equals(byte(i))
			v.Release()
		}
	})
}

func TestStream_GetLength(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			randLen := rand.Int() % 10000
			stream := NewStream()
			bytes := []byte(base.GetRandString(randLen))
			stream.PutBytes(bytes)
			stream.BuildStreamCheck()
			assert(int(stream.GetLength())).Equals(stream.GetWritePos())
			stream.Release()
		}
	})
}

func TestStream_GetKind(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			kind := uint8(rand.Uint32())
			v.SetKind(kind)
			assert(v.GetKind()).Equals(kind)
			v.Release()
		}
	})
}

func TestStream_SetKind(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			kind := uint8(rand.Uint32())
			v.SetKind(kind)
			assert(v.GetKind()).Equals(kind)
			v.Release()
		}
	})
}

func TestStream_GetPriority(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			priority := uint8(rand.Uint32())
			v.SetPriority(priority)
			assert(v.GetPriority()).Equals(priority)
			v.Release()
		}
	})
}

func TestStream_SetPriority(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			priority := uint8(rand.Uint32())
			v.SetPriority(priority)
			assert(v.GetPriority()).Equals(priority)
			v.Release()
		}
	})
}

func TestStream_getCheckSum(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			randLen := rand.Int() % 10000
			stream := NewStream()
			bytes := []byte(base.GetRandString(randLen))
			stream.PutBytes(bytes)
			copy(stream.writeFrame[stream.writeIndex:], bytes)
			stream.BuildStreamCheck()
			copy(stream.writeFrame[stream.writeIndex:], bytes)
			assert(stream.getCheckSum()).Equals(uint64(0))
			stream.Release()
		}
	})
}

func TestStream_BuildStreamCheck(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			randLen := rand.Int() % 10000
			stream := NewStream()
			bytes := []byte(base.GetRandString(randLen))
			stream.PutBytes(bytes)
			stream.BuildStreamCheck()
			assert(stream.getCheckSum()).Equals(uint64(0))
			stream.Release()
		}
	})
}

func TestStream_CheckStream(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			randLen := rand.Int() % 10000
			stream := NewStream()
			bytes := []byte(base.GetRandString(randLen))
			stream.PutBytes(bytes)
			stream.BuildStreamCheck()
			assert(stream.CheckStream()).IsTrue()
			assert(int(stream.GetLength())).Equals(stream.GetWritePos())
			stream.Release()
		}
	})

	t.Run("bytes is change", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			randLen := rand.Int() % 10000
			stream := NewStream()
			bytes := []byte(base.GetRandString(randLen))
			stream.PutBytes(bytes)
			stream.BuildStreamCheck()
			// rand change
			changePos := rand.Int() % stream.GetWritePos()
			changeSeg := changePos / streamBlockSize
			changeIndex := changePos % streamBlockSize
			(*stream.frames[changeSeg])[changeIndex]++
			assert(stream.CheckStream()).IsFalse()
			stream.Release()
		}
	})
}

func TestStream_GetZoneID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := uint16(i)
			v.SetZoneID(id)
			assert(v.GetZoneID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetZoneID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := uint16(i)
			v.SetZoneID(id)
			assert(v.GetZoneID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetTargetID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetTargetID(id)
			assert(v.GetTargetID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetTargetID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetTargetID(id)
			assert(v.GetTargetID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetSourceID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSourceID(id)
			assert(v.GetSourceID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetSourceID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSourceID(id)
			assert(v.GetSourceID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetGatewayID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetGatewayID(id)
			assert(v.GetGatewayID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetGatewayID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetGatewayID(id)
			assert(v.GetGatewayID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSessionID(id)
			assert(v.GetSessionID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetSessionID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetSessionID(id)
			assert(v.GetSessionID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetCallbackID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetCallbackID(id)
			assert(v.GetCallbackID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_SetCallbackID(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			id := rand.Uint64()
			v.SetCallbackID(id)
			assert(v.GetCallbackID()).Equals(id)
			v.Release()
		}
	})
}

func TestStream_GetDepth(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			depth := uint16(i)
			v.SetDepth(depth)
			assert(v.GetDepth()).Equals(depth)
			v.Release()
		}
	})
}

func TestStream_SetDepth(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 1000; i++ {
			v := NewStream()
			depth := uint16(i)
			v.SetDepth(depth)
			assert(v.GetDepth()).Equals(depth)
			v.Release()
		}
	})
}

func TestStream_GetReadPos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i <= 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.GetReadPos()).Equals(streamPosBody)
			} else {
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetReadPos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.SetReadPos(i)).IsFalse()
				assert(stream.GetReadPos()).Equals(streamPosBody)
			} else if i == streamPosBody {
				stream.SetWritePos(i)
				assert(stream.SetReadPos(i)).IsTrue()
				assert(stream.GetReadPos()).Equals(i)
				assert(stream.SetReadPos(i + 1)).IsFalse()
				assert(stream.GetReadPos()).Equals(i)
			} else {
				stream.SetWritePos(i)
				assert(stream.SetReadPos(i - 1)).IsTrue()
				assert(stream.GetReadPos()).Equals(i - 1)
				assert(stream.SetReadPos(i)).IsTrue()
				assert(stream.GetReadPos()).Equals(i)
				assert(stream.SetReadPos(i + 1)).IsFalse()
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetReadPosToBodyStart(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < 3*streamBlockSize; i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.GetReadPos()).Equals(i)
			}
			stream.SetReadPosToBodyStart()
			assert(stream.GetReadPos()).Equals(streamPosBody)
			stream.Release()
		}
	})
}

func TestStream_GetWritePos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i <= 3*streamBlockSize; i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.GetWritePos()).Equals(streamPosBody)
			} else {
				stream.SetWritePos(i)
				assert(stream.GetWritePos()).Equals(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetWritePos(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i++ {
			stream := NewStream()
			if i < streamPosBody {
				assert(stream.SetWritePos(i)).IsFalse()
				assert(stream.GetWritePos()).Equals(streamPosBody)
			} else {
				assert(stream.SetWritePos(i)).IsTrue()
				assert(stream.GetWritePos()).Equals(i)
			}
			stream.Release()
		}
	})
}

func TestStream_SetWritePosToBodyStart(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamFrameArrayInitSize*(streamBlockSize+2); i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
				assert(stream.GetWritePos()).Equals(i)
			}
			stream.SetWritePosToBodyStart()
			assert(stream.GetWritePos()).Equals(streamPosBody)
			stream.Release()
		}
	})
}

func TestStream_CanRead(t *testing.T) {
	t.Run("test can not read", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			assert(stream.SetReadPos(i)).IsTrue()
			assert(stream.CanRead()).IsFalse()
			if (i+1)%streamBlockSize != 0 {
				stream.readIndex = (i + 1) % streamBlockSize
				assert(stream.CanRead()).IsFalse()
			}
			stream.Release()
		}
	})

	t.Run("test can read", func(t *testing.T) {
		assert := base.NewAssert(t)

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i + 1)
			stream.SetReadPos(i)
			assert(stream.CanRead()).IsTrue()
			stream.Release()
		}
	})
}

func TestStream_IsReadFinish(t *testing.T) {
	t.Run("test read is not finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := streamPosBody + 1; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i - 1)
			assert(stream.IsReadFinish()).IsFalse()
			if (i+1)%streamBlockSize != 0 {
				stream.readIndex = (i + 1) % streamBlockSize
				assert(stream.IsReadFinish()).IsFalse()
			}
			stream.Release()
		}
	})

	t.Run("test read is finish", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			assert(stream.IsReadFinish()).IsTrue()
			stream.Release()
		}
	})
}

func TestStream_gotoNextWriteFrame(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		for i := 0; i < streamBlockSize*(streamFrameArrayInitSize+2); i++ {
			stream := NewStream()
			if i >= streamPosBody {
				stream.SetWritePos(i)
			}

			curSeg := stream.writeSeg
			assert(stream.writeSeg).Equals(len(stream.frames) - 1)
			stream.gotoNextWriteFrame()
			assert(stream.writeSeg).Equals(len(stream.frames) - 1)
			assert(stream.writeSeg).Equals(curSeg + 1)
			stream.Release()
		}
	})
}

func TestStream_gotoNextReadFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		stream := NewStream()
		stream.SetWritePos(streamBlockSize * (streamFrameArrayInitSize + 3))
		for i := streamPosBody; i < testEnd; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			stream.gotoNextReadFrameUnsafe()
			assert(stream.GetReadPos()).
				Equals((i/streamBlockSize + 1) * streamBlockSize)
		}
		stream.Release()
	})
}

func TestStream_gotoNextReadByteUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		stream := NewStream()
		stream.SetWritePos(streamBlockSize * (streamFrameArrayInitSize + 3))
		for i := streamPosBody; i < testEnd; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			stream.gotoNextReadByteUnsafe()
			assert(stream.GetReadPos()).Equals(i + 1)
		}
		stream.Release()
	})
}

func TestStream_hasNBytesToRead(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetWritePos(3 * streamBlockSize)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			for n := 0; n < 3*streamBlockSize; n++ {
				assert(stream.hasNBytesToRead(n)).
					Equals(i+n <= 3*streamBlockSize)
			}
		}
		stream.Release()
	})
}

func TestStream_isSafetyReadNBytesInCurrentFrame(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.SetWritePos(3 * streamBlockSize)
		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			assert(stream.SetReadPos(i)).IsTrue()
			for n := 0; n < 2*streamBlockSize; n++ {
				assert(stream.isSafetyReadNBytesInCurrentFrame(n)).
					Equals(streamBlockSize-i%streamBlockSize > n)
			}
		}
		stream.Release()
	})
}

func TestStream_peekNBytesCrossFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(n int) {
			if n > streamBlockSize || n <= 0 {
				panic("error")
			}

			buf := make([]byte, n)
			for i := 0; i < n; i++ {
				buf[i] = byte(i)
			}

			for i := 2*streamBlockSize - n; i < 2*streamBlockSize; i++ {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(buf)
				assert(stream.peekNBytesCrossFrameUnsafe(n)).Equals(buf)
				assert(stream.GetReadPos()).Equals(i)
				stream.Release()
			}
		}

		for _, i := range getTestRange(1, streamBlockSize, 12, 1, 93) {
			fnTest(i)
		}
	})
}

func TestStream_readNBytesCrossFrameUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		fnTest := func(n int) {
			if n > streamBlockSize || n <= 0 {
				panic("error")
			}

			buf := make([]byte, n)
			for i := 0; i < n; i++ {
				buf[i] = byte(i)
			}

			for i := 2*streamBlockSize - n; i < 2*streamBlockSize; i++ {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(buf)
				assert(stream.readNBytesCrossFrameUnsafe(n)).Equals(buf)
				assert(stream.GetReadPos()).Equals(i + n)
				stream.Release()
			}
		}

		for _, i := range getTestRange(1, streamBlockSize, 12, 1, 93) {
			fnTest(i)
		}
	})
}

func TestRpcStream_peekSkip(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)

		testCollection := Array{
			Array{[]byte{0}, 0, byte(0)},
			Array{[]byte{1}, 1, byte(1)},
			Array{[]byte{2}, 1, byte(2)},
			Array{[]byte{3}, 1, byte(3)},
			Array{[]byte{4}, 1, byte(4)},
			Array{[]byte{5}, 9, byte(5)},
			Array{[]byte{6}, 3, byte(6)},
			Array{[]byte{7}, 5, byte(7)},
			Array{[]byte{8}, 9, byte(8)},
			Array{[]byte{9}, 3, byte(9)},
			Array{[]byte{10}, 5, byte(10)},
			Array{[]byte{11}, 9, byte(11)},
			Array{[]byte{12}, 0, byte(12)},
			Array{[]byte{13}, 0, byte(13)},
			Array{[]byte{14}, 1, byte(14)},
			Array{[]byte{63}, 1, byte(63)},
			Array{[]byte{64}, 1, byte(64)},
			Array{[]byte{65, 6, 0, 0, 0}, 6, byte(65)},
			Array{[]byte{94, 6, 0, 0, 0}, 6, byte(94)},
			Array{[]byte{95, 6, 0, 0, 0}, 6, byte(95)},
			Array{[]byte{96, 6, 0, 0, 0}, 1, byte(96)},
			Array{[]byte{97, 6, 0, 0, 0}, 6, byte(97)},
			Array{[]byte{126, 6, 0, 0, 0}, 6, byte(126)},
			Array{[]byte{127, 6, 0, 0, 0}, 6, byte(127)},
			Array{[]byte{128, 6, 0, 0, 0}, 1, byte(128)},
			Array{[]byte{129, 6, 0, 0, 0}, 3, byte(129)},
			Array{[]byte{190, 6, 0, 0, 0}, 64, byte(190)},
			Array{[]byte{191, 80, 0, 0, 0}, 80, byte(191)},
			Array{[]byte{192, 6, 0, 0, 0}, 1, byte(192)},
			Array{[]byte{193, 6, 0, 0, 0}, 2, byte(193)},
			Array{[]byte{254, 6, 0, 0, 0}, 63, byte(254)},
			Array{[]byte{255, 80, 0, 0, 0}, 80, byte(255)},
			Array{[]byte{255, 80, 0}, 0, byte(255)},
		}

		for i := streamPosBody; i < 2*streamBlockSize; i++ {
			for _, item := range testCollection {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.PutBytes(item.(Array)[0].([]byte))
				assert(stream.peekSkip()).
					Equals(item.(Array)[1], item.(Array)[2])
				assert(stream.GetReadPos()).Equals(i)
				stream.Release()
			}
		}
	})
}

func TestRpcStream_writeStreamNext(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 2*streamBlockSize, 16, 16, 61)

		for _, i := range testRange {
			bytes := make([]byte, i)
			dataStream := NewStream()
			for n := 0; n < i; n++ {
				bytes[n] = byte(n)
			}
			dataStream.WriteBytes(bytes)

			// invalid code
			bugStream0 := NewStream()
			bugStream0.PutBytes([]byte{13})

			// length overflow
			bugStream1 := NewStream()
			bugStream1.PutBytes([]byte{65, 6, 0, 0, 0})

			for j := streamPosBody; j < streamBlockSize+20; j++ {
				stream := NewStream()
				stream.SetWritePos(j)
				dataStream.SetReadPos(streamPosBody)

				// dataStream
				assert(stream.writeStreamNext(dataStream)).IsTrue()
				assert(dataStream.GetReadPos()).Equals(dataStream.GetWritePos())
				assert(stream.GetWritePos()).
					Equals(dataStream.GetWritePos() + j - streamPosBody)
				// bugStream0
				assert(stream.writeStreamNext(bugStream0)).IsFalse()
				assert(bugStream0.GetReadPos()).Equals(streamPosBody)
				assert(stream.GetWritePos()).
					Equals(dataStream.GetWritePos() + j - streamPosBody)
				// bugStream1
				assert(stream.writeStreamNext(bugStream1)).IsFalse()
				assert(bugStream1.GetReadPos()).Equals(streamPosBody)
				assert(stream.GetWritePos()).
					Equals(dataStream.GetWritePos() + j - streamPosBody)

				stream.Release()
			}
		}
	})
}

func TestStream_GetBuffer(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		bytes := make([]byte, testEnd)
		for n := 0; n < testEnd; n++ {
			bytes[n] = byte(n)
		}
		for i := 0; i < testEnd; i += 7 {
			stream := NewStream()
			stream.PutBytes(bytes[:i])
			assert(stream.GetBuffer()[streamPosBody:]).Equals(bytes[:i])
			stream.Release()
		}
	})
}

func TestStream_GetBufferUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testEnd := streamBlockSize * (streamFrameArrayInitSize + 2)
		bytes := make([]byte, testEnd)
		for n := 0; n < testEnd; n++ {
			bytes[n] = byte(n)
		}
		for i := 0; i < testEnd; i += 7 {
			stream := NewStream()
			stream.PutBytes(bytes[:i])
			assert(stream.GetBufferUnsafe()[streamPosBody:]).Equals(bytes[:i])
			stream.Release()
		}
	})
}

func TestStream_PeekBufferSlice(t *testing.T) {
	t.Run("pos < 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		assert(stream.PeekBufferSlice(-1, 10)).Equals(nil, true)
		stream.Release()
	})

	t.Run("pos > stream length", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		for i := stream.GetWritePos(); i < 1200; i++ {
			assert(stream.PeekBufferSlice(i, 10)).Equals(nil, true)
		}
		stream.Release()
	})

	t.Run("max < 0", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		assert(stream.PeekBufferSlice(0, -1)).Equals(nil, true)
		stream.Release()
	})

	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		totalBytes := 700
		buffer := make([]byte, totalBytes)
		for i := 0; i < totalBytes; i++ {
			buffer[i] = byte(i)
		}

		for i := streamPosBody; i < totalBytes; i += 97 {
			v := NewStream()
			v.PutBytesTo(buffer[:i], 0)

			for j := 0; j < i; j += 93 {
				for k := 0; j+k <= i; k++ {
					bytes, isFinish := v.PeekBufferSlice(j, k)
					if k == 0 {
						assert(bytes, isFinish).Equals(nil, true)
					} else {
						maxJ := ((j / streamBlockSize) + 1) * streamBlockSize
						assert(bytes, isFinish).Equals(
							buffer[j:base.MinInt(j+k, maxJ)],
							j+len(bytes) == i,
						)
					}
				}
			}

			v.Release()
		}
	})
}

func TestStream_PutBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 10, 10, 93)
		for _, i := range testRange {
			for _, n := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				bytes := make([]byte, n)
				for z := 0; z < n; z++ {
					bytes[z] = byte(z)
				}
				stream.PutBytes(bytes)
				assert(stream.GetBuffer()[i:]).Equals(bytes)
				stream.Release()
			}
		}
	})
}

func TestStream_PutBytesTo(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(0, 3*streamBlockSize, 10, 10, 93)

		for _, i := range testRange {
			bytes := make([]byte, i)
			for n := 0; n < i; n++ {
				bytes[n] = byte(n)
			}

			for _, j := range testRange {
				stream := NewStream()
				if i+j < streamPosBody {
					assert(stream.PutBytesTo(bytes, j)).IsFalse()
				} else {
					assert(stream.PutBytesTo(bytes, j)).IsTrue()
					assert(stream.GetBuffer()[j:]).Equals(bytes)
					assert(stream.GetWritePos()).Equals(i + j)
				}

				stream.Release()
			}
		}
	})
}

func TestStream_WriteNil(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["nil"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(testData[0] == nil).IsTrue()
				stream.WriteNil()
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteBool(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bool"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteBool(testData[0].(bool))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteFloat64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["float64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteFloat64(testData[0].(float64))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteInt64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["int64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteInt64(testData[0].(int64))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteUint64(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["uint64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteUint64(testData[0].(uint64))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteString(testData[0].(string))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_WriteBytes(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.WriteBytes(testData[0].([]byte))
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_writeArray(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).
					Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).
					Equals(StreamWriteOK)
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_writeMap(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).
					Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).
					Equals(StreamWriteOK)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.SetReadPos(i)
				assert(stream.ReadMap()).Equals(testData[0].(Map), nil)
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTArray(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtArray"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTArray(testData[0].(RTArray))).
					Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.writeArray(testData[0].(Array), 64)).
					Equals(StreamWriteOK)
				rtArray, err := stream.ReadRTArray(testRuntime)
				assert(err).IsNil()
				stream.SetWritePos(i)
				assert(stream.writeRTArray(rtArray)).Equals(StreamWriteOK)
				assert(stream.GetBuffer()[i:]).Equals(testData[1])
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTMap(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtMap"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTMap(testData[0].(RTMap))).
					Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				assert(stream.writeMap(testData[0].(Map), 64)).
					Equals(StreamWriteOK)
				rtMap, err := stream.ReadRTMap(testRuntime)
				assert(err).IsNil()
				stream.SetWritePos(i)
				assert(stream.writeRTMap(rtMap)).Equals(StreamWriteOK)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.SetReadPos(i)
				assert(stream.ReadMap()).Equals(testData[0].(Map), nil)
				stream.Release()
			}
		}
	})
}

func TestStream_writeRTValue(t *testing.T) {
	t.Run("test write failed", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["rtValue"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.writeRTValue(testData[0].(RTValue))).
					Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test write ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for key := range streamTestSuccessCollections {
			for _, testData := range streamTestSuccessCollections[key] {
				for _, i := range testRange {
					testRuntime.thread.Reset()
					stream := NewStream()
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					assert(stream.Write(testData[0])).Equals(StreamWriteOK)
					rtValue, _ := stream.ReadRTValue(testRuntime)
					stream.SetWritePos(i)
					assert(stream.writeRTValue(rtValue)).Equals(StreamWriteOK)
					assert(stream.GetWritePos()).
						Equals(len(testData[1].([]byte)) + i)
					stream.SetReadPos(i)
					assert(stream.Read()).Equals(testData[0], nil)
					stream.Release()
				}
			}
		}
	})
}

func TestStream_Write(t *testing.T) {
	t.Run("test write", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestWriteCollections["value"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				assert(stream.Write(testData[0])).Equals(testData[1])
				if testData[1].(string) != StreamWriteOK {
					assert(stream.GetWritePos()).Equals(i)
				}
				stream.Release()
			}
		}
	})
}

func TestStream_ReadNil(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["nil"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadNil()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["nil"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadNil()).Equals(nil, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadNil()).Equals(nil, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadBool(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bool"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadBool()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bool"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadBool()).Equals(false, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadBool()).Equals(false, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadFloat64(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["float64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadFloat64()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["float64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadFloat64()).
						Equals(float64(0), base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadFloat64()).
				Equals(float64(0), base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadInt64(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["int64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadInt64()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["int64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadInt64()).Equals(int64(0), base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadInt64()).Equals(int64(0), base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadUint64(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["uint64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadUint64()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["uint64"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadUint64()).Equals(uint64(0), base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadUint64()).Equals(uint64(0), base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadString()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadString()).Equals("", base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadString()).Equals("", base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("read tail is not zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				stream.SetWritePos(stream.GetWritePos() - 1)
				stream.PutBytes([]byte{1})
				assert(stream.ReadString()).Equals("", base.ErrStream)
				assert(stream.GetReadPos()).Equals(i)
				stream.Release()
			}
		}
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
			0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
			0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
			0xA9, 0x00,
		})
		assert(stream.ReadString()).Equals("", base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
			0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
			0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
			0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
			0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
			0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
			0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
			0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
			0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
			0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
			0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
			0xF0, 0x9F, 0x8D, 0x8A, 0x00,
		})
		assert(stream.ReadString()).Equals("", base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		})
		assert(stream.ReadString()).Equals("", base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})
}

func TestStream_GetNodePathUnsafe(t *testing.T) {
	testCollection := make([][2]interface{}, 0)
	for _, sLen := range []int{0, 1, 61, 62, 63} {
		s := base.GetRandString(sLen)
		testCollection = append(testCollection, [2]interface{}{s, s})
		testCollection = append(testCollection, [2]interface{}{":" + s, ""})
		testCollection = append(testCollection, [2]interface{}{s + ":", s})
		testCollection = append(testCollection, [2]interface{}{
			s + "😊" + ":", s,
		})
		testCollection = append(testCollection, [2]interface{}{
			"😊" + s + ":", "",
		})
	}

	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range testCollection {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.GetNodePathUnsafe()).
					Equals(testData[1], nil)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range testCollection {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.GetNodePathUnsafe()).
						Equals("", base.ErrStream)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.GetNodePathUnsafe()).
				Equals("", base.ErrStream)
			stream.Release()
		}
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		})
		assert(stream.GetNodePathUnsafe()).Equals("", base.ErrStream)
		stream.Release()
	})
}

func TestStream_ReadUnsafeString(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])

				var notSafe bool
				sLen := len(testData[0].(string))
				if sLen == 0 {
					notSafe = false
				} else if sLen < 62 {
					notSafe = stream.readIndex+sLen < streamBlockSize-2
				} else {
					notSafe = (stream.readIndex+5)%streamBlockSize+sLen <
						streamBlockSize-1
				}
				assert(stream.readUnsafeString()).
					Equals(testData[0], !notSafe, nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.readUnsafeString()).
						Equals("", true, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.readUnsafeString()).
				Equals("", true, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("read tail is not zero", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["string"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				stream.SetWritePos(stream.GetWritePos() - 1)
				stream.PutBytes([]byte{1})
				assert(stream.readUnsafeString()).
					Equals("", true, base.ErrStream)
				assert(stream.GetReadPos()).Equals(i)
				stream.Release()
			}
		}
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0x9E, 0xFF, 0x9F, 0x98, 0x80, 0xE2, 0x98, 0x98, 0xEF, 0xB8,
			0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF, 0xB8, 0x8F, 0xC2, 0xA9,
			0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C, 0x88, 0xF0, 0x9F, 0x8E,
			0xA9, 0x00,
		})
		assert(stream.readUnsafeString()).Equals("", true, base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})

	t.Run("read string utf8 error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x6D, 0x00, 0x00, 0x00, 0xFF, 0x9F, 0x98, 0x80, 0xE2,
			0x98, 0x98, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x80, 0x84, 0xEF,
			0xB8, 0x8F, 0xC2, 0xA9, 0xEF, 0xB8, 0x8F, 0xF0, 0x9F, 0x8C,
			0x88, 0xF0, 0x9F, 0x8E, 0xA9, 0xF0, 0x9F, 0x98, 0x9B, 0xF0,
			0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9,
			0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xF0, 0x9F, 0x91,
			0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80,
			0x8D, 0xF0, 0x9F, 0x91, 0xA6, 0xE2, 0x80, 0x8D, 0xF0, 0x9F,
			0x91, 0xA6, 0xF0, 0x9F, 0x91, 0xBC, 0xF0, 0x9F, 0x97, 0xA3,
			0xF0, 0x9F, 0x91, 0x91, 0xF0, 0x9F, 0x91, 0x9A, 0xF0, 0x9F,
			0x91, 0xB9, 0xF0, 0x9F, 0x91, 0xBA, 0xF0, 0x9F, 0x8C, 0xB3,
			0xF0, 0x9F, 0x8D, 0x8A, 0x00,
		})
		assert(stream.readUnsafeString()).Equals("", true, base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xBF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x00,
		})
		assert(stream.readUnsafeString()).Equals("", true, base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})
}

func TestStream_ReadBytes(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadBytes()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["bytes"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadBytes()).Equals(Bytes{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadBytes()).Equals(Bytes{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("read string length error", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{
			0xFF, 0x2F, 0x00, 0x00, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
			0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61,
		})
		assert(stream.ReadBytes()).Equals(Bytes{}, base.ErrStream)
		assert(stream.GetReadPos()).Equals(streamPosBody)
		stream.Release()
	})
}

func TestStream_ReadArray(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadArray()).Equals(testData[0].(Array), nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadArray()).Equals(Array{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadArray()).Equals(Array{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				if len(testData[0].(Array)) > 0 {
					stream.SetWritePos(stream.GetWritePos() - 1)
					stream.PutBytes([]byte{13})
					assert(stream.ReadArray()).Equals(Array{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{0x41, 0x07, 0x00, 0x00, 0x00, 0x02, 0x02})
			assert(stream.ReadArray()).Equals(Array{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})
}

func TestStream_ReadMap(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				assert(stream.ReadMap()).Equals(testData[0], nil)
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 20, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				if len(testData[0].(Map)) > 0 {
					stream.SetWritePos(stream.GetWritePos() - 1)
					stream.PutBytes([]byte{13})
					assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("error in stream (length error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x61, 0x0A, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x02,
			})
			assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream (key error)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				wPos := stream.GetWritePos()
				mapSize := len(testData[0].(Map))

				if mapSize > 30 {
					stream.SetWritePos(i + 9)
					stream.PutBytes([]byte{13})
					stream.SetWritePos(wPos)
					assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				} else if mapSize > 0 {
					stream.SetWritePos(i + 5)
					stream.PutBytes([]byte{13})
					stream.SetWritePos(wPos)
					assert(stream.ReadMap()).Equals(Map{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})
}

func TestStream_Read(t *testing.T) {
	assert := base.NewAssert(t)

	for key := range streamTestSuccessCollections {
		for _, item := range streamTestSuccessCollections[key] {
			stream := NewStream()
			stream.PutBytes(item[1].([]byte))
			assert(stream.Read()).Equals(item[0], nil)
			stream.Release()
		}
	}

	t.Run("unsupported code 12", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{12})
		assert(stream.Read()).Equals(nil, base.ErrStream)
		stream.Release()
	})

	t.Run("unsupported code 13", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		stream.PutBytes([]byte{13})
		assert(stream.Read()).Equals(nil, base.ErrStream)
		stream.Release()
	})
}

func TestStream_ReadRTArray(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				rtArray, err := stream.ReadRTArray(testRuntime)
				assert(err).IsNil()
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.writeRTArray(rtArray)
				assert(stream.ReadArray()).Equals(testData[0], nil)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (outer stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTArray(testRuntime)).
						Equals(RTArray{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (runtime stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["array"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := testRuntime.thread.rtStream
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTArray(testRuntime)).
						Equals(RTArray{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Reset()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadRTArray(testRuntime)).
				Equals(RTArray{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{0x41, 0x07, 0x00, 0x00, 0x00, 0x02, 0x02})
			assert(stream.ReadRTArray(testRuntime)).
				Equals(RTArray{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{0x41, 0x06, 0x00, 0x00, 0x00, 0x0D})
			assert(stream.ReadRTArray(testRuntime)).
				Equals(RTArray{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x41, 0x08, 0x00, 0x00, 0x00, 0x82, 0x61, 0x00,
			})
			assert(stream.ReadRTArray(testRuntime)).
				Equals(RTArray{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTArray((func() R { s = f(0); return R{} })())).
			Equals(
				RTArray{},
				base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s),
			)
	})
}

func TestStream_ReadRTMap(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				rtMap, err := stream.ReadRTMap(testRuntime)
				assert(err).IsNil()
				assert(stream.GetWritePos()).
					Equals(len(testData[1].([]byte)) + i)
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.writeRTMap(rtMap)
				assert(stream.ReadMap()).Equals(testData[0], nil)
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (outer stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := NewStream()
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTMap(testRuntime)).
						Equals(RTMap{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Release()
			}
		}
	})

	t.Run("test readIndex overflow (runtime stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 20, 20, 61)
		for _, testData := range streamTestSuccessCollections["map"] {
			for _, i := range testRange {
				testRuntime.thread.Reset()
				stream := testRuntime.thread.rtStream
				stream.SetWritePos(i)
				stream.SetReadPos(i)
				stream.Write(testData[0])
				writePos := stream.GetWritePos()
				for idx := i; idx < writePos-1; idx++ {
					stream.SetReadPos(i)
					stream.SetWritePos(idx)
					assert(stream.ReadRTMap(testRuntime)).
						Equals(RTMap{}, base.ErrStream)
					assert(stream.GetReadPos()).Equals(i)
				}
				stream.Reset()
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadRTMap(testRuntime)).
				Equals(RTMap{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x61, 0x0A, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x02, 0x02,
			})
			assert(stream.ReadRTMap(testRuntime)).
				Equals(RTMap{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x61, 0x09, 0x00, 0x00, 0x00, 0x81, 0x31, 0x00, 0x0D,
			})
			assert(stream.ReadRTMap(testRuntime)).
				Equals(RTMap{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("error in stream", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{
				0x61, 0x0B, 0x00, 0x00, 0x00,
				0x81, 0x31, 0x00, 0x82, 0x61, 0x00,
			})
			assert(stream.ReadRTMap(testRuntime)).
				Equals(RTMap{}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		type R = Runtime
		s := ""
		f := base.GetFileLine
		assert(stream.ReadRTMap((func() R { s = f(0); return R{} })())).
			Equals(RTMap{}, base.ErrRuntimeIllegalInCurrentGoroutine.AddDebug(s))
	})
}

func TestStream_ReadRTValue(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for key := range streamTestSuccessCollections {
			for _, testData := range streamTestSuccessCollections[key] {
				for _, i := range testRange {
					testRuntime.thread.Reset()
					stream := NewStream()
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					stream.Write(testData[0])
					rtValue, _ := stream.ReadRTValue(testRuntime)
					switch testData[0].(type) {
					case string:
						assert(string(rtValue.cacheBytes), rtValue.cacheError).
							Equals(testData[0], nil)
					}
					assert(stream.GetWritePos()).
						Equals(len(testData[1].([]byte)) + i)
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					stream.writeRTValue(rtValue)
					assert(stream.Read()).Equals(testData[0], nil)
					stream.Release()
				}
			}
		}
	})

	t.Run("test readIndex overflow (outer stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for key := range streamTestSuccessCollections {
			for _, testData := range streamTestSuccessCollections[key] {
				for _, i := range testRange {
					testRuntime.thread.Reset()
					stream := NewStream()
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					stream.Write(testData[0])
					writePos := stream.GetWritePos()
					for idx := i; idx < writePos-1; idx++ {
						stream.SetReadPos(i)
						stream.SetWritePos(idx)
						assert(stream.ReadRTValue(testRuntime)).
							Equals(RTValue{err: base.ErrStream}, base.ErrStream)
						assert(stream.GetReadPos()).Equals(i)
					}
					stream.Release()
				}
			}
		}
	})

	t.Run("test readIndex overflow (runtime stream)", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 20, 20, 61)
		for key := range streamTestSuccessCollections {
			for _, testData := range streamTestSuccessCollections[key] {
				for _, i := range testRange {
					testRuntime.thread.Reset()
					stream := testRuntime.thread.rtStream
					stream.SetWritePos(i)
					stream.SetReadPos(i)
					stream.Write(testData[0])
					writePos := stream.GetWritePos()
					for idx := i; idx < writePos-1; idx++ {
						stream.SetReadPos(i)
						stream.SetWritePos(idx)
						switch testData[0].(type) {
						case string:
							rtValue, _ := stream.ReadRTValue(testRuntime)
							assert(rtValue.ToString()).
								Equals("", base.ErrStream)
						default:
							assert(stream.ReadRTValue(testRuntime)).Equals(
								RTValue{err: base.ErrStream}, base.ErrStream,
							)
						}
						assert(stream.GetReadPos()).Equals(i)
					}
					stream.Reset()
				}
			}
		}
	})

	t.Run("test type not match", func(t *testing.T) {
		assert := base.NewAssert(t)
		testRange := getTestRange(streamPosBody, 3*streamBlockSize, 80, 80, 61)
		for _, i := range testRange {
			testRuntime.thread.Reset()
			stream := NewStream()
			stream.SetWritePos(i)
			stream.SetReadPos(i)
			stream.PutBytes([]byte{13})
			assert(stream.ReadRTValue(testRuntime)).
				Equals(RTValue{err: base.ErrStream}, base.ErrStream)
			assert(stream.GetReadPos()).Equals(i)
			stream.Release()
		}
	})

	t.Run("runtime is not available", func(t *testing.T) {
		assert := base.NewAssert(t)
		stream := NewStream()
		assert(stream.ReadRTValue(Runtime{})).
			Equals(
				RTValue{err: base.ErrRuntimeIllegalInCurrentGoroutine},
				base.ErrRuntimeIllegalInCurrentGoroutine,
			)
	})
}

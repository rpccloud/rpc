package base

import (
	"strings"
	"testing"
	"unsafe"
)

func TestIsNil(t *testing.T) {
	assert := NewAssert(t)
	assert(IsNil(nil)).IsTrue()
	assert(IsNil(t)).IsFalse()
	assert(IsNil(3)).IsFalse()
	assert(IsNil(0)).IsFalse()
	assert(IsNil(uintptr(0))).IsFalse()
	assert(IsNil(uintptr(1))).IsFalse()
	assert(IsNil(unsafe.Pointer(nil))).IsTrue()
	assert(IsNil(unsafe.Pointer(t))).IsFalse()
}

func TestIsUTF8Bytes(t *testing.T) {
	assert := NewAssert(t)

	assert(IsUTF8Bytes(([]byte)("abc"))).IsTrue()
	assert(IsUTF8Bytes(([]byte)("abc！#@¥#%#%#¥%"))).IsTrue()
	assert(IsUTF8Bytes(([]byte)("中文"))).IsTrue()
	assert(IsUTF8Bytes(([]byte)("🀄️文👃d"))).IsTrue()
	assert(IsUTF8Bytes(([]byte)("🀄️文👃"))).IsTrue()

	assert(IsUTF8Bytes(([]byte)(`
    😀 😁 😂 🤣 😃 😄 😅 😆 😉 😊 😋 😎 😍 😘 🥰 😗 😙 😚 ☺️ 🙂 🤗 🤩 🤔 🤨
    🙄 😏 😣 😥 😮 🤐 😯 😪 😫 😴 😌 😛 😜 😝 🤤 😒 😓 😔 😕 🙃 🤑 😲 ☹️ 🙁
    😤 😢 😭 😦 😧 😨 😩 🤯 😬 😰 😱 🥵 🥶 😳 🤪 😵 😡 😠 🤬 😷 🤒 🤕 🤢
    🤡 🥳 🥴 🥺 🤥 🤫 🤭 🧐 🤓 😈 👿 👹 👺 💀 👻 👽 🤖 💩 😺 😸 😹 😻 😼 😽
    👶 👧 🧒 👦 👩 🧑 👨 👵 🧓 👴 👲 👳‍♀️ 👳‍♂️ 🧕 🧔 👱‍♂️ 👱‍♀️ 👨‍🦰 👩‍🦰 👨‍🦱 👩‍🦱 👨‍🦲 👩‍🦲 👨‍🦳
    👩‍🦳 🦸‍♀️ 🦸‍♂️ 🦹‍♀️ 🦹‍♂️ 👮‍♀️ 👮‍♂️ 👷‍♀️ 👷‍♂️ 💂‍♀️ 💂‍♂️ 🕵️‍♀️ 🕵️‍♂️ 👩‍⚕️ 👨‍⚕️ 👩‍🌾 👨‍🌾 👩‍🍳
    👨‍🍳 👩‍🎓 👨‍🎓 👩‍🎤 👨‍🎤 👩‍🏫 👨‍🏫 👩‍🏭 👨‍🏭 👩‍💻 👨‍💻 👩‍💼 👨‍💼 👩‍🔧 👨‍🔧 👩‍🔬 👨‍🔬 👩‍🎨 👨‍🎨 👩‍🚒 👨‍🚒 👩‍✈️ 👨‍✈️ 👩‍🚀
    👩‍⚖️ 👨‍⚖️ 👰 🤵 👸 🤴 🤶 🎅 🧙‍♀️ 🧙‍♂️ 🧝‍♀️ 🧝‍♂️ 🧛‍♀️ 🧛‍♂️ 🧟‍♀️ 🧟‍♂️ 🧞‍♀️ 🧞‍♂️ 🧜‍♀️
    🧜‍♂️ 🧚‍♀️ 🧚‍♂️ 👼 🤰 🤱 🙇‍♀️ 🙇‍♂️ 💁‍♀️ 💁‍♂️ 🙅‍♀️ 🙅‍♂️ 🙆‍♀️ 🙆‍♂️ 🙋‍♀️ 🙋‍♂️ 🤦‍♀️ 🤦‍♂️
    🤷‍♀️ 🤷‍♂️ 🙎‍♀️ 🙎‍♂️ 🙍‍♀️ 🙍‍♂️ 💇‍♀️ 💇‍♂️ 💆‍♀️ 💆‍♂️ 🧖‍♀️ 🧖‍♂️ 💅 🤳 💃 🕺 👯‍♀️ 👯‍♂️
    🕴 🚶‍♀️ 🚶‍♂️ 🏃‍♀️ 🏃‍♂️ 👫 👭 👬 💑 👩‍❤️‍👩 👨‍❤️‍👨 💏 👩‍❤️‍💋‍👩 👨‍❤️‍💋‍👨 👪 👨‍👩‍👧 👨‍👩‍👧‍👦 👨‍👩‍👦‍👦
    👨‍👩‍👧‍👧 👩‍👩‍👦 👩‍👩‍👧 👩‍👩‍👧‍👦 👩‍👩‍👦‍👦 👩‍👩‍👧‍👧 👨‍👨‍👦 👨‍👨‍👧 👨‍👨‍👧‍👦 👨‍👨‍👦‍👦 👨‍👨‍👧‍👧 👩‍👦 👩‍👧 👩‍👧‍👦 👩‍👦‍👦 👩‍👧‍👧 👨‍👦 👨‍👧 👨‍👧‍👦 👨‍👦‍👦 👨‍👧‍👧 🤲 👐
    👎 👊 ✊ 🤛 🤜 🤞 ✌️ 🤟 🤘 👌 👈 👉 👆 👇 ☝️ ✋ 🤚 🖐 🖖 👋 🤙 💪 🦵 🦶
    💍 💄 💋 👄 👅 👂 👃 👣 👁 👀 🧠 🦴 🦷 🗣 👤 👥
  `))).IsTrue()

	assert(IsUTF8Bytes([]byte{0xC1})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xC1, 0x01})).IsFalse()

	assert(IsUTF8Bytes([]byte{0xE1, 0x80})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xE1, 0x01, 0x81})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xE1, 0x80, 0x01})).IsFalse()

	assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x80})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xF1, 0x70, 0x80, 0x80})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x70, 0x80})).IsFalse()
	assert(IsUTF8Bytes([]byte{0xF1, 0x80, 0x80, 0x70})).IsFalse()

	assert(IsUTF8Bytes([]byte{0xFF, 0x80, 0x80, 0x70})).IsFalse()
}

func TestGetRandString(t *testing.T) {
	assert := NewAssert(t)
	assert(GetRandString(-1)).Equal("")
	for i := 0; i < 100; i++ {
		assert(len(GetRandString(i))).Equal(i)
	}
}
func TestGetSeed(t *testing.T) {
	assert := NewAssert(t)
	seed := GetSeed()
	assert(seed > 10000).IsTrue()

	for i := int64(0); i < 1000; i++ {
		assert(GetSeed()).Equal(seed + 1 + i)
	}
}

func TestAddPrefixPerLine(t *testing.T) {
	assert := NewAssert(t)

	assert(AddPrefixPerLine("", "")).Equal("")
	assert(AddPrefixPerLine("a", "")).Equal("a")
	assert(AddPrefixPerLine("\n", "")).Equal("\n")
	assert(AddPrefixPerLine("a\n", "")).Equal("a\n")
	assert(AddPrefixPerLine("a\nb", "")).Equal("a\nb")
	assert(AddPrefixPerLine("", "-")).Equal("-")
	assert(AddPrefixPerLine("a", "-")).Equal("-a")
	assert(AddPrefixPerLine("\n", "-")).Equal("-\n")
	assert(AddPrefixPerLine("a\n", "-")).Equal("-a\n")
	assert(AddPrefixPerLine("a\nb", "-")).Equal("-a\n-b")
}

func TestConcatString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConcatString("", "")).Equal("")
	assert(ConcatString("a", "")).Equal("a")
	assert(ConcatString("", "b")).Equal("b")
	assert(ConcatString("a", "b")).Equal("ab")
	assert(ConcatString("a", "b", "")).Equal("ab")
	assert(ConcatString("a", "b", "c")).Equal("abc")
}

func TestConvertOrdinalToString(t *testing.T) {
	assert := NewAssert(t)

	assert(ConvertOrdinalToString(0)).Equal("")
	assert(ConvertOrdinalToString(1)).Equal("1st")
	assert(ConvertOrdinalToString(2)).Equal("2nd")
	assert(ConvertOrdinalToString(3)).Equal("3rd")
	assert(ConvertOrdinalToString(4)).Equal("4th")
	assert(ConvertOrdinalToString(10)).Equal("10th")
	assert(ConvertOrdinalToString(100)).Equal("100th")
}

func TestAddFileLine(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	fileLine1 := AddFileLine("header", 0)
	assert(strings.HasPrefix(fileLine1, "header ")).IsTrue()
	assert(strings.Contains(fileLine1, "util_test.go")).IsTrue()

	// Test(2)
	fileLine2 := AddFileLine("", 0)
	assert(strings.HasPrefix(fileLine2, " ")).IsFalse()
	assert(strings.Contains(fileLine2, "util_test.go")).IsTrue()

	// Test(3)
	assert(AddFileLine("header", 1000)).Equal("header")
}

func TestGetFileLine(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	fileLine1 := GetFileLine(0)
	assert(strings.Contains(fileLine1, "util_test.go")).IsTrue()
}

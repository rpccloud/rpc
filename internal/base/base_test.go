package base

import (
	"crypto/tls"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestInitAESCipherAndNonce(t *testing.T) {
	t.Run("test basic", func(t *testing.T) {
		assert := NewAssert(t)
		assert(aesCipher).IsNotNil()
		assert(len(aesNonce)).Equals(12)
	})

	t.Run("get keyBuffer error", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {
			saveFNGetRandBytes := fnGetRandBytes
			fnGetRandBytes = func(_ uint32) ([]byte, error) {
				return nil, io.EOF
			}
			defer func() {
				fnGetRandBytes = saveFNGetRandBytes
			}()

			initAESCipherAndNonce()
		})).Equals(io.EOF.Error())
	})

	t.Run("get nonceBuffer error", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {
			saveFNGetRandBytes := fnGetRandBytes
			fnGetRandBytes = func(n uint32) ([]byte, error) {
				if n == 12 {
					return nil, io.EOF
				}

				return saveFNGetRandBytes(n)
			}
			defer func() {
				fnGetRandBytes = saveFNGetRandBytes
			}()

			initAESCipherAndNonce()
		})).Equals(io.EOF.Error())
	})

	t.Run("block size error", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {
			saveFNGetRandBytes := fnGetRandBytes
			fnGetRandBytes = func(n uint32) ([]byte, error) {
				if n == 32 {
					return []byte{1, 2, 3, 4, 5, 6, 7}, nil
				}

				return saveFNGetRandBytes(n)
			}
			defer func() {
				fnGetRandBytes = saveFNGetRandBytes
			}()

			initAESCipherAndNonce()
		})).IsNotNil()
	})

	t.Run("NonceSize error", func(t *testing.T) {
		assert := NewAssert(t)
		assert(RunWithCatchPanic(func() {
			saveFNGetRandBytes := fnGetRandBytes
			fnGetRandBytes = func(n uint32) ([]byte, error) {
				if n == 12 {
					return make([]byte, 0), nil
				}

				return saveFNGetRandBytes(n)
			}
			defer func() {
				fnGetRandBytes = saveFNGetRandBytes
			}()

			initAESCipherAndNonce()
		})).IsNotNil()
	})

}

func TestIsNil(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsNil(nil)).IsTrue()
		assert(IsNil(t)).IsFalse()
		assert(IsNil(3)).IsFalse()
		assert(IsNil(0)).IsFalse()
		assert(IsNil(uintptr(0))).IsFalse()
		assert(IsNil(uintptr(1))).IsFalse()
		assert(IsNil(unsafe.Pointer(nil))).IsTrue()
		assert(IsNil(unsafe.Pointer(t))).IsFalse()
	})
}

func TestMinInt(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(MinInt(1, 2)).Equals(1)
		assert(MinInt(2, 1)).Equals(1)
	})
}

func TestMaxInt(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(MaxInt(1, 2)).Equals(2)
		assert(MaxInt(2, 1)).Equals(2)
	})
}

func TestStringToBytesUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(cap(StringToBytesUnsafe("hello"))).Equals(5)
		assert(len(StringToBytesUnsafe("hello"))).Equals(5)
		assert(string(StringToBytesUnsafe("hello"))).Equals("hello")
	})
}

func TestBytesToStringUnsafe(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(len(BytesToStringUnsafe([]byte("hello")))).Equals(5)
		assert(BytesToStringUnsafe([]byte("hello"))).Equals("hello")
	})
}

func TestGetSeed(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		seed := GetSeed()
		assert(seed > 10000).IsTrue()
		for i := int64(0); i < 500; i++ {
			assert(GetSeed()).Equals(seed + 1 + i)
		}
	})
}

func TestGetRandString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(GetRandString(-1)).Equals("")
		for i := 0; i < 500; i++ {
			assert(len(GetRandString(i))).Equals(i)
		}
	})
}

func TestAddPrefixPerLine(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(AddPrefixPerLine("", "")).Equals("")
		assert(AddPrefixPerLine("a", "")).Equals("a")
		assert(AddPrefixPerLine("\n", "")).Equals("\n")
		assert(AddPrefixPerLine("a\n", "")).Equals("a\n")
		assert(AddPrefixPerLine("a\nb", "")).Equals("a\nb")
		assert(AddPrefixPerLine("", "-")).Equals("-")
		assert(AddPrefixPerLine("a", "-")).Equals("-a")
		assert(AddPrefixPerLine("\n", "-")).Equals("-\n")
		assert(AddPrefixPerLine("a\n", "-")).Equals("-a\n")
		assert(AddPrefixPerLine("a\nb", "-")).Equals("-a\n-b")
	})
}

func TestConcatString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(ConcatString("", "")).Equals("")
		assert(ConcatString("a", "")).Equals("a")
		assert(ConcatString("", "b")).Equals("b")
		assert(ConcatString("a", "b")).Equals("ab")
		assert(ConcatString("a", "b", "")).Equals("ab")
		assert(ConcatString("a", "b", "c")).Equals("abc")
	})
}

func TestGetFileLine(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		fileLine1 := GetFileLine(0)
		assert(strings.Contains(fileLine1, "base_test.go")).IsTrue()
	})
}

func TestAddFileLine(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := AddFileLine("", 0)
		assert(strings.HasPrefix(v1, " ")).IsFalse()
		assert(strings.Contains(v1, "base_test.go")).IsTrue()
	})

	t.Run("skip overflow", func(t *testing.T) {
		assert := NewAssert(t)
		assert(AddFileLine("header", 1000)).Equals("header")
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		v1 := AddFileLine("header", 0)
		assert(strings.HasPrefix(v1, "header ")).IsTrue()
		assert(strings.Contains(v1, "base_test.go")).IsTrue()
	})
}

func TestConvertOrdinalToString(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		assert(ConvertOrdinalToString(0)).Equals("")
		assert(ConvertOrdinalToString(1)).Equals("1st")
		assert(ConvertOrdinalToString(2)).Equals("2nd")
		assert(ConvertOrdinalToString(3)).Equals("3rd")
		assert(ConvertOrdinalToString(4)).Equals("4th")
		assert(ConvertOrdinalToString(10)).Equals("10th")
		assert(ConvertOrdinalToString(100)).Equals("100th")
	})
}

func TestWaitWhenRunning(t *testing.T) {
	t.Run("test isRunning return true", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				assert := NewAssert(t)
				startTime := TimeNow()
				WaitWhileRunning(
					TimeNow().UnixNano(),
					func() bool { return true },
					500*time.Millisecond,
				)
				interval := TimeNow().Sub(startTime)
				assert(interval > 480*time.Millisecond).IsTrue()
				assert(interval < 880*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
	})

	t.Run("test isRunning return false 1", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				assert := NewAssert(t)
				startTime := TimeNow()
				WaitWhileRunning(
					TimeNow().UnixNano(),
					func() bool { return false },
					500*time.Millisecond,
				)
				interval := TimeNow().Sub(startTime)
				assert(interval > -180*time.Millisecond).IsTrue()
				assert(interval < 180*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
	})

	t.Run("test isRunning return false 2", func(t *testing.T) {
		waitCH := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				assert := NewAssert(t)
				count := 0
				startTime := TimeNow()
				WaitWhileRunning(
					startTime.UnixNano(),
					func() bool {
						count++
						return count < 3
					},
					500*time.Millisecond,
				)
				interval := TimeNow().Sub(startTime)
				assert(interval >= 40*time.Millisecond).IsTrue()
				assert(interval < 300*time.Millisecond).IsTrue()
				waitCH <- true
			}()
		}
		for i := 0; i < 100; i++ {
			<-waitCH
		}
	})
}

func TestIsTCPPortOccupied(t *testing.T) {
	t.Run("not occupied", func(t *testing.T) {
		assert := NewAssert(t)
		assert(IsTCPPortOccupied(65535)).Equals(false)
	})

	t.Run("occupied", func(t *testing.T) {
		assert := NewAssert(t)
		Listener, _ := net.Listen("tcp", "127.0.0.1:65535")
		assert(IsTCPPortOccupied(65535)).Equals(true)
		_ = Listener.Close()
	})
}

func TestReadFromFile(t *testing.T) {
	t.Run("file not exist", func(t *testing.T) {
		assert := NewAssert(t)
		v1, err1 := ReadFromFile("./no_file")
		assert(v1).Equals("")
		assert(err1).IsNotNil()
		assert(strings.Contains(err1.Error(), "no_file")).IsTrue()
	})

	t.Run("file exist", func(t *testing.T) {
		assert := NewAssert(t)
		_ = ioutil.WriteFile("./tmp_file", []byte("hello"), 0666)
		assert(ReadFromFile("./tmp_file")).Equals("hello", nil)
		_ = os.Remove("./tmp_file")
	})
}

func TestGetTLSServerConfig(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("cert or key error", func(t *testing.T) {
		assert := NewAssert(t)
		ret, e := GetTLSServerConfig(
			path.Join(curDir, "_cert_", "error.crt"),
			path.Join(curDir, "_cert_", "error.key"),
		)
		assert(ret).IsNil()
		assert(e).IsNotNil()
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		ret, e := GetTLSServerConfig(
			path.Join(curDir, "_cert_", "test.crt"),
			path.Join(curDir, "_cert_", "test.key"),
		)
		cert, _ := tls.LoadX509KeyPair(
			path.Join(curDir, "_cert_", "test.crt"),
			path.Join(curDir, "_cert_", "test.key"),
		)
		assert(ret).IsNotNil()
		assert(ret).Equals(&tls.Config{
			Certificates: []tls.Certificate{cert},
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519, // Go 1.8 only
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		})
		assert(e).IsNil()
	})
}

func TestGetTLSClientConfig(t *testing.T) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	t.Run("ca error 01", func(t *testing.T) {
		assert := NewAssert(t)
		ret, e := GetTLSClientConfig(
			true,
			[]string{path.Join(curDir, "_cert_", "not_exist.ca")},
		)
		assert(ret).IsNil()
		assert(e).IsNotNil()
	})

	t.Run("ca error 02", func(t *testing.T) {
		assert := NewAssert(t)
		ret, e := GetTLSClientConfig(
			true,
			[]string{
				path.Join(curDir, "_cert_", "ca.crt"),
				path.Join(curDir, "_cert_", "test.crt"),
				path.Join(curDir, "_cert_", "error.crt"),
			},
		)

		assert(ret).IsNil()
		assert(e).IsNotNil()
		if e != nil {
			assert(strings.HasSuffix(
				e.Error(),
				"error.crt is not a valid certificate",
			)).IsTrue()
		}
	})

	t.Run("test", func(t *testing.T) {
		assert := NewAssert(t)
		ret, e := GetTLSClientConfig(
			true,
			[]string{
				path.Join(curDir, "_cert_", "ca.crt"),
				path.Join(curDir, "_cert_", "test.crt"),
			},
		)
		assert(ret).IsNotNil()
		if ret != nil {
			assert(ret.InsecureSkipVerify).IsFalse()
			assert(len(ret.RootCAs.Subjects())).Equals(2)
		}
		assert(e).IsNil()
	})
}

func TestEncryptSessionEndpoint(t *testing.T) {
	t.Run("aesCipher == nil", func(t *testing.T) {
		saveCipher := aesCipher
		aesCipher = nil
		defer func() {
			aesCipher = saveCipher
		}()

		assert := NewAssert(t)
		assert(EncryptSessionEndpoint(1, 32)).Equals("", false)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := NewAssert(t)
		v, ok := EncryptSessionEndpoint(1, 32)
		assert(len(v) > 0).IsTrue()
		assert(ok).IsTrue()
		assert(DecryptSessionEndpoint(v)).Equals(uint64(1), uint64(32), true)
	})
}

func TestDecryptSessionEndpoint(t *testing.T) {
	t.Run("aesCipher == nil", func(t *testing.T) {
		saveCipher := aesCipher
		aesCipher = nil
		defer func() {
			aesCipher = saveCipher
		}()

		assert := NewAssert(t)
		assert(DecryptSessionEndpoint("")).Equals(uint64(0), uint64(0), false)
	})

	t.Run("base64 decode error", func(t *testing.T) {
		assert := NewAssert(t)
		assert(DecryptSessionEndpoint("err")).Equals(uint64(0), uint64(0), false)
	})

	t.Run("cipher open error", func(t *testing.T) {
		assert := NewAssert(t)
		errString := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
		assert(DecryptSessionEndpoint(errString)).
			Equals(uint64(0), uint64(0), false)
	})

	t.Run("split arr length error", func(t *testing.T) {
		assert := NewAssert(t)
		errString := base64.StdEncoding.EncodeToString(aesCipher.Seal(
			nil,
			aesNonce,
			[]byte("ho-la-la"),
			nil,
		))
		assert(DecryptSessionEndpoint(errString)).
			Equals(uint64(0), uint64(0), false)
	})

	t.Run("gatewayID parse error", func(t *testing.T) {
		assert := NewAssert(t)
		errString := base64.StdEncoding.EncodeToString(aesCipher.Seal(
			nil,
			aesNonce,
			[]byte("9876543210-32"),
			nil,
		))
		assert(DecryptSessionEndpoint(errString)).
			Equals(uint64(0), uint64(0), false)
	})

	t.Run("sessionID parse error", func(t *testing.T) {
		assert := NewAssert(t)
		errString := base64.StdEncoding.EncodeToString(aesCipher.Seal(
			nil,
			aesNonce,
			[]byte("123-32.5"),
			nil,
		))
		assert(DecryptSessionEndpoint(errString)).
			Equals(uint64(0), uint64(0), false)
	})

	t.Run("test ok", func(t *testing.T) {
		assert := NewAssert(t)
		errString := base64.StdEncoding.EncodeToString(aesCipher.Seal(
			nil,
			aesNonce,
			[]byte("321-123456"),
			nil,
		))
		assert(DecryptSessionEndpoint(errString)).
			Equals(uint64(321), uint64(123456), true)
	})
}

func TestRunWithLogOutput(t *testing.T) {
	t.Run("test CaptureLogOutput", func(t *testing.T) {
		assert := NewAssert(t)
		assert(strings.HasSuffix(
			RunWithLogOutput(func() {
				log.Print("Hello world")
			}),
			"Hello world\n",
		)).IsTrue()
	})
}

func BenchmarkGetFileLine(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			AddFileLine("hello", 1)
		}
	})
}

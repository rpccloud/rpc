// Package base ...
package base

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mathRand "math/rand"
	"net"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// ErrNetClosingSuffix ...
const ErrNetClosingSuffix = "use of closed network connection"

var (
	seedInt64    = int64(10000)
	base64String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"+/"
	fnGetRandBytes = func(n uint32) ([]byte, error) {
		ret := make([]byte, n)
		_, e := io.ReadFull(cryptoRand.Reader, ret)
		return ret, e
	}
	aesCipher, aesNonce = initAESCipherAndNonce()
)

func initAESCipherAndNonce() (cipher.AEAD, []byte) {
	if keyBuffer, e := fnGetRandBytes(32); e != nil {
		panic(e.Error())
	} else if nonceBuffer, e := fnGetRandBytes(12); e != nil {
		panic(e.Error())
	} else if block, e := aes.NewCipher(keyBuffer); e != nil {
		panic(e.Error())
	} else if gcmCipher, e := cipher.NewGCMWithNonceSize(
		block, len(nonceBuffer),
	); e != nil {
		panic(e.Error())
	} else {
		return gcmCipher, nonceBuffer
	}
}

// IsNil ...
func IsNil(val interface{}) (ret bool) {
	defer func() {
		if e := recover(); e != nil {
			ret = false
		}
	}()

	if val == nil {
		return true
	}

	return reflect.ValueOf(val).IsNil()
}

// MinInt64 ...
func MinInt64(v1 int64, v2 int64) int64 {
	if v1 < v2 {
		return v1
	}

	return v2
}

// MaxInt64 ...
func MaxInt64(v1 int64, v2 int64) int64 {
	if v1 < v2 {
		return v2
	}

	return v1
}

// MinInt ...
func MinInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v1
	}

	return v2
}

// MaxInt ...
func MaxInt(v1 int, v2 int) int {
	if v1 < v2 {
		return v2
	}

	return v1
}

// StringToBytesUnsafe ...
func StringToBytesUnsafe(s string) (ret []byte) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bytesHeader.Len = stringHeader.Len
	bytesHeader.Cap = stringHeader.Len
	bytesHeader.Data = stringHeader.Data
	return
}

// BytesToStringUnsafe ...
func BytesToStringUnsafe(bytes []byte) (ret string) {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&ret))
	stringHeader.Len = bytesHeader.Len
	stringHeader.Data = bytesHeader.Data
	return
}

// GetSeed get int64 seed, it is goroutine safety
func GetSeed() int64 {
	return atomic.AddInt64(&seedInt64, 1)
}

// GetRandString get random string
func GetRandString(strLen int) string {
	sb := NewStringBuilder()
	defer sb.Release()

	for strLen > 0 {
		rand64 := mathRand.Uint64()
		for used := 0; used < 10 && strLen > 0; used++ {
			sb.AppendByte(base64String[rand64%64])
			rand64 = rand64 / 64
			strLen--
		}
	}
	return sb.String()
}

// AddPrefixPerLine ...
func AddPrefixPerLine(text string, prefix string) string {
	sb := NewStringBuilder()
	defer sb.Release()

	first := true
	array := strings.Split(text, "\n")
	for idx, v := range array {
		if first {
			first = false
		} else {
			sb.AppendByte('\n')
		}

		if v != "" || idx == 0 || idx != len(array)-1 {
			sb.AppendString(prefix)
			sb.AppendString(v)
		}
	}
	return sb.String()
}

// ConcatString ...
func ConcatString(args ...string) string {
	sb := NewStringBuilder()
	defer sb.Release()
	for _, v := range args {
		sb.AppendString(v)
	}
	return sb.String()
}

// GetFileLine ...
func GetFileLine(skip uint) string {
	return AddFileLine("", skip+1)
}

// AddFileLine ...
func AddFileLine(header string, skip uint) string {
	sb := NewStringBuilder()
	defer sb.Release()

	if _, file, line, ok := runtime.Caller(int(skip) + 1); ok && line > 0 {
		if header != "" {
			sb.AppendString(header)
			sb.AppendByte(' ')
		}

		sb.AppendString(file)
		sb.AppendByte(':')
		sb.AppendString(strconv.Itoa(line))
	} else {
		sb.AppendString(header)
	}

	return sb.String()
}

// ConvertOrdinalToString ...
func ConvertOrdinalToString(n uint) string {
	if n == 0 {
		return ""
	}

	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return strconv.Itoa(int(n)) + "th"
	}
}

// IsTCPPortOccupied ...
func IsTCPPortOccupied(port uint16) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 0)
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	return err == nil
}

// ReadFromFile ...
func ReadFromFile(filePath string) (string, error) {
	ret, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// for windows, remove \r
	return strings.Replace(string(ret), "\r", "", -1), nil
}

// WaitWhileRunning ...
func WaitWhileRunning(
	startNS int64,
	isRunning func() bool,
	duration time.Duration,
) {
	for isRunning() {
		remainNS := startNS + int64(duration) - TimeNow().UnixNano()

		if remainNS <= 0 {
			return
		}

		time.Sleep(
			time.Duration(MinInt64(remainNS, int64(30*time.Millisecond))),
		)
	}
}

// GetServerTLSConfig ...
func GetServerTLSConfig(certFile string, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Causes servers to use Go's default cipher suite preferences,
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

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// GetClientTLSConfig ...
func GetClientTLSConfig(
	verifyServerCert bool,
	caFiles []string,
) (*tls.Config, error) {
	caPool := (*x509.CertPool)(nil)

	if len(caFiles) > 0 {
		caPool = x509.NewCertPool()

		for _, caFile := range caFiles {
			if caCert, e := ioutil.ReadFile(caFile); e == nil {
				if !caPool.AppendCertsFromPEM(caCert) {
					return nil, fmt.Errorf(
						"%s is not a valid certificate",
						caFile,
					)
				}
			} else {
				return nil, e
			}
		}
	}

	return &tls.Config{
		InsecureSkipVerify: !verifyServerCert,
		RootCAs:            caPool,
	}, nil
}

// EncryptSessionEndpoint ...
func EncryptSessionEndpoint(gatewayID uint64, sessionID uint64) (string, bool) {
	if aesCipher != nil {
		return base64.StdEncoding.EncodeToString(aesCipher.Seal(
			nil,
			aesNonce,
			[]byte(fmt.Sprintf("%d-%d", gatewayID, sessionID)),
			nil,
		)), true
	}

	return "", false
}

// DecryptSessionEndpoint ...
func DecryptSessionEndpoint(sessionEndpoint string) (uint64, uint64, bool) {
	if aesCipher == nil {
		return 0, 0, false
	} else if sessionBuf, e := base64.StdEncoding.DecodeString(
		sessionEndpoint,
	); e != nil {
		return 0, 0, false
	} else if sessionBytes, e := aesCipher.Open(
		nil, aesNonce, sessionBuf, nil,
	); e != nil {
		return 0, 0, false
	} else if sArr := strings.Split(string(sessionBytes), "-"); len(sArr) != 2 {
		return 0, 0, false
	} else if gatewayID, e := strconv.ParseUint(sArr[0], 10, 32); e != nil {
		return 0, 0, false
	} else if sessionID, e := strconv.ParseUint(sArr[1], 10, 64); e != nil {
		return 0, 0, false
	} else {
		return gatewayID, sessionID, true
	}
}

// RunWithLogOutput ...
func RunWithLogOutput(fn func()) string {
	r, w, _ := os.Pipe()
	defer func() {
		log.SetOutput(os.Stderr)
		_ = r.Close()
	}()
	log.SetOutput(w)
	fn()
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

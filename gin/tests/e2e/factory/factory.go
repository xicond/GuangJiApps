package factory

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const digitCharset = "0123456789"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = charset[i%len(charset)]
		} else {
			b[i] = charset[num.Int64()]
		}
	}
	return string(b)
}

func RandDigits(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digitCharset))))
		if err != nil {
			b[i] = digitCharset[i%len(digitCharset)]
		} else {
			b[i] = digitCharset[num.Int64()]
		}
	}
	return string(b)
}

func RandCode(prefix string, n int) string {
	return fmt.Sprintf("%s%s", prefix, RandDigits(n))
}

func RandPhone() string {
	return fmt.Sprintf("08%s", RandDigits(9))
}

func RandEmail(prefix string) string {
	return fmt.Sprintf("%s_%s@guangji.test", prefix, RandString(6))
}

func UniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

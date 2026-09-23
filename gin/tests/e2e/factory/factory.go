package factory

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const digitCharset = "0123456789"

// RandString generates a cryptographically secure random alphanumeric string of length n.
//
// Parameters:
//   - n: length of the generated string.
//
// Returns:
//   - string: random alphanumeric string.
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

// RandDigits generates a cryptographically secure random numeric string of length n.
//
// Parameters:
//   - n: number of digits to generate.
//
// Returns:
//   - string: random numeric string.
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

// RandCode generates a formatted code combining a string prefix and a random digit sequence.
//
// Parameters:
//   - prefix: string prefix for the code.
//   - n: number of random digits to append.
//
// Returns:
//   - string: formatted code identifier.
func RandCode(prefix string, n int) string {
	return fmt.Sprintf("%s%s", prefix, RandDigits(n))
}

// RandPhone generates a randomized Indonesian mobile phone number prefixed with "08".
//
// Returns:
//   - string: randomized phone number string.
func RandPhone() string {
	return fmt.Sprintf("08%s", RandDigits(9))
}

// RandEmail generates a unique test email address with the provided prefix.
//
// Parameters:
//   - prefix: local-part identifier prefix.
//
// Returns:
//   - string: generated test email address.
func RandEmail(prefix string) string {
	return fmt.Sprintf("%s_%s@guangji.test", prefix, RandString(6))
}

// UniqueSuffix generates a pseudo-unique numeric suffix based on the current Unix nano timestamp.
//
// Returns:
//   - string: timestamp-based unique suffix string.
func UniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

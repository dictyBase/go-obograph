package generate

import (
	"crypto/rand"
)

func stringWithCharset(strSize int) string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ls := len(charSet)
	b := make([]byte, strSize)
	rand.Read(b)
	for k, v := range b {
		b[k] = charSet[v%byte(ls)]
	}
	return string(b)
}

// RandString generates a random string of a specific length
func RandString(length int) string {
	return stringWithCharset(length)
}

package generate

import (
	"crypto/rand"
	"fmt"
)

func stringWithCharset(strSize int) (string, error) {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ls := len(charSet)
	b := make([]byte, strSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error in reading random data %s", err)
	}
	for k, v := range b {
		b[k] = charSet[v%byte(ls)]
	}
	return string(b), nil
}

// RandString generates a random string of a specific length
func RandString(length int) (string, error) {
	return stringWithCharset(length)
}

package generate

import (
	"crypto/rand"
	"fmt"
)

func stringWithCharset(strSize int) (string, error) {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lsc := len(charSet)
	bys := make([]byte, strSize)
	_, err := rand.Read(bys)
	if err != nil {
		return "", fmt.Errorf("error in reading random data %s", err)
	}
	for k, v := range bys {
		bys[k] = charSet[v%byte(lsc)]
	}

	return string(bys), nil
}

// RandString generates a random string of a specific length.
func RandString(length int) (string, error) {
	return stringWithCharset(length)
}

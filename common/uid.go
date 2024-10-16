package common

import (
	"math/rand"
)

type UID [10]byte

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const charsetLen = len(charset)

func GenerateUID() string {

	randomBytes := make([]byte, 10)

	for i := 0; i < 10; i++ {
		randomBytes[i] = charset[rand.Intn(charsetLen)]
	}

	return string(randomBytes)
}

func ReadUID(uidstr string) UID {
	uidbytes := []byte(uidstr)
	var res [10]byte
	for i := 0; i < 10; i++ {
		res[i] = uidbytes[i]
	}
	return res
}

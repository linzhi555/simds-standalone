package common

import (
	"crypto/rand"
)

type UID [10]byte

func GenerateUID() string {

	// Generate a random number
	randomBytes := make([]byte, 10)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
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

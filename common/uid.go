package common

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"
)

func GenerateUID() string {

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate a random number
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	randomPart := fmt.Sprintf("%x", randomBytes)

	uid := strconv.FormatInt(timestamp, 10) + randomPart

	return uid

}

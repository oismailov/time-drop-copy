package helpers

import (
	"math/rand"
	"strconv"
	"time"
)

//GenerateOneTimeToken returns a random and fixed length int for login
func GenerateOneTimeToken() string {
	rand.Seed(time.Now().UTC().UnixNano())
	var (
		low  = 100000
		high = 999999
	)
	randToken := low + rand.Intn(high-low)
	return strconv.Itoa(randToken)
}

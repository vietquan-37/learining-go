package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const aplphabet = "qwertyuiopasdfghjklzxcvbnm"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// random int value between min and max in parameter
func RandomInit(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}
func RandomString(n int) string {
	var sb strings.Builder
	k := len(aplphabet)
	for i := 0; i < n; i++ {
		c := aplphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}
func RandomOwner() string {
	return RandomString(6)
}
func RandomMoney() int64 {
	return RandomInit(0, 1000)
}
func RandomCurrency() string {
	currency := []string{"EUR", "USD", "CAD"}
	n := len(currency)
	return currency[rand.Intn(n)]
}
func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(6))
}

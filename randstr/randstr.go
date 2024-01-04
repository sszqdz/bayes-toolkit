package randstr

import "math/rand"

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	size    = int64(len(letters))
)

func RandStr(n uint32) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%size]
	}
	return string(b)
}

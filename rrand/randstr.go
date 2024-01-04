// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rrand

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

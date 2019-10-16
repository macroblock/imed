package hash

import (
	"crypto/sha256"
	"fmt"
)

var digits = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

var hashSize = 10

func rolInt64(i int64) int64 {
	b := int64(0)
	if i < 0 {
		b = 1
	}
	i = (i << 1) | b
	return i
}

// String -
// func String(str string, num int) string {
// 	// x := int64(^uint64(0) >> 1)
// 	// fmt.Printf("0: %b\n", uint64(x))
// 	// x = rolInt64(x)
// 	// fmt.Printf("1: %b\n", uint64(x))
// 	// x = rolInt64(x)
// 	// fmt.Printf("2: %b\n", uint64(x))
// 	// x = rolInt64(x)
// 	// fmt.Printf("3: %b\n", uint64(x))

// 	seed := int64(0)
// 	bs := []byte(str)
// 	for _, b := range bs {
// 		seed = rolInt64(seed)
// 		seed ^= int64(b)
// 	}
// 	rand.Seed(seed)
// 	bs = make([]byte, num)
// 	for i := 0; i < num; i++ {
// 		bs[len(bs)-i-1] = digits[rand.Intn(len(digits))]
// 	}
// 	return string(bs)
// }

// SetSize -
func SetSize(size int) {
	if size > sha256.Size {
		panic(fmt.Sprintf("size can not be greater than %v\n", size))
	}
	if size < 1 {
		panic(fmt.Sprintf("size can not be lesser than %v\n", size))
	}
	hashSize = size
}

// Get -
func Get(str string) string {
	sha := sha256.Sum256([]byte(str))
	bs := make([]byte, hashSize)
	for i := 0; i < hashSize; i++ {
		d := int(sha[i]) % len(digits)
		if d < 0 {
			d += len(digits)
		}
		bs[len(bs)-i-1] = digits[d]
	}
	return string(bs)
}

// Key -
func Key(name, season, year, sdhd, comment string) string {
	return name + "_" + season + "_" + year + "_" + sdhd + "_" + comment
}

// func key(s string, size int) []bite {
// 	bs := []byte(str)
// 	for _, b := range bs {
// 		seed = rolInt64(seed)
// 		seed ^= int64(b)
// 	}
// }

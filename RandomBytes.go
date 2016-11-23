package aero

import (
	"math/rand"
	"time"
)

const (
	characterPool       = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	characterPoolLength = uint64(len(characterPool))
)

var primeNumbers = []int64{
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43,
	47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103,
	107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163,
	167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227,
	229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281,
	283, 293, 307, 311,
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomBytes can generate up to 64 random bytes.
func RandomBytes(length uint) []byte {
	b := make([]byte, length)

	// Every bit decides whether to include the indexed prime or not
	bits := rand.Int63()

	// We just want a pretty random start number. So why not the bitmask itself?
	offset := bits

	// Test bit
	testBit := int64(0)

	for i := uint(0); i < length; i++ {
		testBit = int64(1) << i

		// Test the bit
		if bits&testBit != 0 {
			offset += primeNumbers[i]
		} else {
			offset += testBit
		}

		b[i] = characterPool[uint64(offset)%characterPoolLength]
	}

	return b
}

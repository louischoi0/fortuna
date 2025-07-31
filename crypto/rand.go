package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

type CSPRNG struct {
	seed		[]byte
	counter 	uint64
	mu		sync.Mutex
}

func NewCSPRNG(seed []byte) *CSPRNG {
	return &CSPRNG{
		seed: seed,
		counter: 0,
	}
}

func (c *CSPRNG) NextInt64() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, c.counter)
	mac := hmac.New(sha256.New, c.seed)
	mac.Write(counterBytes)
	c.counter++

	digest := mac.Sum(nil)
	randomInt := int64(binary.BigEndian.Uint64(digest[:8]))
	return randomInt
}

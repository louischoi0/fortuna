package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

type CSPRNG struct {
	seed    []byte
	counter uint64
	mu      sync.Mutex
}

func NewCSPRNG(seed []byte) *CSPRNG {
	return &CSPRNG{
		seed:    seed,
		counter: 0,
	}
}

func (c *CSPRNG) NextInt64() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	counterBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(counterBytes, c.counter)
	mac := hmac.New(sha256.New, c.seed)
	mac.Write(counterBytes)
	c.counter++

	digest := mac.Sum(nil)
	randomInt := int64(binary.LittleEndian.Uint64(digest[:8]))
	return randomInt
}

type CSPRNGU struct {
	seed    	[]byte
	counter 	uint64
	mod 		uint64
	mu      	sync.Mutex
	sequences	map[uint64]int8
}

func NewCSPRNGU(seed []byte, mod uint64) *CSPRNGU {
	return &CSPRNGU{
		seed:    seed,
		counter: 0,
		mod:mod,
		sequences: make(map[uint64]int8),
	}
}

func (c *CSPRNGU) NextInt64() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	counterBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(counterBytes, c.counter)
	mac := hmac.New(sha256.New, c.seed)
	mac.Write(counterBytes)

	digest := mac.Sum(nil)
	randomInt := binary.LittleEndian.Uint64(digest[:8])

	if _, ok := c.sequences[randomInt]; ok {
		return c.NextInt64()
	}

	c.counter++
	return randomInt
}

package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"fortuna/crypto"
	"time"
)

type KernelVersion string

const (
	BaseV000 KernelVersion = "base:v.0.0"
)

type StateKernel interface {
	GenVector(seed string, size int64) []int64
	VerifyVector(seed string, vector []int64) bool
	RandInt(seed string, minValue int64, maxValue int64) int64
	GenIndex(seed string, minValue int64, maxValue int64, size int64) []int64

	GenStateSeed() (string, string)
	GenStateSeedPayload(payload string) string
	VerifySeed(seed string, payload string) bool

	Hash(data string) (string, error)
	// HashState(state []int64) string
}

func LoadKernel(version KernelVersion) StateKernel {
	switch version {
	case BaseV000:
		return &BasicStateKernel{}
	}
	return nil
}

type BasicStateKernel struct{}

func (ck *BasicStateKernel) GenStateSeedPayload(payload string) string {
	return ""
}

func (ck *BasicStateKernel) GenStateSeed() (string, string) {
	payload := string(time.Now().UnixNano())
	seed := ck.GenStateSeedPayload(payload)
	return seed, payload
}

func (ck *BasicStateKernel) GenVector(seed string, size int64) []int64 {
	res := make([]int64, size)

	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range size {
		res[i] = rng.NextInt64()
	}
	return res
}

func (ck *BasicStateKernel) VerifyVector(seed string, vec []int64) bool {
	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range len(vec) {
		expected := rng.NextInt64()

		if expected != vec[i] {
			return false
		}
	}
	return true
}

func (ck *BasicStateKernel) GenIndex(seed string, s int64, e int64, size int64) []int64 {
	res := make([]int64, size)

	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range size {
		v := rng.NextInt64()
		res[i] = int64((uint64(v) % uint64(e)) + uint64(s))
	}
	return res
}

func (ck *BasicStateKernel) RandInt(seed string, s int64, e int64) int64 {
	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)
	v := rng.NextInt64()

	return (v % e) + s
}

func (ck *BasicStateKernel) VerifySeed(seed string, payload string) bool {
	return true
}

func (hk *BasicStateKernel) Hash(input string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(input))

	hashedBytes := hash.Sum(nil)

	return hex.EncodeToString(hashedBytes), nil
}

func (hk *BasicStateKernel) HashState(state []int64) string {
	return ""
}

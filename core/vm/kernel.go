package vm

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"
)

type KernelVersion string

const (
	BaseV000 KernelVersion = "base-v0.0.0"
)

type StateKernel interface {
	GenVector(seed int64, size int64) []float64
	GenIndex(seed, size, minValue, maxValue int64) []int64
	RandInt(seed int64, minValue int64, maxValue int64) int64

	GenStateSeed() (int64, string)
	GetEventSeed(eventSpec *EventSpec) (int64, string)

	VerifyNodeStateSeed(seed int64, payload string) bool
	StringToSeed(payload string) (int64, error)

	Hash(data string) (string, error)
	HashState(state []float64) string
}

func LoadKernel(version KernelVersion) StateKernel {
	switch version {
	case BaseV000:
		return &BasicStateKernel{}
	}
	return nil
}

type BasicStateKernel struct{}

func (ck *BasicStateKernel) GenIndex(seed, size, minValue, maxValue int64) []int64 {
	indices := make([]int64, size)

	for i := range size {
		indices[i] = ck.RandInt(seed+i, minValue, maxValue)
	}
	return indices
}

func (ck *BasicStateKernel) GenStateSeed() (int64, string) {
	payload := string(time.Now().UnixNano())
	seed, _ := ck.StringToSeed(payload)
	return seed, payload
}

func (ck *BasicStateKernel) GenVector(seed int64, size int64) []float64 {
	vec := make([]float64, size)

	for i := range size {
		vec[i] = rand.Float64()
	}
	return vec
}

func (ck *BasicStateKernel) RandInt(seed int64, s int64, e int64) int64 {
	rand.Seed(seed)
	if s > e {
		s, e = e, s
	}
	return int64(rand.Intn(int(e-s+1))) + s
}

func (ck *BasicStateKernel) VerifyNodeStateSeed(seed int64, payload string) bool {
	return true
}

func (hk *BasicStateKernel) Hash(input string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(input))

	hashedBytes := hash.Sum(nil)

	return hex.EncodeToString(hashedBytes), nil
}

func (hk *BasicStateKernel) StringToSeed(payload string) (int64, error) {
	hasher := fnv.New64a()

	_, err := hasher.Write([]byte(payload))
	if err != nil {
		return 0, err
	}

	hashValue := hasher.Sum64()

	return int64(hashValue), nil
}

func (hk *BasicStateKernel) HashState(state []float64) string {
	acc := float64(0)

	for idx := range len(state) {
		acc += state[idx] * float64(idx)
	}

	return strconv.FormatFloat(acc, 'f', -1, 64)
}

func (hk *BasicStateKernel) GetEventSeed(eventSpec *EventSpec) (int64, string) {
	return 1, "1"
}

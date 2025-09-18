package vm

import (
	"fortuna/core/model"
	"fortuna/crypto"
)

type KernelVersion string

const (
	BaseV000 KernelVersion = "base:v.0.0"
)

type StateKernel interface {
	GenVector(seed string, size int64) *model.StateVector
	VerifyVector(seed string, vector *model.StateVector) bool
	RandInt(seed string, minValue int64, maxValue int64) int64
	GenIndex(seed string, minValue uint64, maxValue uint64, size uint64) []uint64
	GenIndexU(seed string, minValue uint64, maxValue uint64, size uint64) []uint64

	PayloadIntoSeed(payload string, salt uint64) string
	VerifySeed(seed string, payload string, salt uint64) bool
}

func LoadKernel(version KernelVersion) StateKernel {
	switch version {
	case BaseV000:
		return &BasicStateKernel{}
	}
	return nil
}

type BasicStateKernel struct{}

func (ck *BasicStateKernel) GenVector(seed string, size int64) *model.StateVector {
	res := make([]int64, size)

	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range size {
		res[i] = rng.NextInt64()
	}

	return model.NewStateVector(res)
}

func (ck *BasicStateKernel) VerifyVector(seed string, vec *model.StateVector) bool {
	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range vec.Size() {
		expected := rng.NextInt64()

		if expected != vec.Get(int64(i)) {
			return false
		}
	}
	return true
}

func (ck *BasicStateKernel) GenIndexU(seed string, s uint64, e uint64, size uint64) []uint64 {
	res := make([]uint64, size)

	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNGU(seed_bytes, e)

	for i := range size {
		v := rng.NextInt64()
		// TODO: check if v is negative
		res[i] = uint64((v % e) + s)
	}
	return res
}

func (ck *BasicStateKernel) GenIndex(seed string, s uint64, e uint64, size uint64) []uint64 {
	res := make([]uint64, size)

	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)

	for i := range size {
		v := rng.NextInt64()
		// TODO: check if v is negative
		res[i] = uint64((uint64(v) % uint64(e)) + uint64(s))
	}
	return res
}


func (ck *BasicStateKernel) RandInt(seed string, s int64, e int64) int64 {
	seed_bytes := []byte(seed)
	rng := crypto.NewCSPRNG(seed_bytes)
	v := rng.NextInt64()

	return (v % e) + s
}

func (ck *BasicStateKernel) VerifySeed(seed string, payload string, salt uint64) bool {
	return ck.PayloadIntoSeed(payload, salt) == seed
}

func (ck *BasicStateKernel) PayloadIntoSeed(payload string, salt uint64) string {
	return ""
}


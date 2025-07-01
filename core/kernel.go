package core

type HashKernel interface {
	StringToSeed(payload string) (int64, error)
	Hash(data string) (string, error)
}

type StateKernel interface {
	GenVector(seed int64, size int64) []float64
	GenIndex(size int64, minValue int64, maxValue int64) []int64
	RandInt(minValue int64, maxValue int64) int64
}

type BasicHashKernel struct {}

func (hk *BasicHashKernel) Hash(input string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(input))

	hashedBytes := hash.Sum(nil)

	return hex.EncodeToString(hashedBytes), nil
}

func (hk *BasicHashKernel) StringToSeed(payload string) (int64, error) {
	hasher := fnv.New64a()

	_, err := hasher.Write([]byte(payload))
	if err != nil {
		return 0, err
	}

	hashValue := hasher.Sum64()

	return int64(hashValue), nil
}

type BasicStateKernel struct {}

func (ck *BasicStateKernel) GenVector(size int64) []float64 {
	rand.Seed(time.Now().UnixNano())
	vec := make([]float64, size)

	for i := range(size) {
		vec[i] = rand.Float64()
	}
	return vec
}

func (ck *BasicStateKernel) RandInt(seed int64, min int64, max int64) []int64 {
	rand.Seed(seed)
	if start > end {
        	start, end = end, start
    	}
    	return rand.Intn(end-start+1) + start
}

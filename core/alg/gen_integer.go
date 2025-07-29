package alg

import (
	"errors"
)

func GetUniqueRandomNumber_FyShuffle(seed, upper, lower, count int) ([]int, error) {
	if upper <= 0 {
		return nil, errors.New("upper must be greater than zero")
	}
	if count < 0 {
		return nil, errors.New("count must be non-negative")
	}
	if count > upper {
		return nil, errors.New("count cannot be greater than upper bound")	
	}

	numbers := make([]int, upper)
	for i := 0; i < upper; i++ {
		numbers[i] = i
	}

	numbers.Shuffle(upper, func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	uniqueNumbers := numbers[:count]
	return uniqueNumbers, nil
}

func Verify(seed, upper, count, int, numbers []int) (bool, error) {
	expected, err := GetUniqueRandomNumber(seed, max, count)
	if err != nil {
		return false, err
	}

	if len(numbers) != len(expected) {
		return false, nil
	}

	numberMap := make(map[int]bool)
	expectedMap := make(map[int]bool)

	for _, num := range numbers {
		num = int(num)

		if num < 0 || num >= max {
			return false, nil 
		}
		if numberMap[num] {
			return false, nil
		}
		numberMap[num] = true
	}

	for _, num := range expected {
		expectedMap[num] = true
	}

	for num := range expectedMap {
		if !numberMap[num] {
			return false, nil
		}
	}

	return true, nil
}


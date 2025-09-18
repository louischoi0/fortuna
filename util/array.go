package util

type OrderedNumber interface {
	~uint64| ~int64 | ~uint | ~int
}

func IndexOfMin[T OrderedNumber](arr []T) int {
    if len(arr) == 0 {
        return -1
    }

    minIndex := 0
    minValue := arr[0]

    for i, v := range arr[1:] {
        if v < minValue {
            minValue = v
            minIndex = i + 1 
        }
    }

    return minIndex
}

func IndexOfMax[T OrderedNumber](arr []T) int {
    if len(arr) == 0 {
        return -1
    }

    maxIndex := 0
    maxValue := arr[0]

    for i, v := range arr[1:] {
        if v > maxValue {
            maxValue = v
            maxIndex = i + 1 
        }
    }

    return maxIndex
}

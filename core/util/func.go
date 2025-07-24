package util

func floatEquals(a, b, epsilon float64) bool {
    if a == b {
        return true
    }
    delta := a - b
    if delta < 0 {
        delta = -delta
    }
    return delta < epsilon
}

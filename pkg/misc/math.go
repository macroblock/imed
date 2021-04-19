package misc

// MaxInt -
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt -
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}


func AbsInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

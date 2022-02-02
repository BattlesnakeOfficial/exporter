package render

// abs returns the absolute value of an integer
func abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

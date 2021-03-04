package bot

// calcLimit just rounds up num/div
func calcLimit(num, div int) int {
	n := num / div
	if num%div != 0 {
		n += 1
	}
	return n
}

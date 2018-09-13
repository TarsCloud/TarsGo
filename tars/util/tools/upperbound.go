package tools

func UpperBound(sortedArray []int, in int) (index int) {
	for i, v := range sortedArray {
		if in <= v {
			index = i
			return
		}
		index = i
	}
	return
}

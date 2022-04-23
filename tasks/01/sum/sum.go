package sum


func add(x, y int) int {
	return x + y;
}

func Sum(values []int) int {
	var sum int;
	for i := 0; i < len(values); i++ {
		sum = add(sum, values[i])
	}
	return sum
}

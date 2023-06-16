package command

func calculateBlockSize(fileSize int64, numBlocks int) int64 {
	return int64(RoundUpToNearestMultiple(int(fileSize), numBlocks) / numBlocks)
	//return int64(math.Ceil(float64(fileSize/int64(numBlocks))/10)) * 10
}

func calculateNumBlocks(fileSize int64, numServers int) int {
	return numServers
}

// Алгоритм, который находит минимальное число X, большее заданного числа N и кратное ему
func RoundUpToNearestMultiple(n, k int) int {
	return ((n + k - 1) / k) * k
}

package kimseok

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func jaro(s1, s2 string) float64 {
	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 && len2 == 0 {
		return 1
	}

	matchDistance := max(len1, len2)/2 - 1
	matches := 0
	hashS1 := make([]bool, len1)
	hashS2 := make([]bool, len2)

	for i := 0; i < len1; i++ {
		for j := max(0, i-matchDistance); j < min(len2, i+matchDistance+1); j++ {
			if s1[i] == s2[j] && !hashS2[j] {
				hashS1[i] = true
				hashS2[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0
	}

	t := 0
	point := 0

	for i := 0; i < len1; i++ {
		if hashS1[i] {
			for !hashS2[point] {
				point++
			}
			if s1[i] != s2[point] {
				t++
			}
			point++
		}
	}
	t /= 2

	return (float64(matches)/float64(len1) + float64(matches)/float64(len2) + float64(matches-t)/float64(matches)) / 3.0
}

func jaroWinkler(s1, s2 string) float64 {
	jaroDist := jaro(s1, s2)

	prefix := 0
	for i := 0; i < min(len(s1), len(s2)); i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}
	prefix = min(prefix, 4)

	return jaroDist + 0.1*float64(prefix)*(1-jaroDist)
}

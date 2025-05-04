package match

func minLen(ms []Matcher) (minimum int) {
	for i, m := range ms {
		n := m.MinLen()
		if i == 0 || n < minimum {
			minimum = n
		}
	}
	return minimum
}

func maxLen(ms []Matcher) (maximum int) {
	for i, m := range ms {
		n := m.MinLen()
		if i == 0 || n > maximum {
			maximum = n
		}
	}
	return maximum
}

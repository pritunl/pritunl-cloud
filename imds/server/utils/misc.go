package utils

func StringsContains(val []string, str string) bool {
	if val == nil {
		return false
	}
	for _, v := range val {
		if v == str {
			return true
		}
	}
	return false
}

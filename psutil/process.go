package psutil

func GetProcessNames() (names []string, err error) {
	names, err = processNames()
	if err != nil {
		return
	}

	return
}

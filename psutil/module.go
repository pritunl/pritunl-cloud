package psutil

func GetModuleNames() (names []string, err error) {
	names, err = moduleNames()
	if err != nil {
		return
	}

	return
}

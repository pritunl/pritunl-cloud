package sysctl

func Sysctl() (err error) {
	err = Optimize()
	if err != nil {
		return
	}

	err = Nested()
	if err != nil {
		return
	}

	return
}

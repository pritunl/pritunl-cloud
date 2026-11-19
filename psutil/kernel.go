package psutil

func GetKernel() (release string, err error) {
	release, err = kernelRelease()
	if err != nil {
		return
	}

	return
}

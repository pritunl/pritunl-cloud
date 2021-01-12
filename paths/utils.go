package paths

import (
	"os"

	"github.com/pritunl/pritunl-cloud/settings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	ovmfCodePaths = []string{
		"/usr/share/edk2/ovmf/OVMF_CODE.fd",
		"/usr/share/OVMF/OVMF_CODE.pure-efi.fd",
		"/usr/share/OVMF/OVMF_CODE.fd",
	}
	ovmfVarsPaths = []string{
		"/usr/share/edk2/ovmf/OVMF_VARS.fd",
		"/usr/share/OVMF/OVMF_VARS.pure-efi.fd",
		"/usr/share/OVMF/OVMF_VARS.fd",
	}
)

func existsFile(pth string) (exists bool, err error) {
	_, err = os.Stat(pth)
	if err == nil {
		exists = true
		return
	}

	if os.IsNotExist(err) {
		err = nil
		return
	}

	err = &errortypes.ReadError{
		errors.Wrapf(err, "paths: Failed to stat %s", pth),
	}
	return
}

func FindOvmfCodePath() (pth string, err error) {
	pth = settings.Hypervisor.OvmfCodePath
	if pth != "" {
		return
	}

	for _, pth = range ovmfCodePaths {
		exists, e := existsFile(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			return
		}
	}

	pth = ""
	err = &errortypes.NotFoundError{
		errors.New("paths: Failed to find OVMF code file"),
	}
	return
}

func FindOvmfVarsPath() (pth string, err error) {
	pth = settings.Hypervisor.OvmfVarsPath
	if pth != "" {
		return
	}

	for _, pth = range ovmfVarsPaths {
		exists, e := existsFile(pth)
		if e != nil {
			err = e
			return
		}

		if exists {
			return
		}
	}

	pth = ""
	err = &errortypes.NotFoundError{
		errors.New("paths: Failed to find OVMF vars file"),
	}
	return
}

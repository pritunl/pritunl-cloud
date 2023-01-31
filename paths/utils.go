package paths

import (
	"os"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
)

var (
	ovmfCodePaths = []string{
		"/usr/share/edk2/ovmf/OVMF_CODE.fd",
		"/usr/share/edk2/ovmf/OVMF_CODE.cc.fd",
		"/usr/share/OVMF/OVMF_CODE.pure-efi.fd",
		"/usr/share/OVMF/OVMF_CODE.fd",
	}
	ovmfVarsPaths = []string{
		"/usr/share/edk2/ovmf/OVMF_VARS.fd",
		"/usr/share/OVMF/OVMF_VARS.pure-efi.fd",
		"/usr/share/OVMF/OVMF_VARS.fd",
	}
	ovmfSecureCodePaths = []string{
		"/usr/share/edk2/ovmf/OVMF_CODE.secboot.fd",
		"/usr/share/OVMF/OVMF_CODE.secboot.fd",
	}
	ovmfSecureVarsPaths = []string{
		"/usr/share/edk2/ovmf/OVMF_VARS.secboot.fd",
		"/usr/share/OVMF/OVMF_VARS.secboot.fd",
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

func FindOvmfCodePath(secureBoot bool) (pth string, err error) {
	if secureBoot {
		pth = settings.Hypervisor.OvmfSecureCodePath
		if pth != "" {
			return
		}

		for _, pth = range ovmfSecureCodePaths {
			exists, e := existsFile(pth)
			if e != nil {
				err = e
				return
			}

			if exists {
				return
			}
		}
	} else {
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
	}

	pth = ""
	err = &errortypes.NotFoundError{
		errors.New("paths: Failed to find OVMF code file"),
	}
	return
}

func FindOvmfVarsPath(secureBoot bool) (pth string, err error) {
	if secureBoot {
		pth = settings.Hypervisor.OvmfSecureVarsPath
		if pth != "" {
			return
		}

		for _, pth = range ovmfSecureVarsPaths {
			exists, e := existsFile(pth)
			if e != nil {
				err = e
				return
			}

			if exists {
				return
			}
		}
	} else {
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
	}

	pth = ""
	err = &errortypes.NotFoundError{
		errors.New("paths: Failed to find OVMF vars file"),
	}
	return
}

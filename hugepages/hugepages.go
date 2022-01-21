package hugepages

import (
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

func HugepageSize() (count int, size uint64, err error) {
	virt, err := utils.GetMemInfo()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "hugepages: Failed to read virtual memory"),
		}
		return
	}

	count = int(virt.HugePagesTotal)
	size = virt.HugePageSize

	if size < 1024 {
		err = &errortypes.ReadError{
			errors.Newf("hugepages: Invalid hugepage size %d", size),
		}
		return
	}

	return
}

func UpdateHugepagesSize() (err error) {
	err = utils.ExistsMkdir(settings.Hypervisor.HugepagesPath, 0755)
	if err != nil {
		return
	}

	nodeHugepagesSize := node.Self.HugepagesSize
	if nodeHugepagesSize == 0 {
		return
	}

	curHugepagesCount, hugepageSize, err := HugepageSize()
	if err != nil {
		return
	}

	hugepagesSize := uint64(nodeHugepagesSize) * 1024
	hugepagesCount := int(hugepagesSize / hugepageSize)

	if curHugepagesCount != hugepagesCount {
		logrus.WithFields(logrus.Fields{
			"cur_nr_hugepages": curHugepagesCount,
			"new_nr_hugepages": hugepagesCount,
		}).Info("hugepages: Updating hugepages size")

		_, err = utils.ExecCombinedOutputLogged(
			nil,
			"sysctl",
			"-w",
			fmt.Sprintf("vm.nr_hugepages=%d", hugepagesCount),
		)
		if err != nil {
			return
		}
	}

	return
}

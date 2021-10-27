package oracle

import (
	"strings"
	"sync"

	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Iface struct {
	Name       string
	Address    string
	Namespace  string
	MacAddress string
	VnicId     string
}

var (
	ifaceLock = sync.Mutex{}
)

func GetIfaces(logOutput bool) (ifaces []*Iface, err error) {
	ifaceLock.Lock()
	defer ifaceLock.Unlock()

	output, err := utils.ExecCombinedOutputLogged(
		[]string{
			"does not have",
		},
		"/usr/bin/bash",
		"/home/opc/secondary_vnic_all_configure.sh",
	)
	if err != nil {
		return
	}

	if logOutput {
		logrus.WithFields(logrus.Fields{
			"output": output,
		}).Warn("oracle: Oracle iface output")
	}

	found := false
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 13 {
			if found {
				break
			} else {
				continue
			}
		}

		if found && fields[0] == "-" && fields[5] != "-" {
			iface := &Iface{
				Name:       fields[7],
				Address:    fields[1],
				Namespace:  fields[5],
				MacAddress: fields[11],
				VnicId:     fields[12],
			}

			ifaces = append(ifaces, iface)
		} else if fields[0] == "CONFIG" && fields[1] == "ADDR" &&
			fields[5] == "NS" && fields[7] == "IFACE" &&
			fields[11] == "MAC" && fields[12] == "VNIC" {

			found = true
			continue
		}

	}

	return
}

func ConfIfaces(logOutput bool) (err error) {
	ifaceLock.Lock()
	defer ifaceLock.Unlock()

	output, err := utils.ExecCombinedOutputLogged(
		[]string{
			"does not have",
		},
		"/usr/bin/bash",
		"/home/opc/secondary_vnic_all_configure.sh",
		"-c",
		"-n",
	)
	if err != nil {
		return
	}

	if logOutput {
		logrus.WithFields(logrus.Fields{
			"output": output,
		}).Warn("oracle: Oracle iface config output")
	}

	return
}

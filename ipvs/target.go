package ipvs

import (
	"fmt"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

type Target struct {
	Service    *Service
	Address    string
	Port       int
	Weight     int
	Masquerade bool
}

func (t *Target) Key() string {
	return fmt.Sprintf("%s:%d", t.Address, t.Port)
}

func (t *Target) Add() (err error) {
	if t.Weight == 0 {
		t.Weight = 1
	}

	args := []string{
		"-a",
		t.Service.Protocol, t.Service.Key(),
		"-r", t.Key(),
	}

	if t.Masquerade {
		args = append(args, "-m")
	}

	args = append(args, "-w", fmt.Sprintf("%d", t.Weight))

	resp, err := commander.Exec(&commander.Opt{
		Name:    "ipvsadm",
		Args:    args,
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(resp.Map()).Error("ipvs: Failed to add target")
		return
	}

	return
}

func (t *Target) Delete() (err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name: "ipvsadm",
		Args: []string{
			"-d",
			t.Service.Protocol, t.Service.Key(),
			"-r", t.Key(),
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(resp.Map()).Error("ipvs: Failed to remove target")
		return
	}

	return
}

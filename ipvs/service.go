package ipvs

import (
	"fmt"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	hasSysctl = false
)

type Service struct {
	Scheduler string
	Protocol  string
	Address   string
	Port      int
	Targets   []*Target
}

func (s *Service) Key() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

func (s *Service) Add() (err error) {
	if s.Scheduler == "" {
		s.Scheduler = RoundRobin
	}

	if !hasSysctl {
		resp, err := commander.Exec(&commander.Opt{
			Name: "sysctl",
			Args: []string{
				"-w", "net.ipv4.vs.conntrack=1",
			},
			PipeOut: true,
			PipeErr: true,
		})
		if err != nil {
			logrus.WithFields(resp.Map()).Error(
				"ipvs: Failed to set ipvs sysctl")
			err = nil
		}

		hasSysctl = true
	}

	resp, err := commander.Exec(&commander.Opt{
		Name: "ipvsadm",
		Args: []string{
			"-A",
			s.Protocol, s.Key(),
			"-s", s.Scheduler,
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(resp.Map()).Error("ipvs: Failed to add service")
		return
	}

	return
}

func (s *Service) Delete() (err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name: "ipvsadm",
		Args: []string{
			"-D",
			s.Protocol, s.Key(),
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(resp.Map()).Error("ipvs: Failed to remove service")
		return
	}

	return
}

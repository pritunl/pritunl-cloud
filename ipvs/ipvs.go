package ipvs

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var curState *State

type State struct {
	Services map[string]*Service
}

func (s *State) Print() string {
	var output strings.Builder

	output.WriteString("IPVS Configuration:\n")
	output.WriteString("=================\n\n")

	if len(s.Services) == 0 {
		output.WriteString("No services configured.\n")
		return output.String()
	}

	serviceKeys := make([]string, 0, len(s.Services))
	for key := range s.Services {
		serviceKeys = append(serviceKeys, key)
	}
	sort.Strings(serviceKeys)

	for _, key := range serviceKeys {
		service := s.Services[key]
		output.WriteString(fmt.Sprintf("Service: %s%s\n", service.Protocol, service.Key()))
		output.WriteString(fmt.Sprintf("  Key: %s\n", key))
		output.WriteString(fmt.Sprintf("  Scheduler: %s\n", service.Scheduler))

		if len(service.Targets) == 0 {
			output.WriteString("  No targets configured.\n")
		} else {
			output.WriteString("  Targets:\n")

			sort.Slice(service.Targets, func(i, j int) bool {
				if service.Targets[i].Address != service.Targets[j].Address {
					return service.Targets[i].Address < service.Targets[j].Address
				}
				return service.Targets[i].Port < service.Targets[j].Port
			})

			for _, target := range service.Targets {
				masq := "No"
				if target.Masquerade {
					masq = "Yes"
				}
				output.WriteString(fmt.Sprintf("    - %s (Weight: %d, Masquerade: %s)\n",
					target.Key(), target.Weight, masq))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

func (s *State) AddTarget(serviceAddr, targetAddr string,
	port int, protocol string) (err error) {

	if s.Services == nil {
		s.Services = map[string]*Service{}
	}

	serviceKey := fmt.Sprintf("%s%s:%d", protocol, serviceAddr, port)
	service := s.Services[serviceKey]
	if service == nil {
		service = &Service{
			Scheduler: RoundRobin,
			Protocol:  protocol,
			Address:   serviceAddr,
			Port:      port,
		}
		s.Services[serviceKey] = service
	}

	target := &Target{
		Service:    service,
		Address:    targetAddr,
		Port:       port,
		Weight:     1,
		Masquerade: true,
	}
	service.Targets = append(service.Targets, target)

	return
}

func UpdateState(newState *State) (err error) {
	updated := false

	if curState == nil {
		var state *State
		state, err = LoadState()
		if err != nil {
			return
		}

		curState = state
	}

	for serviceKey, service := range curState.Services {
		newService := newState.Services[serviceKey]

		if newService == nil {
			if !updated {
				logrus.WithFields(logrus.Fields{
					"reason": "unknown_service",
				}).Info("ipvs: Updating ipvs state")
				updated = true
			}
			err = service.Delete()
			if err != nil {
				return
			}

			continue
		}

		for _, target := range service.Targets {
			found := false
			for _, newTarget := range newService.Targets {
				if target.Address == newTarget.Address &&
					target.Port == newTarget.Port {

					if target.Weight != newTarget.Weight ||
						target.Masquerade != newTarget.Masquerade {

						if !updated {
							logrus.WithFields(logrus.Fields{
								"reason": "weight_masquerade",
							}).Info("ipvs: Updating ipvs state")
							updated = true
						}
						err = target.Delete()
						if err != nil {
							return
						}

						found = false
					} else {
						found = true
					}
					break
				}
			}

			if !found {
				if !updated {
					logrus.WithFields(logrus.Fields{
						"reason": "target_unknown",
					}).Info("ipvs: Updating ipvs state")
					updated = true
				}
				err = target.Delete()
				if err != nil {
					return
				}
			}
		}

		if service.Scheduler != newService.Scheduler {
			if !updated {
				logrus.WithFields(logrus.Fields{
					"reason": "scheduler",
				}).Info("ipvs: Updating ipvs state")
				updated = true
			}
			err = service.Delete()
			if err != nil {
				return
			}

			err = newService.Add()
			if err != nil {
				return
			}

			for _, target := range newService.Targets {
				target.Service = newService
				err = target.Add()
				if err != nil {
					return
				}
			}
		}
	}

	for serviceKey, newService := range newState.Services {
		service := curState.Services[serviceKey]

		if service == nil {
			if !updated {
				logrus.WithFields(logrus.Fields{
					"reason": "new_service",
				}).Info("ipvs: Updating ipvs state")
				updated = true
			}
			err = newService.Add()
			if err != nil {
				return
			}

			for _, target := range newService.Targets {
				target.Service = newService
				err = target.Add()
				if err != nil {
					return
				}
			}
		} else if service.Scheduler == newService.Scheduler {
			for _, newTarget := range newService.Targets {
				found := false
				needsUpdate := false

				for _, target := range service.Targets {
					if target.Address == newTarget.Address &&
						target.Port == newTarget.Port {

						found = true
						if target.Weight != newTarget.Weight ||
							target.Masquerade != newTarget.Masquerade {

							needsUpdate = true
						}
						break
					}
				}

				if !found || needsUpdate {
					newTarget.Service = service
					if !updated {
						logrus.WithFields(logrus.Fields{
							"reason": "new_target",
						}).Info("ipvs: Updating ipvs state")
						updated = true
					}
					err = newTarget.Add()
					if err != nil {
						return
					}
				}
			}
		}
	}

	curState = newState
	return
}

func LoadState() (state *State, err error) {
	resp, err := commander.Exec(&commander.Opt{
		Name: "ipvsadm-save",
		Args: []string{
			"-n",
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		logrus.WithFields(resp.Map()).Error("ipvs: Failed to load state")
		return
	}

	state = &State{
		Services: map[string]*Service{},
	}

	for _, line := range strings.Split(string(resp.Output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "-A" {
			serviceKey := ""
			service := &Service{
				Targets: []*Target{},
			}

			for i := 0; i < len(parts); i++ {
				switch parts[i] {
				case "-t":
					if i+1 < len(parts) {
						serviceKey = parts[i+1]
						addrPort := strings.Split(serviceKey, ":")
						serviceKey = Tcp + serviceKey

						service.Protocol = Tcp
						if len(addrPort) == 2 {
							service.Address = addrPort[0]
							service.Port, _ = strconv.Atoi(addrPort[1])
						}
						i++
					}
				case "-u":
					if i+1 < len(parts) {
						serviceKey = parts[i+1]
						addrPort := strings.Split(serviceKey, ":")
						serviceKey = Udp + serviceKey

						service.Protocol = Udp
						if len(addrPort) == 2 {
							service.Address = addrPort[0]
							service.Port, _ = strconv.Atoi(addrPort[1])
						}
						i++
					}
				case "-s":
					if i+1 < len(parts) {
						switch parts[i+1] {
						case "rr":
							service.Scheduler = RoundRobin
							break
						}
						i++
					}
				}
			}

			if serviceKey != "" {
				state.Services[serviceKey] = service
			}
		} else if parts[0] == "-a" {
			target := &Target{}

			for i := 0; i < len(parts); i++ {
				switch parts[i] {
				case "-t":
					if i+1 < len(parts) {
						target.Service = state.Services[Tcp+parts[i+1]]
						i++
					}
				case "-u":
					if i+1 < len(parts) {
						target.Service = state.Services[Udp+parts[i+1]]
						i++
					}
				case "-r":
					if i+1 < len(parts) {
						addrPort := strings.Split(parts[i+1], ":")
						if len(addrPort) == 2 {
							target.Address = addrPort[0]
							target.Port, _ = strconv.Atoi(addrPort[1])
						}
						i++
					}
				case "-w":
					if i+1 < len(parts) {
						target.Weight, _ = strconv.Atoi(parts[i+1])
						i++
					}
				case "-m":
					target.Masquerade = true
				}
			}

			if target.Service != nil {
				target.Service.Targets = append(target.Service.Targets, target)
			}
		}
	}

	return
}

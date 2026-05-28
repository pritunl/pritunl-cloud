package deploy

import (
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/netconf"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/state"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

var (
	annouceLock  = utils.NewTimeoutLock(5 * time.Minute)
	annouceStore = sync.Map{}
)

type Announce struct {
	stat *state.State
}

func (a *Announce) annouce(inst *instance.Instance) (err error) {
	virt := a.stat.GetVirt(inst.Id)
	if virt == nil {
		return
	}

	acquired, lockId := annouceLock.LockOpen()
	if !acquired {
		return
	}

	annouceStore.Store(inst.Id, time.Now())

	go func() {
		defer utils.RecoverLog("deploy: Panic in annouce action")
		defer func() {
			time.Sleep(200 * time.Millisecond)
			annouceLock.Unlock(lockId)
		}()

		nc := netconf.New(virt)

		err = nc.Iface1()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"instance_id": inst.Id.Hex(),
				"error":       err,
			}).Error("deploy: Failed to create netconf for announce")
			return
		}

		addr := ""
		addr6 := ""
		pubAddr := ""
		pubAddr6 := ""
		gatewayAddr := ""
		gatewayAddr6 := ""
		if len(inst.PrivateIps) != 0 {
			addr = strings.Split(inst.PrivateIps[0], "/")[0]
		}
		if len(inst.PrivateIps6) != 0 {
			addr6 = strings.Split(inst.PrivateIps6[0], "/")[0]
		}
		if len(inst.PublicIps) != 0 {
			pubAddr = strings.Split(inst.PublicIps[0], "/")[0]
		}
		if len(inst.PublicIps6) != 0 {
			pubAddr6 = strings.Split(inst.PublicIps6[0], "/")[0]
		} else if len(inst.CloudPublicIps6) != 0 {
			pubAddr6 = strings.Split(inst.CloudPublicIps6[0], "/")[0]
		}
		if len(inst.GatewayIps) != 0 {
			gatewayAddr = strings.Split(inst.GatewayIps[0], "/")[0]
		}
		if len(inst.GatewayIps6) != 0 {
			gatewayAddr6 = strings.Split(inst.GatewayIps6[0], "/")[0]
		}

		_ = addr
		_ = addr6

		if nc.NetworkMode == node.Static {
			iface := nc.SpaceExternalIfaceMod
			if iface == "" {
				iface = nc.SpaceExternalIface
			}

			_, _ = commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", nc.Namespace, "arping",
					"-U", "-I", iface, "-c", "2", pubAddr,
				},
				Timeout: 6 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})

			_, _ = commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", nc.Namespace, "arping",
					"-I", iface, "-c", "2",
					gatewayAddr,
				},
				Timeout: 6 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})
		}

		if nc.NetworkMode6 == node.Static {
			iface := nc.SpaceExternalIfaceMod6
			if iface == "" {
				iface = nc.SpaceExternalIface
			}

			_, _ = commander.Exec(&commander.Opt{
				Name: "ip",
				Args: []string{
					"netns", "exec", nc.Namespace, "ndisc6",
					"-r", "2", pubAddr6, iface,
				},
				Timeout: 6 * time.Second,
				PipeOut: true,
				PipeErr: true,
			})

			if nc.ExternalGatewayAddr6 != nil {
				_, _ = commander.Exec(&commander.Opt{
					Name: "ip",
					Args: []string{
						"netns", "exec", nc.Namespace, "ping6",
						"-c", "2", "-i", "0.5", "-w", "6", "-I", iface,
						gatewayAddr6,
					},
					Timeout: 8 * time.Second,
					PipeOut: true,
					PipeErr: true,
				})
			}
		}
	}()

	return
}

func (a *Announce) Deploy() (err error) {
	instances := a.stat.Instances()
	rate := time.Duration(settings.Hypervisor.AnnounceRate) * time.Second

	for _, inst := range instances {
		last, ok := annouceStore.Load(inst.Id)
		if !ok || time.Since(last.(time.Time)) > rate {
			err = a.annouce(inst)
			if err != nil {
				return
			}
			break
		}
	}

	return
}

func NewAnnounce(stat *state.State) *Announce {
	return &Announce{
		stat: stat,
	}
}

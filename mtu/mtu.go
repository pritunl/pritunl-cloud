package mtu

import (
	"fmt"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/ip"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/zone"
)

type Check struct {
	node        *node.Node
	mtuInternal int
	mtuExternal int
	mtuInstance int
	mtuHost     int
	Instances   []*instance.Instance
}

func (c *Check) host(db *database.Database) (err error) {
	ifaces, e := ip.GetIfaces("")
	if e != nil {
		err = e
		return
	}

	internalIfaces := set.NewSet()
	externalIfaces := set.NewSet()

	for _, iface := range c.node.InternalInterfaces {
		internalIfaces.Add(iface)
	}
	for _, iface := range c.node.ExternalInterfaces {
		externalIfaces.Add(iface)
	}
	for _, iface := range c.node.ExternalInterfaces6 {
		externalIfaces.Add(iface)
	}
	for _, blck := range c.node.Blocks {
		externalIfaces.Add(blck.Interface)
	}
	for _, blck := range c.node.Blocks6 {
		externalIfaces.Add(blck.Interface)
	}

	fmt.Println("*******************************************")
	fmt.Printf("host: %s\n", c.node.Name)

	for _, iface := range ifaces {
		mtu := 0

		if iface.Ifname == settings.Hypervisor.HostNetworkName {
			mtu = c.mtuHost
		} else if iface.Ifname == settings.Hypervisor.NodePortNetworkName {
			mtu = c.mtuHost
		} else if internalIfaces.Contains(iface.Ifname) {
			mtu = c.mtuHost
		} else if externalIfaces.Contains(iface.Ifname) {
			mtu = c.mtuExternal
		} else if len(iface.Ifname) != 14 {
			continue
		} else if strings.HasPrefix(iface.Ifname, "b") {
			mtu = c.mtuInternal
		} else if strings.HasPrefix(iface.Ifname, "k") {
			mtu = c.mtuInternal
		} else if strings.HasPrefix(iface.Ifname, "v") {
			if strings.HasSuffix(iface.Ifname, "0") {
				mtu = c.mtuExternal
			} else if strings.HasSuffix(iface.Ifname, "1") {
				mtu = c.mtuInternal
			} else if strings.HasSuffix(iface.Ifname, "2") {
				mtu = c.mtuHost
			}
		} else {
			continue
		}

		if iface.Mtu != mtu {
			fmt.Printf("◆◆◆ERROR◆◆◆\n%s: %d (%d)\n",
				iface.Ifname, iface.Mtu, mtu)
		} else {
			fmt.Printf("%s: %d\n", iface.Ifname, iface.Mtu)
		}
	}

	fmt.Println("*******************************************")

	return
}

func (c *Check) instances(db *database.Database) (err error) {
	insts, err := instance.GetAll(db, &bson.M{
		"node": c.node.Id,
	})

	for _, inst := range insts {
		if inst.VirtState != vm.Running {
			continue
		}

		namespace := inst.NetworkNamespace
		if namespace == "" {
			continue
		}

		ifaces, e := ip.GetIfaces(namespace)
		if e != nil {
			err = e
			return
		}

		fmt.Println("*******************************************")
		fmt.Printf("instance: %s\n", inst.Name)

		for _, iface := range ifaces {
			mtu := 0

			if iface.Ifname == "br0" {
				mtu = c.mtuInstance
			} else if iface.Ifname == "lo" {
				continue
			} else if strings.HasPrefix(iface.Ifname, "p") {
				mtu = c.mtuInstance
			} else if strings.HasPrefix(iface.Ifname, "e") {
				mtu = c.mtuExternal
			} else if strings.HasPrefix(iface.Ifname, "i") {
				mtu = c.mtuInternal
			} else if strings.HasPrefix(iface.Ifname, "x") {
				mtu = c.mtuInternal
			} else if strings.HasPrefix(iface.Ifname, "h") {
				mtu = c.mtuHost
			} else {
				fmt.Println("◆◆◆UNKNOWN IFACE◆◆◆")
			}

			if iface.Mtu != mtu {
				fmt.Printf("◆◆◆ERROR◆◆◆\n%s-%s: %d (%d)\n", namespace,
					iface.Ifname, iface.Mtu, mtu)
			} else {
				fmt.Printf("%s-%s: %d\n", namespace,
					iface.Ifname, iface.Mtu)
			}
		}

		fmt.Println("*******************************************")
	}

	return
}

func (c *Check) Run() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	ndeId, err := primitive.ObjectIDFromHex(config.Config.NodeId)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "backup: Failed to parse ObjectId"),
		}
		return
	}

	c.node, err = node.Get(db, ndeId)
	if err != nil {
		return
	}

	zne, err := zone.Get(db, c.node.Zone)
	if err != nil {
		return
	}

	if c.node.JumboFrames {
		c.mtuExternal = settings.Hypervisor.JumboMtu
	} else {
		c.mtuExternal = settings.Hypervisor.NormalMtu
	}
	if c.node.JumboFrames || c.node.JumboFramesInternal {
		c.mtuInternal = settings.Hypervisor.JumboMtu
		c.mtuInstance = settings.Hypervisor.JumboMtu
		c.mtuHost = settings.Hypervisor.JumboMtu
	} else {
		c.mtuInternal = settings.Hypervisor.NormalMtu
		c.mtuInstance = settings.Hypervisor.NormalMtu
		c.mtuHost = settings.Hypervisor.NormalMtu
	}

	if zne.NetworkMode == zone.VxlanVlan {
		c.mtuInternal -= 50
		c.mtuInstance -= 54
	}

	err = c.host(db)
	if err != nil {
		return
	}

	err = c.instances(db)
	if err != nil {
		return
	}

	return
}

func NewCheck() (chk *Check) {
	return &Check{}
}

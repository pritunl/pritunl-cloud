package qemu

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/usb"
)

type Disk struct {
	Id     string
	Index  int
	File   string
	Format string
}

type Network struct {
	Iface      string
	MacAddress string
}

type Iso struct {
	Name string
}

type UsbDevice struct {
	Vendor  string
	Product string
	Bus     string
	Address string
}

type PciDevice struct {
	Slot string
}

type DriveDevice struct {
	Id string
}

type IscsiDevice struct {
	Uri string
}

type Qemu struct {
	Id           primitive.ObjectID
	Data         string
	Kvm          bool
	Machine      string
	Cpu          string
	Cpus         int
	Cores        int
	Threads      int
	Boot         string
	Uefi         bool
	OvmfCodePath string
	OvmfVarsPath string
	Memory       int
	Vnc          bool
	VncDisplay   int
	Disks        Disks
	Networks     []*Network
	Isos         []*Iso
	UsbDevices   []*UsbDevice
	PciDevices   []*PciDevice
	DriveDevices []*DriveDevice
	IscsiDevices []*IscsiDevice
}

func (q *Qemu) GetDiskQueues() (queues int) {
	queues = q.Cpus

	if queues > settings.Hypervisor.DiskQueuesMax {
		queues = settings.Hypervisor.DiskQueuesMax
	} else if queues < settings.Hypervisor.DiskQueuesMin {
		queues = settings.Hypervisor.DiskQueuesMin
	}

	return
}

func (q *Qemu) GetNetworkQueues() (queues int) {
	queues = q.Cpus

	if queues > settings.Hypervisor.NetworkQueuesMax {
		queues = settings.Hypervisor.NetworkQueuesMax
	} else if queues < settings.Hypervisor.NetworkQueuesMin {
		queues = settings.Hypervisor.NetworkQueuesMin
	}

	return
}

func (q *Qemu) Marshal() (output string, err error) {
	localIsosPath := paths.GetLocalIsosPath()

	cmd := []string{
		"/usr/bin/qemu-system-x86_64",
		"-nographic",
	}

	nodeVga := node.Self.Vga
	if nodeVga == "" {
		nodeVga = node.Vmware
	}

	gpuPassthrough := false
	if node.Self.PciPassthrough && len(q.PciDevices) > 0 {
		gpuPassthrough = true

		for i, device := range q.PciDevices {
			cmd = append(cmd, "-device")

			if i == 0 {
				cmd = append(cmd, fmt.Sprintf(
					"vfio-pci,host=0000:%s,multifunction=on,x-vga=on",
					device.Slot,
				))
			} else {
				cmd = append(cmd, fmt.Sprintf(
					"vfio-pci,host=0000:%s",
					device.Slot,
				))
			}
		}

		cmd = append(cmd, "-display")
		cmd = append(cmd, "none")
		cmd = append(cmd, "-vga")
		cmd = append(cmd, "none")
	}

	if !gpuPassthrough && q.Vnc && q.VncDisplay != 0 {
		cmd = append(cmd, "-vga")
		cmd = append(cmd, nodeVga)
		cmd = append(cmd, "-vnc")
		cmd = append(cmd, fmt.Sprintf(
			":%d,websocket=%d,password,share=allow-exclusive",
			q.VncDisplay,
			q.VncDisplay+15900,
		))
	}

	if q.Uefi {
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,readonly,file=%s",
			q.OvmfCodePath,
		))
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,file=%s",
			q.OvmfVarsPath,
		))
	}

	if q.Kvm {
		cmd = append(cmd, "-enable-kvm")
	}

	cmd = append(cmd, "-name")
	cmd = append(cmd, fmt.Sprintf("pritunl_%s", q.Id.Hex()))

	cmd = append(cmd, "-machine")
	options := ""
	if q.Kvm {
		options += ",accel=kvm"
	}
	cmd = append(cmd, fmt.Sprintf("type=%s%s", q.Machine, options))

	if q.Kvm {
		cmd = append(cmd, "-cpu")
		if gpuPassthrough {
			cmd = append(cmd, q.Cpu+",kvm=off,hv_vendor_id=null")
		} else {
			cmd = append(cmd, q.Cpu)
		}
	}

	cmd = append(cmd, "-smp")
	cmd = append(cmd, fmt.Sprintf(
		"cpus=%d,cores=%d,threads=%d",
		q.Cpus,
		q.Cores,
		q.Threads,
	))

	cmd = append(cmd, "-boot")
	cmd = append(cmd, q.Boot)

	cmd = append(cmd, "-m")
	cmd = append(cmd, fmt.Sprintf("%dM", q.Memory))

	for _, disk := range q.Disks {
		dskId := fmt.Sprintf("disk_%s", disk.Id)
		dskDevId := fmt.Sprintf("diskdev_%s", disk.Id)

		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"file=%s,media=disk,format=%s,cache=none,"+
				"discard=unmap,if=none,id=%s",
			disk.File,
			disk.Format,
			dskId,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s",
			dskId,
			q.GetDiskQueues(),
			dskDevId,
		))
	}

	for _, device := range q.DriveDevices {
		dskHashId := drive.GetDriveHashId(device.Id)
		dskId := fmt.Sprintf("physicaldisk_%s", dskHashId)
		dskDevId := fmt.Sprintf("physicaldiskdev_%s", dskHashId)

		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"file=%s,media=disk,format=raw,cache=none,"+
				"discard=on,if=none,id=%s",
			path.Join("/dev/disk/by-id", device.Id),
			dskId,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s",
			dskId,
			q.GetDiskQueues(),
			dskDevId,
		))
	}

	hasIscsi := false
	if node.Self.Iscsi {
		for _, device := range q.IscsiDevices {
			if !hasIscsi {
				cmd = append(cmd, "-iscsi")
				cmd = append(cmd, fmt.Sprintf(
					"initiator-name=iqn.2008-11.org.linux-kvm:%s",
					q.Id.Hex(),
				))
				hasIscsi = true
			}

			iscsiHash := md5.New()
			iscsiHash.Write([]byte(device.Uri))
			iscsiId := fmt.Sprintf("%x", iscsiHash.Sum(nil))

			dskId := fmt.Sprintf("iscsidisk_%s", iscsiId)
			dskDevId := fmt.Sprintf("iscsidiskdev_%s", iscsiId)

			cmd = append(cmd, "-drive")
			cmd = append(cmd, fmt.Sprintf(
				"file=%s,media=disk,format=raw,cache=none,"+
					"discard=on,if=none,id=%s",
				device.Uri,
				dskId,
			))

			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-blk-pci,drive=%s,num-queues=%d,id=%s",
				dskId,
				q.GetDiskQueues(),
				dskDevId,
			))
		}
	}

	count := 0
	for _, network := range q.Networks {
		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-net-pci,netdev=net%d,mac=%s",
			count,
			network.MacAddress,
		))

		cmd = append(cmd, "-netdev")
		cmd = append(cmd, fmt.Sprintf(
			"tap,id=net%d,ifname=%s,script=no,vhost=on,queues=%d",
			count,
			network.Iface,
			q.GetNetworkQueues(),
		))
	}

	cmd = append(cmd, "-cdrom")
	cmd = append(cmd, paths.GetInitPath(q.Id))

	cmd = append(cmd, "-monitor")
	cmd = append(cmd, fmt.Sprintf(
		"unix:%s,server,nowait",
		paths.GetSockPath(q.Id),
	))

	cmd = append(cmd, "-qmp")
	cmd = append(cmd, fmt.Sprintf(
		"unix:%s,server,nowait",
		paths.GetQmpSockPath(q.Id),
	))

	cmd = append(cmd, "-pidfile")
	cmd = append(cmd, paths.GetPidPath(q.Id))

	guestPath := paths.GetGuestPath(q.Id)
	cmd = append(cmd, "-chardev")
	cmd = append(cmd, fmt.Sprintf(
		"socket,path=%s,server,nowait,id=guest", guestPath))
	cmd = append(cmd, "-device")
	cmd = append(cmd, "virtio-serial")
	cmd = append(cmd, "-device")
	cmd = append(cmd,
		"virtserialport,chardev=guest,name=org.qemu.guest_agent.0")

	if node.Self.UsbPassthrough {
		if len(q.UsbDevices) > 0 {
			cmd = append(cmd, "-device")
			cmd = append(cmd, "qemu-xhci")
		}

		for _, device := range q.UsbDevices {
			vendor := usb.FilterId(device.Vendor)
			product := usb.FilterId(device.Product)
			bus := usb.FilterAddr(device.Bus)
			address := usb.FilterAddr(device.Address)

			if vendor != "" && product != "" {
				cmd = append(cmd,
					"-device",
					fmt.Sprintf(
						"usb-host,vendorid=0x%s,productid=0x%s,id=usbv_%s_%s",
						vendor, product,
						vendor, product,
					),
				)
			} else if bus != "" && address != "" {
				cmd = append(cmd,
					"-device",
					fmt.Sprintf(
						"usb-host,hostbus=%s,hostaddr=%s,id=usbb_%s_%s",
						strings.TrimLeft(bus, "0"),
						strings.TrimLeft(address, "0"),
						bus,
						address,
					),
				)
			}
		}
	}

	output = fmt.Sprintf(
		systemdTemplate,
		q.Data,
		strings.Join(cmd, " "),
	)
	return
}

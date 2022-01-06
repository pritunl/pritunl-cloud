package qemu

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/usb"
	"github.com/pritunl/pritunl-cloud/utils"
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
	Dies         int
	Sockets      int
	Boot         string
	Uefi         bool
	SecureBoot   bool
	OvmfCodePath string
	OvmfVarsPath string
	Memory       int
	Hugepages    bool
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

	qemuPath, err := features.GetQemuPath()
	if err != nil {
		return
	}

	cmd := []string{
		qemuPath,
		"-nographic",
	}

	nodeVga := node.Self.Vga
	nodeEgl := false
	if nodeVga == "" {
		nodeVga = node.Vmware
	}

	if nodeVga == node.VirtioEgl {
		nodeVga = node.Virtio
		nodeEgl = true
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
		if nodeEgl {
			cmd = append(cmd, "-display")
			cmd = append(cmd, "egl-headless")
		}
		cmd = append(cmd, "-vga")
		cmd = append(cmd, nodeVga)
		cmd = append(cmd, "-vnc")
		cmd = append(cmd, fmt.Sprintf(
			":%d,websocket=%d,password=on,share=allow-exclusive",
			q.VncDisplay,
			q.VncDisplay+15900,
		))
	}

	if q.Uefi {
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,readonly=on,file=%s",
			q.OvmfCodePath,
		))
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,readonly=on,file=%s",
			q.OvmfVarsPath,
		))
	}

	if q.Kvm {
		cmd = append(cmd, "-enable-kvm")
	}

	cmd = append(cmd, "-name")
	cmd = append(cmd, fmt.Sprintf("pritunl_%s", q.Id.Hex()))

	slot := -1
	for i := 0; i < 10; i++ {
		slot += 1
		cmd = append(cmd, "-device")
		cmd = append(cmd,
			fmt.Sprintf("pcie-root-port,id=diskbus%d,slot=%d", slot, slot))
	}

	cmd = append(cmd, "-machine")
	options := ",mem-merge=on"
	if q.SecureBoot {
		options += ",smm=on"
	}
	if q.Hugepages {
		options += ",memory-backend=pc.ram"
	}
	if q.Kvm {
		options += ",accel=kvm"
	}
	if !q.Vnc {
		options += ",vmport=off"
	}
	cmd = append(cmd, fmt.Sprintf("type=%s%s", q.Machine, options))

	if q.Kvm {
		cmd = append(cmd, "-cpu")
		cmd = append(cmd, q.Cpu)
	}

	cmd = append(cmd, "-no-hpet")
	cmd = append(cmd, "-rtc", "base=utc,driftfix=slew")
	cmd = append(cmd, "-msg", "timestamp=on")
	cmd = append(cmd, "-global", "kvm-pit.lost_tick_policy=delay")
	cmd = append(cmd, "-global", "ICH9-LPC.disable_s3=1")
	cmd = append(cmd, "-global", "ICH9-LPC.disable_s4=1")
	if q.SecureBoot {
		cmd = append(
			cmd,
			"-global",
			"driver=cfi.pflash01,property=secure,value=on",
		)
	}

	cmd = append(cmd, "-smp")
	cmd = append(cmd, fmt.Sprintf(
		"cores=%d,threads=%d,dies=%d,sockets=%d",
		q.Cores,
		q.Threads,
		q.Dies,
		q.Sockets,
	))

	if q.Isos != nil && len(q.Isos) > 0 {
		cmd = append(cmd, "-boot")
		cmd = append(
			cmd,
			fmt.Sprintf(
				"order=d,menu=on,splash-time=%d",
				settings.Hypervisor.SplashTime*1000,
			),
		)
	} else {
		cmd = append(cmd, "-boot")
		cmd = append(cmd, q.Boot)
	}

	cmd = append(cmd, "-m")
	cmd = append(cmd, fmt.Sprintf("%dM", q.Memory))

	if q.Hugepages {
		cmd = append(cmd, "-object")
		cmd = append(cmd, fmt.Sprintf(
			"memory-backend-file,id=pc.ram,"+
				"size=%dM,mem-path=%s,prealloc=off,share=off,merge=on",
			q.Memory,
			paths.GetHugepagePath(q.Id),
		))
	}

	diskAio := settings.Hypervisor.DiskAio
	if diskAio == "" {
		supported, e := features.GetUringSupport()
		if e != nil {
			err = e
			return
		}

		if supported {
			diskAio = "io_uring"
		} else {
			diskAio = "threads"
		}
	}

	for _, disk := range q.Disks {
		dskId := fmt.Sprintf("fd_%s", disk.Id)
		dskFileId := fmt.Sprintf("fdf_%s", disk.Id)
		dskDevId := fmt.Sprintf("fdd_%s", disk.Id)

		cmd = append(cmd, "-blockdev")
		cmd = append(cmd, fmt.Sprintf(
			"driver=file,node-name=%s,filename=%s,aio=%s,"+
				"discard=unmap,cache.direct=on,cache.no-flush=off",
			dskFileId,
			disk.File,
			diskAio,
		))

		cmd = append(cmd, "-blockdev")
		cmd = append(cmd, fmt.Sprintf(
			"driver=%s,node-name=%s,file=%s,"+
				"cache.direct=on,cache.no-flush=off",
			disk.Format,
			dskId,
			dskFileId,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,bus=diskbus%d",
			dskId,
			q.GetDiskQueues(),
			dskDevId,
			disk.Index,
		))
	}

	for _, device := range q.DriveDevices {
		dskHashId := drive.GetDriveHashId(device.Id)
		dskId := fmt.Sprintf("pd_%s", dskHashId)
		dskDevId := fmt.Sprintf("pdd_%s", dskHashId)
		dskBusId := fmt.Sprintf("pdb_%s", dskHashId)
		slot += 1

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"pcie-root-port,id=%s,slot=%d",
			dskBusId, slot,
		))

		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"file=%s,media=disk,format=raw,cache=none,"+
				"discard=on,if=none,id=%s",
			path.Join("/dev/disk/by-id", device.Id),
			dskId,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,bus=%s",
			dskId,
			q.GetDiskQueues(),
			dskDevId,
			dskBusId,
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

			dskId := fmt.Sprintf("id_%s", iscsiId)
			dskDevId := fmt.Sprintf("idd_%s", iscsiId)
			dskBusId := fmt.Sprintf("idb_%s", iscsiId)
			slot += 1

			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"pcie-root-port,id=%s,slot=%d",
				dskBusId, slot,
			))

			cmd = append(cmd, "-drive")
			cmd = append(cmd, fmt.Sprintf(
				"file=%s,media=disk,format=raw,cache=none,"+
					"discard=on,if=none,id=%s",
				device.Uri,
				dskId,
			))

			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,bus=%s",
				dskId,
				q.GetDiskQueues(),
				dskDevId,
				dskBusId,
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

	cmd = append(cmd, "-drive")
	cmd = append(cmd, fmt.Sprintf(
		"file=%s,media=cdrom,index=0",
		paths.GetInitPath(q.Id),
	))

	if q.Isos != nil && len(q.Isos) > 0 {
		for i, iso := range q.Isos {
			cmd = append(cmd, "-drive")
			cmd = append(cmd, fmt.Sprintf(
				"file=%s,media=cdrom,index=%d",
				path.Join(
					localIsosPath,
					path.Base(utils.FilterPath(iso.Name, 128)),
				),
				i+1,
			))
		}
	}

	cmd = append(cmd, "-monitor")
	cmd = append(cmd, fmt.Sprintf(
		"unix:%s,server=on,wait=off",
		paths.GetSockPath(q.Id),
	))

	cmd = append(cmd, "-qmp")
	cmd = append(cmd, fmt.Sprintf(
		"unix:%s,server=on,wait=off",
		paths.GetQmpSockPath(q.Id),
	))

	cmd = append(cmd, "-pidfile")
	cmd = append(cmd, paths.GetPidPath(q.Id))

	guestPath := paths.GetGuestPath(q.Id)
	cmd = append(cmd, "-chardev")
	cmd = append(cmd, fmt.Sprintf(
		"socket,path=%s,server=on,wait=off,id=guest", guestPath))
	cmd = append(cmd, "-device")
	cmd = append(cmd, "virtio-serial")
	cmd = append(cmd, "-device")
	cmd = append(cmd,
		"virtserialport,chardev=guest,name=org.qemu.guest_agent.0")

	if !settings.Hypervisor.NoSandbox {
		cmd = append(cmd, "-sandbox")
		cmd = append(cmd, "on,obsolete=deny,elevateprivileges=deny,"+
			"spawn=deny,resourcecontrol=deny")
	}

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

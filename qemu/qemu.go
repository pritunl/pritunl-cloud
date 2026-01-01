package qemu

import (
	"crypto/md5"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/compositor"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/features"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/permission"
	"github.com/pritunl/pritunl-cloud/render"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
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
	Id      string
	Vendor  string
	Product string
	Bus     string
	Address string
	BusPath string
}

type PciDevice struct {
	Slot string
	Gpu  bool
}

type DriveDevice struct {
	Id     string
	Type   string
	VgName string
	LvName string
}

type Mount struct {
	Id   string
	Name string
	Sock string
}

type IscsiDevice struct {
	Uri string
}

type Qemu struct {
	Id           bson.ObjectID
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
	Tpm          bool
	OvmfCodePath string
	OvmfVarsPath string
	Memory       int
	Hugepages    bool
	Vnc          bool
	VncDisplay   int
	Spice        bool
	SpicePort    int
	Gui          bool
	GuiUser      string
	GuiMode      string
	ProtectHome  bool
	ProtectTmp   bool
	Namespace    string
	Disks        Disks
	Networks     []*Network
	Isos         []*Iso
	UsbDevices   []*UsbDevice
	PciDevices   []*PciDevice
	DriveDevices []*DriveDevice
	IscsiDevices []*IscsiDevice
	Mounts       []*Mount
}

func (q *Qemu) GetDiskQueues() (queues int) {
	queues = int(math.Ceil(float64(q.Cores) / 2))

	if queues > settings.Hypervisor.DiskQueuesMax {
		queues = settings.Hypervisor.DiskQueuesMax
	} else if queues < settings.Hypervisor.DiskQueuesMin {
		queues = settings.Hypervisor.DiskQueuesMin
	}

	return
}

func (q *Qemu) GetNetworkQueues() (queues int) {
	queues = int(math.Ceil(float64(q.Cores) / 2))

	if queues > settings.Hypervisor.NetworkQueuesMax {
		queues = settings.Hypervisor.NetworkQueuesMax
	} else if queues < settings.Hypervisor.NetworkQueuesMin {
		queues = settings.Hypervisor.NetworkQueuesMin
	}

	return
}

func (q *Qemu) GetNetworkVectors() (vectors int) {
	vectors = int(math.Ceil(float64(q.Cores) / 2))

	if vectors > settings.Hypervisor.NetworkQueuesMax {
		vectors = settings.Hypervisor.NetworkQueuesMax
	} else if vectors < settings.Hypervisor.NetworkQueuesMin {
		vectors = settings.Hypervisor.NetworkQueuesMin
	}

	vectors = (2 * vectors) + 2

	return
}

func (q *Qemu) Marshal() (output string, err error) {
	localIsosPath := paths.GetLocalIsosPath()
	slot := -1

	qemuPath, err := features.GetQemuPath()
	if err != nil {
		return
	}

	cmd := []string{
		qemuPath,
		"-nographic",
	}

	cmd = append(cmd, "-uuid")
	cmd = append(cmd, paths.GetVmUuid(q.Id))

	nodeVga := node.Self.Vga
	nodeVgaRenderPath := ""
	if nodeVga == "" {
		nodeVga = node.Virtio
	}

	if node.VgaRenderModes.Contains(nodeVga) {
		nodeVgaRender := node.Self.VgaRender
		if nodeVgaRender != "" {
			nodeVgaRenderPath, err = render.GetRender(nodeVgaRender)
			if err != nil {
				return
			}
		}
	}

	memoryBackend, err := features.GetMemoryBackendSupport()
	if err != nil {
		return
	}

	pciPassthrough := false
	gpuPassthrough := false
	if node.Self.PciPassthrough && len(q.PciDevices) > 0 {
		pciPassthrough = true

		for i, device := range q.PciDevices {
			slot += 1
			cmd = append(cmd, "-device")
			cmd = append(cmd,
				fmt.Sprintf("pcie-root-port,id=pcibus%d,slot=%d", i, slot))

			cmd = append(cmd, "-device")
			if device.Gpu {
				gpuPassthrough = true
				cmd = append(cmd, fmt.Sprintf(
					"vfio-pci,host=0000:%s,bus=pcibus%d,"+
						"multifunction=on,x-vga=on",
					device.Slot,
					i,
				))
			} else {
				cmd = append(cmd, fmt.Sprintf(
					"vfio-pci,host=0000:%s,bus=pcibus%d",
					device.Slot, i,
				))
			}
		}

		if gpuPassthrough {
			cmd = append(cmd, "-display")
			cmd = append(cmd, "none")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		}
	}

	vgaPrime := false
	if !gpuPassthrough && (q.Vnc || q.Spice || q.Gui) {
		if q.Gui {
			cmd = append(cmd, "-display")
			if q.GuiMode == node.Gtk && !settings.Hypervisor.NoGuiFullscreen {
				cmd = append(cmd, fmt.Sprintf(
					"%s,gl=on,window-close=off,full-screen=on", q.GuiMode))
			} else {
				cmd = append(cmd, fmt.Sprintf(
					"%s,gl=on,window-close=off", q.GuiMode))
			}
		} else if node.VgaRenderModes.Contains(nodeVga) {
			cmd = append(cmd, "-display")
			options := "egl-headless"
			if nodeVgaRenderPath != "" {
				options += fmt.Sprintf(",rendernode=%s", nodeVgaRenderPath)
			}
			cmd = append(cmd, options)
		}

		switch nodeVga {
		case node.Std:
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "std")
		case node.Vmware:
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "vmware")
		case node.Virtio:
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "virtio")
		case node.VirtioPci:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-pci")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioPciPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-pci")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioVgaGl:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-vga-gl")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioVgaGlPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-vga-gl")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioVgaGlVulkan:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-vga-gl,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioVgaGlVulkanPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-vga-gl,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioGl:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-gl")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioGlPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-gl")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioGlVulkan:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-gpu-gl,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioGlVulkanPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-gpu-gl,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioPciGl:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-gl-pci")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioPciGlPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, "virtio-gpu-gl-pci")
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		case node.VirtioPciGlVulkan:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-gpu-gl-pci,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
		case node.VirtioPciGlVulkanPrime:
			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-gpu-gl-pci,blob=true,hostmem=%dM,venus=true",
				settings.Hypervisor.GlHostMem,
			))
			cmd = append(cmd, "-vga")
			cmd = append(cmd, "none")
			vgaPrime = true
		default:
			cmd = append(cmd, "-vga")
			cmd = append(cmd, nodeVga)
		}

		if q.Vnc {
			cmd = append(cmd, "-vnc")
			cmd = append(cmd, fmt.Sprintf(
				":%d,websocket=%d,password=on,share=allow-exclusive",
				q.VncDisplay,
				q.VncDisplay+15900,
			))
		}

		if q.Spice {
			cmd = append(cmd, "-spice")
			cmd = append(cmd, fmt.Sprintf(
				"ipv4=on,port=%d,image-compression=off",
				q.SpicePort,
			))
		}
	}

	if q.Uefi {
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,unit=0,readonly=on,file=%s",
			q.OvmfCodePath,
		))
		cmd = append(cmd, "-drive")
		cmd = append(cmd, fmt.Sprintf(
			"if=pflash,format=raw,unit=1,file=%s",
			q.OvmfVarsPath,
		))
	}

	if q.Kvm {
		cmd = append(cmd, "-enable-kvm")
	}

	cmd = append(cmd, "-name")
	cmd = append(cmd, fmt.Sprintf("pritunl_%s", q.Id.Hex()))

	if !pciPassthrough {
		supported, e := features.GetRunWithSupport()
		if e != nil {
			err = e
			return
		}

		if supported {
			cmd = append(cmd, "-run-with")
			cmd = append(cmd, fmt.Sprintf(
				"user=%s", permission.GetUserName(q.Id)))
		} else {
			cmd = append(cmd, "-runas")
			cmd = append(cmd, permission.GetUserName(q.Id))
		}
	}

	for i := 0; i < 10; i++ {
		slot += 1
		cmd = append(cmd, "-device")
		cmd = append(cmd,
			fmt.Sprintf("pcie-root-port,id=diskbus%d,slot=%d", i, slot))
	}

	cmd = append(cmd, "-machine")
	options := ",mem-merge=on"
	if q.SecureBoot {
		options += ",smm=on"
	}
	if q.Hugepages && memoryBackend {
		options += ",memory-backend=pc.ram"
	}
	if q.Kvm {
		options += ",accel=kvm"
	}
	if gpuPassthrough || (!q.Vnc && !q.Spice && !q.Gui) {
		options += ",vmport=off"
	}
	cmd = append(cmd, fmt.Sprintf("type=%s%s", q.Machine, options))

	if q.Kvm {
		cmd = append(cmd, "-cpu")
		cmd = append(cmd, q.Cpu)
	}

	//cmd = append(cmd, "-no-hpet")
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

	memShare := "off"
	if len(q.Mounts) > 0 {
		memShare = "on"
	}

	if q.Hugepages {
		if memoryBackend {
			cmd = append(cmd, "-object")
			cmd = append(cmd, fmt.Sprintf(
				"memory-backend-file,id=pc.ram,"+
					"size=%dM,mem-path=%s,prealloc=on,share=%s,merge=off",
				q.Memory,
				paths.GetHugepagePath(q.Id),
				memShare,
			))
		} else {
			cmd = append(cmd, "-mem-path")
			cmd = append(cmd, paths.GetHugepagePath(q.Id))
		}
	}

	if settings.Hypervisor.VirtRng {
		cmd = append(cmd, "-object",
			"rng-random,filename=/dev/random,id=rng0")
		cmd = append(cmd, "-device", "virtio-rng-pci,rng=rng0")
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
			diskAio = "native"
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
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,"+
				"bus=diskbus%d,write-cache=on,packed=on",
			dskId,
			q.GetDiskQueues(),
			dskDevId,
			disk.Index,
		))
	}

	for _, device := range q.DriveDevices {
		drivePth := ""
		if device.Type == vm.Lvm {
			drivePth = filepath.Join("/dev/mapper",
				fmt.Sprintf("%s-%s", device.VgName, device.LvName))
		} else {
			drivePth = paths.GetDrivePath(device.Id)
		}

		dskHashId := drive.GetDriveHashId(device.Id)
		dskId := fmt.Sprintf("pd_%s", dskHashId)
		dskFileId := fmt.Sprintf("pdf_%s", dskHashId)
		dskDevId := fmt.Sprintf("pdd_%s", dskHashId)
		dskBusId := fmt.Sprintf("pdb_%s", dskHashId)
		slot += 1

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"pcie-root-port,id=%s,slot=%d",
			dskBusId, slot,
		))

		cmd = append(cmd, "-blockdev")
		cmd = append(cmd, fmt.Sprintf(
			"driver=file,node-name=%s,filename=%s,aio=%s,"+
				"discard=unmap,cache.direct=on,cache.no-flush=off",
			dskFileId,
			drivePth,
			diskAio,
		))

		cmd = append(cmd, "-blockdev")
		cmd = append(cmd, fmt.Sprintf(
			"driver=raw,node-name=%s,file=%s,"+
				"cache.direct=on,cache.no-flush=off",
			dskId,
			dskFileId,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,"+
				"bus=%s,write-cache=on,packed=on",
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
			dskFileId := fmt.Sprintf("idf_%s", iscsiId)
			dskDevId := fmt.Sprintf("idd_%s", iscsiId)
			dskBusId := fmt.Sprintf("idb_%s", iscsiId)
			slot += 1

			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"pcie-root-port,id=%s,slot=%d",
				dskBusId, slot,
			))

			cmd = append(cmd, "-blockdev")
			cmd = append(cmd, fmt.Sprintf(
				"driver=iscsi,node-name=%s,transport=tcp,"+
					"url=%s,cache.direct=on",
				dskFileId,
				device.Uri,
			))

			cmd = append(cmd, "-blockdev")
			cmd = append(cmd, fmt.Sprintf(
				"driver=raw,node-name=%s,file=%s,"+
					"cache.direct=on,cache.no-flush=off",
				dskId,
				dskFileId,
			))

			cmd = append(cmd, "-device")
			cmd = append(cmd, fmt.Sprintf(
				"virtio-blk-pci,drive=%s,num-queues=%d,id=%s,"+
					"bus=%s,write-cache=on,packed=on",
				dskId,
				q.GetDiskQueues(),
				dskDevId,
				dskBusId,
			))
		}
	}

	for _, mount := range q.Mounts {
		vfsId := fmt.Sprintf("vfs_%s", mount.Id)
		vfsDevId := fmt.Sprintf("vfsd_%s", mount.Id)
		vfsBusId := fmt.Sprintf("vfsb_%s", mount.Id)
		slot += 1

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"pcie-root-port,id=%s,slot=%d",
			vfsBusId, slot,
		))

		cmd = append(cmd, "-chardev")
		cmd = append(cmd, fmt.Sprintf(
			"socket,id=%s,path=%s",
			vfsId,
			mount.Sock,
		))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"vhost-user-fs-pci,chardev=%s,tag=\"%s\",id=%s,bus=%s",
			vfsId,
			mount.Name,
			vfsDevId,
			vfsBusId,
		))
	}

	for i, network := range q.Networks {
		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf(
			"virtio-net-pci,netdev=net%d,mac=%s,mq=on,"+
				"packed=on,rss=on,vectors=%d",
			i,
			network.MacAddress,
			q.GetNetworkVectors(),
		))

		cmd = append(cmd, "-netdev")
		cmd = append(cmd, fmt.Sprintf(
			"tap,id=net%d,ifname=%s,script=no,vhost=on,queues=%d",
			i,
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
					path.Base(utils.FilterRelPath(iso.Name)),
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

	if q.Tpm {
		cmd = append(cmd, "-chardev")
		cmd = append(cmd, fmt.Sprintf(
			"socket,id=tpmsock0,path=%s",
			paths.GetTpmSockPath(q.Id),
		))
		cmd = append(cmd, "-tpmdev")
		cmd = append(cmd, "emulator,id=tpmdev0,chardev=tpmsock0")
		cmd = append(cmd, "-device")
		cmd = append(cmd, "tpm-tis,tpmdev=tpmdev0")
	}

	guestPath := paths.GetGuestPath(q.Id)
	cmd = append(cmd, "-chardev")
	cmd = append(cmd, fmt.Sprintf(
		"socket,path=%s,server=on,wait=off,id=guest",
		guestPath,
	))
	cmd = append(cmd, "-device")
	cmd = append(cmd, "virtio-serial")
	cmd = append(cmd, "-device")
	cmd = append(cmd,
		"virtserialport,chardev=guest,name=org.qemu.guest_agent.0")

	if !settings.Hypervisor.NoSandbox {
		cmd = append(cmd, "-sandbox")
		if q.Gui {
			cmd = append(cmd, "on,obsolete=deny,elevateprivileges=allow,"+
				"spawn=allow,resourcecontrol=deny")
		} else {
			cmd = append(cmd, "on,obsolete=deny,elevateprivileges=allow,"+
				"spawn=deny,resourcecontrol=deny")
		}
	}

	if (q.Vnc || q.Spice || q.Gui) && !settings.Hypervisor.NoVirtioHid {
		cmd = append(cmd, "-device")
		cmd = append(cmd, "virtio-tablet-pci")
		cmd = append(cmd, "-device")
		cmd = append(cmd, "virtio-keyboard-pci")
	}

	if node.Self.UsbPassthrough || q.Vnc || q.Spice || q.Gui {
		slot += 1
		cmd = append(cmd, "-device")
		cmd = append(cmd,
			fmt.Sprintf("pcie-root-port,id=usbbus,slot=%d", slot))

		cmd = append(cmd, "-device")
		cmd = append(cmd, fmt.Sprintf("qemu-xhci,bus=usbbus,p2=%d,p3=%d",
			settings.Hypervisor.UsbHsPorts,
			settings.Hypervisor.UsbSsPorts,
		))

		if (q.Vnc || q.Spice || q.Gui) && settings.Hypervisor.NoVirtioHid {
			cmd = append(cmd, "-device")
			cmd = append(cmd, "usb-tablet")
			cmd = append(cmd, "-device")
			cmd = append(cmd, "usb-kbd")
		}

		for _, device := range q.UsbDevices {
			cmd = append(cmd,
				"-device",
				fmt.Sprintf(
					"usb-host,hostdevice=%s,id=%s",
					device.BusPath,
					device.Id,
				),
			)
		}
	}

	compositorEnv := ""
	if q.Gui {
		compositorEnv, err = compositor.GetEnv(
			q.GuiUser, nodeVgaRenderPath, vgaPrime)
		if err != nil {
			return
		}
	}

	protectTmp := ""
	if q.ProtectTmp {
		protectTmp = "true"
	} else {
		protectTmp = "false"
	}

	protectHome := ""
	if q.ProtectHome {
		protectHome = "true"
	} else {
		protectHome = "read-only"
	}

	if q.Namespace == "" {
		output = fmt.Sprintf(
			systemdTemplateExternalNet,
			q.Data,
			compositorEnv,
			paths.GetCacheDir(q.Id),
			strings.Join(cmd, " "),
			protectTmp,
			protectHome,
		)
	} else {
		output = fmt.Sprintf(
			systemdTemplate,
			q.Data,
			compositorEnv,
			paths.GetCacheDir(q.Id),
			strings.Join(cmd, " "),
			protectTmp,
			protectHome,
			q.Namespace,
		)
	}

	return
}

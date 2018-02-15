package virtualbox

import (
	"fmt"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/vmdk"
	"github.com/satori/go.uuid"
	"time"
)

func NewVirtualBox(v vm.VirtualMachine) (vbox VirtualBox, err error) {
	now := time.Now()

	vboxUuid, err := uuid.NewV4()
	if err != nil {
		return
	}

	vbox = VirtualBox{
		XmlNs:   "http://www.virtualbox.org/",
		Version: "1.16-linux",
		Machine: Machine{
			Uuid:            fmt.Sprintf("{%s}", vboxUuid),
			Name:            v.Name,
			OsType:          "RedHat_64",
			SnapshotFolder:  "Snapshots",
			LastStateChange: now.UTC().Format("2006-01-02T15:04:05Z"),
			MediaRegistry: MediaRegistry{
				HardDisks: []HardDisk{},
			},
			ExtraData: []ExtraDataItem{},
			Hardware: Hardware{
				Cpu: Cpu{
					Count: v.Processors,
					Pae: Option{
						Enabled: true,
					},
					LongMode: Option{
						Enabled: true,
					},
					X2Apic: Option{
						Enabled: true,
					},
					HardwareVirtExLargePages: Option{
						Enabled: false,
					},
				},
				Memory: Memory{
					RamSize: v.Memory,
				},
				Display: Display{
					VramSize: 16,
				},
				VideoCapture: VideoCapture{
					Fps:     25,
					Options: "ac_enabled=false",
				},
				Bios: Bios{
					IoApic: Option{
						Enabled: true,
					},
				},
				Adapters: []Adapter{},
				AudioAdapter: AudioAdapter{
					Codec:     "AD1980",
					Driver:    "Pulse",
					EnabledIn: false,
				},
				Rtc: Rtc{
					LocalOrUtc: "UTC",
				},
				GuestProperties: []GuestProperty{
					GuestProperty{
						Name:      "/VirtualBox/HostInfo/GUI/LanguageID",
						Value:     "en_US",
						Timestamp: 1518515420232646000,
					},
				},
			},
			StorageControllers: []StorageController{
				StorageController{
					Name:                    "SATA",
					Type:                    "AHCI",
					PortCount:               0,
					UseHostIoCache:          false,
					Bootable:                true,
					Ide0MasterEmulationPort: "0",
					Ide0SlaveEmulationPort:  "1",
					Ide1MasterEmulationPort: "2",
					Ide1SlaveEmulationPort:  "3",
					AttachedDevices:         []AttachedDevice{},
				},
			},
		},
	}

	for _, disk := range v.Disks {
		diskUuid, e := vmdk.GetUuid(disk.Path)
		if e != nil {
			err = e
			return
		}

		vbox.Machine.MediaRegistry.HardDisks = append(
			vbox.Machine.MediaRegistry.HardDisks,
			HardDisk{
				Uuid:     fmt.Sprintf("{%s}", diskUuid),
				Location: disk.Path,
				Format:   "VMDK",
				Type:     "Normal",
			},
		)

		vbox.Machine.StorageControllers[0].PortCount += 1
		vbox.Machine.StorageControllers[0].AttachedDevices = append(
			vbox.Machine.StorageControllers[0].AttachedDevices,
			AttachedDevice{
				Type:         "HardDisk",
				HotPluggable: false,
				Port:         0,
				Device:       0,
				Image: Image{
					Uuid: fmt.Sprintf("{%s}", diskUuid),
				},
			},
		)

		for i, adapter := range v.NetworkAdapters {
			vbox.Machine.Hardware.Adapters = append(
				vbox.Machine.Hardware.Adapters,
				Adapter{
					Slot:       i,
					Enabled:    true,
					MacAddress: adapter.MacAddress,
					Type:       "82540EM",
					DisabledModes: DisabledModes{
						InternalNetwork: Name{
							Name: "intnet",
						},
						NatNetwork: Name{
							Name: "NatNetwork",
						},
					},
					BridgedInterface: Name{
						Name: adapter.BridgedInterface,
					},
				},
			)
		}
	}

	return
}

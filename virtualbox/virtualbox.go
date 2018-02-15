package virtualbox

import (
	"encoding/xml"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type AttachedDevice struct {
	Passthrough  string `xml:"passthrough,attr,omitempty"`
	Type         string `xml:"type,attr"`
	HotPluggable bool   `xml:"hotpluggable,attr"`
	Port         int    `xml:"port,attr"`
	Device       int    `xml:"device,attr"`
	Image        Image  `xml:"Image"`
}

type StorageController struct {
	Name                    string           `xml:"name,attr"`
	Type                    string           `xml:"type,attr"`
	PortCount               int              `xml:"PortCount,attr"`
	UseHostIoCache          bool             `xml:"useHostIOCache,attr"`
	Bootable                bool             `xml:"Bootable,attr"`
	Ide0MasterEmulationPort string           `xml:"IDE0MasterEmulationPort,attr,omitempty"`
	Ide0SlaveEmulationPort  string           `xml:"IDE0SlaveEmulationPort,attr,omitempty"`
	Ide1MasterEmulationPort string           `xml:"IDE1MasterEmulationPort,attr,omitempty"`
	Ide1SlaveEmulationPort  string           `xml:"IDE1SlaveEmulationPort,attr,omitempty"`
	AttachedDevices         []AttachedDevice `xml:"AttachedDevice"`
}

type GuestProperty struct {
	Name      string `xml:"name,attr"`
	Value     string `xml:"value,attr"`
	Timestamp int    `xml:"timestamp,attr"`
	Flags     string `xml:"flags,attr"`
}

type Rtc struct {
	LocalOrUtc string `xml:"localOrUTC,attr"`
}

type AudioAdapter struct {
	Codec     string `xml:"codec,attr"`
	Driver    string `xml:"driver,attr"`
	EnabledIn bool   `xml:"enabledIn,attr"`
}

type DisabledModes struct {
	InternalNetwork Name `xml:"InternalNetwork"`
	NatNetwork      Name `xml:"NATNetwork"`
}

type Adapter struct {
	Slot             int           `xml:"slot,attr"`
	Enabled          bool          `xml:"enabled,attr"`
	MacAddress       string        `xml:"MACAddress,attr"`
	Type             string        `xml:"type,attr"`
	DisabledModes    DisabledModes `xml:"DisabledModes"`
	BridgedInterface Name          `xml:"BridgedInterface"`
}

type Bios struct {
	IoApic Option `xml:"IOAPIC"`
}

type VideoCapture struct {
	Fps     int    `xml:"fps,attr"`
	Options string `xml:"options,attr"`
}

type Display struct {
	VramSize int `xml:"VRAMSize,attr"`
}

type Memory struct {
	RamSize int `xml:"RAMSize,attr"`
}

type Option struct {
	Enabled bool `xml:"enabled,attr"`
}

type Name struct {
	Name string `xml:"name,attr"`
}

type Cpu struct {
	Count                    int    `xml:"count,attr,omitempty"`
	Pae                      Option `xml:"PAE,allowempty"`
	LongMode                 Option `xml:"LongMode"`
	X2Apic                   Option `xml:"X2APIC"`
	HardwareVirtExLargePages Option `xml:"HardwareVirtExLargePages"`
}

type Hardware struct {
	Cpu             Cpu             `xml:"CPU"`
	Memory          Memory          `xml:"Memory"`
	Display         Display         `xml:"Display"`
	VideoCapture    VideoCapture    `xml:"VideoCapture"`
	Bios            Bios            `xml:"BIOS"`
	Adapters        []Adapter       `xml:"Network>Adapter"`
	AudioAdapter    AudioAdapter    `xml:"AudioAdapter"`
	Rtc             Rtc             `xml:"RTC"`
	GuestProperties []GuestProperty `xml:"GuestProperties>GuestProperty"`
}

type ExtraDataItem struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Image struct {
	Uuid     string `xml:"uuid,attr"`
	Location string `xml:"location,attr,omitempty"`
}

type HardDisk struct {
	Uuid     string `xml:"uuid,attr"`
	Location string `xml:"location,attr"`
	Format   string `xml:"format,attr"`
	Type     string `xml:"type,attr"`
}

type MediaRegistry struct {
	HardDisks []HardDisk `xml:"HardDisks>HardDisk"`
}

type Machine struct {
	Uuid               string              `xml:"uuid,attr"`
	Name               string              `xml:"name,attr"`
	OsType             string              `xml:"OSType,attr"`
	SnapshotFolder     string              `xml:"snapshotFolder,attr"`
	LastStateChange    string              `xml:"lastStateChange,attr"`
	MediaRegistry      MediaRegistry       `xml:"MediaRegistry"`
	ExtraData          []ExtraDataItem     `xml:"ExtraData>ExtraDataItem"`
	Hardware           Hardware            `xml:"Hardware"`
	StorageControllers []StorageController `xml:"StorageControllers>StorageController"`
}

type VirtualBox struct {
	XMLName xml.Name `xml:"VirtualBox"`
	XmlNs   string   `xml:"xmlns,attr"`
	Version string   `xml:"version,attr"`
	Machine Machine  `xml:"Machine"`
}

func (v *VirtualBox) Marshal() (output string, err error) {
	outputByt, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "virtualbox: Marshal error"),
		}
		return
	}

	output = "<?xml version=\"1.0\"?>\n" + string(outputByt)

	return
}

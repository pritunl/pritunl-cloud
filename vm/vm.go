package vm

type VirtualMachine struct {
	Name            string
	Processors      int
	Memory          int
	Disks           []Disk
	NetworkAdapters []NetworkAdapter
}

type Disk struct {
	Path string
}

type NetworkAdapter struct {
	MacAddress       string
	BridgedInterface string
}

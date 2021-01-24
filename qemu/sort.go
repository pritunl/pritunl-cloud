package qemu

type Disks []*Disk

func (d Disks) Len() int {
	return len(d)
}

func (d Disks) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d Disks) Less(i, j int) bool {
	return d[i].Index < d[j].Index
}

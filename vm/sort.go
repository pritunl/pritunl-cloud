package vm

type SortDisks []*Disk

func (d SortDisks) Len() int {
	return len(d)
}

func (d SortDisks) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d SortDisks) Less(i, j int) bool {
	return d[i].Index < d[j].Index
}

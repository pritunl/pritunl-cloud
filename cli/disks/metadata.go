package disks

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/disk"
)

const (
	// metadataMinGap hides the partition alignment gaps like the web
	// layout, unallocated space below it is not shown.
	metadataMinGap = 16 * 1024 * 1024
	lvmMember      = "LVM2_member"
)

// segment is one block of a layout bar, a partition, a logical volume or
// free space.
type segment struct {
	name  string
	fs    string
	label string
	uuid  string
	size  int64
	free  bool
}

// color matches the file system colors of the web layout bar.
func (s segment) color() color.Color {
	if s.free {
		return widget.ColorGray
	}

	switch s.fs {
	case "xfs":
		return widget.ColorAccent
	case "ext4", "ext3", "ext2":
		return widget.ColorGreen
	case "btrfs", "vfat":
		return widget.ColorCyan
	case "swap":
		return widget.ColorYellow
	case lvmMember:
		return widget.ColorOrange
	}

	return widget.ColorLabel
}

func (s segment) block() string {
	if s.free {
		return "░"
	}
	return "█"
}

// fsName shows the LVM physical volume signature as lvm and an empty
// file system as unformatted like the web layout.
func fsName(fs string) string {
	if fs == "" {
		return "unformatted"
	}
	if fs == lvmMember {
		return "lvm"
	}
	return fs
}

// renderBar renders the segments proportional to their size across the
// width with a space between them, every segment gets at least one cell
// so small partitions stay visible.
func renderBar(segments []segment, total int64, width int) string {
	n := len(segments)
	if n == 0 || width <= 0 {
		return ""
	}

	sep := " "
	avail := width - (n - 1)
	if avail < n {
		sep = ""
		avail = max(width, n)
	}

	widths := make([]int, n)
	assigned := 0
	largest := 0
	for i, seg := range segments {
		w := 0
		if total > 0 {
			w = int(float64(seg.size) / float64(total) * float64(avail))
		}
		if w < 1 {
			w = 1
		}
		widths[i] = w
		assigned += w
		if seg.size > segments[largest].size {
			largest = i
		}
	}

	// The rounding remainder goes to the largest segment
	widths[largest] = max(widths[largest]+avail-assigned, 1)

	parts := make([]string, 0, n)
	for i, seg := range segments {
		parts = append(parts, lipgloss.NewStyle().Foreground(seg.color()).
			Render(strings.Repeat(seg.block(), widths[i])))
	}

	return strings.Join(parts, sep)
}

// renderRow renders the legend line of a segment, the colored block,
// name, file system and size followed by the label and UUID when they
// fit in the width.
func renderRow(seg segment, width int) string {
	fs := "-"
	if !seg.free {
		fs = fsName(seg.fs)
	}

	head := fmt.Sprintf("%-16s %-11s %10s",
		widget.Truncate(16, "%s", seg.name), fs, widget.FormatBytes(seg.size))
	head = widget.Truncate(max(width-2, 0), "%s", head)

	line := lipgloss.NewStyle().Foreground(seg.color()).Render(
		seg.block()) + " " + head

	tail := seg.label
	if seg.uuid != "" {
		if tail != "" {
			tail += "  "
		}
		tail += seg.uuid
	}

	tailWidth := width - 4 - len([]rune(head))
	if tail != "" && tailWidth > 3 {
		line += "  " + widget.MutedStyle.Render(
			widget.Truncate(tailWidth, "%s", tail))
	}

	return line
}

// Sections renders the scanned disk layout like the file system card of
// the web interface, the disk bar with its partitions and unallocated
// space followed by a bar for every volume group with its logical
// volumes and free space.
func (i *Item) Sections(width int) []resource.Section {
	meta := i.dsk.Metadata
	if meta == nil || meta.Size <= 0 {
		return nil
	}

	partitions := append([]*disk.Partition{}, meta.Partitions...)
	sort.SliceStable(partitions, func(a, b int) bool {
		return partitions[a].Start < partitions[b].Start
	})

	header := widget.FormatBytes(meta.Size)
	if meta.Table != "" {
		header = strings.ToUpper(meta.Table) + " · " + header
	}
	if !meta.Timestamp.IsZero() {
		header += " · Scanned " + resource.FormatTime(meta.Timestamp)
	}

	lines := []string{
		widget.MutedStyle.Render(widget.Truncate(width, "%s", header)),
	}

	diskSegments := []segment{}
	groups := []*disk.VolumeGroup{}
	pos := int64(0)

	addFree := func(size int64) {
		if size >= metadataMinGap {
			diskSegments = append(diskSegments, segment{
				name: "Unallocated",
				size: size,
				free: true,
			})
		}
	}

	for _, part := range partitions {
		if part.Start > pos {
			addFree(part.Start - pos)
		}

		name := "Disk"
		if part.Number > 0 {
			name = fmt.Sprintf("Partition %d", part.Number)
		}

		diskSegments = append(diskSegments, segment{
			name:  name,
			fs:    part.FileSystem,
			label: part.Label,
			uuid:  part.Uuid,
			size:  part.Size,
		})

		if part.VolumeGroup != nil {
			groups = append(groups, part.VolumeGroup)
		}

		pos = max(pos, part.Start+part.Size)
	}
	if meta.Size > pos {
		addFree(meta.Size - pos)
	}

	lines = append(lines, renderBar(diskSegments, meta.Size, width))
	for _, seg := range diskSegments {
		lines = append(lines, renderRow(seg, width))
	}

	for _, vg := range groups {
		vgSegments := []segment{}
		for _, lv := range vg.Volumes {
			vgSegments = append(vgSegments, segment{
				name:  lv.Name,
				fs:    lv.FileSystem,
				label: lv.Label,
				uuid:  lv.Uuid,
				size:  lv.Size,
			})
		}
		if vg.Free > 0 {
			vgSegments = append(vgSegments, segment{
				name: "Free",
				size: vg.Free,
				free: true,
			})
		}

		title := fmt.Sprintf("Volume Group %s · %s · %s free", vg.Name,
			widget.FormatBytes(vg.Size), widget.FormatBytes(vg.Free))

		lines = append(lines,
			"",
			widget.LabelStyle.Render(widget.Truncate(width, "%s", title)),
			renderBar(vgSegments, vg.Size, width),
		)
		for _, seg := range vgSegments {
			lines = append(lines, renderRow(seg, width))
		}
	}

	return []resource.Section{
		{
			Title: "File System",
			Body:  strings.Join(lines, "\n"),
		},
	}
}

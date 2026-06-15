package psutil

import (
	"bufio"
	"os"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/metric"
	"github.com/pritunl/pritunl-cloud/utils"
	"golang.org/x/sys/unix"
)

var diskIgnoreFs = map[string]bool{
	"autofs":      true,
	"binfmt_misc": true,
	"bpf":         true,
	"cgroup":      true,
	"cgroup2":     true,
	"configfs":    true,
	"debugfs":     true,
	"devpts":      true,
	"devtmpfs":    true,
	"efivarfs":    true,
	"fusectl":     true,
	"hugetlbfs":   true,
	"iso9660":     true,
	"mqueue":      true,
	"nsfs":        true,
	"overlay":     true,
	"proc":        true,
	"procfs":      true,
	"pstore":      true,
	"ramfs":       true,
	"rpc_pipefs":  true,
	"securityfs":  true,
	"selinuxfs":   true,
	"squashfs":    true,
	"sysfs":       true,
	"tmpfs":       true,
	"tracefs":     true,
	"virtiofs":    true,
}

func diskIgnoreMount(mount string) bool {
	switch {
	case mount == "/dev" || strings.HasPrefix(mount, "/dev/"):
		return true
	case mount == "/proc" || strings.HasPrefix(mount, "/proc/"):
		return true
	case mount == "/sys" || strings.HasPrefix(mount, "/sys/"):
		return true
	case strings.HasPrefix(mount, "/run/"):
		return true
	case strings.HasPrefix(mount, "/var/lib/docker/"):
		return true
	case strings.HasPrefix(mount, "/var/lib/containers/storage/"):
		return true
	case strings.HasPrefix(mount, "/var/lib/kubelet/"):
		return true
	}
	return false
}

func diskUnescapeMount(mount string) string {
	if !strings.ContainsRune(mount, '\\') {
		return mount
	}

	builder := strings.Builder{}
	for i := 0; i < len(mount); i++ {
		if mount[i] == '\\' && i+3 < len(mount) &&
			mount[i+1] >= '0' && mount[i+1] <= '7' &&
			mount[i+2] >= '0' && mount[i+2] <= '7' &&
			mount[i+3] >= '0' && mount[i+3] <= '7' {

			val := (int(mount[i+1]-'0') << 6) |
				(int(mount[i+2]-'0') << 3) |
				int(mount[i+3]-'0')
			builder.WriteByte(byte(val))
			i += 3
			continue
		}
		builder.WriteByte(mount[i])
	}

	return builder.String()
}

func disksList() (disks []*metric.Mount, err error) {
	file, err := os.Open("/proc/self/mounts")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to open mounts"),
		}
		return
	}
	defer file.Close()

	disks = []*metric.Mount{}
	seenMount := map[string]bool{}
	seenDev := map[string]bool{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mount := diskUnescapeMount(fields[1])
		fsType := fields[2]

		if diskIgnoreFs[fsType] || strings.HasPrefix(fsType, "fuse.") {
			continue
		}
		if diskIgnoreMount(mount) {
			continue
		}
		if seenMount[mount] {
			continue
		}

		if strings.HasPrefix(device, "/dev/") && seenDev[device] {
			continue
		}

		stat := &unix.Statfs_t{}
		e := unix.Statfs(mount, stat)
		if e != nil {
			continue
		}
		if stat.Blocks == 0 {
			continue
		}

		bsize := uint64(stat.Bsize)
		size := stat.Blocks * bsize
		used := (stat.Blocks - stat.Bfree) * bsize

		if size < diskMinSize {
			continue
		}

		seenMount[mount] = true
		if strings.HasPrefix(device, "/dev/") {
			seenDev[device] = true
		}

		disks = append(disks, &metric.Mount{
			Mount: mount,
			Used:  utils.ToFixed(float64(used)/float64(size)*100, 2),
			Size:  size,
		})
	}

	err = scanner.Err()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "psutil: Failed to read mounts"),
		}
		return
	}

	return
}

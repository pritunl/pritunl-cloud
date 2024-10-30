package systemd

import (
	"fmt"
	"time"
)

func FormatUptime(timestamp time.Time) (uptime string) {
	since := time.Since(timestamp)
	minutes := int64(since.Minutes())
	days := minutes / 1440
	hours := (minutes % 1440) / 60
	minutes = (minutes % 1440) % 60

	if days > 0 {
		uptime = fmt.Sprintf("%d days", days)
	}
	if hours > 0 || uptime != "" {
		if uptime != "" {
			uptime += " "
		}
		uptime += fmt.Sprintf("%d hours", hours)
	}
	if uptime != "" {
		uptime += " "
	}
	uptime += fmt.Sprintf("%d mins", minutes)

	return
}

func FormatUptimeShort(timestamp time.Time) (uptime string) {
	since := time.Since(timestamp)
	minutes := int64(since.Minutes())
	days := minutes / 1440
	hours := (minutes % 1440) / 60
	minutes = (minutes % 1440) % 60

	if days > 0 {
		uptime = fmt.Sprintf("%d dy", days)
	}
	if hours > 0 || uptime != "" {
		if uptime != "" {
			uptime += " "
		}
		uptime += fmt.Sprintf("%d hr", hours)
	}
	if uptime != "" {
		uptime += " "
	}
	uptime += fmt.Sprintf("%d mn", minutes)

	return
}

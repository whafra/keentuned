package utils

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// TimeSpend struct for time-consuming
type TimeSpend struct {
	Desc  string
	Count time.Duration
}

// Runtime Runtime Statistics
func Runtime(start time.Time) TimeSpend {
	tc := time.Since(start)
	return TimeSpend{
		Desc:  fmt.Sprintf("use time %.3fs", tc.Seconds()),
		Count: tc}
}

// Fluctuation count the fluctuation range of data slice by average
func Fluctuation(data []float32, average float32) string {
	if average == 0 {
		return "fluctuation = [average = 0, can not calculate]"
	}

	var ratios []float32
	for _, value := range data {
		ratios = append(ratios, (value-average)*100/average)
	}

	var minimum, maximum float32
	for _, value := range ratios {
		if value < minimum {
			minimum = value
		}

		if value > maximum {
			maximum = value
		}
	}

	if maximum > 0 {
		if minimum > 0 {
			return fmt.Sprintf("fluctuation = [+%.2f%%, +%.2f%%]", minimum, maximum)
		}

		return fmt.Sprintf("fluctuation = [%.2f%%, +%.2f%%]", minimum, maximum)
	}

	return fmt.Sprintf("fluctuation = [%.2f%%, %.2f%%]", minimum, maximum)
}

// IncreaseRatio calculate the percentage after the change relative to that before the change
func IncreaseRatio(after, before float32) float32 {
	if before == 0 {
		return 0.0
	}

	return (after - before) * 100 / before
}

// ColorString print content string by color
func ColorString(color string, content string) string {
	// 其中0x1B是标记，[开始定义颜色，1代表高亮，40代表黑色背景，32代表绿色前景，0代表恢复默认颜色
	switch strings.ToUpper(color) {
	case "RED":
		return fmt.Sprintf("%c[1;40;31m%s%c[0m", 0x1B, content, 0x1B)
	case "GREEN":
		return fmt.Sprintf("%c[1;40;32m%s%c[0m", 0x1B, content, 0x1B)
	default:
		return content
	}
}

// GetExternalIP Get the real IP address of the device
func GetExternalIP() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("net.Interfaces failed, err:", err.Error())
	}

	for _, iface := range netInterfaces {
		if iface.Flags&net.FlagUp != 0 {
			addrs, _ := iface.Addrs()
			if err != nil {
				return "", err
			}

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("ip not found")
}

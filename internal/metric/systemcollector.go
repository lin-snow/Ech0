package metric

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemCollector struct {
	lastNetBytesSent uint64
	lastNetBytesRecv uint64
	lastCollectTime  time.Time
}

func NewSystemCollector() MetricCollector {
	return &SystemCollector{}
}

func (sc *SystemCollector) Collect() (model.Metrics, error) {
	var m model.Metrics
	now := time.Now().UTC()

	// ---------- CPU ----------
	// Windows 第一次采样会返回 0，所以使用短暂采样间隔
	cpuPercent, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil {
		fmt.Println("[WARN] cpu.Percent error:", err)
	} else if len(cpuPercent) > 0 {
		m.CPU.UsagePercent = cpuPercent[0]
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		m.CPU.Cores = runtime.NumCPU()
		m.CPU.FrequencyMHz = uint64(cpuInfo[0].Mhz)
	}

	// ---------- Memory ----------
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("[WARN] mem.VirtualMemory error:", err)
	} else {
		m.Memory.Total = vmStat.Total
		m.Memory.Used = vmStat.Used
		m.Memory.Available = vmStat.Available
		m.Memory.Percentage = vmStat.UsedPercent
	}

	// ---------- Disk ----------
	rootPath := getRootPath()
	diskStat, err := disk.Usage(rootPath)
	if err != nil {
		fmt.Println("[WARN] disk.Usage error:", err)
	} else {
		m.Disk.Total = diskStat.Total
		m.Disk.Used = diskStat.Used
		m.Disk.Available = diskStat.Free
		m.Disk.Percentage = diskStat.UsedPercent
	}

	// ---------- Network ----------
	netIOs, err := net.IOCounters(true) // true = 所有网卡
	if err != nil {
		fmt.Println("[WARN] net.IOCounters error:", err)
	} else if len(netIOs) > 0 {
		var sent, recv uint64
		for _, nic := range netIOs {
			// 忽略 loopback 和 docker 虚拟接口
			if strings.HasPrefix(nic.Name, "lo") || strings.HasPrefix(nic.Name, "veth") || strings.HasPrefix(nic.Name, "docker") {
				continue
			}
			sent += nic.BytesSent
			recv += nic.BytesRecv
		}

		if !sc.lastCollectTime.IsZero() {
			duration := now.Sub(sc.lastCollectTime).Seconds()
			if duration > 0 {
				m.Network.BytesSentPerSecond = float64(sent-sc.lastNetBytesSent) / duration
				m.Network.BytesReceivedPerSecond = float64(recv-sc.lastNetBytesRecv) / duration
			}
		}

		m.Network.TotalBytesSent = sent
		m.Network.TotalBytesReceived = recv

		sc.lastNetBytesSent = sent
		sc.lastNetBytesRecv = recv
		sc.lastCollectTime = now
	}

	// ---------- System ----------
	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("[WARN] host.Info error:", err)
	} else {
		m.System.Hostname = hostInfo.Hostname
		m.System.OsName = hostInfo.Platform
		m.System.KernelVersion = hostInfo.KernelVersion
		m.System.KernelArch = hostInfo.KernelArch
		m.System.Uptime = time.Duration(hostInfo.Uptime) * time.Second
		m.System.Time = now
		m.System.TimeZone = now.Location().String()
		m.System.ProcessCount = int(hostInfo.Procs)
		m.System.GolangVersion = runtime.Version()
		m.System.GoRoutineCount = runtime.NumGoroutine()
	}

	return m, nil
}

func (sc *SystemCollector) Reset() error {
	sc.lastNetBytesRecv = 0
	sc.lastNetBytesSent = 0
	sc.lastCollectTime = time.Time{}
	return nil
}

// getRootPath 根据系统类型返回合适的根路径
func getRootPath() string {
	switch runtime.GOOS {
	case "windows":
		return "C:\\"
	case "darwin":
		return "/"
	default:
		return "/" // Linux 默认
	}
}

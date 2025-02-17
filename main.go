package main

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

const logo = `
 ██████   ██████  
████████ ████████ 
████████████████ 
 ██████████████  
  ████████████   
    ████████     
      ████       
       ██        
`

func main() {
	userName := getUserName()
	separator := strings.Repeat("-", len(userName))

	osInfo := getOSInfo()
	hostName := getHostNameInfo()
	resolutionInfo := getResolutionInfo()
	cpuInfo := getCPUInfo()
	gpuInfo := getGPUInfo()
	memInfo := getMemoryInfo()
	diskInfo := getDiskInfo()
	uptime := getUptime()

	info := []string{
		userName,
		separator,
		osInfo,
		hostName,
		resolutionInfo,
		cpuInfo,
		gpuInfo,
		memInfo,
		diskInfo,
		uptime,
	}

	logoLines := strings.Split(strings.TrimRight(logo, "\n"), "\n")
	infoLines := info

	maxLogoWidth := 0
	for _, line := range logoLines {
		trimmedLine := strings.TrimRight(line, " ")
		if len(trimmedLine) > maxLogoWidth {
			maxLogoWidth = len(trimmedLine)
		}
	}

	for i := 0; i < max(len(logoLines), len(infoLines)); i++ {
		logoLine := ""
		if i < len(logoLines) {
			logoLine = strings.TrimRight(logoLines[i], " ")
		}

		infoLine := ""
		if i < len(infoLines) {
			infoLine = infoLines[i]
		}

		fmt.Printf("%-*s %s\n", maxLogoWidth, logoLine, infoLine)
	}
}

func getUserName() string {
	currentUser, err := user.Current()
	if err != nil {
		return "User: Not available"
	}

	parts := strings.Split(currentUser.Username, "\\")
	username := parts[len(parts)-1]

	return fmt.Sprintf("User: %s", username)
}

func getOSInfo() string {
	key, _ := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	productName, _, _ := key.GetStringValue("ProductName")
	buildNumber, _, _ := key.GetStringValue("CurrentBuildNumber")
	return fmt.Sprintf("OS: %s (Build %s)", productName, buildNumber)
}

func getHostNameInfo() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Host: Not available"
	}
	return fmt.Sprintf("Host: %s", hostname)
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetDC         = user32.NewProc("GetDC")
	procReleaseDC     = user32.NewProc("ReleaseDC")
	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	procGetDeviceCaps = gdi32.NewProc("GetDeviceCaps")
)

const (
	HORZRES = 8
	VERTRES = 10
)

func getResolutionInfo() string {
	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return "Resolution: Not available"
	}
	defer procReleaseDC.Call(0, hdc)

	width, _, _ := procGetDeviceCaps.Call(hdc, HORZRES)
	height, _, _ := procGetDeviceCaps.Call(hdc, VERTRES)

	return fmt.Sprintf("Resolution: %dx%d", width, height)
}

func getCPUInfo() string {
	cpuModel, _ := cpu.Info()
	cores := runtime.NumCPU()
	return fmt.Sprintf("CPU: %s (%d cores)", cpuModel[0].ModelName, cores)
}

func getGPUInfo() string {
	type Win32_VideoController struct {
		Name string
	}

	var gpus []Win32_VideoController
	query := "SELECT Name FROM Win32_VideoController"
	err := wmi.Query(query, &gpus)
	if err != nil {
		return "GPU: Not available"
	}

	var gpuNames []string
	for _, gpu := range gpus {
		if gpu.Name != "" {
			gpuNames = append(gpuNames, gpu.Name)
		}
	}

	if len(gpuNames) == 0 {
		return "GPU: Not detected"
	}

	return fmt.Sprintf("GPU: %s", strings.Join(gpuNames, ", "))
}

func getMemoryInfo() string {
	memStat, _ := mem.VirtualMemory()
	total := memStat.Total / (1024 * 1024 * 1024)
	used := memStat.Used / (1024 * 1024 * 1024)
	return fmt.Sprintf("Memory: %dGB / %dGB", used, total)
}

func getDiskInfo() string {
	diskStat, _ := disk.Usage("C:")
	total := diskStat.Total / (1024 * 1024 * 1024)
	used := diskStat.Used / (1024 * 1024 * 1024)
	return fmt.Sprintf("Disk (C:): %dGB / %dGB", used, total)
}

func getUptime() string {
	uptimeSec, _ := host.Uptime()
	uptime := time.Duration(uptimeSec) * time.Second

	uptimeStr := uptime.String()
	uptimeStr = strings.ReplaceAll(uptimeStr, "h", "h ")
	uptimeStr = strings.ReplaceAll(uptimeStr, "m", "m ")
	uptimeStr = strings.ReplaceAll(uptimeStr, "s", "s ")
	uptimeStr = strings.TrimSpace(uptimeStr)

	return fmt.Sprintf("Uptime: %s", uptimeStr)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

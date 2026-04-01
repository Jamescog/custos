package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type UPowerData struct {
	Devices []Device
	Daemon  DaemonInfo
}

type Device struct {
	Path          string
	NativePath    string
	Vendor        string
	Model         string
	Serial        string
	PowerSupply   bool
	Updated       time.Time
	UpdatedAgo    time.Duration
	HasHistory    bool
	HasStatistics bool
	Battery       *BatteryInfo
	LinePower     *LinePowerInfo
	DisplayDevice *DisplayDeviceInfo
}

type BatteryInfo struct {
	Present          bool
	Rechargeable     bool
	State            string
	WarningLevel     string
	Energy           float64
	EnergyEmpty      float64
	EnergyFull       float64
	EnergyFullDesign float64
	EnergyRate       float64
	Voltage          float64
	ChargeCycles     int
	Percentage       float64
	Capacity         float64
	Technology       string
	IconName         string
}

type LinePowerInfo struct {
	WarningLevel string
	Online       bool
	IconName     string
}

type DisplayDeviceInfo struct {
	Present      bool
	State        string
	WarningLevel string
	Energy       float64
	EnergyFull   float64
	EnergyRate   float64
	ChargeCycles string // Can be "N/A" or number
	Percentage   float64
	IconName     string
}

type DaemonInfo struct {
	DaemonVersion  string
	OnBattery      bool
	LidIsClosed    bool
	LidIsPresent   bool
	CriticalAction string
}

func ParseUPowerOutput(output string) (*UPowerData, error) {
	data := &UPowerData{
		Devices: []Device{},
	}

	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentDevice *Device
	var section string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "Device: ") {
			if currentDevice != nil {
				data.Devices = append(data.Devices, *currentDevice)
			}

			currentDevice = &Device{
				Path: strings.TrimPrefix(line, "Device: "),
			}
			section = ""
			continue
		}

		if strings.HasPrefix(line, "Daemon:") {
			if currentDevice != nil {
				data.Devices = append(data.Devices, *currentDevice)
				currentDevice = nil
			}
			parseDaemonSection(scanner, data)
			break
		}

		if currentDevice != nil {
			if strings.TrimSpace(line) == "battery" {
				section = "battery"
				currentDevice.Battery = &BatteryInfo{}
				continue
			} else if strings.TrimSpace(line) == "line-power" {
				section = "line-power"
				currentDevice.LinePower = &LinePowerInfo{}
				continue
			} else if strings.TrimSpace(line) == "DisplayDevice" {
				currentDevice.DisplayDevice = &DisplayDeviceInfo{}
				section = "display"
				continue
			}

			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				value = strings.Trim(value, "'")

				switch section {
				case "battery":
					parseBatteryProperty(currentDevice.Battery, key, value)
				case "line-power":
					parseLinePowerProperty(currentDevice.LinePower, key, value)
				case "display":
					parseDisplayDeviceProperty(currentDevice.DisplayDevice, key, value)
				default:
					parseDeviceProperty(currentDevice, key, value)
				}
			}
		}
	}

	if currentDevice != nil {
		data.Devices = append(data.Devices, *currentDevice)
	}

	return data, nil
}

func parseDeviceProperty(device *Device, key, value string) {
	switch key {
	case "native-path":
		device.NativePath = value
	case "vendor":
		device.Vendor = value
	case "model":
		device.Model = value
	case "serial":
		device.Serial = value
	case "power supply":
		device.PowerSupply = value == "yes"
	case "updated":
		if idx := strings.Index(value, "("); idx != -1 {
			timeStr := strings.TrimSpace(value[:idx])
			if t, err := time.Parse("Mon 02 Jan 2006 03:04:05 PM MST", timeStr); err == nil {
				device.Updated = t
			}
			agoStr := strings.Trim(strings.TrimPrefix(value[idx:], "("), ")")
			if strings.Contains(agoStr, "seconds ago") {
				seconds := strings.TrimSpace(strings.Split(agoStr, " ")[0])
				if sec, err := strconv.Atoi(seconds); err == nil {
					device.UpdatedAgo = time.Duration(sec) * time.Second
				}
			}
		}
	case "has history":
		device.HasHistory = value == "yes"
	case "has statistics":
		device.HasStatistics = value == "yes"
	}
}

func parseBatteryProperty(battery *BatteryInfo, key, value string) {
	switch key {
	case "present":
		battery.Present = value == "yes"
	case "rechargeable":
		battery.Rechargeable = value == "yes"
	case "state":
		battery.State = value
	case "warning-level":
		battery.WarningLevel = value
	case "energy":
		battery.Energy = parseFloat(value)
	case "energy-empty":
		battery.EnergyEmpty = parseFloat(value)
	case "energy-full":
		battery.EnergyFull = parseFloat(value)
	case "energy-full-design":
		battery.EnergyFullDesign = parseFloat(value)
	case "energy-rate":
		battery.EnergyRate = parseFloat(value)
	case "voltage":
		battery.Voltage = parseFloat(value)
	case "charge-cycles":
		battery.ChargeCycles = parseInt(value)
	case "percentage":
		battery.Percentage = parsePercentage(value)
	case "capacity":
		battery.Capacity = parsePercentage(value)
	case "technology":
		battery.Technology = value
	case "icon-name":
		battery.IconName = value
	}
}

func parseLinePowerProperty(linePower *LinePowerInfo, key, value string) {
	switch key {
	case "warning-level":
		linePower.WarningLevel = value
	case "online":
		linePower.Online = value == "yes"
	case "icon-name":
		linePower.IconName = value
	}
}

func parseDisplayDeviceProperty(display *DisplayDeviceInfo, key, value string) {
	switch key {
	case "present":
		display.Present = value == "yes"
	case "state":
		display.State = value
	case "warning-level":
		display.WarningLevel = value
	case "energy":
		display.Energy = parseFloat(value)
	case "energy-full":
		display.EnergyFull = parseFloat(value)
	case "energy-rate":
		display.EnergyRate = parseFloat(value)
	case "charge-cycles":
		display.ChargeCycles = value
	case "percentage":
		display.Percentage = parsePercentage(value)
	case "icon-name":
		display.IconName = value
	}
}

func parseDaemonSection(scanner *bufio.Scanner, data *UPowerData) {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "daemon-version":
				data.Daemon.DaemonVersion = value
			case "on-battery":
				data.Daemon.OnBattery = value == "yes"
			case "lid-is-closed":
				data.Daemon.LidIsClosed = value == "yes"
			case "lid-is-present":
				data.Daemon.LidIsPresent = value == "yes"
			case "critical-action":
				data.Daemon.CriticalAction = value
			}
		}
	}
}

func parseFloat(value string) float64 {
	// Remove units
	parts := strings.Fields(value)
	if len(parts) > 0 {
		val, err := strconv.ParseFloat(parts[0], 64)
		if err == nil {
			return val
		}
	}
	return 0
}

func parseInt(value string) int {
	val, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return val
}

func parsePercentage(value string) float64 {
	// Remove % sign
	value = strings.TrimSuffix(value, "%")
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return val
}

// QueryUPower runs the upower command and returns parsed data
func QueryUPower() (*UPowerData, error) {
	cmd := exec.Command("upower", "-d")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run upower: %w", err)
	}

	return ParseUPowerOutput(string(output))
}

// Helper methods for easy access to common information

// GetBattery returns the main battery device (if exists)
func (d *UPowerData) GetBattery() *Device {
	for i := range d.Devices {
		if d.Devices[i].Battery != nil && d.Devices[i].NativePath == "BAT0" {
			return &d.Devices[i]
		}
	}
	return nil
}

// GetACAdapter returns the AC adapter device (if exists)
func (d *UPowerData) GetACAdapter() *Device {
	for i := range d.Devices {
		if d.Devices[i].LinePower != nil {
			return &d.Devices[i]
		}
	}
	return nil
}

// GetDisplayDevice returns the display device (if exists)
func (d *UPowerData) GetDisplayDevice() *Device {
	for i := range d.Devices {
		if d.Devices[i].DisplayDevice != nil {
			return &d.Devices[i]
		}
	}
	return nil
}

// IsOnBattery returns true if system is running on battery
func (d *UPowerData) IsOnBattery() bool {
	return d.Daemon.OnBattery
}

// GetBatteryPercentage returns the current battery percentage
func (d *UPowerData) GetBatteryPercentage() float64 {
	if battery := d.GetBattery(); battery != nil && battery.Battery != nil {
		return battery.Battery.Percentage
	}
	if display := d.GetDisplayDevice(); display != nil && display.DisplayDevice != nil {
		return display.DisplayDevice.Percentage
	}
	return 0
}

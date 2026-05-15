package main

import "time"

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
	ChargeCycles string
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

func (d *UPowerData) GetBattery() *Device {
	var fallback *Device
	for i := range d.Devices {
		if d.Devices[i].Battery == nil {
			continue
		}
		if d.Devices[i].NativePath == "BAT0" {
			return &d.Devices[i]
		}
		if fallback == nil {
			fallback = &d.Devices[i]
		}
	}
	return fallback
}

func (d *UPowerData) GetACAdapter() *Device {
	for i := range d.Devices {
		if d.Devices[i].LinePower != nil {
			return &d.Devices[i]
		}
	}
	return nil
}

func (d *UPowerData) GetDisplayDevice() *Device {
	for i := range d.Devices {
		if d.Devices[i].DisplayDevice != nil {
			return &d.Devices[i]
		}
	}
	return nil
}

func (d *UPowerData) IsOnBattery() bool {
	return d.Daemon.OnBattery
}

func (d *UPowerData) GetBatteryPercentage() float64 {
	if battery := d.GetBattery(); battery != nil && battery.Battery != nil {
		return battery.Battery.Percentage
	}
	if display := d.GetDisplayDevice(); display != nil && display.DisplayDevice != nil {
		return display.DisplayDevice.Percentage
	}
	return 0
}

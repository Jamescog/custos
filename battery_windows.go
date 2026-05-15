//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	Reserved1           byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

var getSystemPowerStatusProc = syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemPowerStatus")

func QueryUPower() (*UPowerData, error) {
	status, err := getSystemPowerStatus()
	if err != nil {
		return nil, err
	}

	percentage := float64(status.BatteryLifePercent)
	if status.BatteryLifePercent == 255 {
		percentage = 0
	}

	onBattery := status.ACLineStatus == 0
	state := "discharging"
	if status.BatteryFlag&8 != 0 {
		state = "charging"
	} else if !onBattery && percentage >= 99 {
		state = "fully-charged"
	}

	warningLevel := batteryWarningLevel(status.BatteryFlag)

	data := &UPowerData{
		Devices: []Device{{
			NativePath: "BAT0",
			Model:      "Battery",
			Battery: &BatteryInfo{
				Present:      status.BatteryFlag != 128,
				Rechargeable: true,
				State:        state,
				WarningLevel: warningLevel,
				Percentage:   percentage,
				Capacity:     percentage,
			},
		}, {
			NativePath: "AC",
			LinePower: &LinePowerInfo{
				WarningLevel: warningLevel,
				Online:       !onBattery,
			},
		}, {
			NativePath: "DisplayBattery",
			DisplayDevice: &DisplayDeviceInfo{
				Present:      status.BatteryFlag != 128,
				State:        state,
				WarningLevel: warningLevel,
				Percentage:   percentage,
			},
		}},
		Daemon: DaemonInfo{OnBattery: onBattery},
	}

	return data, nil
}

func getSystemPowerStatus() (*systemPowerStatus, error) {
	var status systemPowerStatus
	r1, _, callErr := getSystemPowerStatusProc.Call(uintptr(unsafe.Pointer(&status)))
	if r1 == 0 {
		if callErr != syscall.Errno(0) {
			return nil, fmt.Errorf("failed to query system power status: %w", callErr)
		}
		return nil, fmt.Errorf("failed to query system power status")
	}

	return &status, nil
}

func batteryWarningLevel(flags byte) string {
	if flags&4 != 0 {
		return "critical"
	}
	if flags&2 != 0 {
		return "low"
	}
	return "none"
}

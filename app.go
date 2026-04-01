package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
)

type App struct {
	ctx          context.Context
	config       AppConfig
	configMutex  sync.RWMutex
	lastNotified time.Time
}

type AppConfig struct {
	UpperLimit float64 `json:"upperLimit"`
	LowerLimit float64 `json:"lowerLimit"`
	Interval   int     `json:"intervalMinutes"`
}

type BatterySnapshot struct {
	OK             bool    `json:"ok"`
	Error          string  `json:"error,omitempty"`
	DeviceName     string  `json:"deviceName,omitempty"`
	Percentage     float64 `json:"percentage"`
	State          string  `json:"state,omitempty"`
	StateLabel     string  `json:"stateLabel,omitempty"`
	OnBattery      bool    `json:"onBattery"`
	ACOnline       bool    `json:"acOnline"`
	Technology     string  `json:"technology,omitempty"`
	Capacity       float64 `json:"capacity,omitempty"`
	Energy         float64 `json:"energy,omitempty"`
	EnergyRate     float64 `json:"energyRate,omitempty"`
	Voltage        float64 `json:"voltage,omitempty"`
	ChargeCycles   int     `json:"chargeCycles,omitempty"`
	WarningLevel   string  `json:"warningLevel,omitempty"`
	UpdatedAgoText string  `json:"updatedAgoText,omitempty"`
}

func NewApp() *App {
	a := &App{
		config: AppConfig{
			UpperLimit: 80.0,
			LowerLimit: 20.0,
			Interval:   10,
		},
	}
	a.loadConfig()
	return a
}

func (a *App) getConfigPath() string {
	configDir, _ := os.UserConfigDir()
	dir := filepath.Join(configDir, "custos")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

func (a *App) loadConfig() {
	path := a.getConfigPath()
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &a.config)
	}
}

func (a *App) SaveConfig(config AppConfig) {
	a.configMutex.Lock()
	a.config = config
	a.configMutex.Unlock()

	path := a.getConfigPath()
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(path, data, 0644)
}

func (a *App) GetConfig() AppConfig {
	a.configMutex.RLock()
	defer a.configMutex.RUnlock()
	return a.config
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	beeep.AppName = "Custos"
	go a.monitorBattery()
	a.installDesktopFiles()
}

func (a *App) installDesktopFiles() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	// Skip writing if it's just `go run ...` which often outputs to /tmp/go-build*
	if filepath.Base(execPath) == "main" || filepath.Ext(execPath) != "" {
		return
	}

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=Custos
Comment=System Battery Status & Notifications
Exec=%s
Terminal=false
Type=Application
Categories=Utility;HardwareSettings;
Icon=battery
Keywords=power;battery;charging;status;
`, execPath)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	appsDir := filepath.Join(homeDir, ".local", "share", "applications")
	os.MkdirAll(appsDir, 0755)
	appDesktopPath := filepath.Join(appsDir, "custos.desktop")
	if _, err := os.Stat(appDesktopPath); os.IsNotExist(err) {
		os.WriteFile(appDesktopPath, []byte(desktopContent), 0644)
	}

	autostartDir := filepath.Join(homeDir, ".config", "autostart")
	os.MkdirAll(autostartDir, 0755)
	startDesktopPath := filepath.Join(autostartDir, "custos.desktop")
	if _, err := os.Stat(startDesktopPath); os.IsNotExist(err) {
		os.WriteFile(startDesktopPath, []byte(desktopContent), 0644)
	}
}

func (a *App) monitorBattery() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			upowerData, err := QueryUPower()
			if err != nil {
				continue
			}

			percentage := upowerData.GetBatteryPercentage()
			isOnBattery := upowerData.IsOnBattery()

			a.configMutex.RLock()
			thresholdHigh := a.config.UpperLimit
			thresholdLow := a.config.LowerLimit
			interval := time.Duration(a.config.Interval) * time.Minute
			a.configMutex.RUnlock()

			now := time.Now()

			if now.Sub(a.lastNotified) < interval {
				continue
			}

			if !isOnBattery && percentage >= thresholdHigh {
				title := "Battery Charged"
				message := fmt.Sprintf("Battery has reached %.0f%%. Please disconnect the charger.", percentage)
				err := beeep.Notify(title, message, "")
				if err == nil {
					a.lastNotified = now
				}
			} else if isOnBattery && percentage <= thresholdLow {
				title := "Battery Low"
				message := fmt.Sprintf("Battery has dropped to %.0f%%. Please connect the charger.", percentage)
				err := beeep.Notify(title, message, "")
				if err == nil {
					a.lastNotified = now
				}
			}
		}
	}
}

func (a *App) GetBatteryStatus() string {
	upowerData, err := QueryUPower()
	if err != nil {
		fmt.Println("Error querying UPower:", err)
		snapshot := BatterySnapshot{OK: false, Error: err.Error()}
		payload, _ := json.Marshal(snapshot)
		return string(payload)
	}

	battery := upowerData.GetBattery()
	display := upowerData.GetDisplayDevice()
	linePower := upowerData.GetACAdapter()

	snapshot := BatterySnapshot{
		OK:         true,
		OnBattery:  upowerData.IsOnBattery(),
		ACOnline:   linePower != nil && linePower.LinePower != nil && linePower.LinePower.Online,
		Percentage: upowerData.GetBatteryPercentage(),
	}

	if battery != nil && battery.Battery != nil {
		snapshot.DeviceName = battery.Model
		if snapshot.DeviceName == "" {
			snapshot.DeviceName = battery.NativePath
		}
		snapshot.State = battery.Battery.State
		snapshot.StateLabel = batteryStateLabel(battery.Battery.State)
		snapshot.Technology = battery.Battery.Technology
		snapshot.Capacity = battery.Battery.Capacity
		snapshot.Energy = battery.Battery.Energy
		snapshot.EnergyRate = battery.Battery.EnergyRate
		snapshot.Voltage = battery.Battery.Voltage
		snapshot.ChargeCycles = battery.Battery.ChargeCycles
		snapshot.WarningLevel = battery.Battery.WarningLevel
		return marshalSnapshot(snapshot)
	}

	if display != nil && display.DisplayDevice != nil {
		snapshot.DeviceName = "Display battery"
		snapshot.State = display.DisplayDevice.State
		snapshot.StateLabel = batteryStateLabel(display.DisplayDevice.State)
		snapshot.Energy = display.DisplayDevice.Energy
		snapshot.EnergyRate = display.DisplayDevice.EnergyRate
		snapshot.WarningLevel = display.DisplayDevice.WarningLevel
		return marshalSnapshot(snapshot)
	}

	snapshot.OK = false
	snapshot.Error = "no battery data found"
	return marshalSnapshot(snapshot)
}

func marshalSnapshot(snapshot BatterySnapshot) string {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Sprintf("{\"ok\":false,\"error\":%q}", err.Error())
	}
	return string(payload)
}

func batteryStateLabel(state string) string {
	switch state {
	case "charging":
		return "Charging"
	case "discharging":
		return "Discharging"
	case "fully-charged":
		return "Fully charged"
	case "pending-charge":
		return "Pending charge"
	case "pending-discharge":
		return "Pending discharge"
	default:
		return "Unknown"
	}
}

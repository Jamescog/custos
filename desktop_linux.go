//go:build linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) installDesktopFiles() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	// Skip when running in dev mode (wails dev runs as "main" or has an extension)
	if filepath.Base(execPath) == "main" || filepath.Ext(execPath) != "" {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// Install the app icon into the user's hicolor icon theme so that
	// "Icon=custos" resolves correctly in any desktop environment.
	iconDir := filepath.Join(homeDir, ".local", "share", "icons", "hicolor", "256x256", "apps")
	if err := os.MkdirAll(iconDir, 0755); err == nil {
		// Always overwrite so the icon stays in sync with the installed binary.
		os.WriteFile(filepath.Join(iconDir, "custos.png"), appIcon, 0644)
	}

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Name=Custos
GenericName=Battery Monitor
Comment=System Battery Status & Notifications
Exec=%s
Terminal=false
Type=Application
Categories=Utility;HardwareSettings;
Icon=custos
Keywords=power;battery;charging;status;
StartupNotify=false
`, execPath)

	appsDir := filepath.Join(homeDir, ".local", "share", "applications")
	os.MkdirAll(appsDir, 0755)
	os.WriteFile(filepath.Join(appsDir, "custos.desktop"), []byte(desktopContent), 0644)

	autostartDir := filepath.Join(homeDir, ".config", "autostart")
	os.MkdirAll(autostartDir, 0755)
	os.WriteFile(filepath.Join(autostartDir, "custos.desktop"), []byte(desktopContent), 0644)
}

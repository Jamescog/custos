package main

import (
	"context"
	"embed"
	"os"
	"sync"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed trayicon.png
var trayIcon []byte

//go:embed build/appicon.png
var appIcon []byte

var (
	appCtx context.Context
	ctxMu  sync.RWMutex
)

func main() {
	if NotifyExistingInstance() {
		return
	}

	app := NewApp()

	systray.Register(onSystrayReady, onSystrayExit)

	err := wails.Run(&options.App{
		Title:         "custos",
		Width:         420,
		Height:        560,
		DisableResize: true,
		AlwaysOnTop:   true,
		Frameless:     true,
		StartHidden:   false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(c context.Context) {
			ctxMu.Lock()
			appCtx = c
			ctxMu.Unlock()
			app.startup(c)
			StartIPC(c)
		},
		OnBeforeClose: func(ctx context.Context) bool {
			// Return true to cancel the close event and keep the app running.
			// The × button in the UI calls WindowHide directly.
			return true
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func onSystrayReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip("Custos - Battery Monitor")

	showItem := systray.AddMenuItem("Show", "Show Custos")
	quitItem := systray.AddMenuItem("Quit", "Quit Custos")

	for {
		select {
		case <-showItem.ClickedCh:
			ctxMu.RLock()
			c := appCtx
			ctxMu.RUnlock()
			if c != nil {
				runtime.WindowShow(c)
				runtime.WindowUnminimise(c)
				runtime.WindowCenter(c)
			}
		case <-quitItem.ClickedCh:
			os.Exit(0)
		}
	}
}

func onSystrayExit() {
}

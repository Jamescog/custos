# Custos

Custos is a sleek, lightweight system-tray/background battery monitor and notifier built with Wails and React. It was designed to replace heavy web-based dashboards with a frameless, native-feeling utility widget.

> **Note:** Currently Custos fully supports Linux environments (using `UPower` and D-Bus notifications).

## Features

- **Stealth Background Process:** Monitor battery metrics silently without keeping a window open.
- **Smart Single Instance Mode:** Clicking the app shortcut bounces the hidden window up rather than opening multiple copies.
- **Customizable Limits:** Set your own thresholds to get alerted when your battery dips too low or hits peak charge. 
- **Notification Cooldowns:** Ensures you won't be spammed with consecutive alerts.

## Installation

Ensure you have [Go](https://go.dev/) and [Wails v2](https://wails.io/) installed.

```bash
git clone https://github.com/Jamescog/custos.git
cd custos
wails build
```

On your first run of the compiled binary, it will automatically generate a `.desktop` integration file so that the application opens silently on boot and is easily searchable through your desktop environment's app drawer.

## Usage

* `Settings (⚙️)`: Modify the charging limit, discharging limit, and repeated notification delays securely.
* `Hide (-)`: Dismiss the panel and return to silent tracking mode.
* `Quit (x)`: Fully exit the daemon process.

---

Built with Go, React, and Wails. 

# svcm (Service Manager)
[AUR Version](https://img.shields.io/aur/version/svcm)


**svcm** is a lightweight systemd service manager for Wayland, written in Go. It offers a unified experience across multiple interfaces:
- **CLI**: Standard terminal commands.
- **TUI**: Interactive k9s-style terminal UI.
- **GUI**: System tray and window management.
- **MCP**: Integration for AI assistants.

## Installation

```bash
git clone https://github.com/your/svcm.git
cd svcm
go build -o svcm ./src/cmd/svcm
```

## Usage

### User Services (Default)
Manage services for the current user (Systemd User Bus).

```bash
# Interact with TUI
./svcm tui

# List services
./svcm list

# Start/Stop
./svcm start pipewire
./svcm stop pipewire

# View Logs
./svcm logs pipewire

# Launch GUI (Tray)
./svcm gui
```

### System Services (Privileged)
Manage system-wide services (Systemd System Bus). Requires `sudo` and the `--privileged` flag.

```bash
# TUI for system services
sudo ./svcm tui --privileged

# List system services
sudo ./svcm list -P

# Restart a system service
sudo ./svcm restart bluetooth -P
```

## Modules

The project is structured into modular components in `src/internal`:
- **Core**: Systemd DBus interactions.
- **CLI**: Cobra-based command line interface.
- **TUI**: `tview`-based terminal UI.
- **GUI**: `fyne`-based graphical UI.
- **MCP**: Model Context Protocol server.

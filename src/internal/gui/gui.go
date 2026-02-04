package gui

import (
	"log"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"svcm/src/internal/core"
)

func Run(systemMode bool) {
	a := app.NewWithID("com.arya.lsysctl")
	w := a.NewWindow("lsysctl - Service Manager")

	// Verify we can support system tray
	var desk desktop.App
	if d, ok := a.(desktop.App); ok {
		desk = d
	} else {
		log.Println("System tray not supported on this platform")
	}

	// Service Manager Connection
	manager, err := core.NewSystemdManager(systemMode)
	if err != nil {
		w.SetContent(widget.NewLabel("Failed to connect to systemd: " + err.Error()))
		w.ShowAndRun()
		return
	}
	defer manager.Close()

	// UI Components
	listContainer := container.NewVBox()
	scroll := container.NewVScroll(listContainer)

	statusLabel := widget.NewLabel("Ready")

	refreshServices := func() {
		listContainer.Objects = nil
		services, err := manager.ListServices()
		if err != nil {
			statusLabel.SetText("Error listing services: " + err.Error())
			return
		}

		// Sort by name
		sort.Slice(services, func(i, j int) bool {
			return services[i].Name < services[j].Name
		})

		for _, s := range services {
			// Capture variable for closure
			svcName := s.Name
			svcActive := s.ActiveState

			nameLabel := widget.NewLabel(svcName)
			stateLabel := widget.NewLabel(svcActive)
			descLabel := widget.NewLabel(s.Description)

			// Truncate description if too long
			if len(s.Description) > 50 {
				descLabel.SetText(s.Description[:47] + "...")
			}

			var actionBtn *widget.Button
			if svcActive == "active" {
				actionBtn = widget.NewButton("Stop", func() {
					if err := manager.StopService(svcName); err != nil {
						statusLabel.SetText("Failed to stop " + svcName + ": " + err.Error())
					} else {
						statusLabel.SetText("Stopped " + svcName)
						// Simple refresh after action
						// In a real app we might watch signals
					}
				})
			} else {
				actionBtn = widget.NewButton("Start", func() {
					if err := manager.StartService(svcName); err != nil {
						statusLabel.SetText("Failed to start " + svcName + ": " + err.Error())
					} else {
						statusLabel.SetText("Started " + svcName)
					}
				})
			}

			// Row layout
			row := container.New(layout.NewGridLayout(4), nameLabel, stateLabel, descLabel, actionBtn)
			listContainer.Add(row)
		}
		listContainer.Refresh()
	}

	// Initial load
	refreshServices()

	refreshBtn := widget.NewButton("Refresh", refreshServices)

	// Main Layout
	content := container.NewBorder(
		nil,
		container.NewVBox(statusLabel, refreshBtn),
		nil, nil,
		scroll,
	)
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))

	// Setup Tray
	if desk != nil {
		menu := fyne.NewMenu("lsysctl",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}),
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	w.ShowAndRun()
}

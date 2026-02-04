package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"svcm/src/internal/core"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	tviewApp *tview.Application
	table    *tview.Table
	infoBox  *tview.TextView
	manager  *core.SystemdManager
	services []core.ServiceUnit
	filter   string
}

func Run(systemMode bool) error {
	manager, err := core.NewSystemdManager(systemMode)
	if err != nil {
		return err
	}
	// We don't defer manager.Close() here because tview takes over the UI loop.
	// We should close it on exit.

	app := &App{
		tviewApp: tview.NewApplication(),
		table:    tview.NewTable(),
		infoBox:  tview.NewTextView(),
		manager:  manager,
	}

	return app.run()
}

func (a *App) run() error {
	a.refreshServices()

	// Auto refresh every 2 seconds
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.tviewApp.QueueUpdateDraw(func() {
					a.refreshServices()
				})
			}
		}
	}()

	if err := a.tviewApp.SetRoot(a.layout(), true).EnableMouse(true).Run(); err != nil {
		return err
	}

	// Cleanup
	a.manager.Close()
	return nil
}

func (a *App) layout() tview.Primitive {
	// Header
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("lsysctl - System Service Manager (k9s-style)")
	header.SetBackgroundColor(tcell.ColorDarkBlue)
	header.SetTextColor(tcell.ColorWhite)

	// Table styling
	a.table.SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 1)

	// Footer/Help
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]s[white]tart [yellow]x[white]stop [yellow]r[white]estart [yellow]l[white]ogs [yellow]/[white]filter [yellow]q[white]uit")
	footer.SetBackgroundColor(tcell.ColorDarkGray)

	// Main Flex Layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(a.table, 0, 1, true).
		AddItem(footer, 1, 1, false)

	// Keybindings for table
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := a.table.GetSelection()
		if row <= 0 || row >= a.table.GetRowCount() {
			return event
		}

		// Context is the service name stored in the first cell
		cell := a.table.GetCell(row, 0)
		serviceName := cell.Text

		switch event.Rune() {
		case 's':
			a.performAction("Starting", serviceName, a.manager.StartService)
		case 'x':
			a.performAction("Stopping", serviceName, a.manager.StopService)
		case 'r':
			a.performAction("Restarting", serviceName, a.manager.RestartService)
		case 'l':
			a.showLogs(serviceName)
		case '/':
			// Filter implementation placeholder
			// In a real k9s-like app, this opens an input field
		case 'q':
			a.tviewApp.Stop()
		}

		// Also handle Enter for logs
		if event.Key() == tcell.KeyEnter {
			a.showLogs(serviceName)
		}

		return event
	})

	return flex
}

func (a *App) refreshServices() {
	services, err := a.manager.ListServices()
	if err != nil {
		return
	}
	a.services = services

	// Save selection
	row, _ := a.table.GetSelection()
	selectedName := ""
	if row > 0 && row < a.table.GetRowCount() {
		selectedName = a.table.GetCell(row, 0).Text
	}

	a.table.Clear()

	// Header Row
	headers := []string{"NAME", "ACTIVE", "SUB", "LOAD", "DESCRIPTION"}
	for c, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)
		a.table.SetCell(0, c, cell)
	}

	// Data Rows
	currentRow := 1
	newSelectionRow := 0

	for _, s := range services {
		// Apply Filter if needed
		if a.filter != "" && !strings.Contains(s.Name, a.filter) {
			continue
		}

		color := tcell.ColorGreen
		if s.ActiveState != "active" {
			color = tcell.ColorGray
		}
		if s.ActiveState == "failed" {
			color = tcell.ColorRed
		}

		a.table.SetCell(currentRow, 0, tview.NewTableCell(s.Name).SetTextColor(color))
		a.table.SetCell(currentRow, 1, tview.NewTableCell(s.ActiveState).SetTextColor(color))
		a.table.SetCell(currentRow, 2, tview.NewTableCell(s.SubState).SetTextColor(color))
		a.table.SetCell(currentRow, 3, tview.NewTableCell(s.LoadState))
		a.table.SetCell(currentRow, 4, tview.NewTableCell(s.Description))

		if s.Name == selectedName {
			newSelectionRow = currentRow
		}
		currentRow++
	}

	// Restore selection or default to 1
	if newSelectionRow > 0 {
		a.table.Select(newSelectionRow, 0)
	} else if currentRow > 1 {
		// if we lost selection, maybe keep index or reset?
		// keeping pure index might lead to wrong selection if list changes heavily
		if row > 0 && row < currentRow {
			a.table.Select(row, 0)
		} else {
			a.table.Select(1, 0)
		}
	}
}

func (a *App) performAction(actionVerb string, name string, actionFunc func(string) error) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("%s %s...", actionVerb, name)).
		AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.tviewApp.SetRoot(a.layout(), true)
		})

	go func() {
		err := actionFunc(name)
		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				modal.SetText(fmt.Sprintf("Error: %v", err)).
					AddButtons([]string{"OK"})
			} else {
				// Return to main layout on success after brief pause or immediately
				a.tviewApp.SetRoot(a.layout(), true)
				a.refreshServices()
			}
		})
	}()

	a.tviewApp.SetRoot(modal, false)
}

func (a *App) showLogs(name string) {
	// Show a modal with logs
	// We will cheat and just run journalctl and capture output for now
	// Ideally this should stream data into the TextView

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			// Scroll to end
		})

	textView.SetBorder(true).SetTitle(" Logs: " + name + " (Press Esc to close) ")

	// Capture Key to close
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.tviewApp.SetRoot(a.layout(), true)
			return nil
		}
		return event
	})

	// Initial Load
	go func() {
		// Run journalctl -n 200
		// Using exec.Command
		cmd := exec.Command("journalctl", "--user", "-u", name, "-n", "200", "--no-pager")
		out, err := cmd.CombinedOutput()

		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				textView.SetText(fmt.Sprintf("Error fetching logs: %v", err))
			} else {
				textView.SetText(string(out))
				textView.ScrollToEnd()
			}
		})
	}()

	a.tviewApp.SetRoot(textView, true)
}

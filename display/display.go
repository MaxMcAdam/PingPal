package display

import (
	"fmt"
	"time"

	"github.com/gbin/goncurses"
	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/record"
)

type App struct {
	screen  *goncurses.Window
	mainWin *goncurses.Window
	running bool
}

func NewApp() (*App, error) {
	// Initialize ncurses
	screen, err := goncurses.Init()
	if err != nil {
		return nil, err
	}

	// Enable colors if available
	if err := goncurses.StartColor(); err != nil {
		goncurses.End()
		return nil, err
	}

	// Don't echo pressed keys
	goncurses.Echo(false)

	// Enable special keys (like arrow keys)
	screen.Keypad(true)

	// Set non-blocking input
	screen.Timeout(100) // 100ms timeout

	// Get screen dimensions
	maxY, maxX := screen.MaxYX()

	// Create main window (full screen)
	mainWin := screen.Sub(maxY, maxX, 0, 0)
	if mainWin == nil {
		goncurses.End()
		return nil, fmt.Errorf("Error creating main window.")
	}

	return &App{
		screen:  screen,
		mainWin: mainWin,
		running: true,
	}, nil
}

func (app *App) Cleanup() {
	if app.mainWin != nil {
		app.mainWin.Delete()
	}
	goncurses.End()
}

func GenOutputValues(address string, record *record.AddressRecord) (latAvg string, lossAvg string, errStr string) {
	// Get the lock for this record
	avgLat, pktLossAvg, _ := record.GetCurrentStats()

	if record.Health.ErrorCount > 0 {
		// If there was an error on the most recent latency check, display that
		errStr = fmt.Sprintf("%s Error: %s ", address, record.Health.Status)
	} else if len(record.PacketDQ) > 0 && record.PacketsSentSuccess > 0 {
		// Otherwise return a string with the average latency and packet loss
		latAvg = "0.0"
		if record.LatAvgCount > 0 {
			latAvg = fmt.Sprintf("%.2f", avgLat)
		}
		lossAvg = fmt.Sprintf("%.2f%%", 100*pktLossAvg)
	}
	return
}

func (app *App) UpdateMainDisplay(addresses *[]*record.AddressRecord, packetDropTime uint64) {
	// Clear the main window
	app.mainWin.Clear()

	// Draw border
	app.mainWin.Border('|', '|', '-', '-', 0, 0, 0, 0)

	// Get dimensions
	maxY, _ := app.mainWin.MaxYX()

	app.mainWin.MovePrint(1, 2, fmt.Sprintf("IP Address        Latency (Avg %vS)     Lost Packets (Avg %vS)", packetDropTime, packetDropTime))

	// Display address stats
	line := 2
	for _, rec := range *addresses {
		app.mainWin.MovePrint(line, 2, rec.Address)
		latAvg, lossAvg, errStr := GenOutputValues(rec.Address, rec)
		if errStr != "" {
			app.mainWin.MovePrint(line, 10, errStr)
		} else {
			app.mainWin.MovePrint(line, 20, latAvg)
			app.mainWin.MovePrint(line, 43, lossAvg)
		}
		line++
	}
	app.mainWin.MovePrint(maxY-1, 2, "Press 'q' to quit")

	app.mainWin.Refresh()
}

func (app *App) HandleInput() {
	key := app.screen.GetChar()

	// Main window is active
	switch key {
	case 'q', 'Q':
		app.running = false
	}
}

func (app *App) Run(addresses *[]*record.AddressRecord, settings *config.SessionSettings) {
	for app.running {
		// Update the main display
		app.UpdateMainDisplay(addresses, settings.PktDropTimeS)

		// Handle input
		app.HandleInput()

		// Small delay to control update rate
		time.Sleep(5 * time.Millisecond)
	}
}

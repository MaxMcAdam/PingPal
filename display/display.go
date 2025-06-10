package display

import (
	"fmt"
	"time"

	"github.com/gbin/goncurses"
	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/latency"
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

func GenOutputString(address string, record *latency.AddressRecord) string {
	str := fmt.Sprintf("%s", address)

	// Get the lock for this record
	record.Lock.Lock()
	defer record.Lock.Unlock()

	if len(record.PacketDQ) > 0 && record.PacketDQ[len(record.PacketDQ)-1].Err != nil {
		// If there was an error on the most recent latency check, display that
		str = fmt.Sprintf("%s Error: %v", address, record.PacketDQ[len(record.PacketDQ)-1].Err)
	} else if len(record.PacketDQ) > 0 && record.PacketsSentSuccess > 0 {
		// Otherwise return a string with the average latency and packet loss
		latAvg := 0.0
		if record.LatAvgCount > 0 {
			latAvg = record.LatAvgSum / record.LatAvgCount
		}
		str = fmt.Sprintf("%s    %f              %f%%", address, latAvg, float64(record.PacketsDropped)/float64(record.PacketsSentSuccess))
	}
	return str
}

func (app *App) UpdateMainDisplay(addresses *[]*latency.AddressRecord, packetDropTime uint64) {
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
		app.mainWin.MovePrint(line, 2, GenOutputString(rec.Address, rec))
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

func (app *App) Run(addresses *[]*latency.AddressRecord, settings *config.SessionSettings) {
	for app.running {
		// Update the main display
		app.UpdateMainDisplay(addresses, settings.PktDropTimeS)

		// Handle input
		app.HandleInput()

		// Small delay to control update rate
		time.Sleep(5 * time.Millisecond)
	}
}

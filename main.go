package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/gbin/goncurses"
	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/display"
	"github.com/seanmcadam/PingPal/latency"
)

func main() {
	stdscr, err := goncurses.Init()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	defer goncurses.End()

	goncurses.Echo(false)  // Don't echo input characters
	goncurses.CBreak(true) // Don't wait for Enter key
	goncurses.Cursor(0)    // Hide cursor
	stdscr.Keypad(true)    // Enable special keys

	input, err := config.ParseFlagsWithValidation()
	if err != nil {
		fmt.Errorf("Error parsing input flags: %v", err)
		os.Exit(1)
	}

	sessAddr := map[string]*latency.AddressRecord{}

	for _, a := range input.Addresses {
		newRec := latency.AddressRecord{}
		sessAddr[a] = &newRec
	}

	for k, v := range sessAddr {
		go func() {
			latency.MonitorLatency(k, v, &input.Settings)
		}()
	}

	go func() {
		display.UpdateScreen(&sessAddr, stdscr, &input.Settings)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	goncurses.End()
}

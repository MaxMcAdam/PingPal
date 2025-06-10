package main

import (
	"fmt"
	"os"

	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/display"
	"github.com/seanmcadam/PingPal/latency"
)

func main() {
	input, err := config.ParseFlagsWithValidation()
	if err != nil {
		fmt.Errorf("Error parsing input flags: %v", err)
		os.Exit(1)
	}

	sessAddr := []*latency.AddressRecord{}

	for _, a := range input.Addresses {
		newRec := latency.AddressRecord{Address: a}
		sessAddr = append(sessAddr, &newRec)
	}

	for _, v := range sessAddr {
		go func() {
			latency.MonitorLatency(v.Address, v, &input.Settings)
		}()
	}

	app, err := display.NewApp()
	if err != nil {
		fmt.Printf("Error initializing application:", err)
		os.Exit(1)
	}
	defer app.Cleanup()

	app.Run(&sessAddr, &input.Settings)
}

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/display"
	"github.com/seanmcadam/PingPal/latency"
	"github.com/seanmcadam/PingPal/record"
)

func main() {
	input, err := config.ParseFlagsWithValidation()
	if err != nil {
		fmt.Printf("Error parsing input flags: %v\n", err)
		os.Exit(1)
	}

	capNetRaw, err := config.HasCapNetRaw()
	if err != nil {
		fmt.Printf("Error checking for CAP_NET_RAW: %v", err)
		os.Exit(1)
	} else if !capNetRaw {
		fmt.Printf("Creating a raw socket for ICMP packets requires elevated privileges.")
		os.Exit(1)
	}

	sessAddr := []*record.AddressRecord{}

	for _, a := range input.Addresses {
		newRec := record.AddressRecord{Address: a}
		sessAddr = append(sessAddr, &newRec)
	}

	for _, v := range sessAddr {
		time.Sleep(1 * time.Second)
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

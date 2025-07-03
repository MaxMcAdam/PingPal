package config

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Alternative implementation with more explicit validation
func ParseFlagsWithValidation() (*Input, error) {
	settings := &SessionSettings{}

	// Create custom FlagSet for more control
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	fs.Uint64Var(&settings.DisplayRefreshTimeS, "d", 1, "Display refresh rate in seconds (uint64)")
	fs.Uint64Var(&settings.PktDropTimeS, "p", 30, "Time to average latency and packet loss across in seconds (uint64)")
	fs.Uint64Var(&settings.LatencyCheckIntervalS, "l", 1, "Latency check interval seconds (uint64)")
	fs.Uint64Var(&settings.ConnectionTimeoutS, "c", 500, "Connection timeout in milliseconds (uint64)")

	fs.Usage = usage

	// Parse with custom error handling
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	addr := fs.Args()

	if len(addr) == 0 {
		return nil, fmt.Errorf("Must provide at least one address to monitor.")
	}

	input := Input{Addresses: addr, Settings: *settings}

	return &input, nil
}

const usageHeader = `Usage: pingpal [c] [d] [l] [p] (ip address...)
  -c uint
    	Connection timeout in milliseconds (uint64) (default 120)
  -d uint
    	Display refresh rate in seconds (uint64) (default 1)
  -l uint
    	Latency check interval seconds (uint64) (default 1)
  -p uint
    	Time to average latency and packet loss across in seconds (uint64) (default 300)`

func usage() {
	fmt.Println(usageHeader)
}

type Input struct {
	Settings  SessionSettings
	Addresses []string
}

// The settings for this session including update time preferences
type SessionSettings struct {
	DisplayRefreshTimeS   uint64 // how often to refresh the display in seconds
	PktDropTimeS          uint64 // time to retain packets and average latency and loss across in seconds
	LatencyCheckIntervalS uint64 // how often to check latency in seconds
	ConnectionTimeoutS    uint64 // how long to wait for a reply to the ICMP packet in seconds
}

func HasCapNetRaw() (bool, error) {
	caps := unix.Getuid()

	// Check if running as root (has all capabilities)
	if caps == 0 {
		return true, nil
	}

	// For non-root, check specific capability
	hdr := unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     0, // 0 means current process
	}

	var data [2]unix.CapUserData
	err := unix.Capget(&hdr, &data[0])
	if err != nil {
		return false, err
	}

	// CAP_NET_RAW is capability 13
	capNetRaw := uint32(1 << 13)
	return (data[0].Effective & capNetRaw) != 0, nil
}

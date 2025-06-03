package display

import (
	"fmt"
	"time"

	"github.com/gbin/goncurses"
	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/latency"
)

func GenOutputString(address string, record *latency.AddressRecord) string {
	str := fmt.Sprintf("%s", address)

	record.Lock.Lock()
	defer record.Lock.Unlock()

	if len(record.PacketDQ) > 0 && record.PacketDQ[len(record.PacketDQ)-1].Err != nil {
		str = fmt.Sprintf("%s Error: %v", address, record.PacketDQ[len(record.PacketDQ)-1].Err)
	} else if len(record.PacketDQ) > 0 && record.PacketsSentSuccess > 0 {
		latAvg := 0.0
		if record.LatAvgCount > 0 {
			latAvg = record.LatAvgSum / record.LatAvgCount
		}
		str = fmt.Sprintf("%s Latency: %f Loss: %f%%", address, latAvg, float64(record.PacketsDropped)/float64(record.PacketsSentSuccess))
	}
	return str
}

func UpdateScreen(SessionAddresses *map[string]*latency.AddressRecord, stdscr *goncurses.Window, sessConfig *config.SessionSettings) {
	for true {
		stdscr.Clear()
		for addr, rec := range *SessionAddresses {
			stdscr.Println(GenOutputString(addr, rec))
			stdscr.Refresh()
		}
		time.Sleep(time.Duration(sessConfig.DisplayRefreshTimeS * uint64(time.Second)))
	}
}

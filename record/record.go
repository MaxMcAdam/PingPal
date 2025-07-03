package record

import (
	"fmt"
	"sync"
	"time"
)

// PackersSent and PacketsLost are updated everytime a new entry is added to the PacketDQ
// The key for the PacketDQ is an id number
type AddressRecord struct {
	Address            string
	Lock               sync.Mutex
	PacketsSentSuccess uint64
	PacketsDropped     uint64
	PacketDQ           []PacketRecord
	LatAvgSum          float64
	LatAvgCount        float64
	Health             ThreadHealth
}

type ThreadHealth struct {
	LastSeen   time.Time
	ErrorCount int
	Status     string
}

type PacketRecord struct {
	TimeSent time.Time
	Latency  float64
	Dropped  bool
}

func (a *AddressRecord) AddPacketRecord(sentTime time.Time, err error, latency float64, dropped bool, windowDuration time.Duration) {
	a.Lock.Lock()
	defer a.Lock.Unlock()

	if err != nil {
		a.Health.Status = fmt.Sprintf("%v", err)
		a.Health.ErrorCount++
	} else {
		if a.Health.ErrorCount > 0 {
			a.Health.Status = "Healthy"
			a.Health.ErrorCount = 0
		}
		a.Health.LastSeen = time.Now()
	}

	if err != nil {
		return
	}

	// Add new record
	newRecord := PacketRecord{
		TimeSent: sentTime,
		Latency:  latency,
		Dropped:  dropped,
	}
	a.PacketDQ = append(a.PacketDQ, newRecord)

	a.limitDQSize()

	// Clean up expired records and recalculate stats
	a.cleanupAndRecalculate(windowDuration)
}

func (a *AddressRecord) limitDQSize() {
	if len(a.PacketDQ) > 1000 {
		a.PacketDQ = a.PacketDQ[1:]
	}
}

// Separate method for cleanup that recalculates everything to avoid drift
func (a *AddressRecord) cleanupAndRecalculate(windowDuration time.Duration) {
	cutoff := time.Now().Add(-windowDuration)

	// Reset counters
	a.PacketsSentSuccess = 0
	a.PacketsDropped = 0
	a.LatAvgSum = 0
	a.LatAvgCount = 0

	// Find valid records and recalculate stats
	validIndex := 0
	for i, record := range a.PacketDQ {
		if record.TimeSent.After(cutoff) {
			validIndex = i
			break
		}
	}

	// Keep only valid records
	a.PacketDQ = a.PacketDQ[validIndex:]

	// Recalculate all stats from valid records
	for _, record := range a.PacketDQ {
		a.PacketsSentSuccess++
		if record.Dropped {
			a.PacketsDropped++
		} else {
			a.LatAvgSum += record.Latency
			a.LatAvgCount++
		}
	}
}

func (a *AddressRecord) GetCurrentStats() (avgLatency float64, packetLossRate float64, totalPackets uint64) {
	a.Lock.Lock()
	defer a.Lock.Unlock()

	if a.LatAvgCount > 0 {
		avgLatency = a.LatAvgSum / float64(a.LatAvgCount)
	}

	if a.PacketsSentSuccess > 0 {
		packetLossRate = float64(a.PacketsDropped) / float64(a.PacketsSentSuccess)
	}

	return avgLatency, packetLossRate, a.PacketsSentSuccess
}

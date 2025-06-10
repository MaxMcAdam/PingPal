package latency

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/seanmcadam/PingPal/config"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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
}

type PacketRecord struct {
	TimeSent time.Time
	Err      error
	Latency  float64
	Dropped  bool
}

// This loop monitors the latency for a single ip address.
// The latency is updated in a PacketLoss struct shared with the main process.
func MonitorLatency(ipAddr string, packets *AddressRecord, sessConfig *config.SessionSettings) {
	for true {
		// Check latency and update packet record queue
		latency, sentTime, dropped, err := CheckLatencyICMP(ipAddr, time.Duration(sessConfig.ConnectionTimeoutS*uint64(time.Second)))

		packets.Lock.Lock()

		packets.PacketDQ = append(packets.PacketDQ, PacketRecord{TimeSent: sentTime, Err: err, Latency: latency, Dropped: dropped})

		// Update address aggregates with new pack info
		if latency > 0 || dropped {
			packets.PacketsSentSuccess++
		}
		if dropped {
			packets.PacketsDropped++
		}

		packets.LatAvgSum += latency
		packets.LatAvgCount++

		if sessConfig.PktDropTimeS > 0 {
			firstValid := 0
			// Remove PacketRecords that have timed out
			// Oldest records are at the front of the slice
			validAfter := time.Now().Add(-time.Duration(sessConfig.PktDropTimeS * uint64(time.Second)))
			for i, PacketRec := range packets.PacketDQ {
				if PacketRec.TimeSent.Compare(validAfter) > 0 {
					firstValid = i
					break
				}
				packets.LatAvgSum -= PacketRec.Latency
				packets.LatAvgCount--
			}
			packets.PacketDQ = packets.PacketDQ[firstValid:]
		}

		packets.Lock.Unlock()

		time.Sleep(time.Duration(sessConfig.LatencyCheckIntervalS * uint64(time.Second)))
	}
}

// CheckLatency measures the round-trip time to the specified IP address using ICMP echo
// It returns the latency in milliseconds, time the icmp packet was sent, and any error encountered
func CheckLatencyICMP(ipAddr string, timeout time.Duration) (float64, time.Time, bool, error) {
	// Create a new ICMP connection
	// On most systems, you need to run as root/administrator to use raw sockets
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, time.Now(), false, fmt.Errorf("error creating ICMP listener: %w", err)
	}
	defer conn.Close()

	// Resolve the IP address
	dst, err := net.ResolveIPAddr("ip4", ipAddr)
	if err != nil {
		return 0, time.Now(), false, fmt.Errorf("error resolving IP address: %w", err)
	}

	// Create an ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff, // Use process ID as identifier
			Seq:  1,                    // Sequence number
			Data: []byte("ping test"),  // Payload data
		},
	}

	// Marshal the message into binary
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, time.Now(), false, fmt.Errorf("error marshaling ICMP message: %w", err)
	}

	// Set a deadline on the connection
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, time.Now(), false, fmt.Errorf("error setting deadline: %w", err)
	}

	// Record the start time
	startTime := time.Now()

	// Send the ICMP packet
	_, err = conn.WriteTo(msgBytes, dst)
	if err != nil {
		return 0, time.Now(), false, fmt.Errorf("error sending ICMP packet: %w", err)
	}

	// Prepare to receive the reply
	reply := make([]byte, 1500) // Buffer size for reply
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return 0, startTime, true, fmt.Errorf("error receiving ICMP reply: %w", err)
	}

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Parse the reply message
	parsedMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), reply[:n])
	if err != nil {
		return 0, startTime, false, fmt.Errorf("error parsing ICMP reply: %w", err)
	}

	// Check if it's an echo reply
	if parsedMsg.Type != ipv4.ICMPTypeEchoReply {
		return 0, startTime, false, fmt.Errorf("received non-echo reply message: %v", parsedMsg.Type)
	}

	return float64(elapsed.Microseconds()) / 1000.0, startTime, false, nil // Return latency in milliseconds
}

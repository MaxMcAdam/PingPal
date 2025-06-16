package latency

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/seanmcadam/PingPal/config"
	"github.com/seanmcadam/PingPal/record"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// This loop monitors the latency for a single ip address.
// The latency is updated in a PacketLoss struct shared with the main process.
func MonitorLatency(ipAddr string, packets *record.AddressRecord, sessConfig *config.SessionSettings) {
	for true {
		// Check latency and update packet record queue
		latency, sentTime, dropped, err := CheckLatencyICMP(ipAddr, time.Duration(sessConfig.ConnectionTimeoutS*uint64(time.Millisecond)))

		packets.AddPacketRecord(sentTime, err, latency, dropped, time.Duration(sessConfig.PktDropTimeS))

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
		return 0, startTime, true, nil
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

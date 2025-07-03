# PingPal

A real-time network latency monitoring tool that provides continuous ping monitoring with visual feedback and statistical analysis.

## Overview

PingPal is a terminal-based network monitoring application that tracks ping latency and packet loss across multiple IP addresses simultaneously. It provides a live updating display with rolling averages and packet loss statistics, making it ideal for network troubleshooting and performance monitoring.

## Features

- **Multi-target monitoring**: Monitor multiple IP addresses concurrently
- **Real-time display**: Live terminal interface with configurable refresh rates
- **Statistical analysis**: Rolling averages for latency and packet loss over configurable time windows
- **Low-level ICMP**: Uses raw ICMP sockets for accurate latency measurements
- **Configurable parameters**: Adjustable timeout, refresh rates, and averaging windows
- **Robust error handling**: Graceful handling of network failures and transient errors
- **Memory efficient**: Bounded packet history with time-based cleanup

## Installation

### Prerequisites

- Go 1.19 or later
- Root/administrator privileges (required for raw ICMP sockets)
- Linux/Unix-like operating system (for ncurses support)


### System Capabilities

PingPal requires `CAP_NET_RAW` capability to create raw ICMP sockets:

## Usage

### Basic Usage

```bash
# Monitor single address
./pingpal 8.8.8.8

# Monitor multiple addresses
./pingpal 8.8.8.8 1.1.1.1 
```

### Command Line Options

```bash
./pingpal [OPTIONS] <ip_address_1> [ip_address_2] ...

Options:
  -c uint    Connection timeout in milliseconds (default: 500)
  -d uint    Display refresh rate in seconds (default: 1)
  -l uint    Latency check interval in seconds (default: 5)
  -p uint    Packet loss averaging window in seconds (default: 30)
```

## Interface

### Controls

- **q** or **Q**: Quit the application
- **Ctrl+C**: Force quit

## Architecture

### Core Components

1. **Configuration (`config/`)**: Command-line parsing and validation
2. **Latency Monitoring (`latency/`)**: ICMP ping implementation
3. **Data Recording (`record/`)**: Thread-safe statistics tracking
4. **Display (`display/`)**: Terminal UI using ncurses

### Design Principles

#### Concurrent Architecture
- Each monitored address runs in its own goroutine
- Thread-safe data structures with mutex protection
- Non-blocking display updates

#### Memory Management
- Bounded packet history (default: 1000 packets per address)
- Time-based cleanup removes expired records
- Rolling statistics prevent memory leaks during long-running sessions

#### Error Resilience
- Graceful handling of network failures
- Retry logic for transient errors
- Clear error reporting in the interface

#### Performance Optimization
- Delta-based display updates
- Efficient statistics calculation
- Minimal memory allocation in hot paths

### Data Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌─────────────┐
│ Input Args  │ →  │ Config       │ →  │ Monitor     │ →  │ Display     │
│             │    │ Validation   │    │ Goroutines  │    │ Update Loop │
└─────────────┘    └──────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ↓
                                       ┌─────────────┐
                                       │ ICMP Socket │
                                       │ Operations  │
                                       └─────────────┘
                                              │
                                              ↓
                                       ┌─────────────┐
                                       │ Statistics  │
                                       │ Recording   │
                                       └─────────────┘
```

## Configuration

### Default Values

| Parameter | Default | Description |
|-----------|---------|-------------|
| Connection timeout | 500ms | Maximum time to wait for ICMP reply |
| Display refresh | 1s | How often to update the screen |
| Latency check interval | 5s | Time between ping attempts |
| Packet loss window | 30s | Time window for averaging statistics |

### Tuning Guidelines

#### High-Frequency Monitoring
```bash
# For detailed network analysis
./pingpal -l 1 -d 1 -p 60 target.com
```

#### Long-Term Monitoring
```bash
# For server uptime monitoring
./pingpal -l 30 -d 5 -p 3600 server1.com server2.com
```

#### Slow Network Connections
```bash
# For satellite or high-latency links
./pingpal -c 5000 -l 10 satellite.provider.com
```

## Troubleshooting

### Common Issues

#### Permission Denied
```
Error: error creating ICMP listener: permission denied
```
**Solution**: Run with sudo or set capabilities:
```bash
sudo setcap cap_net_raw+ep ./pingpal
```

#### Network Unreachable
```
Error: Network unreachable
```
**Solution**: Check routing, firewall rules, and network connectivity

#### No Response from Host
```
Error: timeout waiting for reply
```
**Solution**: 
- Increase timeout with `-c` flag
- Verify target host accepts ICMP
- Check for packet filtering

### Debug Information

PingPal provides error information directly in the interface:
- Network errors are displayed per-address
- Transient errors are retried automatically
- Persistent errors are clearly indicated

### Performance Considerations

- **Memory usage**: ~40KB per 1000 packets per address
- **CPU usage**: Minimal, most time spent in network I/O
- **Network impact**: Configurable ping frequency minimizes bandwidth usage

## Development

### Project Structure

```
PingPal/
├── config/          # Configuration and command-line parsing
│   └── config.go
├── display/         # Terminal UI and ncurses interface
│   └── display.go
├── latency/         # ICMP ping implementation
│   └── latency.go
├── record/          # Statistics tracking and data structures
│   └── record.go
├── main.go          # Application entry point
├── go.mod           # Go module dependencies
└── README.md        # This file
```

### Dependencies

- `golang.org/x/net/icmp`: ICMP protocol implementation
- `golang.org/x/net/ipv4`: IPv4 packet handling
- `golang.org/x/sys/unix`: Unix system calls for capabilities
- `github.com/gbin/goncurses`: ncurses terminal interface

Built with Go's networking libraries and ncurses for terminal UI.
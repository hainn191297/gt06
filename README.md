# GT06 GPS Tracking Protocol Implementation

A comprehensive implementation of the **GT06 GPS tracking protocol** using **Go** (server) and **Python** (client). This project provides a complete solution for handling GPS device communications, including login, heartbeat, location tracking, and alarm packet processing.

## Overview

The GT06 protocol is widely used in GPS tracking devices (dashcams, vehicle trackers, personal locators, etc.). This implementation includes:

- **Go Server**: TCP server that receives and processes GT06 protocol packets
- **Python Client**: Test client that sends GT06 formatted packets
- **Protocol Parser**: Robust parsing and validation of CONCOX/GT06 packets
- **Data Persistence**: MongoDB integration for storing device data and tracking information
- **CRC Validation**: CRC-ITU checksum calculation and validation for packet integrity

## Project Structure

```text
gt06/
├── main.go                          # Server entry point
├── client_send_packet.py            # Python test client
├── Makefile                         # Build configuration (unused)
├── protocol/
│   ├── concox.go                    # Protocol definitions and parsing
│   └── concox.c                     # C implementation (not used in Go build)
├── services/
│   ├── packet_service.go            # Service interface
│   ├── login_device.go              # Login packet handler
│   ├── heartbeat_packet.go          # Heartbeat packet handler
│   ├── alarm_packet.go              # Alarm packet handler
│   └── svc/
│       └── service_context.go       # Service dependency injection
├── tcp/                             # TCP server implementation
├── common/                          # Utility functions (CRC, helpers)
├── config/                          # Configuration structures
├── conf/                            # Configuration loading
├── database/                        # MongoDB models
├── docker-compose.yaml              # Docker orchestration
└── test/                            # Test utilities
```

## Features

### Supported Packet Types

1. **Login Packet (0x01)**
   - Device registration and authentication
   - IMEI, model code, time zone, language

2. **Heartbeat Packet (0x13)**
   - Periodic device status updates
   - Terminal info, voltage, battery, signal strength

3. **Location Packet (0x22)**
   - GPS location data with timestamp
   - Latitude, longitude, speed, course
   - Network information (MCC, MNC, LAC, Cell ID)

4. **Alarm Packet (0x26)**
   - Emergency or warning alerts
   - Similar to location packet but with alarm flags
   - Battery, signal, and terminal status

### Key Components

- **CRC Validation**: CRC-ITU checksum for data integrity
- **Binary Protocol**: Efficient binary packet format with headers and stop bits
- **MongoDB Integration**: Persistent storage of device and location data
- **Service Architecture**: Modular packet handling with service interfaces

## Getting Started

### Prerequisites

- Go 1.16+
- Python 3.7+
- MongoDB (Docker recommended)

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd gt06
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Start MongoDB (using Docker)**

   ```bash
   docker-compose up -d
   ```

4. **Run the server**

   ```bash
   go run main.go -c etc/server.yaml
   ```

   Or use default configuration:

   ```bash
   go run main.go
   ```

### Testing

Send test packets using the Python client:

```bash
python3 client_send_packet.py
```

This sends:

- Login packet with IMEI `123456789123456`
- Location packet with sample GPS coordinates (Vietnam)
- Alarm packet with simulated emergency status
- Heartbeat packet with device status

## Protocol Details

### Packet Structure

Each GT06 packet follows this format:

```text
[Header(2)] [Length(1)] [Protocol(1)] [Content(N)] [Serial(2)] [CRC(2)] [Stop(2)]
  0x7878     Len         Type           Data        Seq#        Checksum  0x0D0A
```

### Coordinates Format

- **Latitude/Longitude**: Encoded as `decimal_degrees * 1800000` (4 bytes each)
- **Speed**: In km/h (1 byte)
- **Course**: 0-359 degrees (2 bytes)

### Time Format

Date-time encoded as 6 bytes: `[year, month, day, hour, minute, second]`

### CRC Calculation

Uses CRC-ITU (CRC-16 CCITT) with polynomial 0x1021:

```python
# Python example
def calculate_crc(data):
    crc = 0xFFFF
    for byte in data:
        crc = (crc >> 8) ^ CRC_TABLE[(crc ^ byte) & 0xFF]
    return ~crc & 0xFFFF
```

## API Endpoints (Go Server)

The server listens on `0.0.0.0:8000` and processes packets based on protocol type.

### Packet Handlers

- **LoginDeviceService**: Processes login packets, stores device info
- **HeartbeatPacketService**: Handles periodic heartbeat updates
- **LocationPacketService**: Stores GPS location data
- **AlarmPacketService**: Processes alarm/alert packets

## Configuration

Configuration is loaded from `etc/server.yaml`. All fields have sensible defaults:

### Configuration Fields

- **TCPServer**: Server listen address (default: `0.0.0.0:8000`)
- **MongoURI**: MongoDB connection string (default: `mongodb://localhost:27017`)
- **DBName**: Database name (default: `gt06`)
- **LogLevel**: Logging level (default: `info`)
- **Timeout**: Connection timeout in seconds (default: `10`)

### Example Configuration

```yaml
TCPServer: 0.0.0.0:8000
MongoURI: mongodb://admin:admin@mongodb:27017
DBName: gt06
LogLevel: info
Timeout: 10
```

### Usage

```bash
# With custom config file
go run main.go -c etc/server.yaml

# With default config (if file not found)
go run main.go

# With different config file
go run main.go -c /path/to/config.yaml
```

## Code Quality

### Overview of Implementation

**Strengths:**

- Clear packet structure definitions with proper binary layout
- Comprehensive CRC validation for data integrity
- Service-based architecture for modularity
- Proper error handling with context logging
- Protocol parsing with bounds checking

**Areas for Improvement:**

- **Error Handling**: Some error cases silently fail (e.g., `s.svc.MongoDBModel.Insert` errors are ignored)
- **Type Safety**: Python client uses magic numbers; consider enum-like constants
- **Testing**: No unit tests for protocol parsing or CRC calculation
- **Documentation**: Protocol constants (0x78, 0x7878, etc.) could use named constants
- **Concurrency**: TCP server implementation needs thread-safety verification
- **Input Validation**: Client-side lat/lon encoding assumes valid decimal degrees


## Docker Deployment

```bash
docker-compose up -d
```

This starts:

- MongoDB service
- Go GT06 server

## Dependencies

- **go-zero**: Microservice framework
- **mongo-driver**: MongoDB driver for Go
- **socket**: Python standard library for TCP communication
- **struct**: Python standard library for binary packing

## Contributing

When contributing:

1. Follow Go conventions (gofmt, golint)
2. Add unit tests for protocol parsing
3. Update protocol documentation
4. Handle errors explicitly (no silent failures)
5. Add named constants for magic numbers


## References

- [GT06/CONCOX Protocol Specification](docs/ProtocolJM_VL02%2CJM_VG03%2CEG02%2CEG03%2CJM01%2CJV200%2CGT300%2CGT800%2CMT200%2COB22.pdf) - Complete protocol documentation
- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver) - Official Go MongoDB driver
- [go-zero Framework](https://go-zero.dev/) - Microservice framework documentation

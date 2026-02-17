package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	PARSE_PASS       = 0
	PARSE_FAIL       = -1
	MAX_INFO_CONTENT = 256

	// Packet Structure Constants
	PacketStartBit        = 0x78
	PacketStopBit0        = 0x0D
	PacketStopBit1        = 0x0A

	// Protocol Types
	ProtocolLogin         = 0x01
	ProtocolHeartbeat     = 0x13
	ProtocolHeartbeatAlt  = 0x23
	ProtocolLocation      = 0x12
	ProtocolLocationUTC   = 0x22
	ProtocolAlarm         = 0x26
)

var crcTable = [256]uint16{
	0x0000, 0x1189, 0x2312, 0x329B, 0x4624, 0x57AD, 0x6536, 0x74BF,
	0x8C48, 0x9DC1, 0xAF5A, 0xBED3, 0xCA6C, 0xDBE5, 0xE97E, 0xF8F7,
	0x1081, 0x0108, 0x3393, 0x221A, 0x56A5, 0x472C, 0x75B7, 0x643E,
	0x9CC9, 0x8D40, 0xBFDB, 0xAE52, 0xDAED, 0xCB64, 0xF9FF, 0xE876,
	0x2102, 0x308B, 0x0210, 0x1399, 0x6726, 0x76AF, 0x4434, 0x55BD,
	0xAD4A, 0xBCC3, 0x8E58, 0x9FD1, 0xEB6E, 0xFAE7, 0xC87C, 0xD9F5,
	0x3183, 0x200A, 0x1291, 0x0318, 0x77A7, 0x662E, 0x54B5, 0x453C,
	0xBDCB, 0xAC42, 0x9ED9, 0x8F50, 0xFBEF, 0xEA66, 0xD8FD, 0xC974,
	0x4204, 0x538D, 0x6116, 0x709F, 0x0420, 0x15A9, 0x2732, 0x36BB,
	0xCE4C, 0xDFC5, 0xED5E, 0xFCD7, 0x8868, 0x99E1, 0xAB7A, 0xBAF3,
	0x5285, 0x430C, 0x7197, 0x601E, 0x14A1, 0x0528, 0x37B3, 0x263A,
	0xDECD, 0xCF44, 0xFDDF, 0xEC56, 0x98E9, 0x8960, 0xBBFB, 0xAA72,
	0x6306, 0x728F, 0x4014, 0x519D, 0x2522, 0x34AB, 0x0630, 0x17B9,
	0xEF4E, 0xFEC7, 0xCC5C, 0xDDD5, 0xA96A, 0xB8E3, 0x8A78, 0x9BF1,
	0x7387, 0x620E, 0x5095, 0x411C, 0x35A3, 0x242A, 0x16B1, 0x0738,
	0xFFCF, 0xEE46, 0xDCDD, 0xCD54, 0xB9EB, 0xA862, 0x9AF9, 0x8B70,
	0x8408, 0x9581, 0xA71A, 0xB693, 0xC22C, 0xD3A5, 0xE13E, 0xF0B7,
	0x0840, 0x19C9, 0x2B52, 0x3ADB, 0x4E64, 0x5FED, 0x6D76, 0x7CFF,
	0x9489, 0x8500, 0xB79B, 0xA612, 0xD2AD, 0xC324, 0xF1BF, 0xE036,
	0x18C1, 0x0948, 0x3BD3, 0x2A5A, 0x5EE5, 0x4F6C, 0x7DF7, 0x6C7E,
	0xA50A, 0xB483, 0x8618, 0x9791, 0xE32E, 0xF2A7, 0xC03C, 0xD1B5,
	0x2942, 0x38CB, 0x0A50, 0x1BD9, 0x6F66, 0x7EEF, 0x4C74, 0x5DFD,
	0xB58B, 0xA402, 0x9699, 0x8710, 0xF3AF, 0xE226, 0xD0BD, 0xC134,
	0x39C3, 0x284A, 0x1AD1, 0x0B58, 0x7FE7, 0x6E6E, 0x5CF5, 0x4D7C,
	0xC60C, 0xD785, 0xE51E, 0xF497, 0x8028, 0x91A1, 0xA33A, 0xB2B3,
	0x4A44, 0x5BCD, 0x6956, 0x78DF, 0x0C60, 0x1DE9, 0x2F72, 0x3EFB,
	0xD68D, 0xC704, 0xF59F, 0xE416, 0x90A9, 0x8120, 0xB3BB, 0xA232,
	0x5AC5, 0x4B4C, 0x79D7, 0x685E, 0x1CE1, 0x0D68, 0x3FF3, 0x2E7A,
	0xE70E, 0xF687, 0xC41C, 0xD595, 0xA12A, 0xB0A3, 0x8238, 0x93B1,
	0x6B46, 0x7ACF, 0x4854, 0x59DD, 0x2D62, 0x3CEB, 0x0E70, 0x1FF9,
	0xF78F, 0xE606, 0xD49D, 0xC514, 0xB1AB, 0xA022, 0x92B9, 0x8330,
	0x7BC7, 0x6A4E, 0x58D5, 0x495C, 0x3DE3, 0x2C6A, 0x1EF1, 0x0F78,
}

// CONCOXPacket structure
type CONCOXPacket struct {
	StartBit         [2]byte
	PacketLength     uint8
	ProtocolNumber   uint8
	InfoContent      [MAX_INFO_CONTENT]byte
	InfoSerialNumber uint16
	ErrorCheck       uint16
	StopBit          [2]byte
}

// Login Packet Information Content
type CONCOXLoginInfoContent struct {
	TerminalID       [8]byte
	ModelCode        [2]byte
	TimeZoneLanguage [2]byte
}

// Heartbeat Packet Information Content
type CONCOXHeartbeatInfoContent struct {
	TerminalInfo        uint8
	ExternalVoltage     uint16
	BatteryVoltageLevel uint8
	GSMSignalStrength   uint8
	LanguageStatus      uint16
}

// Location Packet Information Content
type CONCOXLocationInfoContent struct {
	DateTime            [6]byte
	GPSSatellites       uint8
	Latitude            uint32
	Longitude           uint32
	Speed               uint8
	CourseStatus        uint16
	MCC                 uint16
	MNC                 uint8
	LAC                 uint16
	CellID              uint32
	ACCStatus           uint8
	UploadMode          uint8
	GPSRealTimeReupload uint8
	Mileage             uint32
}

// Alarm Packet Information Content
type CONCOXAlarmInfoContent struct {
	DateTime          [6]byte
	GPSSatellites     uint8
	Latitude          uint32
	Longitude         uint32
	Speed             uint8
	CourseStatus      uint16
	LBSLength         uint8
	MCC               uint16
	MNC               uint8
	LAC               uint16
	CellID            uint32
	TerminalInfo      uint8
	VoltageLevel      uint8
	GSMSignalStrength uint8
	AlarmLanguage     uint16
	Mileage           uint32
}

func calculateCRC(data []byte) uint16 {
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc = (crc >> 8) ^ crcTable[(crc^uint16(b))&0xFF]
	}

	return ^crc & 0xFFFF
}

func ParseAndValidatePacket(buffer []byte) (*CONCOXPacket, error) {

	if len(buffer) < 10 {
		return nil, fmt.Errorf("buffer too small for packet: %d", len(buffer))
	}

	packet := &CONCOXPacket{}

	copy(packet.StartBit[:], buffer[:2])
	if packet.StartBit[0] != PacketStartBit || packet.StartBit[1] != PacketStartBit {
		return nil, errors.New("invalid start bits")
	}

	packet.PacketLength = buffer[2]
	packet.ProtocolNumber = buffer[3]

	infoLength := int(packet.PacketLength) - 5
	if infoLength < 0 {
		return nil, fmt.Errorf("invalid packet length: %d", packet.PacketLength)
	}
	if infoLength > MAX_INFO_CONTENT {
		return nil, errors.New("info content exceeds maximum size")
	}

	// Total packet size: header(2) + length(1) + protocol(1) + info(infoLength) + serial(2) + crc(2) + stop(2)
	totalPacketSize := 10 + infoLength
	if len(buffer) < totalPacketSize {
		return nil, fmt.Errorf("buffer too small: expected at least %d bytes, got %d", totalPacketSize, len(buffer))
	}

	copy(packet.InfoContent[:infoLength], buffer[4:4+infoLength])

	packet.InfoSerialNumber = binary.BigEndian.Uint16(buffer[4+infoLength : 6+infoLength])

	packet.ErrorCheck = binary.BigEndian.Uint16(buffer[6+infoLength : 8+infoLength])

	copy(packet.StopBit[:], buffer[8+infoLength:10+infoLength])
	if packet.StopBit[0] != PacketStopBit0 || packet.StopBit[1] != PacketStopBit1 {
		return nil, errors.New("invalid stop bits")
	}

	calculatedCRC := calculateCRC(buffer[2 : 2+packet.PacketLength-1])
	if calculatedCRC != packet.ErrorCheck {
		return nil, fmt.Errorf("CRC mismatch: expected 0x%04X got 0x%04X", packet.ErrorCheck, calculatedCRC)
	}

	return packet, nil
}

func ParseCONCOXLoginInfoContent(buffer []byte) (*CONCOXLoginInfoContent, error) {
	if len(buffer) < 12 {
		return nil, fmt.Errorf("buffer too small for login info: %d bytes", len(buffer))
	}

	loginInfo := &CONCOXLoginInfoContent{}

	copy(loginInfo.TerminalID[:], buffer[:8])
	copy(loginInfo.ModelCode[:], buffer[8:10])
	copy(loginInfo.TimeZoneLanguage[:], buffer[10:12])

	return loginInfo, nil
}

func ParseCONCOXAlarmInfoContent(buffer []byte) (*CONCOXAlarmInfoContent, error) {
	// Minimum size required for Alarm Info Packet
	if len(buffer) < 36 {
		return nil, fmt.Errorf("buffer too small for alarm info: %d bytes", len(buffer))
	}

	alarmInfo := &CONCOXAlarmInfoContent{}

	copy(alarmInfo.DateTime[:], buffer[:6])

	alarmInfo.GPSSatellites = buffer[6]
	alarmInfo.Latitude = binary.BigEndian.Uint32(buffer[7:11])
	alarmInfo.Longitude = binary.BigEndian.Uint32(buffer[11:15])
	alarmInfo.Speed = buffer[15]
	alarmInfo.CourseStatus = binary.BigEndian.Uint16(buffer[16:18])
	alarmInfo.LBSLength = buffer[18]
	alarmInfo.MCC = binary.BigEndian.Uint16(buffer[19:21])
	alarmInfo.MNC = buffer[21]
	alarmInfo.LAC = binary.BigEndian.Uint16(buffer[22:24])
	alarmInfo.CellID = uint32(buffer[24])<<16 | uint32(buffer[25])<<8 | uint32(buffer[26])
	alarmInfo.TerminalInfo = buffer[27]
	alarmInfo.VoltageLevel = buffer[28]
	alarmInfo.GSMSignalStrength = buffer[29]
	alarmInfo.AlarmLanguage = binary.BigEndian.Uint16(buffer[30:32])
	alarmInfo.Mileage = binary.BigEndian.Uint32(buffer[32:36])

	return alarmInfo, nil
}

func ParseCONCOXLocationInfoContent(buffer []byte) (*CONCOXLocationInfoContent, error) {

	if len(buffer) < 33 {
		return nil, fmt.Errorf("buffer too small for location info: %d bytes", len(buffer))
	}

	locationInfo := &CONCOXLocationInfoContent{}

	// Copy DateTime (6 bytes)
	copy(locationInfo.DateTime[:], buffer[:6])

	// Parse individual fields
	locationInfo.GPSSatellites = buffer[6]
	locationInfo.Latitude = binary.BigEndian.Uint32(buffer[7:11])
	locationInfo.Longitude = binary.BigEndian.Uint32(buffer[11:15])
	locationInfo.Speed = buffer[15]
	locationInfo.CourseStatus = binary.BigEndian.Uint16(buffer[16:18])
	locationInfo.MCC = binary.BigEndian.Uint16(buffer[18:20])
	locationInfo.MNC = buffer[20]
	locationInfo.LAC = binary.BigEndian.Uint16(buffer[21:23])
	locationInfo.CellID = uint32(buffer[23])<<16 | uint32(buffer[24])<<8 | uint32(buffer[25])
	locationInfo.ACCStatus = buffer[26]
	locationInfo.UploadMode = buffer[27]
	locationInfo.GPSRealTimeReupload = buffer[28]
	locationInfo.Mileage = binary.BigEndian.Uint32(buffer[29:33])

	return locationInfo, nil
}

func ParseCONCOXHeartbeatInfoContent(buffer []byte) (*CONCOXHeartbeatInfoContent, error) {

	if len(buffer) < 7 { // 7 bytes: 1 (TerminalInfo) + 2 (ExternalVoltage) + 1 (BatteryLevel) + 1 (GSM) + 2 (LanguageStatus)
		return nil, fmt.Errorf("buffer too small for heartbeat info: %d bytes", len(buffer))
	}

	heartbeatInfo := &CONCOXHeartbeatInfoContent{}

	heartbeatInfo.TerminalInfo = buffer[0]
	heartbeatInfo.ExternalVoltage = binary.BigEndian.Uint16(buffer[1:3])
	heartbeatInfo.BatteryVoltageLevel = buffer[3]
	heartbeatInfo.GSMSignalStrength = buffer[4]
	heartbeatInfo.LanguageStatus = binary.BigEndian.Uint16(buffer[5:7])

	return heartbeatInfo, nil
}

func BuildCONCOXResponseLogin(receivedPacket *CONCOXPacket) []byte {
	responsePacket := make([]byte, 0)

	// Start Bit
	responsePacket = append(responsePacket, PacketStartBit, PacketStartBit)

	// Packet Length (fixed for login response)
	responsePacket = append(responsePacket, 0x05)

	// Protocol Number
	responsePacket = append(responsePacket, receivedPacket.ProtocolNumber)

	// Information Serial Number
	serialNumber := make([]byte, 2)
	binary.BigEndian.PutUint16(serialNumber, receivedPacket.InfoSerialNumber)
	responsePacket = append(responsePacket, serialNumber...)

	// Calculate CRC
	crc := calculateCRC(responsePacket[2:]) // Length + Protocol Number + Serial Number
	crcBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(crcBytes, crc)
	responsePacket = append(responsePacket, crcBytes...)

	// Stop Bit
	responsePacket = append(responsePacket, PacketStopBit0, PacketStopBit1)

	return responsePacket
}

func BuildCONCOXResponseHeartbeat(packet *CONCOXPacket) []byte {
	response := []byte{
		PacketStartBit, PacketStartBit,     // Start Bit
		0x05,                               // Packet Length (default 5 bytes)
		packet.ProtocolNumber,              // Protocol Number (usually 0x13 or 0x23)
		byte(packet.InfoSerialNumber >> 8), // Serial Number (High Byte)
		byte(packet.InfoSerialNumber),      // Serial Number (Low Byte)
	}

	crc := calculateCRC(response[2:])
	response = append(response, byte(crc>>8), byte(crc))

	response = append(response, PacketStopBit0, PacketStopBit1) // Stop Bit

	return response
}

func BuildCONCOXResponseLocation(receivedPacket *CONCOXPacket) []byte {
	responsePacket := make([]byte, 0)

	// Start Bit
	responsePacket = append(responsePacket, PacketStartBit, PacketStartBit)

	// Packet Length (fixed for location response)
	responsePacket = append(responsePacket, 0x05)

	// Protocol Number
	responsePacket = append(responsePacket, receivedPacket.ProtocolNumber)

	// Information Serial Number
	serialNumber := make([]byte, 2)
	binary.BigEndian.PutUint16(serialNumber, receivedPacket.InfoSerialNumber)
	responsePacket = append(responsePacket, serialNumber...)

	// Calculate CRC
	crc := calculateCRC(responsePacket[2:]) // Length + Protocol Number + Serial Number
	crcBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(crcBytes, crc)
	responsePacket = append(responsePacket, crcBytes...)

	// Stop Bit
	responsePacket = append(responsePacket, PacketStopBit0, PacketStopBit1)

	return responsePacket
}

func BuildCONCOXResponseAlarm(receivedPacket *CONCOXPacket) []byte {
	responsePacket := make([]byte, 0)

	// Start Bit
	responsePacket = append(responsePacket, PacketStartBit, PacketStartBit)

	// Packet Length (fixed for alarm response)
	responsePacket = append(responsePacket, 0x05)

	// Protocol Number
	responsePacket = append(responsePacket, receivedPacket.ProtocolNumber)

	// Information Serial Number
	serialNumber := make([]byte, 2)
	binary.BigEndian.PutUint16(serialNumber, receivedPacket.InfoSerialNumber)
	responsePacket = append(responsePacket, serialNumber...)

	// Calculate CRC
	crc := calculateCRC(responsePacket[2:]) // Length + Protocol Number + Serial Number
	crcBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(crcBytes, crc)
	responsePacket = append(responsePacket, crcBytes...)

	// Stop Bit
	responsePacket = append(responsePacket, PacketStopBit0, PacketStopBit1)

	return responsePacket
}

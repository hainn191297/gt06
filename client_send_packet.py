import socket
import struct
import time
# CRC-ITU Table
CRC_TABLE = [
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
    0x7BC7, 0x6A4E, 0x58D5, 0x495C, 0x3DE3, 0x2C6A, 0x1EF1, 0x0F78
]

# Function to calculate CRC-ITU
def calculate_crc(data):
    print("Data input:", data.hex())
    crc = 0xFFFF
    for byte in data:
        crc = (crc >> 8) ^ CRC_TABLE[(crc ^ byte) & 0xFF]
    return ~crc & 0xFFFF
# Build Login Packet
def build_login_packet(imei, model_code, time_zone, language, serial_number):
    """
    Build the login packet according to the protocol.

    Args:
        imei (str): 15-digit IMEI number of the device.
        model_code (int): 2-byte model identification code.
        time_zone (int): Time zone in 1/100 of an hour (e.g., +8 -> 800, -12:45 -> -1245).
        language (int): Language selection bit (1 for Chinese, 2 for English).
        serial_number (int): 2-byte serial number for the packet.

    Returns:
        bytes: The constructed login packet.
    """
    # Convert IMEI to 8-byte Terminal ID
    imei_bytes = struct.pack('>Q', int(imei))[-8:]

    # Time zone and language calculation
    time_zone_hex = (time_zone * 100) & 0xFFF  # Encode time zone as hex
    time_zone_lang = ((time_zone_hex << 4) | (language & 0xF)).to_bytes(2, 'big')

    # Information Content
    info_content = imei_bytes + struct.pack('>H', model_code) + time_zone_lang

    # Protocol Number
    protocol_number = b'\x01'

    # Packet Length
    packet_length = len(protocol_number + info_content + struct.pack('>H', serial_number)) + 2  # 4 for CRC and Stop Bit

    # Header
    header = b'\x78\x78'

    # Serial Number (2 bytes)
    serial_bytes = struct.pack('>H', serial_number)

    # Construct packet (excluding CRC for now)
    packet_without_crc = (
        header +
        struct.pack('B', packet_length) +
        protocol_number +
        info_content +
        serial_bytes
    )

    # Calculate CRC
    crc = calculate_crc(packet_without_crc[2:])
    crc_bytes = struct.pack('>H', crc)

    # Stop Bit
    stop_bit = b'\x0D\x0A'

    # Final Packet
    packet = packet_without_crc + crc_bytes + stop_bit


    print(f"Login Packet: {packet.hex()}")
    return packet

def build_location_packet(date_time, latitude, longitude, speed, course_status, mcc, mnc, lac, cell_id, acc_status, data_upload_mode, gps_real_time_reupload, mileage, serial_number):
    """
    Build the location packet according to the protocol.

    Args:
        date_time (tuple): Date and time as (year, month, day, hour, minute, second).
        latitude (float): Latitude in decimal degrees.
        longitude (float): Longitude in decimal degrees.
        speed (int): Speed in km/h.
        course_status (int): 2-byte course and status field.
        mcc (int): Mobile country code.
        mnc (int): Mobile network code.
        lac (int): Location area code.
        cell_id (int): Cell tower ID.
        acc_status (int): ACC status (0 for low, 1 for high).
        data_upload_mode (int): GPS data upload mode.
        gps_real_time_reupload (int): GPS real-time re-upload indicator (0 or 1).
        mileage (int): Mileage in meters.
        serial_number (int): 2-byte serial number for the packet.

    Returns:
        bytes: The constructed location packet.
    """
    # Convert date and time to 6 bytes
    date_time_bytes = struct.pack('>BBBBBB', *date_time)

    # Convert latitude and longitude to protocol format
    latitude_bytes = struct.pack('>I', int(latitude * 1800000))
    longitude_bytes = struct.pack('>I', int(longitude * 1800000))

    # MCC, MNC, LAC, and Cell ID
    mcc_bytes = struct.pack('>H', mcc)
    mnc_bytes = struct.pack('>B', mnc)
    lac_bytes = struct.pack('>H', lac)
    cell_id_bytes = struct.pack('>I', cell_id)[1:]  # Use the last 3 bytes

    # Mileage (4 bytes)
    mileage_bytes = struct.pack('>I', mileage)

    # Information Content
    info_content = (
        date_time_bytes +
        struct.pack('>B', 10) +  # Placeholder for satellite count (fixed at 10)
        latitude_bytes +
        longitude_bytes +
        struct.pack('>B', speed) +
        struct.pack('>H', course_status) +
        mcc_bytes +
        mnc_bytes +
        lac_bytes +
        cell_id_bytes +
        struct.pack('>B', acc_status) +
        struct.pack('>B', data_upload_mode) +
        struct.pack('>B', gps_real_time_reupload) +
        mileage_bytes
    )

    # Protocol Number
    protocol_number = b'\x22'

    # Packet Length
    packet_length = len(protocol_number + info_content + struct.pack('>H', serial_number)) + 2  # 4 for CRC and Stop Bit

    # Header
    header = b'\x78\x78'

    # Serial Number (2 bytes)
    serial_bytes = struct.pack('>H', serial_number)

    # Construct packet (excluding CRC for now)
    packet_without_crc = (
        header +
        struct.pack('B', packet_length) +
        protocol_number +
        info_content +
        serial_bytes
    )

    # Calculate CRC
    crc = calculate_crc(packet_without_crc[2:])
    crc_bytes = struct.pack('>H', crc)

    # Stop Bit
    stop_bit = b'\x0D\x0A'

    location_packet = packet_without_crc + crc_bytes + stop_bit
    print(f"Location Packet: {location_packet.hex()}")
    return location_packet

def build_alarm_packet(date_time, latitude, longitude, speed, course_status, mcc, mnc, lac, cell_id, acc_status, terminal_info, battery_level, gsm_signal_strength, alarm_language, mileage, serial_number):
    """
    Build the alarm packet according to the protocol.

    Args:
        date_time (tuple): Date and time as (year, month, day, hour, minute, second).
        latitude (float): Latitude in decimal degrees.
        longitude (float): Longitude in decimal degrees.
        speed (int): Speed in km/h.
        course_status (int): 2-byte course and status field.
        mcc (int): Mobile country code.
        mnc (int): Mobile network code.
        lac (int): Location area code.
        cell_id (int): Cell tower ID.
        acc_status (int): ACC status (0 for low, 1 for high).
        terminal_info (int): 1-byte terminal information flags.
        battery_level (int): 1-byte built-in battery voltage level.
        gsm_signal_strength (int): GSM signal strength (0x00 to 0x04).
        alarm_language (int): 2-byte combined alarm type and language.
        mileage (int): Mileage in meters.
        serial_number (int): 2-byte serial number for the packet.

    Returns:
        bytes: The constructed alarm packet.
    """
    # Convert date and time to 6 bytes
    date_time_bytes = struct.pack('>BBBBBB', *date_time)

    # Convert latitude and longitude to protocol format
    latitude_bytes = struct.pack('>I', int(latitude * 1800000))
    longitude_bytes = struct.pack('>I', int(longitude * 1800000))

    # MCC, MNC, LAC, and Cell ID
    mcc_bytes = struct.pack('>H', mcc)
    mnc_bytes = struct.pack('>B', mnc)
    lac_bytes = struct.pack('>H', lac)
    cell_id_bytes = struct.pack('>I', cell_id)[1:]  # Use the last 3 bytes

    # Terminal Information (1 byte)
    terminal_info_byte = struct.pack('>B', terminal_info)

    # Battery Level (1 byte)
    battery_level_byte = struct.pack('>B', battery_level)

    # GSM Signal Strength (1 byte)
    gsm_signal_strength_byte = struct.pack('>B', gsm_signal_strength)

    # Alarm/Language (2 bytes)
    alarm_language_bytes = struct.pack('>H', alarm_language)

    # Mileage (4 bytes)
    mileage_bytes = struct.pack('>I', mileage)

    # Information Content
    info_content = (
        date_time_bytes +
        struct.pack('>B', 10) +  # Placeholder for satellite count (fixed at 10)
        latitude_bytes +
        longitude_bytes +
        struct.pack('>B', speed) +
        struct.pack('>H', course_status) +
        mcc_bytes +
        mnc_bytes +
        lac_bytes +
        cell_id_bytes +
        struct.pack('>B', acc_status) +
        terminal_info_byte +
        battery_level_byte +
        gsm_signal_strength_byte +
        alarm_language_bytes +
        mileage_bytes
    )

    # Protocol Number
    protocol_number = b'\x26'

    # Packet Length
    packet_length = len(protocol_number + info_content + struct.pack('>H', serial_number)) + 2  # 4 for CRC and Stop Bit

    # Header
    header = b'\x78\x78'

    # Serial Number (2 bytes)
    serial_bytes = struct.pack('>H', serial_number)

    # Construct packet (excluding CRC for now)
    packet_without_crc = (
        header +
        struct.pack('B', packet_length) +
        protocol_number +
        info_content +
        serial_bytes
    )

    # Calculate CRC
    crc = calculate_crc(packet_without_crc[2:])
    crc_bytes = struct.pack('>H', crc)

    # Stop Bit
    stop_bit = b'\x0D\x0A'

    # Complete packet
    alarm_packet = packet_without_crc + crc_bytes + stop_bit
    print(f"Alarm Packet: {alarm_packet.hex()}")
    return alarm_packet



# Build Heartbeat Packet
def build_heartbeat_packet(terminal_info, external_voltage, battery_level, gsm_signal_strength, language_port_status, serial_number):
    """
    Build the heartbeat packet according to the protocol.

    Args:
        terminal_info (int): 1-byte terminal status flags (e.g., oil/electricity, GPS tracking).
        external_voltage (float): External voltage in volts (e.g., 12.34V -> 1234).
        battery_level (int): 1-byte battery level (0x00 to 0x06).
        gsm_signal_strength (int): GSM signal strength (0x00 to 0x04).
        language_port_status (int): 2-byte language and extended port status field.
        serial_number (int): 2-byte serial number for the packet.

    Returns:
        bytes: The constructed heartbeat packet.
    """
    # Terminal Information (1 byte)
    terminal_info_byte = struct.pack('>B', terminal_info)

    # External Voltage (2 bytes, multiplied by 100 and converted to int)
    external_voltage_bytes = struct.pack('>H', int(external_voltage * 100))

    # Battery Level (1 byte)
    battery_level_byte = struct.pack('>B', battery_level)

    # GSM Signal Strength (1 byte)
    gsm_signal_strength_byte = struct.pack('>B', gsm_signal_strength)

    # Language/Port Status (2 bytes)
    language_port_status_bytes = struct.pack('>H', language_port_status)

    # Information Content
    info_content = (
        terminal_info_byte +
        external_voltage_bytes +
        battery_level_byte +
        gsm_signal_strength_byte +
        language_port_status_bytes
    )

    # Protocol Number
    protocol_number = b'\x13'

    # Packet Length
    packet_length = len(protocol_number + info_content + struct.pack('>H', serial_number)) + 2  # 4 for CRC and Stop Bit

    # Header
    header = b'\x78\x78'

    # Serial Number (2 bytes)
    serial_bytes = struct.pack('>H', serial_number)

    # Construct packet (excluding CRC for now)
    packet_without_crc = (
        header +
        struct.pack('B', packet_length) +
        protocol_number +
        info_content +
        serial_bytes
    )

    # Calculate CRC
    crc = calculate_crc(packet_without_crc[2:])
    crc_bytes = struct.pack('>H', crc)

    # Stop Bit
    stop_bit = b'\x0D\x0A'

    heartbeat_packet = packet_without_crc + crc_bytes + stop_bit
    print(f"Heartbeat Packet: {heartbeat_packet.hex()}")
    return heartbeat_packet

def main():
    #server_ip = '20.212.168.51'
    server_ip = '127.0.0.1'  # Replace with the actual server IP
    server_port = 8000  # Replace with the actual server port

    # Connect to the server
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as client:
        try:
            client.connect((server_ip, server_port))
            print(f"Connected to server {server_ip}:{server_port}")
        except socket.error as e:
            print(f"Connection error: {e}")
            return

        # Send login packet
        imei = "123456789123456"
        model_code = 0x0242  # Example model code
        time_zone = 8  # GMT+8
        language = 2  # English
        serial_number = 1  # Serial number for the packet

        login_packet = build_login_packet(imei, model_code, time_zone, language, serial_number)
        print(f"Sending login packet: {login_packet.hex()}")
        client.sendall(login_packet)

        # Receive server response
        try:
            response = client.recv(1024)
            print(f"Received response: {response.hex()}")
            print(f"Response Details: Length={len(response)}, Content={response.hex()}")
            if response[3:4] != b'\x01':  # Check if response protocol matches login
                print("Login failed. Aborting further communication.")
                return
        except socket.error as e:
            print(f"Error receiving response: {e}")
            return

        # Build and send location packet
        date_time = (24, 1, 12, 12, 30, 45)  # Example date-time
        latitude = 10.762622
        longitude = 106.660172
        speed = 60  # 60 km/h
        course_status = 0x1234  # Example course and status
        mcc = 452  # Vietnam
        mnc = 1
        lac = 12345
        cell_id = 67890
        acc_status = 1  # ACC high
        data_upload_mode = 0x00  # Upload by time interval
        gps_real_time_reupload = 0x01  # Real-time upload
        mileage = 123456  # 123.456 km
        serial_number = 2

        location_packet = build_location_packet(date_time, latitude, longitude, speed, course_status, mcc, mnc, lac, cell_id, acc_status, data_upload_mode, gps_real_time_reupload, mileage, serial_number)

        # Send location data packet
        print(f"Sending location packet: {location_packet.hex()}")
        client.sendall(location_packet)
        #time.sleep(3)
        

        # Receive server response
        # try:
        #     response = client.recv(1024)
        #     print(f"Received response: {response.hex()}")
        #     print(f"Response Details: Length={len(response)}, Content={response.hex()}")
        # except socket.error as e:
        #     print(f"Error receiving response: {e}")
        #     return
        
        date_time = (24, 1, 12, 12, 30, 45)  # Example date-time
        latitude = 10.762622
        longitude = 106.660172
        speed = 60  # 60 km/h
        course_status = 0x1234  # Example course and status
        mcc = 452  # Vietnam
        mnc = 1
        lac = 12345
        cell_id = 67890
        acc_status = 1  # ACC high
        terminal_info = 0x01  # Example terminal info
        battery_level = 0x05  # High battery level
        gsm_signal_strength = 0x03  # Good signal
        alarm_language = 0x0102  # Alarm type and language combined
        mileage = 123456  # Example mileage (123.456 km)
        serial_number = 3

        alarm_packet = build_alarm_packet(date_time, latitude, longitude, speed, course_status, mcc, mnc, lac, cell_id, acc_status, terminal_info, battery_level, gsm_signal_strength, alarm_language, mileage, serial_number)
        print("Alarm Packet:", alarm_packet.hex())
        client.sendall(alarm_packet)
        
        # Build and send heartbeat packet
        terminal_info = 0b01100000  # Example terminal info (GPS tracking on, oil/electricity connected)
        external_voltage = 12.34  # Example external voltage
        battery_level = 0x05  # High battery level
        gsm_signal_strength = 0x03  # Good signal
        language_port_status = 0x0102  # English, extended port status
        serial_number = 1

        heartbeat_packet = build_heartbeat_packet(terminal_info, external_voltage, battery_level, gsm_signal_strength, language_port_status, serial_number)

        print("Heartbeat Packet:", heartbeat_packet.hex())
        client.sendall(heartbeat_packet)
        # Receive server response
        # try:
        #     response = client.recv(1024)
        #     print(f"Received response: {response.hex()}")
        #     print(f"Response Details: Length={len(response)}, Content={response.hex()}")
        # except socket.error as e:
        #     print(f"Error receiving response: {e}")
        #     return
        time.sleep(20)
if __name__ == "__main__":
    main()

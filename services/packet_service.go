package services

import "gt06/protocol"

type PacketService interface {
	ProcessPacket(packet *protocol.CONCOXPacket) ([]byte, error)
}

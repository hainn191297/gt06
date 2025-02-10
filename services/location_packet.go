package services

import (
	"context"
	"fmt"
	"gt06/common"
	"gt06/protocol"

	"github.com/zeromicro/go-zero/core/logx"
)

type LocationService struct {
	context context.Context
	log     logx.Logger
}

func NewLocationService(c context.Context) *LocationService {
	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &LocationService{
		context: c,
		log:     logx.WithContext(c),
	}
}

func (s *LocationService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Infof("Processing Location Packet")

	locationInfo, err := protocol.ParseCONCOXLocationInfoContent(packet.InfoContent[:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse location info: %w", err)
	}
	s.log.Infof("Parsed Location Info: %+v", locationInfo)

	// response := protocol.BuildCONCOXResponseLocation(packet)
	return nil, nil
}

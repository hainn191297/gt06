package services

import (
	"context"
	"gt06/common"
	"gt06/protocol"

	"github.com/zeromicro/go-zero/core/logx"
)

type HeartbeatService struct {
	context context.Context
	log     logx.Logger
}

func NewHeartbeatService(c context.Context) *HeartbeatService {
	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &HeartbeatService{
		context: c,
		log:     logx.WithContext(c),
	}
}

func (s *HeartbeatService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Info("Processing Heartbeat Packet")

	response := protocol.BuildCONCOXResponseHeartbeat(packet)
	return response, nil
}

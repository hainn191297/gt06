package services

import (
	"context"
	"fmt"
	"gt06/common"
	"gt06/protocol"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlarmService struct {
	context context.Context
	log     logx.Logger
}

func NewAlarmService(c context.Context) *AlarmService {
	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &AlarmService{
		context: c,
		log:     logx.WithContext(c),
	}
}

func (s *AlarmService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Info("Processing Alarm Packet")

	alarmInfo, err := protocol.ParseCONCOXAlarmInfoContent(packet.InfoContent[:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse alarm info: %w", err)
	}
	s.log.Infof("Parsed Alarm Info: %+v", alarmInfo)

	// response := protocol.BuildCONCOXResponseAlarm(packet)
	return nil, nil
}

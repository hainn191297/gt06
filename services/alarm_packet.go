package services

import (
	"context"
	"fmt"
	"gt06/common"
	"gt06/protocol"
	"gt06/services/svc"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type AlarmService struct {
	context context.Context
	log     logx.Logger
	svc     *svc.ServiceContext
}

func NewAlarmService(c context.Context, svc *svc.ServiceContext) *AlarmService {
	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &AlarmService{
		context: c,
		log:     logx.WithContext(c),
		svc:     svc,
	}
}

func (s *AlarmService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Info("Processing Alarm Packet")

	alarmInfo, err := protocol.ParseCONCOXAlarmInfoContent(packet.InfoContent[:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse alarm info: %w", err)
	}
	s.log.Infof("Parsed Alarm Info: %+v", alarmInfo)

	// Decode datetime from the 6-byte format [year, month, day, hour, minute, second]
	dateTime := time.Date(
		2000+int(alarmInfo.DateTime[0]),
		time.Month(alarmInfo.DateTime[1]),
		int(alarmInfo.DateTime[2]),
		int(alarmInfo.DateTime[3]),
		int(alarmInfo.DateTime[4]),
		int(alarmInfo.DateTime[5]),
		0,
		time.UTC,
	)

	// Decode latitude and longitude from fixed-point format (decimal_degrees * 1800000)
	latitude := float64(alarmInfo.Latitude) / 1800000.0
	longitude := float64(alarmInfo.Longitude) / 1800000.0

	document := bson.M{
		"date_time":        dateTime,
		"gps_satellites":   alarmInfo.GPSSatellites,
		"latitude":         latitude,
		"longitude":        longitude,
		"speed":            alarmInfo.Speed,
		"course_status":    alarmInfo.CourseStatus,
		"lbs_length":       alarmInfo.LBSLength,
		"mcc":              alarmInfo.MCC,
		"mnc":              alarmInfo.MNC,
		"lac":              alarmInfo.LAC,
		"cell_id":          alarmInfo.CellID,
		"terminal_info":    alarmInfo.TerminalInfo,
		"voltage_level":    alarmInfo.VoltageLevel,
		"gsm_signal":       alarmInfo.GSMSignalStrength,
		"alarm_language":   alarmInfo.AlarmLanguage,
		"mileage":          alarmInfo.Mileage,
		"created_at":       time.Now(),
	}

	_, err = s.svc.MongoDBModel.Insert(s.context, "CONCOXAlarmInfoContent", document)
	if err != nil {
		s.log.Errorf("Failed to insert alarm info: %w", err)
		return nil, fmt.Errorf("failed to save alarm data: %w", err)
	}

	s.log.Infof("Alarm data saved: lat=%.6f, lng=%.6f, voltage=0x%02X", latitude, longitude, alarmInfo.VoltageLevel)

	response := protocol.BuildCONCOXResponseAlarm(packet)
	return response, nil
}

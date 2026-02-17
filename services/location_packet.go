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

type LocationService struct {
	context context.Context
	log     logx.Logger
	svc     *svc.ServiceContext
}

func NewLocationService(c context.Context, svc *svc.ServiceContext) *LocationService {
	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &LocationService{
		context: c,
		log:     logx.WithContext(c),
		svc:     svc,
	}
}

func (s *LocationService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Infof("Processing Location Packet")

	locationInfo, err := protocol.ParseCONCOXLocationInfoContent(packet.InfoContent[:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse location info: %w", err)
	}
	s.log.Infof("Parsed Location Info: %+v", locationInfo)

	// Decode datetime from the 6-byte format [year, month, day, hour, minute, second]
	dateTime := time.Date(
		2000+int(locationInfo.DateTime[0]),
		time.Month(locationInfo.DateTime[1]),
		int(locationInfo.DateTime[2]),
		int(locationInfo.DateTime[3]),
		int(locationInfo.DateTime[4]),
		int(locationInfo.DateTime[5]),
		0,
		time.UTC,
	)

	// Decode latitude and longitude from fixed-point format (decimal_degrees * 1800000)
	latitude := float64(locationInfo.Latitude) / 1800000.0
	longitude := float64(locationInfo.Longitude) / 1800000.0

	document := bson.M{
		"date_time":              dateTime,
		"gps_satellites":         locationInfo.GPSSatellites,
		"latitude":               latitude,
		"longitude":              longitude,
		"speed":                  locationInfo.Speed,
		"course_status":          locationInfo.CourseStatus,
		"mcc":                    locationInfo.MCC,
		"mnc":                    locationInfo.MNC,
		"lac":                    locationInfo.LAC,
		"cell_id":                locationInfo.CellID,
		"acc_status":             locationInfo.ACCStatus,
		"upload_mode":            locationInfo.UploadMode,
		"gps_real_time_reupload": locationInfo.GPSRealTimeReupload,
		"mileage":                locationInfo.Mileage,
		"created_at":             time.Now(),
	}

	_, err = s.svc.MongoDBModel.Insert(s.context, "CONCOXLocationInfoContent", document)
	if err != nil {
		s.log.Errorf("Failed to insert location info: %w", err)
		return nil, fmt.Errorf("failed to save location data: %w", err)
	}

	s.log.Infof("Location data saved: lat=%.6f, lng=%.6f", latitude, longitude)

	response := protocol.BuildCONCOXResponseLocation(packet)
	return response, nil
}

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

type LoginDeviceService struct {
	context context.Context
	log     logx.Logger
	svc     *svc.ServiceContext
}

func NewLoginDeviceService(c context.Context, svc *svc.ServiceContext) *LoginDeviceService {

	c = logx.ContextWithFields(c, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	})

	return &LoginDeviceService{
		context: c,
		log:     logx.WithContext(c),
		svc:     svc,
	}
}

func (s *LoginDeviceService) ProcessPacket(packet *protocol.CONCOXPacket) (buf []byte, err error) {
	s.log.Info("Processing Login Packet")

	var infoContent *protocol.CONCOXLoginInfoContent
	infoContent, err = protocol.ParseCONCOXLoginInfoContent(packet.InfoContent[:])

	if err != nil {
		return nil, fmt.Errorf("failed to parse login info: %w", err)
	}

	// build login info
	gmt, region, language, _ := decodeRawTimeZone(infoContent.TimeZoneLanguage)
	document := bson.M{
		"terminal_id":        common.ConvertToHexString(infoContent.TerminalID[:]),
		"model_code":         common.ConvertToHexString(infoContent.TerminalID[:]),
		"time_zone_language": infoContent.TimeZoneLanguage,
		"gmt":                gmt,
		"region":             region,
		"language":           language,
		"created_at":         time.Now(),
	}

	ret, _ := s.svc.MongoDBModel.Insert(s.context, "CONCOXLoginInfoContent", document)

	// logx.Info(ret, err)

	// get
	filter := bson.M{"_id": ret.InsertedID}
	a, _ := s.svc.MongoDBModel.Get(s.context, "CONCOXLoginInfoContent", filter)

	logx.Infof(" %+v", a)

	// parse data

	v := make(map[string]interface{}, 0)
	common.ParseBsonMReflect(a, &v)

	logx.Info(v)
	buildLoginInfo := protocol.BuildCONCOXResponseLogin(packet)

	bson.Marshal(a)
	return buildLoginInfo, nil
}

func decodeRawTimeZone(data [2]byte) (float64, string, string, error) {
	rawValue := int(data[0])<<8 | int(data[1])

	gmtRaw := (rawValue >> 4) & 0xFFF
	gmt := float64(gmtRaw) / 100.0

	isEastern := (rawValue>>4)&0x1 == 0
	region := "Eastern"
	if !isEastern {
		region = "Western"
	}

	if region == "Western" {
		gmt = -gmt
	}

	languageBits := rawValue & 0x3
	language := ""
	switch languageBits {
	case 0x1:
		language = "English"
	case 0x2:
		language = "Chinese"
	case 0x3:
		return 0, "", "", fmt.Errorf("invalid language bits: %02b", languageBits)
	}

	return gmt, region, language, nil
}

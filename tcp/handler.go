package tcp

import (
	"context"
	"gt06/common"
	"gt06/protocol"
	"gt06/services"
	"gt06/services/svc"
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProtocolHandler struct {
	sessions sync.Map // use sync.Map to store sessions instead of a map [con]session
	eng      gnet.Engine
	mu       sync.Mutex
	c        context.Context
	svc      *svc.ServiceContext
}

func (ph *ProtocolHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.eng = eng
	ph.c = context.Background()
	logx.Info("Server booted")
	return
}

func (ph *ProtocolHandler) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {

	// root context
	ctx := context.Background()

	ctx = logx.ContextWithFields(ctx, logx.LogField{
		Key:   string(common.SpanID),
		Value: common.GenerateSpanID(),
	}, logx.LogField{
		Key:   string(common.TraceID),
		Value: common.GenerateTraceID(),
	})

	session := &Session{
		Context:    ctx,
		Conn:       c,
		LastActive: time.Now(),
	}

	ph.sessions.Store(c, session)

	logx.WithContext(ctx).Infof("New connection from %s, fd %d", c.RemoteAddr(), c.Fd())
	return
}

func (ph *ProtocolHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	value, ok := ph.sessions.Load(c)
	if ok {
		session := value.(*Session)
		logx.WithContext(session.Context).Infof("Client disconnected: %s", c.RemoteAddr())
		ph.sessions.Delete(c)
	}
	return
}

func (ph *ProtocolHandler) OnTraffic(c gnet.Conn) (action gnet.Action) {
	// run safe
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("Recovered from panic: %v", r)
			action = gnet.Close
		}
	}()
	// Peek header to determine packet size (4 bytes: 0x78 0x78 length protocol)
	headerData, err := c.Peek(4)
	if err != nil || len(headerData) < 4 {
		if err != nil {
			logx.Errorf("Failed to peek header: %v", err)
		}
		return gnet.None
	}

	// Extract packet length from byte 2
	packetLength := int(headerData[2])
	if packetLength < 5 || packetLength > 255 {
		logx.Errorf("Invalid packet length: %d", packetLength)
		// Try to skip this invalid packet
		return gnet.None
	}

	// Total packet size: header(2) + length(1) + protocol(1) + info(length-5) + serial(2) + crc(2) + stop(2)
	// Which simplifies to: 10 + (length - 5) = length + 5
	totalPacketSize := packetLength + 5

	// Peek exactly the amount needed for this packet
	data, err := c.Peek(totalPacketSize)
	if err != nil || len(data) < totalPacketSize {
		// Not enough data yet, wait for more
		return gnet.None
	}

	packet, err := protocol.ParseAndValidatePacket(data)
	if err != nil {
		logx.Errorf("Failed to parse packet: %v", err)
		// Discard invalid data to prevent infinite loop
		c.Discard(1)
		return gnet.None
	}

	value, exists := ph.sessions.Load(c)
	if !exists {
		return gnet.Close
	}
	session := value.(*Session)
	session.LastActive = time.Now()

	var service services.PacketService

	switch packet.ProtocolNumber {
	case protocol.ProtocolLogin:
		service = services.NewLoginDeviceService(session.Context, ph.svc)
	case protocol.ProtocolHeartbeat, protocol.ProtocolHeartbeatAlt:
		service = services.NewHeartbeatService(session.Context)
	case protocol.ProtocolLocation, protocol.ProtocolLocationUTC:
		service = services.NewLocationService(session.Context, ph.svc)
	case protocol.ProtocolAlarm:
		service = services.NewAlarmService(session.Context, ph.svc)
	default:
		logx.WithContext(session.Context).Errorf("Unknown Protocol Number: 0x%02X", packet.ProtocolNumber)
		return gnet.Close
	}

	out, err := service.ProcessPacket(packet)
	if err != nil {
		logx.WithContext(session.Context).Errorf("Packet processing failed: %v", err)
		return gnet.Close
	}

	if out != nil {
		// use callback function when handle error like retrying, push error notification or dead letter queue, etc..
		if err := c.AsyncWrite(out, nil); err != nil {
			logx.WithContext(session.Context).Errorf("Failed to send response: %v", err)
			return gnet.Close
		}
	}

	// discard only this packet (not all buffered data) to allow processing of next packets
	c.Discard(totalPacketSize)
	// logx.Info(common.ConvertToHexString(out))
	return gnet.None
}

func (ph *ProtocolHandler) OnShutdown(gnet.Engine) {
	// TODO: implement graceful shutdown in many cases
	ph.mu.Lock()
	logx.Info("Server is shutting down...")

	// close all connections
	ph.sessions.Range(func(key, value interface{}) bool {
		conn := key.(gnet.Conn)

		err := conn.Close()
		if err != nil {
			logx.Errorf("Error closing connection %v: %v", conn.RemoteAddr(), err)
		}
		ph.sessions.Delete(conn)
		return true
	})

	// stop engine
	ph.eng.Stop(ph.c)
	ph.mu.Unlock()
}

func (ph *ProtocolHandler) OnTick() (delay time.Duration, action gnet.Action) {
	delay = 10 * time.Second
	ph.sessions.Range(func(key, value interface{}) bool {
		conn := key.(gnet.Conn)
		session := value.(*Session)

		// Close inactive connections after 1 minute
		if time.Since(session.LastActive) > time.Minute {
			logx.WithContext(session.Context).Infof("Closing inactive connection: %s", conn.RemoteAddr())
			conn.Close()
			ph.sessions.Delete(conn)
		}
		return true
	})
	return
}

func (ph *ProtocolHandler) SetServiceContext(svc *svc.ServiceContext) {
	ph.svc = svc
}

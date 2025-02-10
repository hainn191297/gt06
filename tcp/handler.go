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
	// get data from packet
	data, _ := c.Peek(-1)

	packet, err := protocol.ParseAndValidatePacket(data)
	if err != nil {
		logx.Error(err)
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
	case 0x01: // Login Packet
		service = services.NewLoginDeviceService(session.Context, ph.svc)
	case 0x13, 0x23: // Heartbeat Packet (Disassemble alarm/Fall alarm)
		service = services.NewHeartbeatService(session.Context)
	case 0x12, 0x22: // Location Packet (/UTC)
		service = services.NewLocationService(session.Context)
	case 0x26:
		service = services.NewAlarmService(session.Context)
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

	// discard received data after sending response
	c.Discard(len(data))
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

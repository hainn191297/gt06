package tcp

import (
	"context"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Session struct {
	Context    context.Context
	Conn       gnet.Conn
	LastActive time.Time
}

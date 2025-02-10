package tcp

import (
	"fmt"
	"gt06/config"
	"gt06/services/svc"
	"log"
	"runtime"

	"github.com/panjf2000/gnet/v2"
)

type TCPServer struct {
	Address         string
	ProtocolHandler *ProtocolHandler
}

func NewTCPServer(address string) *TCPServer {
	return &TCPServer{
		Address: address,
	}
}

func (s *TCPServer) Start(c config.Config) error {
	protocolHandler := &ProtocolHandler{}
	serviceContext := svc.NewServiceContext(c)

	if serviceContext == nil {
		log.Fatalf("Failed to create service context")
	}

	protocolHandler.SetServiceContext(serviceContext)

	options := gnet.WithOptions(
		gnet.Options{
			Multicore:    true,
			Ticker:       true,
			NumEventLoop: runtime.NumCPU(),
		})
	s.ProtocolHandler = protocolHandler

	fmt.Printf("Starting server on %s\n", s.Address)
	return gnet.Run(protocolHandler, fmt.Sprintf("tcp://%s", s.Address), options)
}

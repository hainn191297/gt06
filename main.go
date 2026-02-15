package main

import (
	"flag"
	"gt06/conf"
	"gt06/config"
	"gt06/tcp"
)

var configFile = flag.String("c", "etc/server.yaml", "the config file path")

func main() {
	flag.Parse()

	c := config.Default()
	conf.MustLoad(*configFile, &c)

	tcpServer := tcp.NewTCPServer(c.TCPServer)
	tcpServer.Start(c)
}

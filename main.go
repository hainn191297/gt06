package main

import (
	"flag"
	"gt06/conf"
	"gt06/config"
	"gt06/tcp"
)

var configFile = flag.String("server config", "etc/server.yaml", "the config file")

func main() {
	c := config.Config{}

	conf.MustLoad(*configFile, &c)
	tcp := tcp.NewTCPServer("0.0.0.0:8000")

	tcp.Start(c)
}

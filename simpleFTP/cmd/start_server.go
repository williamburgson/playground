package main

import (
	flag "github.com/spf13/pflag"

	"simpleFTP/pkg"
)

func main() {
	addr := flag.String("host", "127.0.0.1", "Server host")
	port := flag.Int("port", 9001, "Server port")
	flag.Parse()

	cf := server.ServerConfig{Addr: *addr, Port: *port}

	server.Serve(&cf)
}

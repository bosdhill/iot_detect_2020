package main

import (
	"flag"
)

var (
	serverAddr = flag.String("server_addr", "192.168.1.72:12034", "The app server address in the format of host:port")
)

func main() {
	flag.Parse()
	ec := NewEdgeCommunication(*serverAddr)
	ec.ServeEdge()
}

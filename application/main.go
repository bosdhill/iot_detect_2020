package main

import (
	"flag"
)

var (
	serverAddr         = flag.String("server-addr", "192.168.1.72:12034", "The app server address in the format of host:port")
	appQueryServerAddr = flag.String("app-query-addr", "localhost:4204", "The app query server address in the format of host:port")
)

func main() {
	flag.Parse()

	ec := NewEdgeCommunication(*serverAddr)
	go ec.ServeEdge()

	eq, err := NewEdgeQuery(*appQueryServerAddr)
	if err != nil {
		panic(err)
	}

	eq.TestQuery()
}

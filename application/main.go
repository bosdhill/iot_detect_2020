package main

import (
	"flag"
	"log"
)
var serverAddr = flag.String("server_addr", "192.168.1.121:10000", "The server address in the format of host:port")

func main() {
	flag.Parse()
	ec, err := NewEdgeComm(*serverAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	ec.UploadImage(nil)
	//ds.Show(c)
}

package main

import (
	"flag"
	"gocv.io/x/gocv"
	"log"
)
var serverAddr = flag.String("server_addr", "192.168.1.121:10000", "The server address in the format of host:port")

const DATA_SOURCE_PATH string = "data/traffic-mini.mp4"

func main() {
	flag.Parse()
	ds, err := NewDataSource(DATA_SOURCE_PATH)
	if err != nil {
		log.Fatal(err)
		return
	}
	ec, err := NewEdgeComm(*serverAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	c := make(chan gocv.Mat)
	go ds.GetFrames(c)
	ec.UploadImage(c)
	//ds.Show(c)
}

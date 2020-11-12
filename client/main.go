package main

import (
	"flag"
	"gocv.io/x/gocv"
	"log"
)

var (
	edgeServerAddr = flag.String("edge-server-addr", "192.168.1.121:10000", "The server address in the format of host:port")
	dataSourcePath = flag.String("datapath", "data/traffic-mini.mp4", "Either the file path to the mp4 data source or 0 for webcam")
)

func main() {
	flag.Parse()

	ds, err := NewDataSource(*dataSourcePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	ec, err := NewEdgeComm(*edgeServerAddr)
	if err != nil {
		log.Fatal(err)
		return
	}

	c := make(chan gocv.Mat)
	go ds.GetFrames(c)
	ec.UploadImage(c)
}

package main

import (
	"flag"
	"gocv.io/x/gocv"
	"log"
)

var edgeServerAddr = flag.String("edge_server_addr", "192.168.1.121:10000", "The server address in the format of host:port")

const dataSourcePath string = "data/traffic-mini.mp4"

func main() {
	flag.Parse()
	ds, err := NewDataSource(dataSourcePath)
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
	//ds.Show(c)
}

package main

import (
	"flag"
	"gocv.io/x/gocv"
	"log"
	"os"
)

var (
	edgeServerAddr = flag.String("edge-server-addr", "localhost:4200", "The server address in the format of host:port")
	dataSourcePath = flag.String("datapath", "data/traffic-mini.mp4", "Either the file path to the mp4 data source or 0 for webcam")
	logsPath	   = flag.String("logs-path", "logs/", "path to logs directory")
	withContStream = flag.Bool("cont-stream", false, "whether or not to repeat the mp4 video stream indefinitely")
)

func main() {
	flag.Parse()

	file, err := os.OpenFile(*logsPath + "logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

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

	if *withContStream {
		go ds.GetFramesContinuous(c)
	} else {
		go ds.GetFrames(c)
	}

	ec.UploadImageFrames(c)
}

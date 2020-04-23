package main

import (
	"gocv.io/x/gocv"
	"log"
)

const DATA_SOURCE_PATH string = "data/traffic-mini.mp4"

func main() {
	ds, err := NewDataSource(DATA_SOURCE_PATH)
	if err != nil {
		log.Fatal(err)
		return
	}
	ec, err := NewEdgeComm()
	if err != nil {
		log.Fatal(err)
		return
	}
	c := make(chan gocv.Mat)
	go ds.GetFrames(c)
	ec.UploadImage(c)
	//ds.Show(c)
}

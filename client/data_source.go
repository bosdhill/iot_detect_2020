package main

import (
	"gocv.io/x/gocv"
	"log"
)

type dataSource struct {
	capt *gocv.VideoCapture
	Fc int
}

func NewDataSource(filePath string) (*dataSource, error) {
	log.Println("NewDataSource")
	capt, err := gocv.OpenVideoCapture(filePath)
	if err != nil {
		return nil, err
	}
	fc := int(capt.Get(gocv.VideoCaptureFrameCount))
	ds := dataSource{capt, fc}
	return &ds, nil
}

func (ds *dataSource) GetFrames(c chan <- gocv.Mat) {
	log.Println("GetFrames")
	log.Println("numFrames", ds.Fc)
	count := 0
	for {
		if count < ds.Fc {
			img := gocv.NewMat()
			ds.capt.Read(&img)
			c <- img
			count++
		} else {
			log.Println("frames read", count)
			close(c)
			break
		}
	}
}


// Note: There is an occasional NSInternalInconsistencyException on MacOS
// see https://github.com/swook/GazeML/issues/17
func (ds *dataSource) Show(c chan gocv.Mat) {
	log.Println("Show")
	window := gocv.NewWindow("client")
	for img := range c {
		window.IMShow(img)
		window.WaitKey(1)
	}
	err := window.Close()
	if err != nil {
		log.Fatal(err)
	}
}

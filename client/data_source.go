package main

import (
	"github.com/bosdhill/iot_detect_2020/sdl"
	"gocv.io/x/gocv"
	"log"
)

// DataSource used gocv for caputring image frames via a webcam.
type DataSource struct {
	capt *gocv.VideoCapture
	fc   int
}

// NewDataSource returns a new data source component.
func NewDataSource(filePath string) (*DataSource, error) {
	log.Println("NewDataSource")
	capt, err := gocv.OpenVideoCapture(filePath)
	if err != nil {
		return nil, err
	}
	fc := int(capt.Get(gocv.VideoCaptureFrameCount))
	ds := DataSource{capt, fc}
	return &ds, nil
}

// GetFrames reads frames into a channel
func (ds *DataSource) GetFrames(c chan<- gocv.Mat) {
	log.Println("GetFrames")
	log.Println("numFrames", ds.fc)

	count := 0
	// fixed number of frames from mp4
	if ds.fc != 0 {
		for i := 0; i < ds.fc; i++ {
			img := gocv.NewMat()
			ds.capt.Read(&img)
			c <- img
			count++
		}
	} else {
		// unknown number of frames from webcam
		for {
			img := gocv.NewMat()
			if ok := ds.capt.Read(&img); !ok {
				log.Println("webcam closed")
				return
			}
			if !img.Empty() {
				c <- img
				count++
			}
		}
	}
	log.Println("frames read", count)
}

// Show is used for testing and displays the image frames locally
// Note: There is an occasional NSInternalInconsistencyException on MacOS
// see https://github.com/hybridgroup/gocv/issues/599
// and https://github.com/golang/go/wiki/LockOSThread
func (ds *DataSource) Show(c chan gocv.Mat) {
	log.Println("Show")
	window := gocv.NewWindow("client")
	for img := range c {
		sdl.Show(window, &img)
	}
	err := window.Close()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"gocv.io/x/gocv"
)

type DetectionResult struct {
	detectionTime int64
	labels []string
	img gocv.Mat
	detections map[string]*BoundingBox
}

// Below can be done on request
type BoundingBox struct {
	topLeftX     int
	topLeftY     int
	bottomRightX int
	bottomRightY int
	confidence   float32
}

func (dr DetectionResult) String() string {
	ret := "\n"
	format := "detection: %s\ntopLeftX: %d\ntopLeftY: %d\nbottomRightX: %d\nbottomRightY: %d\nconf: %f\n"
	for label, b := range dr.detections {
		ret += fmt.Sprintf(format, label, b.topLeftX, b.topLeftY, b.bottomRightX, b.bottomRightY, b.confidence)
	}
	return ret
}
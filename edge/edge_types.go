package main

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
)

// EdgeContext is used to terminate go routines
type EdgeContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// DetectionResult represents the result yielded by the object detection model
type DetectionResult struct {

	// Whether or not the result has no objects
	Empty bool

	// The time of detection
	DetectionTime int64

	// A bitmap of detected labels
	Labels map[string]int

	// The matrix representation of the image frame
	Img gocv.Mat

	// A map from label to its bounding box
	LabelBoxes map[string]([]*BoundingBox)
}

// BoundingBox represents the dimensions of the detected object's bounding box
type BoundingBox struct {
	TopLeftX     int
	TopLeftY     int
	BottomRightX int
	BottomRightY int
	Confidence   float32
}

func (dr DetectionResult) String() string {
	ret := fmt.Sprintf("\n%v\n", dr.Labels)
	format := "detection: %s\ntopLeftX: %d\ntopLeftY: %d\nbottomRightX: %d\nbottomRightY: %d\nconf: %f\n"
	for label, boxSl := range dr.LabelBoxes {
		ret += label + "\n"
		for _, b := range boxSl {
			ret += fmt.Sprintf(format, label, b.TopLeftX, b.TopLeftY,
				b.BottomRightX, b.BottomRightY, b.Confidence)
		}
		ret += "\n"
	}
	return ret
}

// TODO implement below Indexers
// Need Indexer that yields DetectionResults that have LabelBoxes with label with Confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have LabelBoxes with multiple Labels with Confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have LabelBoxes with multiple Labels - Done

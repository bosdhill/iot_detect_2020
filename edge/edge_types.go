package main

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
)

// TODO create a separate file for global variables
const numClasses = 80

type EdgeContext struct {
	ctx context.Context
	cancel context.CancelFunc
}

type DetectionResult struct {
	Empty         bool
	DetectionTime int64
	Labels        map[string]bool
	Img           gocv.Mat
	LabelBoxes    map[string]([]*BoundingBox)
}

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
			ret += fmt.Sprintf(format, label, b.TopLeftX, b.TopLeftY, b.BottomRightX, b.BottomRightY, b.Confidence)
		}
		ret += "\n"
	}
	return ret
}

// TODO implement below Indexers
// Need Indexer that yields DetectionResults that have LabelBoxes with label with Confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have LabelBoxes with multiple Labels with Confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have LabelBoxes with multiple Labels - Done

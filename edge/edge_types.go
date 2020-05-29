package main

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
)

type EdgeContext struct {
	ctx context.Context
	cancel context.CancelFunc
}

type DetectionResult struct {
	empty bool
	detectionTime int64
	labels map[string]bool
	img gocv.Mat
	labelBoxes map[string]([]*BoundingBox)
}

type BoundingBox struct {
	topLeftX     int
	topLeftY     int
	bottomRightX int
	bottomRightY int
	confidence   float32
}

func (dr DetectionResult) String() string {
	ret := fmt.Sprintf("\n%v\n", dr.labels)
	format := "detection: %s\ntopLeftX: %d\ntopLeftY: %d\nbottomRightX: %d\nbottomRightY: %d\nconf: %f\n"
	for label, boxSl := range dr.labelBoxes {
		ret += label + "\n"
		for _, b := range boxSl {
			ret += fmt.Sprintf(format, label, b.topLeftX, b.topLeftY, b.bottomRightX, b.bottomRightY, b.confidence)
		}
		ret += "\n"
	}
	return ret
}

// TODO implement below Indexers
// Need Indexer that yields DetectionResults that have labelBoxes with label with confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have labelBoxes with multiple labels with confidence greater than or equal to threshold
// Need Indexer that yields DetectionResults that have labelBoxes with multiple labels - Done

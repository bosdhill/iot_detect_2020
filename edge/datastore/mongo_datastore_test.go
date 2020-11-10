package datastore

import (
	"golang.org/x/net/context"
	"log"
	"testing"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"time"
)

func TestNewMongoDataStore(t *testing.T) {
	ctx := context.Background()
	ds, err := NewMongoDataStore()

	if err != nil {
		t.Errorf("%v", err)
	}

	if err := ds.db.Ping(ctx, nil); err != nil {
		t.Errorf("%v", err)
	}
}

func TestMongoDataStore_InsertDetectionResult(t *testing.T) {
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		Labels:        nil,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := NewMongoDataStore()

	if err != nil {
		t.Errorf("%v", err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		t.Errorf("%v", err)
	}
}

func TestMongoDataStore_FilterByTime(t *testing.T) {
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		Labels:        map[string]int32{"person" : 1, "bus" : 4},
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := NewMongoDataStore()

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	// Filter by detection results from last 30 minutes in nanoseconds
	drSl, err := ds.FilterByTime(time.Duration(30 * time.Minute).Nanoseconds())

	if len(drSl) == 0 {
		t.Error(err)
	}
}
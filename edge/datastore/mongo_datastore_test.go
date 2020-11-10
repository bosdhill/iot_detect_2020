package datastore

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"golang.org/x/net/context"
	"log"
	"testing"
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

func TestMongoDataStore_DurationFilter(t *testing.T) {
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelMap:      map[string]int32{"person": 1, "bus": 4},
		Labels:        []string{"person", "bus"},
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
	drSl, err := ds.QueryBy(ds.DurationFilter(time.Duration(30 * time.Minute).Nanoseconds()))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelFilter(t *testing.T) {
	labels := []string{"person", "bus"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelMap:      map[string]int32{"person": 1, "bus": 4},
		Labels:        labels,
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

	drSl, err := ds.QueryBy(ds.LabelFilter(labels))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelMapFilter(t *testing.T) {
	labelMap := map[string]int32{"person": 1, "bus": 4}
	labels := []string{"person", "bus"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelMap:      labelMap,
		Labels:        labels,
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

	drSl, err := ds.QueryBy(ds.LabelMapFilter(labelMap))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

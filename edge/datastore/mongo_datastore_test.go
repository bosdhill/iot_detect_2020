package datastore

import (
	"golang.org/x/net/context"
	"testing"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
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
		DetectionTime: 0,
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
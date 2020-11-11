package datastore

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/context"
	"log"
	"os"
	"testing"
	"time"
)

var mongoUri string
var mongoAtlasUri string

func getEnvVars() {
	err := godotenv.Load("../credentials.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	mongoAtlasUri = os.Getenv("MONGO_ATLAS_URI")
	mongoUri = os.Getenv("MONGO_LOCAL_URI")
}

func TestMongoDataStore_InsertDetectionResult(t *testing.T) {
	getEnvVars()
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

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

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

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	// Filter by detection results from last 30 minutes in nanoseconds
	drSl, err := ds.FilterBy(ds.DurationQuery(time.Duration(30 * time.Minute).Nanoseconds()))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelsIntersectFilter(t *testing.T) {
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

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.FilterBy(ds.LabelsIntersectQuery(labels))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_And(t *testing.T) {
	labelMap := map[string]int32{"person": 1, "dog": 2, "bus": 4}
	labels := []string{"person", "dog", "bus"}
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

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.FilterBy(bson.D{
		ds.DurationQuery(time.Duration(30 * time.Minute).Nanoseconds()),
		ds.LabelsIntersectQuery(labels),
	})

	if len(drSl) == 0 {
		t.Error(err)
	}

	query := []string{"dog"}
	drSl, err = ds.FilterBy(bson.D{
		ds.DurationQuery(time.Duration(30 * time.Minute).Nanoseconds()),
		ds.LabelsIntersectQuery(query),
	})

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelsSubsetFilter(t *testing.T) {
	labels := []string{"person", "bus", "bike"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelMap:      map[string]int32{"person": 1, "bus": 4, "bike": 1},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.FilterBy(ds.LabelsSubsetQuery(labels))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelMapQuery(t *testing.T) {
	labels := []string{"person", "bus", "bike"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelMap:      map[string]int32{"person": 1, "bus": 10, "bike": 100},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := NewMongoDataStore(context.Background(), mongoUri, mongoAtlasUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.InsertDetectionResult(dr); err != nil {
		log.Printf("%v", err)
	}

	labelQuery := map[string]int32{"bike": 100}
	drSl, err := ds.FilterBy(ds.LabelMapQuery(labelQuery, GreaterThanOrEqual))

	if len(drSl) == 0 {
		t.Error(err)
	}

	drSl, err = ds.FilterBy(ds.LabelMapQuery(labelQuery, Equal))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bike": 200}
	drSl, err = ds.FilterBy(ds.LabelMapQuery(labelQuery, LessThan))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bus": 1}
	drSl, err = ds.FilterBy(ds.LabelMapQuery(labelQuery, GreaterThan))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bus": 10}
	drSl, err = ds.FilterBy(ds.LabelMapQuery(labelQuery, LessThanOrEqual))

	if len(drSl) == 0 {
		t.Error(err)
	}
}
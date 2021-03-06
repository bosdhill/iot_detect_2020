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

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Errorf("%v", err)
	}

	if err := ds.insert(dr); err != nil {
		t.Errorf("%v", err)
	}
}

func TestMongoDataStore_DurationFilter(t *testing.T) {
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelNumber:   map[string]int32{"person": 1, "bus": 4},
		Labels:        []string{"person", "bus"},
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	// Filter by detection results from last 30 minutes in nanoseconds
	drSl, err := ds.Find(ds.DurationFilter(time.Duration(30 * time.Minute).Nanoseconds()))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelsIntersectFilter(t *testing.T) {
	labels := []string{"person", "bus"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelNumber:   map[string]int32{"person": 1, "bus": 4},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.Find(ds.LabelsIntersectFilter(labels))

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
		LabelNumber:   labelMap,
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.Find(bson.D{
		ds.DurationFilter(time.Duration(30 * time.Minute).Nanoseconds()),
		ds.LabelsIntersectFilter(labels),
	})

	if len(drSl) == 0 {
		t.Error(err)
	}

	query := []string{"dog"}
	drSl, err = ds.Find(bson.D{
		ds.DurationFilter(time.Duration(30 * time.Minute).Nanoseconds()),
		ds.LabelsIntersectFilter(query),
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
		LabelNumber:   map[string]int32{"person": 1, "bus": 4, "bike": 1},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	drSl, err := ds.Find(ds.LabelsSubsetFilter(labels))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMongoDataStore_LabelMapQuery(t *testing.T) {
	labels := []string{"person", "bus", "bike"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelNumber:   map[string]int32{"person": 1, "bus": 10, "bike": 100},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	labelQuery := map[string]int32{"bike": 100}
	drSl, err := ds.Find(ds.LabelMapFilter(labelQuery, GreaterThanOrEqual))

	if len(drSl) == 0 {
		t.Error(err)
	}

	drSl, err = ds.Find(ds.LabelMapFilter(labelQuery, Equal))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bike": 200}
	drSl, err = ds.Find(ds.LabelMapFilter(labelQuery, LessThan))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bus": 1}
	drSl, err = ds.Find(ds.LabelMapFilter(labelQuery, GreaterThan))

	if len(drSl) == 0 {
		t.Error(err)
	}

	labelQuery = map[string]int32{"bus": 10}
	drSl, err = ds.Find(ds.LabelMapFilter(labelQuery, LessThanOrEqual))

	if len(drSl) == 0 {
		t.Error(err)
	}
}

func TestMarshal(t *testing.T) {
	labels := []string{"person", "bus", "bike"}
	dr := pb.DetectionResult{
		Empty:         false,
		DetectionTime: time.Now().UnixNano(),
		LabelNumber:   map[string]int32{"person": 1, "bus": 10, "bike": 100},
		Labels:        labels,
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	ds, err := New(context.Background(), mongoUri)

	if err != nil {
		t.Error(err)
	}

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	//var query = bson.D{{
	//		"$or",
	//		[]bson.E{bson.E{"labelnumber.person", bson.D{{"$gte", 1}}}},
	//	},
	//}
	//var query = bson.E{Key: "labelnumber.person", Value: bson.D{{"$gte", 1}}}
	//
	//var queryBsonE = []bson.E{query}
	//
	//var queryOr = bson.D{{"$or", bson.A{queryBsonE}}}

	//var query = bson.D{{"$or", bson.A{[]bson.E{bson.E{Key: "labelnumber.person", Value: bson.D{{"$gte", 1}}}}}}}

	var query = bson.E{"$or", bson.A{[]bson.E{bson.E{Key: "labelnumber.person", Value: bson.D{{"$gte", 1}}}}}}

	filter, err := bson.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("filter", query)

	var bF bson.E
	err = bson.Unmarshal(filter, &bF)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bF", bF)

	if err := ds.insert(dr); err != nil {
		log.Printf("%v", err)
	}

	//drSl, err := ds.Find(ds.Or([]bson.E{query}))
	drSl, err := ds.Find(bF)

	if len(drSl) == 0 {
		t.Error(err)
	}
}

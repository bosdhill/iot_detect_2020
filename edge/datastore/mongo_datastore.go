package datastore

import (
	"context"
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDataStore struct {
	db *mongo.Client
	ctx context.Context
	col *mongo.Collection
}

const dbName = "detections"
const colName = "detection_result"

// NewMongoDataStore returns a connection to the local mongodb instance
func NewMongoDataStore() (*MongoDataStore, error) {
	var err error
	db, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = db.Connect(ctx)

	// Creates database if it doesn't exist
	db.Database(dbName)

	// Creates collection if it doesn't exist
	collection := db.Database(dbName).Collection(colName)

	return &MongoDataStore{db: db, ctx: ctx, col: collection}, nil
}

func (ds *MongoDataStore) InsertWorker(drCh chan pb.DetectionResult) {

}

// InsertDetectionResult
func (ds *MongoDataStore) InsertDetectionResult(dr pb.DetectionResult) error {
	log.Println("InsertDetectionResult")
	res, err := ds.col.InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

// QueryBy queries mongodb by a specific filter or filters chained by Or or And
func (ds *MongoDataStore) QueryBy(f interface{}) ([]pb.DetectionResult, error) {
	var err error
	var query bson.D

	// Check whether its a bson document element or a bson document
	b, ok := f.(bson.E)
	if ok {
		query = append(query, b)
	} else {
		query, ok = f.(bson.D)
		if !ok {
			return nil, fmt.Errorf("filter should be either bson.D or bson.E")
		}
	}

	log.Println("query", query)
	cur, err := ds.col.Find(ds.ctx, query)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	drSl := make([]pb.DetectionResult, 0)

	for cur.Next(context.Background()) {
		dr := pb.DetectionResult{}
		err := cur.Decode(&dr)

		if err != nil {
			return nil, err
		}

		drSl = append(drSl, dr)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return drSl, nil
}

// DurationFilter creates a duration filter query the frames between now and now - duration
func (ds *MongoDataStore) DurationFilter(duration int64) bson.E {
	since := time.Now().UnixNano() - duration
	return  bson.E{"detectiontime", bson.D{{"$gte", since}}}
}

// LabelsIntersectFilter creates a filter for the labels field, where the query labels intersect the labels
func (ds *MongoDataStore) LabelsIntersectFilter(labels []string) bson.E {
	return  bson.E{"labels", bson.D{{"$in", labels}}}
}

// LabelsSubsetFilter creates a filter for the labels field, where the query labels are a subset of the labels
func (ds *MongoDataStore) LabelsSubsetFilter(labels []string) bson.E {
	return  bson.E{"labels", bson.D{ {"$all", labels}}}
}

//// LabelMapFilter creates a filter for the label map field
//func (ds *MongoDataStore) LabelMapFilter(labelMap map[string]int32) bson.E {
//	return  bson.E{"labelmap", bson.D{{"$in", labelMap}}}
//}
//
//// LabelNumberFilter creates a filter for labels that are less/equal to a number
//func (ds *MongoDataStore) LabelNumberFilter(labelMap map[string]int32) bson.E {
//	return  bson.E{"labelmap", bson.D{{"$in", labelMap}}}
//}

// And for chaining together filter queries
func (ds *MongoDataStore) And(param []bson.E) bson.D {
	var b bson.D
	for _, filter := range param {
		b = append(b, filter)
	}
	return b
}

// Or for chaining together filter queries
func (ds *MongoDataStore) Or(param []bson.E) bson.D {
	var b []bson.E
	for _, filter := range param {
		b = append(b, filter)
	}
	return bson.D{{"$or", bson.A{b}}}
}
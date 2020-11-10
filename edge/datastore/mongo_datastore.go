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
	db    *mongo.Client
	ctx   context.Context
	drCol *mongo.Collection
}

type LabelBoxesDoc struct {
	DetectionTime string
	LabelBoxes	  map[string]*pb.BoundingBox
}

const (
	dbName = "detections"
	drColName = "detection_result"
	LessThan = "$lt"
	GreaterThan = "$gt"
	Equal = "$eq"
	GreaterThanOrEqual = "$gte"
	LessThanOrEqual = "$lte"
)


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

	// Creates collections if it doesn't exist
	drCol := db.Database(dbName).Collection(drColName)

	return &MongoDataStore{db: db, ctx: ctx, drCol: drCol}, nil
}

// InsertWorker pulls from the detection result channel and calls InsertDetectionResult
func (ds *MongoDataStore) InsertWorker(drCh chan pb.DetectionResult) {
	log.Println("InsertWorker")
	for dr := range drCh {
		if err := ds.InsertDetectionResult(dr); err != nil {
			log.Printf("could not insert detection result: %v", err)
		}
	}
}

// InsertDetectionResult inserts the detection results into the detection_result
func (ds *MongoDataStore) InsertDetectionResult(dr pb.DetectionResult) error {
	log.Println("InsertDetectionResult")
	res, err := ds.drCol.InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

// FilterBy queries mongodb by a specific filter or filters chained by Or or And
func (ds *MongoDataStore) FilterBy(query interface{}) ([]pb.DetectionResult, error) {
	var err error
	var q bson.D

	// Check whether its a bson document element or a bson document
	b, ok := query.(bson.E)
	if ok {
		q = append(q, b)
	} else {
		q, ok = query.(bson.D)
		if !ok {
			return nil, fmt.Errorf("query should be of type bson.D or bson.E")
		}
	}

	log.Println("query", q)
	cur, err := ds.drCol.Find(ds.ctx, q)
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

// LabelMapQuery takes in a labelQuery, which is  a key value pairs, and comparison which is the method of comparing
// each label key to the value in the map.
//
// For example if labelMap = ["person" : 1, "dog": 3] and comparison = Equal
// will return all detection results where there is exactly one person and 3 dogs.
func (ds *MongoDataStore) LabelMapQuery(labelMap map[string]int32, comparison string) bson.D {
	prefixKey := "labelmap"
	var b bson.D
	for k, v := range labelMap {
		key := fmt.Sprintf("%s.%s", prefixKey, k)
		b = append(b, bson.E{key, bson.D{{comparison, v}}})
	}
	return b
}

// DurationQuery creates a duration filter query the frames between now and now - duration
func (ds *MongoDataStore) DurationQuery(duration int64) bson.E {
	since := time.Now().UnixNano() - duration
	return  bson.E{"detectiontime", bson.D{{GreaterThanOrEqual, since}}}
}

// LabelsIntersectQuery creates a filter for the labels field, where the query labels intersect the labels
func (ds *MongoDataStore) LabelsIntersectQuery(labels []string) bson.E {
	return  bson.E{"labels", bson.D{{"$in", labels}}}
}

// LabelsSubsetQuery creates a filter for the labels field, where the query labels are a subset of the labels
func (ds *MongoDataStore) LabelsSubsetQuery(labels []string) bson.E {
	return  bson.E{"labels", bson.D{ {"$all", labels}}}
}


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
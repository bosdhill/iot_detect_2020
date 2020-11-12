package datastore

import (
	"context"
	"fmt"
	"github.com/bosdhill/iot_detect_2020/edge/connection"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDataStore struct {
	ctx    context.Context
	client *mongo.Client
}

type LabelBoxesDoc struct {
	DetectionTime string
	LabelBoxes    map[string]*pb.BoundingBox
}

const (
	dbName             = "detections"
	drColName          = "detection_result"
	LessThan           = "$lt"
	GreaterThan        = "$gt"
	Equal              = "$eq"
	GreaterThanOrEqual = "$gte"
	LessThanOrEqual    = "$lte"
)

// NewMongoDataStore returns a connection to the local mongodb instance, with uris for local and remote instances, and
// the time to live for the local mongodb instance in seconds.
func NewMongoDataStore(ctx context.Context, mongoUri string) (*MongoDataStore, error) {
	log.Println("NewMongoDataStore")
	client, err := connection.New(ctx, mongoUri)
	if err != nil {
		return nil, err
	}
	return &MongoDataStore{ctx: ctx, client: client}, nil
}

// InsertWorker pulls from the detection result channel and calls insertDetectionResult
func (ds *MongoDataStore) InsertWorker(drCh chan pb.DetectionResult) {
	log.Println("InsertWorker")

	for dr := range drCh {
		if err := ds.localInsert(dr); err != nil {
			log.Printf("could not insert detection result to mongodb: %v", err)
		}
	}
	close(drCh)
}

// localInsert inserts the detection results into the detection_result collection
func (ds *MongoDataStore) localInsert(dr pb.DetectionResult) error {
	log.Println("localInsert")

	res, err := ds.client.Database(dbName).Collection(drColName).InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

// DeleteMany serves a delete filter query locally
func (ds *MongoDataStore) DeleteMany(filter bson.D) (*mongo.DeleteResult, error) {
	deleteRes, err := ds.client.Database(dbName).Collection(drColName).DeleteMany(ds.ctx, filter)

	if err != nil {
		return nil, err
	}
	return deleteRes, nil
}

// Find queries mongodb by a specific filter or filters chained by Or or And
func (ds *MongoDataStore) Find(filter interface{}, opt *options.FindOptions) ([]pb.DetectionResult, error) {
	var err error
	var q bson.D

	// Check whether its a bson document element or a bson document
	b, ok := filter.(bson.E)
	if ok {
		q = append(q, b)
	} else {
		q, ok = filter.(bson.D)
		if !ok {
			return nil, fmt.Errorf("filter should be of type bson.D or bson.E")
		}
	}

	log.Println("filter", q)
	cur, err := ds.client.Database(dbName).Collection(drColName).Find(ds.ctx, q, opt)
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

// LabelMapFilter takes in a labelQuery, which is  a key value pairs, and comparison which is the method of comparing
// each label key to the value in the map.
//
// For example if labelMap = ["person" : 1, "dog": 3] and comparison = Equal
// will return all detection results where there is exactly one person and 3 dogs.
func (ds *MongoDataStore) LabelMapFilter(labelMap map[string]int32, comparison string) bson.D {
	prefixKey := "labelmap"
	var b bson.D
	for k, v := range labelMap {
		key := fmt.Sprintf("%s.%s", prefixKey, k)
		b = append(b, bson.E{key, bson.D{{comparison, v}}})
	}
	return b
}

// TimeRangeFilter creates a filter query for the detection results between a time range
func (ds *MongoDataStore) TimeRangeFilter(lower, upper int64) bson.E {
	return bson.E{"detectiontime", bson.D{{GreaterThanOrEqual, lower}, {LessThan, upper}}}
}

// DurationFilter creates a duration filter query the detection results between now and now - duration
func (ds *MongoDataStore) DurationFilter(duration int64) bson.E {
	since := time.Now().UnixNano() - duration
	return bson.E{"detectiontime", bson.D{{GreaterThanOrEqual, since}}}
}

// LabelsIntersectFilter creates a filter for the labels field, where the query labels intersect the labels
func (ds *MongoDataStore) LabelsIntersectFilter(labels []string) bson.E {
	return bson.E{"labels", bson.D{{"$in", labels}}}
}

// LabelsSubsetFilter creates a filter for the labels field, where the query labels are a subset of the labels
func (ds *MongoDataStore) LabelsSubsetFilter(labels []string) bson.E {
	return bson.E{"labels", bson.D{{"$all", labels}}}
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

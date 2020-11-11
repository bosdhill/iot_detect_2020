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
	ctx         context.Context
	drCol       *mongo.Collection
	drRemoteCol *mongo.Collection
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

type DetectionResult struct {
	Created time.Time
	DetectionResult pb.DetectionResult
}

func initDBCollection(ctx context.Context, uri string) (*mongo.Collection, error) {
	log.Println("initDBCollection")
	db, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := db.Connect(ctx); err != nil {
		return nil, err
	}

	// test connection
	if err := db.Ping(ctx, nil); err != nil {
		return nil, err
	}

	// creates database if it doesn't exist
	db.Database(dbName)

	// creates collections if it doesn't exist
	drCol := db.Database(dbName).Collection(drColName)
	return drCol, nil
}

// create TTL index for retention policy for local db
func createIndex(ctx context.Context, drCol *mongo.Collection) error {
	ttlIdx := mongo.IndexModel{
		Keys:    bson.D{{"created", 1}},
		Options: options.Index().SetExpireAfterSeconds(600),
	}

	cursor, err := drCol.Indexes().List(ctx)
	if err != nil {
		return err
	}

	// find and delete existing created_1 index
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)
	for _, v := range results {
		fmt.Println(v)
		if idx, ok := v["name"]; ok {
			if idx == "created_1" {
				// drop it and recreate it if it exists
				if _, err := drCol.Indexes().DropOne(ctx, "created_1"); err != nil {
					return err
				}
				break
			}
		}
	}

	// create new created_1 index
	idx, err := drCol.Indexes().CreateOne(ctx, ttlIdx)
	if err != nil {
		return err
	}

	log.Println("created index: ", idx)
	return nil
}

// NewMongoDataStore returns a connection to the local mongodb instance
func NewMongoDataStore(ctx context.Context, mongoUri, mongoAtlasUri string) (*MongoDataStore, error) {
	log.Println("NewMongoDataStore")
	drCol, err := initDBCollection(ctx, mongoUri)
	if err != nil {
		return nil, err
	}

	if err := createIndex(ctx, drCol); err != nil {
		return nil, err
	}

	drRemoteCol, err := initDBCollection(ctx, mongoAtlasUri)
	if err != nil {
		return nil, err
	}

	return &MongoDataStore{ctx: ctx, drCol: drCol, drRemoteCol: drRemoteCol}, nil
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

	drBson := DetectionResult{
		Created: time.Now(),
		DetectionResult: dr,
	}

	res, err := ds.drCol.InsertOne(ds.ctx, drBson)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

// FilterBy queries mongodb by a specific filter or filters chained by Or or And
func (ds *MongoDataStore) FilterBy(query interface{}) ([]DetectionResult, error) {
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

	drSl := make([]DetectionResult, 0)

	for cur.Next(context.Background()) {
		dr := DetectionResult{}
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
	prefixKey := "detectionresult.labelmap"
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
	return bson.E{"detectionresult.detectiontime", bson.D{{GreaterThanOrEqual, since}}}
}

// LabelsIntersectQuery creates a filter for the labels field, where the query labels intersect the labels
func (ds *MongoDataStore) LabelsIntersectQuery(labels []string) bson.E {
	return bson.E{"detectionresult.labels", bson.D{{"$in", labels}}}
}

// LabelsSubsetQuery creates a filter for the labels field, where the query labels are a subset of the labels
func (ds *MongoDataStore) LabelsSubsetQuery(labels []string) bson.E {
	return bson.E{"detectionresult.labels", bson.D{{"$all", labels}}}
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

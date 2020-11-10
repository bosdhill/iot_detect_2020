package datastore

import (
	"context"
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

func (ds *MongoDataStore) InsertDetectionResult(dr pb.DetectionResult) error {
	log.Println("InsertDetectionResult")
	res, err := ds.col.InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

func (ds *MongoDataStore) FilterBy(f bson.D) ([]pb.DetectionResult, error) {
	cur, err := ds.col.Find(ds.ctx, f)
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

func (ds *MongoDataStore) DurationFilter(duration int64) bson.D {
	since := time.Now().UnixNano() - duration
	return  bson.D{{"detectiontime", bson.D{{"$gte", since}}}}
}

func (ds *MongoDataStore) LabelFilter(labels []string) bson.D {
	return  bson.D{{"labels", bson.D{{"$in", bson.A{labels}}}}}
}

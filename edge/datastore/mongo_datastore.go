package datastore

import (
	"context"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDataStore struct {
	db *mongo.Client
	ctx context.Context
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

	return &MongoDataStore{db: db, ctx: ctx}, nil
}

func (ds *MongoDataStore) InsertWorker(drCh chan pb.DetectionResult) {

}

func (ds *MongoDataStore) InsertDetectionResult(dr pb.DetectionResult) error {
	log.Println("InsertDetectionResult")
	collection := ds.db.Database(dbName).Collection(colName)

	res, err := collection.InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}
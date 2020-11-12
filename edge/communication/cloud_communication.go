package communication

import (
	"context"
	"github.com/bosdhill/iot_detect_2020/edge/connection"
	"github.com/bosdhill/iot_detect_2020/edge/datastore"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type CloudComm struct {
	ctx    context.Context
	client *mongo.Client
	ds     *datastore.MongoDataStore
}

const (
	dbName    = "detections"
	drColName = "detection_result"
)

// NewCloudCommunication returns a new CloudComm component, used to upload detection results from the cloud and remove
// detection results from the Edge
func NewCloudCommunication(ctx context.Context, ds *datastore.MongoDataStore, mongoAtlasUri string) (*CloudComm, error) {
	log.Println("NewCloudCommunication")
	client, err := connection.New(ctx, mongoAtlasUri)
	if err != nil {
		return nil, err
	}
	return &CloudComm{ctx: ctx, client: client, ds: ds}, nil
}

// CloudInsert creates a cloud insert job scheduled for every TTL seconds. The cloud insert job remotely inserts and
// locally removes a batchSize number of frames.
func (cloudComm *CloudComm) CloudInsert(batchSize int64, ttl int64) {
	log.Println("CloudInsert")
	cloudInsertJob := func() {
		log.Println("cloudInsertJob")

		drSl, err := cloudComm.ds.Find(bson.D{}, options.Find().SetLimit(batchSize))
		if err != nil {
			log.Printf("Error while reading from local database: %v", err)
		} else {
			log.Printf("read %v detection results from local db\n", len(drSl))
		}

		deleteRes, err := cloudComm.remoteInsert(drSl)
		if err != nil {
			log.Printf("Error while remotely inserting: %v", err)
		} else {
			log.Printf("Deleted %v detection results from the local instance\n", deleteRes.DeletedCount)
		}
	}

	// The job may run despite a previously scheduled job not completing, which means it may look like jobs
	// are failing when they're actually just run redundantly.
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(uint64(ttl)).Seconds().Do(cloudInsertJob)

	if err != nil {
		log.Printf("Error while doing job: %v", err)
	}

	s.StartAsync()
}

// remoteInsert inserts the detection results in the mongodb cloud instance, and then deletes the detection results up
// to and including he last detection result that was remotely inserted.
func (cloudComm *CloudComm) remoteInsert(drSl []pb.DetectionResult) (*mongo.DeleteResult, error) {
	log.Println("remoteInsert")
	var drInsert []interface{}
	for _, dr := range drSl {
		drInsert = append(drInsert, dr)
	}

	res, err := cloudComm.client.Database(dbName).Collection(drColName).InsertMany(cloudComm.ctx, drInsert)
	if err != nil {
		return nil, err
	}

	log.Printf("uploaded ids: %v", res.InsertedIDs)

	lastDetectionTime := drSl[len(drSl)-1].DetectionTime

	query := bson.D{{"detectiontime", bson.D{{datastore.LessThanOrEqual, lastDetectionTime}}}}

	return cloudComm.ds.DeleteMany(query)
}

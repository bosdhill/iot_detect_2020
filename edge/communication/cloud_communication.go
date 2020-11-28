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
	"sync"
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
	var mutex = &sync.Mutex{}

	// cloudInsertJob is run every TTL seconds, and handles cloud upload in 3 phases:
	// 		1) read batchSize number of detection results that haven't been uploaded
	//		2) upload batchSize number of detection results
	// 		3) remove the image field from each uploaded detection result
	cloudInsertJob := func() {
		log.Println("cloudInsertJob")
		mutex.Lock()
		defer mutex.Unlock()
		// Find all detection results that haven't been uploaded (img.image not nil)
		drSl, err := cloudComm.ds.Find(bson.D{{"img.image", bson.D{{"$ne", nil}}}}, options.Find().SetLimit(batchSize))
		if err != nil {
			log.Printf("Error while reading from local database: %v", err)
		} else {
			log.Printf("read %v detection results from local db\n", len(drSl))
		}

		// If there are no results to upload, exit
		if len(drSl) == 0 {
			return
		}

		// Upload phase
		dTime, err := cloudComm.remoteInsert(drSl)
		if err != nil {
			log.Printf("Error while remotely inserting: %v", err)
		}
		log.Printf("Last detection time was: %v\n", *dTime)

		// Update phase
		updateRes, err := cloudComm.updateLocalDr(*dTime)
		if err != nil {
			log.Printf("Error while locally updating: %v", err)
		} else {
			log.Printf("Updated %v detection results on the local instance\n", updateRes.ModifiedCount)
		}
	}

	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(uint64(ttl)).Seconds().Do(cloudInsertJob)

	if err != nil {
		log.Printf("Error while doing job: %v", err)
	}

	s.StartAsync()
}

// remoteInsert inserts the detection results in the mongodb cloud instance, and then deletes the detection results up
// to and including he last detection result that was remotely inserted.
func (cloudComm *CloudComm) remoteInsert(drSl []pb.DetectionResult) (*int64, error) {
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
	return &lastDetectionTime, nil
}

// deleteLocalDr deletes the local detection results up to the last detection time that was uploaded
func (cloudComm *CloudComm) deleteLocalDr(lastDetectionTime int64) (*mongo.DeleteResult, error) {
	log.Println("deleteLocalDr")
	query := bson.D{{"detectiontime", bson.D{{datastore.LessThanOrEqual, lastDetectionTime}}}}
	return cloudComm.ds.DeleteMany(query)
}

// updateLocalDr updates the local detection results up to the last detection time that was uploaded in order to remove
// the image binary so the Edge retains metadata only
func (cloudComm *CloudComm) updateLocalDr(lastDetectionTime int64) (*mongo.UpdateResult, error) {
	log.Println("updateLocalDr")
	filter := bson.D{{"detectiontime", bson.D{{datastore.LessThanOrEqual, lastDetectionTime}}}}
	update := bson.D{{"$set", bson.D{{"img.image", nil}}}}
	return cloudComm.ds.UpdateMany(filter, update)
}
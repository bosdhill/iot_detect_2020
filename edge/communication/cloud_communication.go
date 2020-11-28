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

// CloudInsert creates two jobs that are mutually exclusive:
// 		1) a cloud insert job scheduled for every uploadTTL seconds
// 		2) a local delete job scheduled for every deleteTTL seconds
func (cComm *CloudComm) CloudInsert(batchSize int64, uploadTTL, deleteTTL int64) {
	log.Println("CloudInsert")
	var mtx = &sync.Mutex{}

	// cloudInsertJob is run every uploadTTL seconds, and handles cloud upload in 3 phases:
	// 		1) read batchSize number of detection results that haven't been uploaded
	//		2) upload batchSize number of detection results
	// 		3) remove the image field from each uploaded detection result
	cloudInsertJob := func() {
		log.Println("cloudInsertJob")
		mtx.Lock()
		defer mtx.Unlock()

		// Read phase
		drSl, err := cComm.localFind(batchSize)
		if err != nil {
			log.Printf("Error while reading from local database: %v", err)
		} else {
			log.Printf("Found %v detection results from local db\n", len(drSl))
		}

		if len(drSl) == 0 {
			return
		}

		// Upload phase
		dTime, err := cComm.remoteInsert(drSl)
		if err != nil {
			log.Printf("Error while remotely inserting: %v", err)
		}
		log.Printf("Last detection time inserted was: %v\n", *dTime)

		// Update phase
		updateRes, err := cComm.localUpdate(*dTime)
		if err != nil {
			log.Printf("Error while locally updating: %v", err)
		} else {
			log.Printf("Updated %v detection results on the local instance\n", updateRes.ModifiedCount)
		}
	}

	// localDeleteJob is run every deleteTTL seconds, and scans for all detection results with img.image equal to null
	// and deletes them
	localDeleteJob := func() {
		log.Println("localDeleteJob")
		mtx.Lock()
		defer mtx.Unlock()

		filter := bson.D{{"img.image", nil}}
		delRes, err := cComm.ds.DeleteMany(filter)
		if err != nil {
			log.Printf("Error while locally deleting: %v", err)
		} else {
			log.Printf("Deleted %v detection results on the local instance\n", delRes.DeletedCount)
		}
	}

	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(uint64(uploadTTL)).Seconds().Do(cloudInsertJob)
	if err != nil {
		log.Fatalf("Error while doing job: %v", err)
	}

	_, err = s.Every(uint64(deleteTTL)).Seconds().Do(localDeleteJob)
	if err != nil {
		log.Fatalf("Error while doing job: %v", err)
	}

	s.StartAsync()
}

// remoteInsert inserts the detection results in the mongodb cloud instance, and then deletes the detection results up
// to and including he last detection result that was remotely inserted.
func (cComm *CloudComm) remoteInsert(drSl []pb.DetectionResult) (*int64, error) {
	log.Println("remoteInsert")
	var drInsert []interface{}
	for _, dr := range drSl {
		drInsert = append(drInsert, dr)
	}

	res, err := cComm.client.Database(dbName).Collection(drColName).InsertMany(cComm.ctx, drInsert)
	if err != nil {
		return nil, err
	}

	log.Printf("uploaded ids: %v", res.InsertedIDs)

	lastDetectionTime := drSl[len(drSl)-1].DetectionTime
	return &lastDetectionTime, nil
}

// localFind returns all detection results that haven't been uploaded, which is determined by whether img.image is nil
func (cComm *CloudComm) localFind(batchSize int64) ([]pb.DetectionResult, error) {
	filter := bson.D{{"img.image", bson.D{{"$ne", nil}}}}
	drSl, err := cComm.ds.Find(filter, options.Find().SetLimit(batchSize))
	if err != nil {
		return nil, err
	}
	return drSl, nil
}

// localDelete deletes the local detection results up to the last detection time that was uploaded
func (cComm *CloudComm) localDelete(lastDetectionTime int64) (*mongo.DeleteResult, error) {
	log.Println("localDelete")
	filter := bson.D{{"detectiontime", bson.D{{datastore.LessThanOrEqual, lastDetectionTime}}}}
	return cComm.ds.DeleteMany(filter)
}

// localUpdate updates the local detection results up to the last detection time that was uploaded in order to remove
// the image binary so the Edge retains metadata only
func (cComm *CloudComm) localUpdate(lastDetectionTime int64) (*mongo.UpdateResult, error) {
	log.Println("localUpdate")
	filter := bson.D{{"detectiontime", bson.D{{datastore.LessThanOrEqual, lastDetectionTime}}}}
	update := bson.D{{"$set", bson.D{{"img.image", nil}}}}
	return cComm.ds.UpdateMany(filter, update)
}

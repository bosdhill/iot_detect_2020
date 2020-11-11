package datastore

import (
	"context"
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDataStore struct {
	ctx         context.Context
	drLocalCol  *mongo.Collection
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


// initDBCollection connects to the local and remote mongodb instances and creates the detections collection
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

// NewMongoDataStore returns a connection to the local mongodb instance, with uris for local and remote instances, and
// the time to live for the local mongodb instance in seconds.
func NewMongoDataStore(ctx context.Context, mongoUri, mongoAtlasUri string, ttlSec int32) (*MongoDataStore, error) {
	log.Println("NewMongoDataStore")
	drCol, err := initDBCollection(ctx, mongoUri)
	if err != nil {
		return nil, err
	}

	drRemoteCol, err := initDBCollection(ctx, mongoAtlasUri)
	if err != nil {
		return nil, err
	}

	return &MongoDataStore{ctx: ctx, drLocalCol: drCol, drRemoteCol: drRemoteCol}, nil
}

// InsertWorker pulls from the detection result channel and calls insertDetectionResult
func (ds *MongoDataStore) InsertWorker(drCh chan pb.DetectionResult) {
	log.Println("InsertWorker")
	var lower int64
	var upper int64
	lower = time.Now().UnixNano()
	cloudUploadJob := func() {
		log.Println("cloudUploadJob")
		upper = lower + time.Duration(30 * time.Second).Nanoseconds()
		// fetch detection results from last 30 seconds
		log.Printf("fetching detection results in detectiontime range [%v, %v)\n", lower, upper)
		drSl, err := ds.FilterBy(ds.TimeRangeQuery(lower, upper))
		if err != nil {
			log.Printf("Error while reading from local database: %v", err)
		}
		log.Printf("read %v detection results from local db\n", len(drSl))

		// upload to cloud
		if err := ds.cloudUpload(drSl, ds.TimeRangeQuery(lower, upper)); err != nil {
			log.Printf("Error uploading to mongodb atlas: %v", err)
		}
		lower = upper
	}

	// Defines a new scheduler that schedules and runs jobs
	// NOTE: the job may run despite a previously scheduled job not completing, which means it may look like jobs
	// are failing when they're actually just run redundantly. The jobs are run in a way so that they will be able to
	// upload detection results within the entire time range, where some times in the range may or may not have any
	// detection results. Since uploading to the cloud depends on upload speed and network bandwidth if the number of
	// results is large, a single batched upload may take a while.
	//
	// For example, a cloudUploadJob may be scheduled at time T and then reads out entries to upload, which it waits
	// for before issuing a success message.
	//
	// Another job may be scheduled at time T + t that reads an empty slice, and reports a failure message. This is okay
	// because the time range of the previous upload already accounts for the detection results within that range.
	//
	// A possible optimization would be to chunk or batch the uploads to a fixed set instead of batching uploads
	// over a certain fixed time range.
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(30).Seconds().Do(cloudUploadJob)

	if err != nil {
		log.Printf("Error while doing job: %v", err)
	}

	// scheduler starts running jobs and current thread continues to execute
	s.StartAsync()

	for dr := range drCh {
		if err := ds.localInsert(dr); err != nil {
			log.Printf("could not insert detection result to mongodb: %v", err)
		}
	}
	close(drCh)
}

// cloudUpload uploads detection results to mongodb atlas and removes those same detection results from the local
// mongodb instance
func (ds *MongoDataStore) cloudUpload(drSl []pb.DetectionResult, timeRange bson.E) error {
	log.Println("cloudUpload")

	var drInsert []interface{}
	for _, dr := range drSl {
		drInsert = append(drInsert, dr)
	}

	res, err := ds.drRemoteCol.InsertMany(ds.ctx, drInsert)

	if err != nil {
		return err
	}
	log.Printf("uploaded ids: %v", res.InsertedIDs)

	deleteRes, err := ds.drLocalCol.DeleteMany(ds.ctx, elemToDoc(timeRange))

	if err != nil {
		return err
	}

	log.Printf("deleted %v detection results locally", deleteRes.DeletedCount)

	return nil
}

// localInsert inserts the detection results into the detection_result collection
func (ds *MongoDataStore) localInsert(dr pb.DetectionResult) error {
	log.Println("localInsert")

	res, err := ds.drLocalCol.InsertOne(ds.ctx, dr)

	if err != nil {
		return err
	}

	log.Printf("inserted detection result with id: %v", res.InsertedID)
	return nil
}

func elemToDoc(elem bson.E) bson.D {
	var doc bson.D
	doc = append(doc, elem)
	return doc
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
	cur, err := ds.drLocalCol.Find(ds.ctx, q)
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

// TimeRangeQuery creates a filter query for the detection results between a time range
func (ds *MongoDataStore) TimeRangeQuery(lower, upper int64) bson.E {
	return bson.E{"detectiontime", bson.D{{GreaterThanOrEqual, lower}, {LessThan, upper}}}
}

// DurationQuery creates a duration filter query the detection results between now and now - duration
func (ds *MongoDataStore) DurationQuery(duration int64) bson.E {
	since := time.Now().UnixNano() - duration
	return bson.E{"detectiontime", bson.D{{GreaterThanOrEqual, since}}}
}

// LabelsIntersectQuery creates a filter for the labels field, where the query labels intersect the labels
func (ds *MongoDataStore) LabelsIntersectQuery(labels []string) bson.E {
	return bson.E{"labels", bson.D{{"$in", labels}}}
}

// LabelsSubsetQuery creates a filter for the labels field, where the query labels are a subset of the labels
func (ds *MongoDataStore) LabelsSubsetQuery(labels []string) bson.E {
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

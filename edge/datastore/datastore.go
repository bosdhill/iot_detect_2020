package datastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	od "github.com/bosdhill/iot_detect_2020/edge/detection"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var dbTable = "detection"

// DataStore stores the sqllite DB reference
type DataStore struct {
	db  *sql.DB
	ctx context.Context
}

// WriteOutImg handles writing out the annotated image frames
func (ds *DataStore) WriteOutImg() {
	//for label, box := range res.LabelBoxes {
	//	gocv.Rectangle(&Img, image.Rect(box.topleft.X, box.topleft.Y, box.bottomright.X, box.bottomright.Y), color.RGBA{230, 25, 75, 0}, 1)
	//	gocv.PutText(&Img, box.label, image.Point{box.topleft.X, box.topleft.Y - 5}, gocv.FontHersheySimplex, 0.5, color.RGBA{230, 25, 75, 0}, 1)
	//}
	//gocv.IMWrite("detect.jpg", Img)
}

// Get gets a frame from the data store.
// TODO Should have a case when there is an Empty Get
// TODO test Get with a gRPC application interface (next step)
func (ds *DataStore) Get() error {
	log.Println("Get")
	rows, err := ds.db.Query("SELECT * FROM labels WHERE person = FALSE AND bus = TRUE")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var dTime int64
	var labels [od.NumClasses]bool
	ret := []interface{}{&dTime}
	for _, b := range labels {
		ret = append(ret, &b)
	}
	for rows.Next() {
		err = rows.Scan(ret...)
		if err != nil {
			log.Fatalf("InsertWorker: scan error with err = %s", err)
		}
		log.Printf("Row and Col Values")
		dt := ret[0].(*int64)
		log.Println(*dt)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return nil
}

// InsertWorker grabs detection results from the channel and inserts them into
// the tables
func (ds *DataStore) InsertWorker(drCh chan pb.DetectionResult) {
	log.Println("InsertWorker")
	for dr := range drCh {
		ds.InsertImageTable(&dr)
		ds.InsertBoundingBoxTable(&dr)
		ds.InsertLabelsTable(&dr)
	}
	ds.Get()
}

// InsertImageTable inserts detection results into the image table.
func (ds *DataStore) InsertImageTable(dr *pb.DetectionResult) {
	log.Println("InsertImageTable")
	txn, err := ds.db.Begin()
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	// Insert into images table
	stmt, err := txn.Prepare("INSERT INTO images VALUES(?,?,?)")
	defer func() {
		stmt = nil
		txn = nil
		// to prevent memory leak
		//if err := dr.Img.Close(); err != nil {
		//	log.Fatalf("InsertWorker: could not close image with err = %s", err)
		//}
		//runtime.GC()
	}()
	if err != nil {
		log.Fatalf("InsertWorker: could not prepare insert into images table with err = %s", err)
	}
	_, err = stmt.Exec(dr.DetectionTime, dr.Img.Image, dr.Empty)
	if err != nil {
		log.Fatalf("InsertWorker: could not exec insert into images table with err = %s", err)
	}
	err = stmt.Close()
	if err != nil {
		log.Fatalf("InsertWorker: could not close stmt for images table with err = %s", err)
	}
	if err := txn.Commit(); err != nil {
		log.Fatalf("InsertWorker: could not commit images table txn with err = %s", err)
	}
}

// InsertBoundingBoxTable inserts detection results into the bounding box table.
func (ds *DataStore) InsertBoundingBoxTable(dr *pb.DetectionResult) {
	log.Println("InsertBoundingBoxTable")
	txn, err := ds.db.Begin()
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	stmt, err := txn.Prepare("INSERT INTO bounding_box VALUES(?,?,?,?)")
	defer func() {
		stmt = nil
		txn = nil
	}()
	if err != nil {
		log.Fatalf("InsertWorker: could not prepare insert into bounding_box with err = %s", err)
	}
	// Insert into bounding_box table
	for label, bboxSl := range dr.LabelBoxes {
		for _, bbox := range bboxSl.LabelBoxes {
			b, err := json.Marshal(bbox)
			if err != nil {
				log.Fatalf("InsertWorker: could not marshal bbox with err = %s", err)
			}
			fmt.Print(bbox)
			_, err = stmt.Exec(dr.DetectionTime, label, b, bbox.Confidence)
			if err != nil {
				log.Fatalf("InsertWorker: could not exec insert into bounding_box with err = %s", err)
			}
		}
	}
	err = stmt.Close()
	if err != nil {
		log.Fatalf("InsertWorker: could not close stmt with err = %s", err)
	}
	if err := txn.Commit(); err != nil {
		log.Fatalf("InsertWorker: could not commit images table txn with err = %s", err)
	}
}

// InsertLabelsTable inserts detection results into the labels table
func (ds *DataStore) InsertLabelsTable(dr *pb.DetectionResult) {
	log.Println("InsertLabelsTable")
	txn, err := ds.db.Begin()
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	// Insert default row with foreign key of detection time in the labels table
	prepLabels := fmt.Sprintf("INSERT INTO labels (detection_time) VALUES(?)")
	stmt, err := txn.Prepare(prepLabels)
	defer func() {
		stmt = nil
		txn = nil
	}()
	if err != nil {
		log.Fatalf("InsertWorker: could not prepare insert into labels with err = %s", err)
	}
	_, err = stmt.Exec(dr.DetectionTime)
	if err != nil {
		log.Fatalf("InsertWorker: could not exec insert statement into labels with err = %s", err)
	}
	err = stmt.Close()
	if err != nil {
		log.Fatalf("InsertWorker: could not close stmt with err = %s", err)
	}
	err = txn.Commit()
	if err != nil {
		log.Fatalf("InsertWorker: could not commit txn for bounding_box and images with err = %s", err)
	}
	// Update each column of that newly added row
	for label, numDetected := range dr.LabelMap {
		_, err := ds.db.Exec(fmt.Sprintf("UPDATE labels SET %s = %v WHERE detection_time = %d", label, numDetected, dr.DetectionTime))
		if err != nil {
			log.Fatalf("InsertWorker: could not prepare update into labels err = %s", err)
		}
	}
}

// NewDataStore creates a new data store component with a frame db with the
// images, bounding_box, and labels tables. These tables are used to store the
// detection results.
// TODO update schema for labels, and make it number of labels instead of existence of labels
func NewDataStore(ctx context.Context) (*DataStore, error) {
	log.Println("NewDataStore")
	os.Remove("./object_detection.db")
	db, err := sql.Open("sqlite3", "./object_detection.db")
	if err != nil {
		log.Printf("%q: %v\n", err, db)
	}

	// TODO make labels table dynamic so we can swap out models with different label sets
	createImageTable := `
	CREATE TABLE IF NOT EXISTS images (
	  detection_time integer,
	  image blob,
	  objects_detected boolean,
	  PRIMARY KEY(detection_time)
	);
	CREATE TABLE IF NOT EXISTS bounding_box (
	  detection_time integer,
	  label string,
	  dimensions blob,
	  Confidence float,
	  FOREIGN KEY(detection_time) REFERENCES image(detection_time)
	);
	CREATE TABLE IF NOT EXISTS labels (
	 detection_time integer,
	"person" int DEFAULT 0,
	"bicycle" int DEFAULT 0,
	"car" int DEFAULT 0,
	"motorcycle" int DEFAULT 0,
	"airplane" int DEFAULT 0,
	"bus" int DEFAULT 0,
	"train" int DEFAULT 0,
	"truck" int DEFAULT 0,
	"boat" int DEFAULT 0,
	"traffic light" int DEFAULT 0,
	"fire hydrant" int DEFAULT 0,
	"stop sign" int DEFAULT 0,
	"parking meter" int DEFAULT 0,
	"bench" int DEFAULT 0,
	"bird" int DEFAULT 0,
	"cat" int DEFAULT 0,
	"dog" int DEFAULT 0,
	"horse" int DEFAULT 0,
	"sheep" int DEFAULT 0,
	"cow" int DEFAULT 0,
	"elephant" int DEFAULT 0,
	"bear" int DEFAULT 0,
	"zebra" int DEFAULT 0,
	"giraffe" int DEFAULT 0,
	"backpack" int DEFAULT 0,
	"umbrella" int DEFAULT 0,
	"handbag" int DEFAULT 0,
	"tie" int DEFAULT 0,
	"suitcase" int DEFAULT 0,
	"frisbee" int DEFAULT 0,
	"skis" int DEFAULT 0,
	"snowboard" int DEFAULT 0,
	"sports ball" int DEFAULT 0,
	"kite" int DEFAULT 0,
	"baseball bat" int DEFAULT 0,
	"baseball glove" int DEFAULT 0,
	"skateboard" int DEFAULT 0,
	"surfboard" int DEFAULT 0,
	"tennis racket" int DEFAULT 0,
	"bottle" int DEFAULT 0,
	"wine glass" int DEFAULT 0,
	"cup" int DEFAULT 0,
	"fork" int DEFAULT 0,
	"knife" int DEFAULT 0,
	"spoon" int DEFAULT 0,
	"bowl" int DEFAULT 0,
	"banana" int DEFAULT 0,
	"apple" int DEFAULT 0,
	"sandwich" int DEFAULT 0,
	"orange" int DEFAULT 0,
	"broccoli" int DEFAULT 0,
	"carrot" int DEFAULT 0,
	"hot dog" int DEFAULT 0,
	"pizza" int DEFAULT 0,
	"donut" int DEFAULT 0,
	"cake" int DEFAULT 0,
	"chair" int DEFAULT 0,
	"couch" int DEFAULT 0,
	"potted plant" int DEFAULT 0,
	"bed" int DEFAULT 0,
	"dining table" int DEFAULT 0,
	"toilet" int DEFAULT 0,
	"tv" int DEFAULT 0,
	"laptop" int DEFAULT 0,
	"mouse" int DEFAULT 0,
	"remote" int DEFAULT 0,
	"keyboard" int DEFAULT 0,
	"cell phone" int DEFAULT 0,
	"microwave" int DEFAULT 0,
	"oven" int DEFAULT 0,
	"toaster" int DEFAULT 0,
	"sink" int DEFAULT 0,
	"refrigerator" int DEFAULT 0,
	"book" int DEFAULT 0,
	"clock" int DEFAULT 0,
	"vase" int DEFAULT 0,
	"scissors" int DEFAULT 0,
	"teddy bear" int DEFAULT 0,
	"hair drier" int DEFAULT 0,
	"toothbrush" int DEFAULT 0,
	FOREIGN KEY(detection_time) REFERENCES image(detection_time)
	);
	`

	_, err = db.Exec(createImageTable)
	if err != nil {
		log.Printf("%q: %s\n", err, createImageTable)
		return nil, err
	}
	log.Println("Successfully created tables")

	return &DataStore{db: db, ctx: ctx}, nil
}

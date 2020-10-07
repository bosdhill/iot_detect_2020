package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"runtime"
)

var dbTable = "detection"

// DataStore stores the sqllite DB reference
type DataStore struct {
	db   *sql.DB
	eCtx *EdgeContext
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
	var labels [numClasses]bool
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
func (ds *DataStore) InsertWorker(drCh chan DetectionResult) {
	log.Println("InsertWorker")
	for dr := range drCh {
		ds.InsertImageTable(&dr)
		ds.InsertBoundingBoxTable(&dr)
		ds.InsertLabelsTable(&dr)
	}
	ds.Get()
}

// InsertImageTable inserts detection results into the image table.
func (ds *DataStore) InsertImageTable(dr *DetectionResult) {
	txn, err := ds.db.Begin()
	log.Println(dr)
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	// Insert into images table
	stmt, err := txn.Prepare("INSERT INTO images VALUES(?,?,?)")
	defer func() {
		stmt = nil
		txn = nil
		// to prevent memory leak
		if err := dr.Img.Close(); err != nil {
			log.Fatalf("InsertWorker: could not close image with err = %s", err)
		}
		runtime.GC()
	}()
	if err != nil {
		log.Fatalf("InsertWorker: could not prepare insert into images table with err = %s", err)
	}
	_, err = stmt.Exec(dr.DetectionTime, dr.Img.ToBytes(), dr.Empty)
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
func (ds *DataStore) InsertBoundingBoxTable(dr *DetectionResult) {
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
		for _, bbox := range bboxSl {
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
func (ds *DataStore) InsertLabelsTable(dr *DetectionResult) {
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
	for label, exists := range dr.Labels {
		_, err := ds.db.Exec(fmt.Sprintf("UPDATE labels SET %s = %v WHERE detection_time = %d", label, exists, dr.DetectionTime))
		if err != nil {
			log.Fatalf("InsertWorker: could not prepare update into labels err = %s", err)
		}
	}
}

// NewDataStore creates a new data store component with a frame db with the
// images, bounding_box, and labels tables. These tables are used to store the
// detection results.
// TODO update schema for labels, and make it number of labels instead of existence of labels
func NewDataStore(eCtx *EdgeContext) (*DataStore, error) {
	log.Println("NewDataStore")
	os.Remove("./object_detection.db")
	db, err := sql.Open("sqlite3", "./object_detection.db")

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
	"person" boolean DEFAULT false,
	"bicycle" boolean DEFAULT false,
	"car" boolean DEFAULT false,
	"motorcycle" boolean DEFAULT false,
	"airplane" boolean DEFAULT false,
	"bus" boolean DEFAULT false,
	"train" boolean DEFAULT false,
	"truck" boolean DEFAULT false,
	"boat" boolean DEFAULT false,
	"traffic light" boolean DEFAULT false,
	"fire hydrant" boolean DEFAULT false,
	"stop sign" boolean DEFAULT false,
	"parking meter" boolean DEFAULT false,
	"bench" boolean DEFAULT false,
	"bird" boolean DEFAULT false,
	"cat" boolean DEFAULT false,
	"dog" boolean DEFAULT false,
	"horse" boolean DEFAULT false,
	"sheep" boolean DEFAULT false,
	"cow" boolean DEFAULT false,
	"elephant" boolean DEFAULT false,
	"bear" boolean DEFAULT false,
	"zebra" boolean DEFAULT false,
	"giraffe" boolean DEFAULT false,
	"backpack" boolean DEFAULT false,
	"umbrella" boolean DEFAULT false,
	"handbag" boolean DEFAULT false,
	"tie" boolean DEFAULT false,
	"suitcase" boolean DEFAULT false,
	"frisbee" boolean DEFAULT false,
	"skis" boolean DEFAULT false,
	"snowboard" boolean DEFAULT false,
	"sports ball" boolean DEFAULT false,
	"kite" boolean DEFAULT false,
	"baseball bat" boolean DEFAULT false,
	"baseball glove" boolean DEFAULT false,
	"skateboard" boolean DEFAULT false,
	"surfboard" boolean DEFAULT false,
	"tennis racket" boolean DEFAULT false,
	"bottle" boolean DEFAULT false,
	"wine glass" boolean DEFAULT false,
	"cup" boolean DEFAULT false,
	"fork" boolean DEFAULT false,
	"knife" boolean DEFAULT false,
	"spoon" boolean DEFAULT false,
	"bowl" boolean DEFAULT false,
	"banana" boolean DEFAULT false,
	"apple" boolean DEFAULT false,
	"sandwich" boolean DEFAULT false,
	"orange" boolean DEFAULT false,
	"broccoli" boolean DEFAULT false,
	"carrot" boolean DEFAULT false,
	"hot dog" boolean DEFAULT false,
	"pizza" boolean DEFAULT false,
	"donut" boolean DEFAULT false,
	"cake" boolean DEFAULT false,
	"chair" boolean DEFAULT false,
	"couch" boolean DEFAULT false,
	"potted plant" boolean DEFAULT false,
	"bed" boolean DEFAULT false,
	"dining table" boolean DEFAULT false,
	"toilet" boolean DEFAULT false,
	"tv" boolean DEFAULT false,
	"laptop" boolean DEFAULT false,
	"mouse" boolean DEFAULT false,
	"remote" boolean DEFAULT false,
	"keyboard" boolean DEFAULT false,
	"cell phone" boolean DEFAULT false,
	"microwave" boolean DEFAULT false,
	"oven" boolean DEFAULT false,
	"toaster" boolean DEFAULT false,
	"sink" boolean DEFAULT false,
	"refrigerator" boolean DEFAULT false,
	"book" boolean DEFAULT false,
	"clock" boolean DEFAULT false,
	"vase" boolean DEFAULT false,
	"scissors" boolean DEFAULT false,
	"teddy bear" boolean DEFAULT false,
	"hair drier" boolean DEFAULT false,
	"toothbrush" boolean DEFAULT false,
	FOREIGN KEY(detection_time) REFERENCES image(detection_time)
	);
	`

	_, err = db.Exec(createImageTable)
	if err != nil {
		log.Printf("%q: %s\n", err, createImageTable)
		return nil, err
	}
	log.Println("Succesfully created tables")

	return &DataStore{db: db, eCtx: eCtx}, nil
}

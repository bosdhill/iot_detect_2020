package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
	"encoding/json"
)

var dbTable = "detection"

type dataStore struct {
	db *sql.DB
	eCtx *EdgeContext
}

// Data store can handle writing out the annotated image frames, or it can be handled by client
func (ds *dataStore) WriteOutImg() {
	//for label, box := range res.LabelBoxes {
	//	gocv.Rectangle(&Img, image.Rect(box.topleft.X, box.topleft.Y, box.bottomright.X, box.bottomright.Y), color.RGBA{230, 25, 75, 0}, 1)
	//	gocv.PutText(&Img, box.label, image.Point{box.topleft.X, box.topleft.Y - 5}, gocv.FontHersheySimplex, 0.5, color.RGBA{230, 25, 75, 0}, 1)
	//}
	//gocv.IMWrite("detect.jpg", Img)
}

//func (ds *dataStore) GetRange(int64)

// TODO Should have a case when there is an Empty Get
func (ds *dataStore) Get() error {
	log.Println("Get")
	//txn := ds.db.Txn(false)
	//defer txn.Abort()
	//
	//it, err := txn.Get(dbTable, "id")
	//if err != nil {
	//	return err
	//}
	//
	//log.Println("All the LabelBoxes:")
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	dr := obj.(*DetectionResult)
	//	log.Println(dr.Labels)
	//}
	//log.Println("end of get")
	//
	//log.Println("Only LabelBoxes with bus:")
	//it, err = txn.Get(dbTable, "Labels", "bus", "t")
	//if err != nil {
	//	return err
	//}
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	dr := obj.(*DetectionResult)
	//	log.Println(dr.Labels)
	//}
	//
	//// Range scan
	//it, err = txn.LowerBound(dbTable, "TopLeftX",2)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("People aged 25 - 35:")
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	p := obj.(*DetectionResult)
	//	if p.DetectionTime > 35 {
	//		break
	//	}
	//}
	return nil
}

func (ds *dataStore) InsertImageTable(dr *DetectionResult) {
	txn, err := ds.db.Begin()
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	// Insert into images table
	stmt, err := txn.Prepare("INSERT INTO images VALUES(?,?,?)")
	defer func() {
		stmt = nil
		txn = nil
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

func (ds *dataStore) InsertBoundingBoxTable(dr *DetectionResult) {
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

func (ds *dataStore) InsertWorker(drCh chan DetectionResult) {
	log.Println("InsertWorker")
//	db, err := sql.Open("sqlite3", "./object_detection.db")
//	defer db.Close()

	for dr := range drCh {
		ds.InsertImageTable(&dr)
		ds.InsertBoundingBoxTable(&dr)
		ds.InsertLabelsTable(&dr)
	}
	rows, err := ds.db.Query("SELECT * FROM labels")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var dTime int64
	var labels [80] bool
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
		log.Println(dTime, labels)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

}

func (ds *dataStore) InsertLabelsTable(dr *DetectionResult) {
	txn, err := ds.db.Begin()
	if err != nil {
		log.Fatalf("InsertWorker: could not db.begin with err = %s", err)
	}
	//// Commit bounding_box and images row insertions
	//
	//// Insert default row with foreign key of detection time in the Labels table
	//insertLabels := fmt.Sprintf("INSERT INTO labels VALUES (%d, %s", dr.DetectionTime, strings.Repeat("false, ", 79) + "false)")
	//_, err = ds.db.Exec(insertLabels)
	//if err != nil {
	//	log.Fatalf("InsertWorker: could not exec for labels with err = %s", err)
	//}
	// Insert default row with foreign key of detection time in the Labels table
	prepLabels := fmt.Sprintf("INSERT INTO labels VALUES(?, %s", strings.Repeat("?, ", 79)+"?)")
	stmt, err := txn.Prepare(prepLabels)
	defer func() {
		stmt = nil
		txn = nil
	}()
	if err != nil {
		log.Fatalf("InsertWorker: could not prepare insert into labels with err = %s", err)
	}
	_, err = stmt.Exec(dr.DetectionTime, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false)
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
		//_, err = stmt.Exec(dr.DetectionTime, label, exists)
		//if err != nil {
		//	log.Fatalf("InsertWorker: could not exec update into labels err = %s", err)
		//}
	}
}

func (ds *dataStore) Insert(dr []DetectionResult) {
	//for _, p := range dr {
	//	if err := txn.Insert(dbTable, &p); err != nil {
	//		panic(err)
	//	}
	//}
	//txn.Commit()
}

func NewDataStore(eCtx *EdgeContext) (*dataStore, error) {
	log.Println("NewDataStore")

	db, err := sql.Open("sqlite3", "./object_detection.db")
//	defer db.Close()

	createImageTable := `
	CREATE TABLE IF NOT EXISTS images (
	  detection_time integer,
	  image blob,
	  results boolean,
	  PRIMARY KEY(detection_time)
	);
	CREATE TABLE IF NOT EXISTS bounding_box (
	  detection_time integer,
	  label string,
	  dimensions blob,
	  Confidence float, 
	  FOREIGN KEY(detection_time) REFERENCES image(detection_time)
	);
	CREATE TABLE IF NOT EXISTS Labels (
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

	return &dataStore{db: db, eCtx: eCtx}, nil
}

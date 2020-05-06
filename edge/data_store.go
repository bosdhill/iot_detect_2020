package main

import (
	"context"
	"github.com/hashicorp/go-memdb"
	"log"
	"time"
)

// Use watch feature

// Create a sample struct

var dbTable = "detection"

type dataStore struct {
	db *memdb.MemDB
}

// Data store can handle writing out the annotated image frames, or it can be handled by client
func (ds *dataStore) WriteOutImg() {
	//for label, box := range res.detections {
	//	gocv.Rectangle(&img, image.Rect(box.topleft.X, box.topleft.Y, box.bottomright.X, box.bottomright.Y), color.RGBA{230, 25, 75, 0}, 1)
	//	gocv.PutText(&img, box.label, image.Point{box.topleft.X, box.topleft.Y - 5}, gocv.FontHersheySimplex, 0.5, color.RGBA{230, 25, 75, 0}, 1)
	//}
	//gocv.IMWrite("detect.jpg", img)
}

// TODO Should have a case when there is an empty Get
func (ds *dataStore) Get() error {
	log.Println("Get")
	txn := ds.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(dbTable, "id")
	if err != nil {
		return err
	}

	log.Println("All the detections:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		dr := obj.(*DetectionResult)
		log.Println(dr.labels)
	}
	log.Println("end of get")

	log.Println("Only detections with bus:")
	it, err = txn.Get(dbTable, "labels", "bus", "t")
	if err != nil {
		return err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		dr := obj.(*DetectionResult)
		log.Println(dr.labels)
	}

	// Range scan
	//it, err = txn.LowerBound(dbTable, "id", 25)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("People aged 25 - 35:")
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	p := obj.(*DetectionResult)
	//	if p.detectionTime > 35 {
	//		break
	//	}
	//}
	return nil
}

func (ds *dataStore) InsertWorker(ctx context.Context, drCh chan DetectionResult) {
	log.Println("InsertWorker")
	txn := ds.db.Txn(true)
	for {
		select {
			case <-ctx.Done():
				log.Println("commit")
				txn.Commit()
				time.Sleep(5 * time.Second)
				ds.Get()
				return
			case dr := <- drCh:
				//log.Println("inserting...")
				log.Println(dr)
				if err := txn.Insert(dbTable, &dr); err != nil {
					panic(err)
				}
				//log.Println("Finished inserting...")
		}
	}
}

func (ds *dataStore) Insert(dr []DetectionResult) {
	txn := ds.db.Txn(true)
	for _, p := range dr {
		if err := txn.Insert(dbTable, &p); err != nil {
			panic(err)
		}
	}
	txn.Commit()
}

func NewDataStore() (*dataStore, error) {
	log.Println("NewDataStore")

	// TODO use logical clocks (unix time) so it's easier to get ranges
	// 	last 24 hours == curr_unix_time - 24 hours in unix time
	// 	last x unix_t == curr_unix_time - x
	// TODO should there be separate tables for images with detections vs images without detections?
	//  This is explained in paper I wrote.
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			dbTable : {
				Name: dbTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id" : {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "detectionTime"},
					},
					"labels" : {
						Name: "labels",
						Unique: false,
						Indexer: &memdb.StringMapFieldIndex{Field: "labels", Lowercase: true},
						// Some detections may not produce any results, but we should store image frames anyways
						AllowMissing: true,
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &dataStore{db: db}, nil
}

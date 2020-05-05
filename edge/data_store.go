package main

import (
	"github.com/hashicorp/go-memdb"
	"log"
)

// Use watch feature

// Create a sample struct

var dbTable = "detection"

type Detection struct {
	detectionTime int
	labels []string
}

type dataStore struct {
	db *memdb.MemDB
}

func (ds *dataStore) InsertWorker(drCh chan DetectionResult) {
	txn := ds.db.Txn(true)
	for dr := range drCh {
		log.Println(dr)
		if err := txn.Insert(dbTable, dr); err != nil {
			panic(err)
		}
	}
}

func (ds *dataStore) Insert(dr []DetectionResult) {
	txn := ds.db.Txn(true)
	for _, p := range dr {
		if err := txn.Insert(dbTable, p); err != nil {
			panic(err)
		}
	}
	txn.Commit()
}

func NewDataStore() (*dataStore, error) {
	log.Println("data store")

	// TODO use logical clocks (unix time) so it's easier to get ranges
	// 	last 24 hours == curr_unix_time - 24 hours in unix time
	// 	last x unix_t == curr_unix_time - x
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
						Indexer: &memdb.StringSliceFieldIndex{Field: "labels"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
		panic(err)
	}
	return &dataStore{db: db}, nil

	//
	//txn := db.Txn(true)
	//
	//detections := []*Detection{
	//	&Detection{10002320, []string{"person", "bus"} },
	//	&Detection{10005420, []string{"person", "bus", "car"} },
	//}
	//
	//for _, p := range detections {
	//	if err := txn.Insert("detection", p); err != nil {
	//		panic(err)
	//	}
	//}
	//
	//txn.Commit()
	//
	//// Create read-only transaction
	//txn = db.Txn(false)
	//defer txn.Abort()
	//
	//raw, err := txn.First("detection", "labels", "person")
	//
	//// Say hi!
	//fmt.Printf("Detection time %d!\n", raw.(*Detection).detectionTime)
	//
	//// List all the people
	//it, err := txn.Get("detection", "id")
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("All the detections:")
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	p := obj.(*Detection)
	//	fmt.Println(p.labels)
	//}

	//// Range scan over people with ages between 25 and 35 inclusive
	//it, err = txn.LowerBound("person", "age", 25)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("People aged 25 - 35:")
	//for obj := it.Next(); obj != nil; obj = it.Next() {
	//	p := obj.(*Person)
	//	if p.Age > 35 {
	//		break
	//	}
	//	fmt.Printf("  %s is aged %d\n", p.Name, p.Age)
	//}



}

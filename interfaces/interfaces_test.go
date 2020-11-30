package interfaces

import (
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
)

//func TestEvent_Flags_Enum(t *testing.T) {
//	flags := Event_METADATA
//	assert.Equal(t, flags&Event_METADATA, Event_METADATA, "flags should be only Event_METADATA")
//	assert.Equal(t, flags&Event_ANNOTATED == 0, true, "flags should be only Event_METADATA")
//	assert.Equal(t, flags&Event_BOXES == 0, true, "flags should be only Event_METADATA")
//
//	flags = Event_ANNOTATED | Event_BOXES
//	assert.Equal(t, flags&Event_ANNOTATED != 0, true, "flags should contain Event_ANNOTATED")
//	assert.Equal(t, flags&Event_BOXES != 0, true, "flags should contain Event_BOXES")
//	assert.Equal(t, flags&Event_CONFIDENCE == 0, true, "flags should not contain Event_CONFIDENCE")
//
//	flags = Event_CONFIDENCE | Event_ANNOTATED | Event_BOXES | Event_PERSIST | Event_IMAGE
//	assert.Equal(t, flags&Event_CONFIDENCE != 0, true, "flags should contain Event_CONFIDENCE")
//	assert.Equal(t, flags&Event_ANNOTATED != 0, true, "flags should contain Event_ANNOTATED")
//	assert.Equal(t, flags&Event_BOXES != 0, true, "flags should contain Event_BOXES")
//	assert.Equal(t, flags&Event_PERSIST != 0, true, "flags should contain Event_PERSIST")
//	assert.Equal(t, flags&Event_IMAGE != 0, true, "flags should contain Event_IMAGE")
//}

func TestComparisonType_Descriptor(t *testing.T) {
	filterEvent := bson.D{
		bson.E{
			Key: "$or",
			Value: bson.D{
				bson.E{
					Key: "person",
					Value: bson.E{
						Key:   "$gte",
						Value: 1,
					},
				},
				bson.E{
					Key: "bus",
					Value: bson.E{
						Key:   "$gte",
						Value: 1,
					},
				},
			},
		},
	}
	log.Println(filterEvent.Map())

	// If it has a set of or statements
	ma, ok := filterEvent.Map()["$or"]
	if ok {
		b, ok := ma.(bson.D)
		log.Println(b.Map())
		if ok {
			for label, elt := range b.Map() {
				log.Println("label:", label, "elt:", elt)
				b, ok := elt.(bson.E)
				log.Println(ok)
				log.Println(b.Key)
				log.Println(b.Value.(int))
			}
		}
	}

	// filter is detectiontime is greater than or equal to 2414213412122 AND (person is greater than or equal to 1 OR bus is greater than or equal to 1)
	//filter := bson.D{
	//	bson.E{
	//		Key: "detectiontime",
	//		Value: bson.D{
	//			bson.E{
	//				Key:   "$gte",
	//				Value: 2414213412122,
	//			},
	//		},
	//	},
	//	bson.E{
	//		Key: "$or",
	//		Value: bson.D{
	//			bson.E{
	//				Key: "person",
	//				Value: bson.E{
	//					Key:   "$gte",
	//					Value: 1,
	//				},
	//			},
	//			bson.E{
	//				Key: "bus",
	//				Value: bson.E{
	//					Key:   "$gte",
	//					Value: 1,
	//				},
	//			},
	//		},
	//	},
	//}


	//filterEventAnd := bson.D{
	//	bson.E{
	//		Key: "$and",
	//		Value: bson.D{
	//			bson.E{
	//				Key: "person",
	//				Value: bson.E{
	//					Key:   "$gte",
	//					Value: 1,
	//				},
	//			},
	//			bson.E{
	//				Key: "bus",
	//				Value: bson.E{
	//					Key:   "$gte",
	//					Value: 1,
	//				},
	//			},
	//		},
	//	},
	//}
	//
	//log.Println(filterEventAnd.Map()["$and"])
	//
	//log.Println(filter)
	//log.Println(filter.Map())
	//mFilter, err := bson.Marshal(filter)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//var umFilter bson.D
	//err = bson.Unmarshal(mFilter, &umFilter)
	//m := umFilter.Map()
	//log.Println(m["detectiontime"])
	//log.Println(m["$or"])
	//v := m["detectiontime"]
	//b, ok := v.(bson.D)
	//if ok {
	//	log.Println(b.Map())
	//}
	//
	//var allFilter = bson.D{bson.E{
	//	Key: "labels",
	//	Value: bson.D{
	//		bson.E{
	//			Key: "$all",
	//			Value: bson.A{
	//				"person",
	//				"bus",
	//			},
	//		},
	//	},
	//}}
	//
	//log.Println(allFilter)
	//log.Println(allFilter.Map())
	//ma, ok = allFilter.Map()["labels"]
	//if ok {
	//	b, ok := ma.(bson.D)
	//	if ok {
	//		log.Println(b)
	//		log.Println(b.Map())
	//		a, ok := b.Map()["$all"]
	//		if ok {
	//			log.Println(a)
	//			aSl, ok := a.(bson.A)
	//			if ok {
	//				log.Println(aSl)
	//				for _, v := range aSl {
	//					log.Println(v)
	//				}
	//			}
	//		}
	//	}
	//}
	//
	//var ret bool
	//ret = ret && (1 == 4)
	//ret = ret && false
}

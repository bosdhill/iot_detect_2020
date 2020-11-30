package realtimefilter

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
)

func TestNew(t *testing.T) {
	orQuery := bson.D{
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

	andQuery := bson.D{
		bson.E{
			Key: "$and",
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

	arrayQuery := bson.D{
		bson.E{
			Key: "labels",
			Value: bson.D{
				bson.E{
					Key: "$all",
					Value: bson.A{
						"person",
						"bus",
					},
				},
			},
		},
	}

	mOrQuery, _ := bson.Marshal(orQuery)
	mAndQuery, _ := bson.Marshal(andQuery)
	mArrayQuery, _ := bson.Marshal(arrayQuery)

	cases := []struct {
		description    string
		detectedLabels map[string]int32
		eventFilters   *pb.EventFilters
		expected       bool
	}{
		{
			description:    "or query",
			detectedLabels: nil,
			eventFilters: &pb.EventFilters{
				EventFilters: []*pb.EventFilter{
					{
						TimeFilter: 0,
						Name:       "OrQuery",
						Filter:     mOrQuery,
						Flags:      0,
					},
				},
			},
		},
		{
			description:    "and query",
			detectedLabels: nil,
			eventFilters: &pb.EventFilters{
				EventFilters: []*pb.EventFilter{
					{
						TimeFilter: 0,
						Name:       "AndQuery",
						Filter:     mAndQuery,
						Flags:      0,
					},
				},
			},
		},
		{
			description:    "array query",
			detectedLabels: nil,
			eventFilters: &pb.EventFilters{
				EventFilters: []*pb.EventFilter{
					{
						TimeFilter: 0,
						Name:       "ArrayQuery",
						Filter:     mArrayQuery,
						Flags:      0,
					},
				},
			},
		},
	}

	for _, c := range cases {
		_, err := New(c.eventFilters)
		if err != nil {
			t.Errorf(c.description, err)
		}
	}
}

func TestSet_GetEvents(t *testing.T) {
	orQuery := bson.D{
		bson.E{
			Key: "$or",
			Value: bson.D{
				bson.E{
					Key: "person",
					Value: bson.E{
						Key:   "$gte",
						Value: 5,
					},
				},
				bson.E{
					Key: "bus",
					Value: bson.E{
						Key:   "$gte",
						Value: 4,
					},
				},
			},
		},
	}

	andQuery := bson.D{
		bson.E{
			Key: "$and",
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
						Key:   "$lte",
						Value: 10,
					},
				},
			},
		},
	}

	arrayQuery := bson.D{
		bson.E{
			Key: "labels",
			Value: bson.D{
				bson.E{
					Key: "$all",
					Value: bson.A{
						"person",
						"bus",
					},
				},
			},
		},
	}

	mOrQuery, _ := bson.Marshal(orQuery)
	mAndQuery, _ := bson.Marshal(andQuery)
	mArrayQuery, _ := bson.Marshal(arrayQuery)

	dr := &pb.DetectionResult{
		Empty:         false,
		DetectionTime: 0,
		LabelNumber:   map[string]int32{
			"bus" : 2,
			"person" : 3,
		},
		Labels:        []string{"bus", "person"},
		Img: &pb.Image{
			Image: nil,
			Rows:  0,
			Cols:  0,
			Type:  0,
		},
		LabelBoxes: nil,
	}

	cases := []struct {
		description    string
		eventFilters   *pb.EventFilters
		expected       bool
	}{
		{
			description:    "or query",
			eventFilters: &pb.EventFilters{
				EventFilters: []*pb.EventFilter{
					{
						TimeFilter: 0,
						Name:       "OrQuery",
						Filter:     mOrQuery,
						Flags:      0,
					},
					{
						TimeFilter: 0,
						Name:       "AndQuery",
						Filter:     mAndQuery,
						Flags:      0,
					},
					{
						TimeFilter: 0,
						Name:       "ArrayQuery",
						Filter:     mArrayQuery,
						Flags:      0,
					},
				},
			},
		},
	}

	eSet, err := New(cases[0].eventFilters)
	if err != nil {
		log.Fatal(err)
	}

	events := eSet.GetEvents(dr)

	for _, e := range events {
		if e.Name == "OrQuery" {
			log.Fatal()
		}
	}
}



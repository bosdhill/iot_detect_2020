package main

import (
	"context"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"log"
	"math"
	"sync"
	"time"
)

// EdgeComm contains an EventQuery client stub
type EventQuery struct {
	client pb.EventQueryClient
}

// NewEventQueryClient returns an NewEventQueryClient
func NewEventQueryClient(addr string) (*EventQuery, error) {
	log.Println("NewEventQueryClient")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}

	client := pb.NewEventQueryClient(conn)
	return &EventQuery{client}, nil
}

// Find makees a Find event filter query request to the EventQuery server on the Edge
func (eQuery *EventQuery) Find(eFilter *pb.EventFilter) ([]*pb.Event, error) {
	log.Printf("Find")
	ctx, _ := context.WithCancel(context.Background())
	resp, err := eQuery.client.Find(ctx, &pb.FindRequest{EventFilter: eFilter})
	if err != nil {
		return nil, err
	}
	return resp.GetEvents(), nil
}

func TestQuery(group *sync.WaitGroup) {
	defer group.Done()

	eQuery, err := NewEventQueryClient(*appQueryServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("TestQuery")
	// Test query
	time.Sleep(30 * time.Second)
	var query interface{} = bson.E{
		Key: "$or",
		Value: bson.A{[]bson.E{
			{
				Key: "labelnumber.person",
				Value: bson.D{{
					Key:   "$gte",
					Value: 1,
				}},
			},
			{
				Key: "labelnumber.bus",
				Value: bson.D{{
					Key:   "$gte",
					Value: 1,
				}},
			},
		},
		},
	}

	filter, err := bson.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}

	// Get Events from the last 60 seconds that match filter
	eFilter := &pb.EventFilter{
		Seconds: 60,
		Name:    "TestQueryEvent",
		Filter:  filter,
		Flags:   0,
	}

	events, err := eQuery.Find(eFilter)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range events {
		log.Println(e.GetName())
		log.Println(e.GetDetectionResult().GetDetectionTime())
		log.Println(e.GetDetectionResult().GetLabelNumber())
	}
}

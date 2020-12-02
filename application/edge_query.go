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

// EdgeComm contains an EdgeQuery client stub
type EdgeQuery struct {
	client pb.EdgeQueryClient
}

// NewEdgeQuery returns an NewEdgeQueryClient
func NewEdgeQuery(addr string) (*EdgeQuery, error) {
	log.Println("NewEdgeQuery")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}

	client := pb.NewEdgeQueryClient(conn)
	return &EdgeQuery{client}, nil
}

// Find makees a Find event filter query request to the EdgeQuery server on the Edge
func (eQuery *EdgeQuery) Find(eFilter *pb.EventFilter) (*pb.Events, error) {
	log.Printf("Find")
	ctx, _ := context.WithCancel(context.Background())
	events, err := eQuery.client.Find(ctx, eFilter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func TestQuery(group *sync.WaitGroup) {
	defer group.Done()

	eQuery, err := NewEdgeQuery(*appQueryServerAddr)
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

	for _, e := range events.GetEvents() {
		log.Println(e.GetName())
		log.Println(e.GetDetectionResult().GetDetectionTime())
		log.Println(e.GetDetectionResult().GetLabelNumber())
	}
}

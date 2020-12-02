package main

import (
	"context"
	"github.com/bosdhill/iot_detect_2020/edge/realtimefilter"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/golang/protobuf/ptypes/empty"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"io"
	"log"
	"math"
	"sync"
)

// EventOnDetect is used to serve the application's requests
type EventOnDetect struct {
	client       pb.EventOnDetectClient
	ctx          context.Context
	eventFilters *pb.EventFilters
	rtFilter     *realtimefilter.Set
	uuid 		string
}

// New creates an event on detect client
func NewEventOnDetect(ctx context.Context, addr string) (*EventOnDetect, error) {
	log.Println("NewEventOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewEventOnDetectClient(conn)
	return &EventOnDetect{client: client, ctx: ctx}, nil
}

func (eod *EventOnDetect) GetLabels() (*pb.Labels, error) {
	log.Println("GetLabels")
	var err error
	resp, err := eod.client.GetLabels(eod.ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.GetLabels(), nil
}

// RegisterEventFilters is used to register the application's eventFilters
func (eod *EventOnDetect) RegisterApp(labels *pb.Labels) error {
	log.Println("RegisterApp")
	eventFilters := &pb.EventFilters{}
	if labels.Labels["person"] && labels.Labels["bus"] {
		filters := map[string]bson.D{
			"FourPersonsOrFourBuses": {
				bson.E{
					Key: "$or",
					Value: bson.D{
						bson.E{
							Key: "person",
							Value: bson.E{
								Key:   "$gte",
								Value: 2,
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
			},
			"AtLeastOnePersonAndBus": {
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
			},
			"PersonAndBus": {
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
			},
			"AtLeast3People": {
				bson.E{
					Key: "$or",
					Value: bson.D{
						bson.E{
							Key: "person",
							Value: bson.E{
								Key:   "$gte",
								Value: 3,
							},
						},
					},
				},
			},
		}

		for name, filter := range filters {
			mFilter, err := bson.Marshal(filter)
			if err != nil {
				log.Fatal(err)
			}

			eFilter := &pb.EventFilter{
				Name:   name,
				Filter: mFilter,
				Flags:  0,
			}

			eventFilters.EventFilters = append(eventFilters.EventFilters, eFilter)
		}
	}

	req := &pb.RegisterAppRequest{
		EventFilters: eventFilters,
	}

	resp, err := eod.client.RegisterApp(eod.ctx, req)
	if err != nil {
		return err
	}

	eod.uuid = resp.Uuid
	log.Println("received uuid: ", eod.uuid)

	return nil
}

func (eod *EventOnDetect) StreamEvents() {
	log.Println("StreamEvents")
	req := &pb.StreamEventsRequest{
		Uuid: eod.uuid,
	}

	stream, err := eod.client.StreamEvents(eod.ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := stream.Recv()

		if err == io.EOF {
			log.Println("EOF")
			if err = stream.CloseSend(); err != nil {
				log.Printf("error when closing stream: %v", err)
				break
			}
		}

		if err != nil {
			log.Fatal(err)
		}

		for _, e := range resp.GetEvents() {
			log.Println(e.Name)
			log.Println(e.GetDetectionResult().GetDetectionTime())
			log.Println(e.GetDetectionResult().GetLabelNumber())
		}
	}
}

func TestEventOnDetect(group *sync.WaitGroup) {
	log.Println("TestEventOnDetect")
	defer group.Done()

	ctx := context.Background()
	ec, err := NewEventOnDetect(ctx, *eodServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	labels, err := ec.GetLabels()
	if err != nil {
		log.Fatal(err)
	}

	if err = ec.RegisterApp(labels); err != nil {
		log.Fatal(err)
	}

	ec.StreamEvents()
}

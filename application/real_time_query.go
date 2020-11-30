package main

import (
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes/empty"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math"
	"net"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

// EdgeComm is used to wrap the server for serving actions issued by the Edge
type EdgeComm struct {
	server pb.EventOnDetectServer
	lis    net.Listener
}

const (
	LessThan           = "$lt"
	GreaterThan        = "$gt"
	Equal              = "$eq"
	GreaterThanOrEqual = "$gte"
	LessThanOrEqual    = "$lte"
)

// RegisterEventFilters is called by the Edge to filter results for events it cares about
func (comm *EdgeComm) RegisterEventFilters(ctx context.Context, labels *pb.Labels) (*pb.EventFilters, error) {
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

	return eventFilters, nil
}

// SendEvent receives the Events sent by the Edge
func (comm *EdgeComm) SendEvents(ctx context.Context, events *pb.Events) (*empty.Empty, error) {
	log.Println("SendEvent")
	for _, e := range events.GetEvents() {
		log.Println(e.Name)
		log.Println(e.GetDetectionResult().GetDetectionTime())
		log.Println(e.GetDetectionResult().GetLabelNumber())
	}
	return &empty.Empty{}, nil
}

// StreamEvents TODO
func (comm *EdgeComm) StreamEvents(server pb.EventOnDetect_StreamEventsServer) error {
	panic("implement me")
}

// NewEdgeCommunication returns a new edge communication component
func NewEdgeCommunication(addr string) *EdgeComm {
	log.Println("NewEdgeCommunication")
	flag.Parse()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	EdgeComm := &EdgeComm{lis: lis}
	return EdgeComm
}

// ServeEdge serves action events requests from the Edge
func (comm *EdgeComm) ServeEdge() error {
	log.Println("ServeEdge")
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventOnDetectServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

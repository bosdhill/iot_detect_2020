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
	events := &pb.EventFilters{}

	if labels.Labels["person"] {
		// marshal mongodb bson filter
		filter, err := bson.Marshal(bson.D{
			bson.E{
				Key: "person",
				Value: bson.D{
					bson.E{
						Key:   GreaterThanOrEqual,
						Value: 1,
					},
				},
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		event := &pb.EventFilter{
			Name:   "MoreThanOnePersonFound",
			Filter: filter,
			Flags:  0,
		}

		events.EventFilters = append(events.EventFilters, event)
	}

	return events, nil
}

// SendEvent receives the action sent by the Edge
func (comm *EdgeComm) SendEvent(ctx context.Context, event *pb.Event) (*empty.Empty, error) {
	log.Println("SendEvent")
	log.Printf("Received: %v, %v\n", event.GetDetectionResult().Labels, event.GetDetectionResult().GetLabelBoxes())
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

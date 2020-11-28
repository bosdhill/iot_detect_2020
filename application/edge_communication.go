package main

import (
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/hashstructure"
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

// RegisterEventFilters is called by the Edge to filter results for events it cares about
// TODO this should be an event driven pattern, such as in
//  https://stephenafamo.com/blog/implementing-an-event-driven-system-in-go/
func (comm *EdgeComm) RegisterEventFilters(ctx context.Context, labels *pb.Labels) (*pb.EventFilters, error) {
	events := &pb.EventFilters{}

	// Example of application setting an Event with EventConditions specified for triggering an Action
	// TODO need to handle person or bus, not just person and bus
	if labels.Labels["person"] {
		event := &pb.EventFilter{
			LabelEvents: map[string]*pb.EventConditions{
				"person": {
					ConfThreshold: 0.30,
					Quantity:      1,
					QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
					Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
				},
			},
			Labels:          []string{"person"},
			DistanceMeasure: pb.EventFilter_DISTANCE_MEASURE_UNSPECIFIED,
			Flags:           uint32(pb.EventFilter_METADATA),
		}
		uid, err := hashstructure.Hash(event, nil)
		if err != nil {
			log.Println("RegisterEventFilters: uid hash failed")
			return nil, err
		}
		event.Uid = uid
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

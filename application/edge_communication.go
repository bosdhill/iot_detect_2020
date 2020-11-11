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
	server pb.ActionOnDetectServer
	lis    net.Listener
}

// RegisterEvents is called by the Edge to filter results for events it cares about
// TODO this should be an event driven pattern, such as in
//  https://stephenafamo.com/blog/implementing-an-event-driven-system-in-go/
func (comm *EdgeComm) RegisterEvents(ctx context.Context, labels *pb.Labels) (*pb.Events, error) {
	events := &pb.Events{}

	// Example of application setting an Event with EventConditions specified for triggering an Action
	// TODO need to handle person or bus, not just person and bus
	if labels.Labels["person"] && labels.Labels["bus"] {
		event := &pb.Event{
			LabelEvents: map[string]*pb.EventConditions{
				"person": {
					ConfThreshold: 0.30,
					Quantity:      1,
					QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
					Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
				},
				"bus": {
					ConfThreshold: 0.30,
					Quantity:      1,
					QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
					Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
				},
			},
			Labels:          []string{"person", "bus"},
			DistanceMeasure: pb.Event_DISTANCE_MEASURE_UNSPECIFIED,
			Flags:           uint32(pb.Event_METADATA),
		}
		uid, err := hashstructure.Hash(event, nil)
		if err != nil {
			log.Println("RegisterEvents: uid hash failed")
			return nil, err
		}
		event.Uid = uid
		events.Events = append(events.Events, event)
	}

	return events, nil
}

// SendAction receives the action sent by the Edge
func (comm *EdgeComm) SendAction(ctx context.Context, action *pb.Action) (*empty.Empty, error) {
	log.Println("SendAction")
	log.Printf("Received: %v, %v\n", action.GetDetectionResult().Labels, action.GetDetectionResult().GetLabelBoxes())
	return &empty.Empty{}, nil
}

// StreamActions TODO
func (comm *EdgeComm) StreamActions(pb.ActionOnDetect_StreamActionsServer) error {
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
	pb.RegisterActionOnDetectServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

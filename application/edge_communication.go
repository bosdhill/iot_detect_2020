package main

import (
	"context"
	"flag"
	"github.com/golang/protobuf/ptypes/empty"
	"log"
	"net"
	"github.com/mitchellh/hashstructure"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

type edgeComm struct {
	server pb.ActionOnDetectServer
	lis    net.Listener
}

// TODO this should be an event driven pattern, such as in
//  https://stephenafamo.com/blog/implementing-an-event-driven-system-in-go/
// SetEvents is a method implemented by the application that determines what labels the application cares about
func (comm *edgeComm) SetEvents(ctx context.Context, labels *pb.Labels) (*pb.Events, error) {
	events := &pb.Events{}

	// Example of application setting an Event with EventConditions specified for triggering an Action
	if labels.Labels["person"] && labels.Labels["car"] {
		event := &pb.Event{
			LabelEvents: map[string]*pb.EventConditions{
				"person": {
					ConfThreshold: 0.30,
					Quantity:      1,
					QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
					Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
				},
				"car": {
					ConfThreshold: 0.30,
					Quantity:      1,
					QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
					Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
				},
			},
			Labels: []string{"person", "car"},
			DistanceMeasure: pb.Event_DISTANCE_MEASURE_UNSPECIFIED,
			Flags:           uint32(pb.Event_METADATA),
		}
		uid, err := hashstructure.Hash(event, nil)
		if err != nil {
			log.Println("SetEvents: uid hash failed")
			return nil, err
		}
		event.Uid = uid
		events.Events = append(events.Events, event)
	}

	return events, nil
}

func (comm *edgeComm) ActionOnDetect(context.Context, *pb.Action) (*empty.Empty, error) {
	panic("implement me")
}

func (comm *edgeComm) EventStream(pb.ActionOnDetect_EventStreamServer) error {
	panic("implement me")
}

func NewEdgeCommunication(addr string) *edgeComm {
	log.Println("NewEdgeCommunication")
	flag.Parse()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	edgeComm := &edgeComm{lis: lis}
	return edgeComm
}


func (comm *edgeComm) ServeEdge() error {
	log.Println("ServeEdge")
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterActionOnDetectServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

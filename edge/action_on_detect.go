package main

import (
	"github.com/bosdhill/iot_detect_2020/edge/event_set"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
	"log"
)

// ActionOnDetect isused to serve the application's requests
type ActionOnDetect struct {
	client    pb.ActionOnDetectClient
	eCtx      *EdgeContext
	events    *pb.Events
	event_set *event_set.EventSet
}

// NewActionOnDetect starts up a grpc server and
func NewActionOnDetect(eCtx *EdgeContext, addr string) (*ActionOnDetect, error) {
	log.Println("NewActionOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.DialContext(eCtx.ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewActionOnDetectClient(conn)
	return &ActionOnDetect{client: client, eCtx: eCtx}, nil
}

// RegisterEvents is used to register events
func (aod *ActionOnDetect) RegisterEvents(labels map[string]bool) (*event_set.EventSet, error) {
	var err error
	aod.events, err = aod.client.RegisterEvents(aod.eCtx.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}
	aod.event_set = event_set.New(aod.events)
	return aod.event_set, nil
}

// CheckEvents checks whether the detection result sastifies any event set by
// the application. If it does, it creates an action and sends it to the
// application
func (aod *ActionOnDetect) CheckEvents(dr *DetectionResult) {
	// TODO probably slow -- just store []string of labels separately
	// needs list of strings for labels
	events := aod.event_set.Find(dr.Labels)
	log.Println("found", events)

	// return dummy action
	//action := pb.Action{
	//	Labels: map[string]*pb.BoundingBox{
	//		"": {
	//			TopLeftX:     0.0,
	//			TopLeftY:     0.0,
	//			BottomRightX: 0.0,
	//			BottomRightY: 0.0,
	//			Confidence:   0.0,
	//		},
	//	},
	//	Img:          nil,
	//	AnnotatedImg: nil,
	//}

	// create action for each event
	//aod.client.SendAction(aod.eCtx.ctx, &action)
}

// func (aod *Appaod) SendAction(Action) {

// }

// func (aod *Appaod) StreamActions(Action) {

// }

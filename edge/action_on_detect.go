package main

import (
	"github.com/bosdhill/iot_detect_2020/edge/event_set"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
	"log"
	"math"
)

// ActionOnDetect isused to serve the application's requests
type ActionOnDetect struct {
	client   pb.ActionOnDetectClient
	eCtx     *EdgeContext
	events   *pb.Events
	eventSet *event_set.EventSet
}

// NewActionOnDetect starts up a grpc server and
func NewActionOnDetect(eCtx *EdgeContext, addr string) (*ActionOnDetect, error) {
	log.Println("NewActionOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithMaxMsgSize(math.MaxInt32))
	conn, err := grpc.DialContext(eCtx.ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewActionOnDetectClient(conn)
	return &ActionOnDetect{client: client, eCtx: eCtx}, nil
}

// RegisterEvents is used to register the application's events
func (aod *ActionOnDetect) RegisterEvents(labels map[string]bool) (*event_set.EventSet, error) {
	log.Println("RegisterEvents")
	var err error
	aod.events, err = aod.client.RegisterEvents(aod.eCtx.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}
	aod.eventSet = event_set.New(aod.events)
	return aod.eventSet, nil
}

// CheckEvents checks whether the detection result satisfies any event conditions set by
// the application. If it does, it creates an action and sends it to the
// application.
func (aod *ActionOnDetect) CheckEvents(dr *pb.DetectionResult) {
	log.Println("CheckEvents")
	event := aod.eventSet.Find(dr.Labels)
	log.Println("found", event)

	if event != nil {
		action := pb.Action{
			DetectionResult: dr,
			AnnotatedImg:    nil,
		}

		go func() {
			_, err := aod.client.SendAction(aod.eCtx.ctx, &action)
			if err != nil {
				log.Printf("Error while sending action: %v", err)
			}
		}()
	}
}

// func (aod *Appaod) SendAction(Action) {

// }

// func (aod *Appaod) StreamActions(Action) {

// }

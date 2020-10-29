package main

import (
	"log"

	"github.com/bosdhill/iot_detect_2020/edge/settrie"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
)

// ActionOnDetect isused to serve the application's requests
type ActionOnDetect struct {
	client pb.ActionOnDetectClient
	eCtx   *EdgeContext
	events *pb.Events
	trie   *settrie.SetTrie
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
func (aod *ActionOnDetect) RegisterEvents(labels map[string]bool) (*settrie.SetTrie, error) {
	var err error
	aod.events, err = aod.client.RegisterEvents(aod.eCtx.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}
	aod.trie = settrie.New()
	// store events in prefix trie for later subset search and comparison
	for _, event := range aod.events.GetEvents() {
		aod.trie.Add(event.GetLabels(), event)
	}
	aod.trie.Output()
	return aod.trie, nil
}

// CheckEvents checks whether the detection result sastifies any event set by
// the application. If it does, it creates an action and sends it to the
// application
func (aod *ActionOnDetect) CheckEvents(dr *DetectionResult) {
	// TODO probably slow -- just store []string of labels separately
	labels := make([]string, 0, len(dr.Labels))

	for k := range dr.Labels {
		labels = append(labels, k)
	}
	// needs list of strings for labels
	_, err := aod.trie.Find(labels)

	if err != nil {
		panic(err)
	}

	// return dummy action
	action := pb.Action{
		Labels: map[string]*pb.BoundingBox{
			"": {
				TopLeftX:     0.0,
				TopLeftY:     0.0,
				BottomRightX: 0.0,
				BottomRightY: 0.0,
				Confidence:   0.0,
			},
		},
		Img:          nil,
		AnnotatedImg: nil,
	}

	// create action for each event
	aod.client.SendAction(aod.eCtx.ctx, &action)
}

// func (aod *Appaod) SendAction(Action) {

// }

// func (aod *Appaod) StreamActions(Action) {

// }

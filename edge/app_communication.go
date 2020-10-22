package main

import (
	"log"

	"github.com/bosdhill/iot_detect_2020/edge/settrie"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
)

// AppComm isused to serve the application's requests
type AppComm struct {
	client pb.ActionOnDetectClient
	eCtx   *EdgeContext
}

// NewActionOnDetect starts up a grpc server and
func NewActionOnDetect(eCtx *EdgeContext, addr string) (*AppComm, error) {
	log.Println("NewActionOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.DialContext(eCtx.ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewActionOnDetectClient(conn)
	return &AppComm{client: client, eCtx: eCtx}, nil
}

// SetEvents is used to register events
// TODO change SetEvents to RegisterEvents (we're following a service pattern,
// NOT a distributed object pattern)
func (comm *AppComm) SetEvents(labels map[string]bool) (*settrie.SetTrie, error) {
	events, err := comm.client.SetEvents(comm.eCtx.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}
	trie := settrie.New()
	// store events in prefix trie for later subset search and comparison
	for _, event := range events.GetEvents() {
		trie.Add(event.GetLabels(), event)
	}
	trie.Output()
	return trie, nil
}

package main

import (
	"log"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
	"github.com/bosdhill/iot_detect_2020/edge/settrie"
)

type appComm struct {
	client pb.ActionOnDetectClient
	eCtx   *EdgeContext
}

func NewAppCommunication(eCtx *EdgeContext, addr string) (*appComm, error) {
	log.Println("NewAppCommunication")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.DialContext(eCtx.ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewActionOnDetectClient(conn)
	return &appComm{client:client, eCtx: eCtx}, nil
}

func (comm *appComm) SetEvents(labels map[string]bool) (*settrie.SetTrie, error) {
	events, err := comm.client.SetEvents(comm.eCtx.ctx, &pb.Labels{ Labels: labels})
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
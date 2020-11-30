package eventondetect

import (
	"context"
	"github.com/bosdhill/iot_detect_2020/edge/realtimefilter"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
	"log"
	"math"
)

// EventOnDetect is used to serve the application's requests
type EventOnDetect struct {
	client       pb.EventOnDetectClient
	ctx          context.Context
	eventFilters *pb.EventFilters
	rtFilter     *realtimefilter.Set
}

// New creates an event on detect client
func New(ctx context.Context, addr string) (*EventOnDetect, error) {
	log.Println("NewEventOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewEventOnDetectClient(conn)
	return &EventOnDetect{client: client, ctx: ctx}, nil
}

// RegisterEventFilters is used to register the application's eventFilters
func (eod *EventOnDetect) RegisterEventFilters(labels map[string]bool) (*realtimefilter.Set, error) {
	log.Println("RegisterEventFilters")
	var err error
	eod.eventFilters, err = eod.client.RegisterEventFilters(eod.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}

	eod.rtFilter, err = realtimefilter.New(eod.eventFilters)
	if err != nil {
		return nil, err
	}

	return eod.rtFilter, nil
}

// FilterEvents checks whether the detection result satisfies any event conditions set by the application. If it does,
// it creates an event and sends it to the application.
func (eod *EventOnDetect) FilterEvents(dr *pb.DetectionResult) {
	log.Println("FilterEvents")
	events := eod.rtFilter.GetEvents(dr)
	log.Printf("Found %v events", len(events.GetEvents()))

	if events != nil {
		go func() {
			_, err := eod.client.SendEvents(eod.ctx, events)
			if err != nil {
				log.Printf("Error while sending action: %v", err)
			}
		}()
	}
}

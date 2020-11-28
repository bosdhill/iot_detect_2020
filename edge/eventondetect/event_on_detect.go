package eventondetect

import (
	"context"
	"github.com/bosdhill/iot_detect_2020/edge/events"
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
	eventSet     *events.EventFilterSet
}

// NewEventOnDetect starts up a grpc server and
func NewEventOnDetect(ctx context.Context, addr string) (*EventOnDetect, error) {
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
func (aod *EventOnDetect) RegisterEventFilters(labels map[string]bool) (*events.EventFilterSet, error) {
	log.Println("RegisterEventFilters")
	var err error
	aod.eventFilters, err = aod.client.RegisterEventFilters(aod.ctx, &pb.Labels{Labels: labels})
	if err != nil {
		return nil, err
	}
	aod.eventSet = events.New(aod.eventFilters)
	return aod.eventSet, nil
}

// FilterEvents checks whether the detection result satisfies any event conditions set by the application. If it does,
// it creates an event and sends it to the application.
func (aod *EventOnDetect) FilterEvents(dr *pb.DetectionResult) {
	log.Println("FilterEvents")
	event := aod.eventSet.Find(dr.LabelNumber)
	log.Println("found", event)

	if event != nil {
		event := pb.Event{
			DetectionResult: dr,
			AnnotatedImg:    nil,
		}

		go func() {
			_, err := aod.client.SendEvent(aod.ctx, &event)
			if err != nil {
				log.Printf("Error while sending action: %v", err)
			}
		}()
	}
}

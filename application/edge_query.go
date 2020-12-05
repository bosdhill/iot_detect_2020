package main

import (
	"context"
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

// EdgeComm contains an EventQuery client stub
type EventQuery struct {
	client pb.EventQueryClient
}

// NewEventQueryClient returns an NewEventQueryClient
func NewEventQueryClient(addr string) (*EventQuery, error) {
	log.Println("NewEventQueryClient")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}

	client := pb.NewEventQueryClient(conn)
	return &EventQuery{client}, nil
}

// Find makees a Find event filter query request to the EventQuery server on the Edge
func (eQuery *EventQuery) Find(eFilter *pb.EventFilter) ([]*pb.Event, error) {
	log.Printf("Find")
	ctx, _ := context.WithCancel(context.Background())
	resp, err := eQuery.client.Find(ctx, &pb.FindRequest{EventFilter: eFilter})
	if err != nil {
		return nil, err
	}
	return resp.GetEvents(), nil
}

func TimedTestQuery(group *sync.WaitGroup) {
	defer group.Done()
	log.Println("TimedTestQuery")
	timer := time.NewTimer(*testTimeout)
	eQuery, err := NewEventQueryClient(*appQueryServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	totalReq := 0
	totalEvents := 0
	totalReqLatency := time.Duration(0)
	avgReqLatency := time.Duration(0)
	avgEvents := 0.0

	queryLoop:
		for {
			select {
			case <- timer.C:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"AVG Events Recv (events/req)", "AVG Latency (msec/req)", "TOTAL Events Recv", "TOTAL Requests Sent",
					"AVG THROUGHPUT (req/sec)", "PERIOD Timeout (sec)", "EVENTFILTER SECONDS (sec)"})
				table.SetBorder(false)
				data := [][]string{
					{
						fmt.Sprintf("%.2f", avgEvents),
						avgReqLatency.String(),
						strconv.Itoa(totalEvents),
						strconv.Itoa(totalReq),
						fmt.Sprintf("%.2f", float64(totalReq) / float64(testTimeout.Seconds())),
						testTimeout.String(),
						fmt.Sprintf("%vs", *eventQuerySeconds),
					},
				}
				table.AppendBulk(data)
				table.Render()
				break queryLoop
			default:
				time.Sleep(time.Duration(*eventQueryPeriod))
				latency, events := eQuery.TestEventQuery()
				numEvents := len(events)

				totalReq++
				totalReqLatency += latency
				totalEvents += numEvents
				avgEvents = float64(totalEvents) / float64(totalReq)
				avgReqLatency = totalReqLatency / time.Duration(totalReq)

				if *logEvents {
					for _, e := range events {
						log.Println(e.GetName())
						log.Println(e.GetDetectionResult().GetDetectionTime())
						log.Println(e.GetDetectionResult().GetLabelNumber())
					}
				}
			}
		}
}

func (eQuery *EventQuery) TestEventQuery() (time.Duration, []*pb.Event) {
	log.Println("TestEventQuery")
	var query interface{} = bson.E{
		Key: "$or",
		Value: bson.A{[]bson.E{
			{
				Key: "labelnumber.person",
				Value: bson.D{{
					Key:   "$gte",
					Value: 1,
				}},
			},
			{
				Key: "labelnumber.bus",
				Value: bson.D{{
					Key:   "$gte",
					Value: 1,
				}},
			},
		},
		},
	}

	filter, err := bson.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}

	var flags uint32 = 0
	if *metadata {
		flags = uint32(pb.EventFilter_METADATA)
	}

	// Get Events from the last eventQuerySeconds that match filter
	eFilter := &pb.EventFilter{
		Seconds: *eventQuerySeconds,
		Name:    "TestQueryEvent",
		Filter:  filter,
		Flags:   flags,
	}

	t := time.Now()
	events, err := eQuery.Find(eFilter)
	if err != nil {
		log.Fatal(err)
	}
	e := time.Since(t)

	log.Println("current event query latency", e)
	log.Println("current num received events", len(events))

	return e, events
}

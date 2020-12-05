package main

import (
	"context"
	"fmt"
	"github.com/bosdhill/iot_detect_2020/edge/realtimefilter"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

// EventOnDetect is used to serve the application's requests
type EventOnDetect struct {
	client       pb.EventOnDetectClient
	ctx          context.Context
	eventFilters *pb.EventFilters
	rtFilter     *realtimefilter.Set
	uuid         string
	conn 		*grpc.ClientConn
}

// New creates an event on detect client
func NewEventOnDetect(ctx context.Context, addr string) (*EventOnDetect, error) {
	log.Println("NewEventOnDetect")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	client := pb.NewEventOnDetectClient(conn)
	return &EventOnDetect{client: client, ctx: ctx, conn: conn}, nil
}

// GetLabels fetches the object labels that the model on the edge can detect
func (eod *EventOnDetect) GetLabels() (*pb.Labels, error) {
	log.Println("GetLabels")
	var err error
	resp, err := eod.client.GetLabels(eod.ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.GetLabels(), nil
}

// RegisterEventFilters is used to register the application's eventFilters
func (eod *EventOnDetect) RegisterApp(labels *pb.Labels) error {
	log.Println("RegisterApp")
	eventFilters := &pb.EventFilters{}
	if labels.Labels["person"] && labels.Labels["bus"] {
		filters := map[string]bson.D{
			"FourPersonsOrFourBuses": {
				bson.E{
					Key: "$or",
					Value: bson.D{
						bson.E{
							Key: "person",
							Value: bson.E{
								Key:   "$gte",
								Value: 2,
							},
						},
						bson.E{
							Key: "bus",
							Value: bson.E{
								Key:   "$gte",
								Value: 1,
							},
						},
					},
				},
			},
			"AtLeastOnePersonAndBus": {
				bson.E{
					Key: "$and",
					Value: bson.D{
						bson.E{
							Key: "person",
							Value: bson.E{
								Key:   "$gte",
								Value: 1,
							},
						},
						bson.E{
							Key: "bus",
							Value: bson.E{
								Key:   "$lte",
								Value: 10,
							},
						},
					},
				},
			},
			"PersonAndBus": {
				bson.E{
					Key: "labels",
					Value: bson.D{
						bson.E{
							Key: "$all",
							Value: bson.A{
								"person",
								"bus",
							},
						},
					},
				},
			},
			"AtLeast3People": {
				bson.E{
					Key: "$or",
					Value: bson.D{
						bson.E{
							Key: "person",
							Value: bson.E{
								Key:   "$gte",
								Value: 3,
							},
						},
					},
				},
			},
		}

		for name, filter := range filters {
			mFilter, err := bson.Marshal(filter)
			if err != nil {
				log.Fatal(err)
			}

			eFilter := &pb.EventFilter{
				Name:   name,
				Filter: mFilter,
				Flags:  0,
			}

			eventFilters.EventFilters = append(eventFilters.EventFilters, eFilter)
		}
	}

	req := &pb.RegisterAppRequest{
		EventFilters: eventFilters,
	}

	resp, err := eod.client.RegisterApp(eod.ctx, req)
	if err != nil {
		return err
	}

	eod.uuid = resp.Uuid
	log.Println("received uuid:", eod.uuid)

	return nil
}

func (eod *EventOnDetect) GetEvents() pb.EventOnDetect_GetEventsClient {
	req := &pb.GetEventsRequest{
		Uuid: eod.uuid,
	}

	stream, err := eod.client.GetEvents(eod.ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	return stream
}

func TimedTestEventOnDetect(group *sync.WaitGroup) {
	defer group.Done()
	log.Println("TimedTestEventOnDetect")
	timer := time.NewTimer(*testTimeout)
	ctx, cancel := context.WithCancel(context.Background())
	eod, err := NewEventOnDetect(ctx, *eodServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	labels, err := eod.GetLabels()
	if err != nil {
		log.Fatal(err)
	}

	if err = eod.RegisterApp(labels); err != nil {
		log.Fatal(err)
	}

	stream := eod.GetEvents()

	recvEvents := func() (*time.Duration, int) {
		t := time.Now()
		resp, err := stream.Recv()

		if err == io.EOF {
			log.Println("EOF")
			if err = stream.CloseSend(); err != nil {
				log.Printf("error when closing stream: %v", err)
				return nil, 0
			}
		}

		if err != nil {
			log.Fatal(err)
		}
		latency := time.Since(t)
		numEvents := len(resp.GetEvents())

		for _, e := range resp.GetEvents() {
			log.Println(e.Name)
			log.Println(e.GetDetectionResult().GetDetectionTime())
			log.Println(e.GetDetectionResult().GetLabelNumber())
		}
		return &latency, numEvents
	}


	totalResp := 0
	totalEvents := 0
	totalRespLatency := time.Duration(0)
	avgRespLatency := time.Duration(0)
	avgEvents := 0.0

	recvLoop:
		for {
			select {
			case <- timer.C:
				fmt.Println("closing conn")
				if err := eod.conn.Close(); err != nil {
					log.Println("error closing connection:", err)
				}
				cancel()
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"AVG Events Recv", "AVG Latency (msec/resp)", "TOTAL Events Recv", "TOTAL Response Recv",
					"RATE Response (resp/sec)", "PERIOD Timeout (sec)"})
				table.SetBorder(false)
				data := [][]string{
					{
						fmt.Sprintf("%.2f", avgEvents),
						avgRespLatency.String(),
						strconv.Itoa(totalEvents),
						strconv.Itoa(totalResp),
						fmt.Sprintf("%.2f", float64(totalResp) / float64(testTimeout.Seconds())),
						testTimeout.String(),
					},
				}
				table.AppendBulk(data)
				table.Render()
				break recvLoop
			default:
				fmt.Println("default")
				latency, numEvents := recvEvents()
				totalResp++
				totalRespLatency += *latency
				totalEvents += numEvents
				avgEvents = float64(totalEvents) / float64(totalResp)
				avgRespLatency = totalRespLatency / time.Duration(totalResp)
				fmt.Println("end default")
			}
		}
}
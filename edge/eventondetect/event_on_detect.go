package eventondetect

import (
	"context"
	"flag"
	"fmt"
	"github.com/bosdhill/iot_detect_2020/edge/realtimefilter"
	"github.com/golang/protobuf/ptypes/empty"
	"log"
	"math"
	"net"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

// AppComm stores the events channel and real time filter to be used with a specific application
type AppComm struct {
	eventsChan chan []*pb.Event
	rtFilter   *realtimefilter.Set
}

// EventOnDetect is used to wrap the server for serving actions issued by the Edge
type EventOnDetect struct {
	server  pb.EventOnDetectServer
	lis     net.Listener
	appComm map[string]AppComm
	labels  *pb.Labels
}

// GetLabels returns the labels for objects that the object detection model can detect
func (eod *EventOnDetect) GetLabels(context.Context, *empty.Empty) (*pb.GetLabelsResponse, error) {
	log.Println("GetLabels")
	return &pb.GetLabelsResponse{
		Labels: eod.labels,
	}, nil
}

// RegisterApp implements the register app rpc which accepts event filters for an app and assigns it a uuid
func (eod *EventOnDetect) RegisterApp(ctx context.Context, req *pb.RegisterAppRequest) (*pb.RegisterAppResponse, error) {
	log.Println("RegisterApp")
	rtFilter, err := realtimefilter.New(req.EventFilters)
	if err != nil {
		return nil, err
	}

	// create new uuid for application
	uuidWithHyphen := uuid.New()
	appId := fmt.Sprint(uuidWithHyphen)

	// create new channel for this app
	eod.appComm[appId] = AppComm{eventsChan: make(chan []*pb.Event), rtFilter: rtFilter}

	return &pb.RegisterAppResponse{
		Uuid: appId,
	}, nil
}

// GetEvents streams any events detected in real time to the application that calls this rpc if their uuid exists
func (eod *EventOnDetect) GetEvents(req *pb.GetEventsRequest, stream pb.EventOnDetect_GetEventsServer) error {
	log.Println("GetEvents")
	appComm, ok := eod.appComm[req.Uuid]
	if !ok {
		return fmt.Errorf("app with uuid %v not found", req.Uuid)
	}

	for events := range appComm.eventsChan {
		resp := &pb.GetEventsResponse{
			Events: events,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	close(appComm.eventsChan)
	return nil
}

// New returns a new event on detect component
func New(addr string, labels map[string]bool) (*EventOnDetect, error) {
	log.Println("NewEventOnDetect")
	flag.Parse()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	EdgeComm := &EventOnDetect{lis: lis, appComm: make(map[string]AppComm), labels: &pb.Labels{Labels: labels}}
	return EdgeComm, nil
}

// FilterEventsWorker checks whether the detection result satisfies any event conditions set by the application. If it does,
// it creates an event and sends it to the application.
func (eod *EventOnDetect) FilterEventsWorker(drFilterCh chan pb.DetectionResult) {
	log.Println("FilterEventsWorker")
	for dr := range drFilterCh {
		if !dr.Empty {
			for _, appComm := range eod.appComm {
				events := appComm.rtFilter.GetEvents(&dr)
				if events != nil {
					appComm.eventsChan <- events
				}
			}
		}
	}
	close(drFilterCh)
}

// ServeEventOnDetect serves EventOnDetect rpc calls from the App
func (eod *EventOnDetect) ServeEventOnDetect() error {
	log.Println("ServeEventOnDetect")
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventOnDetectServer(grpcServer, eod)
	err := grpcServer.Serve(eod.lis)
	if err != nil {
		return err
	}
	return nil
}

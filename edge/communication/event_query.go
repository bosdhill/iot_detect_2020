package communication

import (
	"context"
	"github.com/bosdhill/iot_detect_2020/edge/datastore"
	"github.com/bosdhill/iot_detect_2020/edge/realtimefilter"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"log"
	"math"
	"net"
	"time"
)

// AppComm is a wrapper around the EdgeQuery server which is used to serve query requests
// from an application
type AppComm struct {
	server pb.EventQueryServer
	ds     *datastore.MongoDataStore
	lis    net.Listener
	eCtx   context.Context
}

// Find extracts the filter from EventFilter and queries the local mongodb instance a given number of seconds back.
// If there are detection results returned, it creates and returns the corresponding Events.
func (comm *AppComm) Find(ctx context.Context, req *pb.FindRequest) (*pb.FindResponse, error) {
	log.Println("Find")
	filter, err := realtimefilter.UnmarshallBsonEFilter(req.GetEventFilter().GetFilter())
	if err != nil {
		return nil, err
	}
	secondsNano := (time.Duration(req.GetEventFilter().GetSeconds()) * time.Second).Nanoseconds()

	log.Printf("Querying from last %v nano seconds\n", secondsNano)

	// Query results from time.now - secondsNano and that satisfy provided filter
	drSl, err := comm.ds.Find(bson.D{
		comm.ds.DurationFilter(secondsNano),
		filter,
	})

	if err != nil {
		return nil, err
	}

	if len(drSl) == 0 {
		// TODO read/redirect to Cloud
	}

	log.Printf("Found %v detection results for filter %v\n", len(drSl), filter)

	events := realtimefilter.NewEvents(req.GetEventFilter(), drSl)
	return &pb.FindResponse{
		Events: events,
	}, nil
}

func (comm *AppComm) EventStream(*pb.EventStreamRequest, pb.EventQuery_EventStreamServer) error {
	panic("implement me")
}

func NewEventQuery(eCtx context.Context, ds *datastore.MongoDataStore, addr string) (*AppComm, error) {
	log.Println("NewEventQuery")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &AppComm{
		ds:   ds,
		lis:  lis,
		eCtx: eCtx,
	}, nil
}

// ServeEventQuery creates a new EdgeQueryServer to serve the application's query requests
func (comm *AppComm) ServeEventQuery() error {
	log.Println("ServeEventQuery")
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventQueryServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

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
	server pb.EdgeQueryServer
	ds     *datastore.MongoDataStore
	lis    net.Listener
	eCtx   context.Context
}

// Find extracts the filter from EventFilter and queries the local mongodb instance a given number of seconds back.
// If there are detection results returned, it creates and returns the corresponding Events.
func (comm *AppComm) Find(ctx context.Context, eFilter *pb.EventFilter) (*pb.Events, error) {
	log.Println("Find")
	filter, err := realtimefilter.UnmarshallBsonEFilter(eFilter.GetFilter())
	if err != nil {
		return nil, err
	}
	secondsNano := (time.Duration(eFilter.GetSeconds()) * time.Second).Nanoseconds()

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

	events := realtimefilter.NewEvents(eFilter, drSl)
	return events, nil
}

func (comm *AppComm) EventStream(*pb.EventFilter, pb.EdgeQuery_EventStreamServer) error {
	panic("implement me")
}

func NewAppQuery(eCtx context.Context, ds *datastore.MongoDataStore, addr string) (*AppComm, error) {
	log.Println("NewAppQuery")
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

// ServeAppQuery creates a new EdgeQueryServer to serve the application's query requests
func (comm *AppComm) ServeAppQuery() error {
	log.Println("ServeAppQuery")
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEdgeQueryServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

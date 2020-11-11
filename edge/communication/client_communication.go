package communication

import (
	"context"
	"flag"
	"github.com/bosdhill/iot_detect_2020/edge/datastore"
	od "github.com/bosdhill/iot_detect_2020/edge/detection"
	"io"
	"log"
	"math"
	"net"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"google.golang.org/grpc"
)

// ClientComm is a wrapper around the uploader server which is used to serve
// client image frame upload requests. It also contains references to a
// SQLDataStore, ObjectDetect, and EdgeContext.
type ClientComm struct {
	server pb.UploaderServer
	ds     *datastore.MongoDataStore
	od     *od.ObjectDetect
	lis    net.Listener
	eCtx   context.Context
	cancel context.CancelFunc
}

// UploadImage serves image frame requests from the client. The image frame is
// then passed to the image detection pipeline, where the frame is then inserted
// into the db.
// TODO: find a way to annotate image frames after object detection.
func (comm *ClientComm) UploadImage(stream pb.Uploader_UploadImageServer) error {
	log.Println("UploadImage")
	count := 0
	resCh := make(chan pb.DetectionResult)
	imgCh := make(chan *pb.Image)
	go comm.od.CaffeWorker(imgCh, resCh)
	go comm.ds.InsertWorker(resCh)
	for {
		img, err := stream.Recv()
		count++
		if err == io.EOF {
			log.Println("EOF")
			close(imgCh)
			return stream.SendAndClose(&pb.ImageResponse{Success: true})
		}
		if err != nil {
			log.Println("err=", err)
			return err
		}
		imgCh <- img
	}
}

// NewClientCommunication returns a new client communication, which wraps around
// a gRPC server to serve the client's image frame upload requests.
func NewClientCommunication(eCtx context.Context, addr string, ds *datastore.MongoDataStore, od *od.ObjectDetect) (*ClientComm, error) {
	log.Println("NewClientCommunication")
	flag.Parse()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	cComm := &ClientComm{ds: ds, od: od, lis: lis, eCtx: eCtx}
	return cComm, nil
}

// ServeClient creates a new server and registers it to serve the client's
// image frame upload requests.
func (comm *ClientComm) ServeClient() error {
	log.Println("ServeClient")
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32), grpc.MaxSendMsgSize(math.MaxInt32))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUploaderServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

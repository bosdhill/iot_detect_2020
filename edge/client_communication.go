package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	//sdl "github.com/bosdhill/iot_detect_2020/sdl"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
	piAddr     = flag.String("pi_addr", "192.168.1.12", "Raspberry Pi IP address")
)

// ClientComm is a wrapper around the uploader server which is used to serve
// client image frame upload requests. It also contains references to a
// DataStore, ObjectDetect, and EdgeContext.
type ClientComm struct {
	server pb.UploaderServer
	ds     *DataStore
	od     *ObjectDetect
	lis    net.Listener
	eCtx   *EdgeContext
	cancel context.CancelFunc
}

func uploadReqToImg(req *pb.Image) *gocv.Mat {
	height := int(req.Rows)
	width := int(req.Cols)
	mType := gocv.MatType(req.Type)
	mat, err := gocv.NewMatFromBytes(height, width, mType, req.Image)
	if mType != gocv.MatTypeCV32F {
		mat.ConvertTo(&mat, gocv.MatTypeCV32F)
	}
	if err != nil {
		log.Fatal(err)
	}
	return &mat
}

// UploadImage serves image frame requests from the client. The image frame is
// then passed to the image detection pipeline, where the frame is then inserted
// into the db.
// TODO: find a way to annotate image frames after object detection.
func (comm *ClientComm) UploadImage(stream pb.Uploader_UploadImageServer) error {
	log.Println("UploadImage")
	count := 0
	resCh := make(chan DetectionResult)
	iCh := make(chan *gocv.Mat)
	go comm.od.caffeWorker(iCh, resCh)
	go comm.ds.InsertWorker(resCh)
	for {
		req, err := stream.Recv()
		count++
		//log.Println("received image from stream", count)
		if err == io.EOF {
			log.Println("EOF")
			close(iCh)
			return stream.SendAndClose(&pb.ImageResponse{Success: true})
		}
		if err != nil {
			log.Println("err=", err)
			return err
		}
		iCh <- uploadReqToImg(req)
	}
}

// NewClientCommunication returns a new client communication, which wraps around
// a gRPC server to serve the client's image frame upload requests.
func NewClientCommunication(eCtx *EdgeContext, addr string, ds *DataStore, od *ObjectDetect) (*ClientComm, error) {
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
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUploaderServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

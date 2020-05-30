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
	piAddr 	   = flag.String("pi_addr", "192.168.1.12", "Raspberry Pi IP address")
)

type clientComm struct {
	server pb.UploaderServer
	ds     *dataStore
	od     *objectDetect
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

// TODO find a way to annotate image frames after object detection
func (comm *clientComm) UploadImage(stream pb.Uploader_UploadImageServer) error {
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

func NewClientCommunication(eCtx *EdgeContext, addr string, ds *dataStore, od *objectDetect) (*clientComm, error) {
	log.Println("NewClientCommunication")
	flag.Parse()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	cComm := &clientComm{ds: ds, od: od, lis: lis, eCtx: eCtx}
	return cComm, nil
}

func (comm *clientComm) ServeClient() error {
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

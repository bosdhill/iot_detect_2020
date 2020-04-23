package main

import (
	"context"
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"log"
	"net"
	yo "github.com/bosdhill/iot_detect_2020/edge/tiny-yolo-v2-coco"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
)

type clientComm struct {
	server pb.UploaderServer
	yo *yo.TinyYolo
}

func (comm *clientComm) UploadImage(ctx context.Context, img *pb.Image) (*pb.ImageResponse, error) {
	log.Println("UploadImage")
	height := int(img.Rows)
	width := int(img.Cols)
	mType := gocv.MatType(img.Type)
	mat, err := gocv.NewMatFromBytes(height, width, mType, img.Image)

	log.Println("Show")
	window := gocv.NewWindow("edge")
	window.IMShow(mat)
	window.WaitKey(1)
	err = window.Close()
	if err != nil {
		log.Fatal(err)
	}

	//if err != nil {
	//	log.Fatal(err)
	//}
	//mImg, err := mat.ToImage()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//comm.yo.Detect(mImg)
	log.Println("received image")
	resp := pb.ImageResponse{Success: true}
	err = mat.Close()
	if err != nil {
		log.Fatal(err)
	}
	return &resp, nil
}

func newServer() *clientComm {
	yo := yo.NewTinyYolo()
	s := &clientComm{yo: yo}
	return s
}

func ServeClient() {
	log.Println("ServeClient")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUploaderServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

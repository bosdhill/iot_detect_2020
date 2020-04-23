package main

import (
	"context"
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"image"
	"log"
	"net"
	yo "github.com/bosdhill/iot_detect_2020/edge/tiny-yolo-v2-coco"
	//sdl "github.com/bosdhill/iot_detect_2020/sdl"
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
	//window *gocv.Window
}

func (comm *clientComm) UploadImage(ctx context.Context, img *pb.Image) (*pb.ImageResponse, error) {
	log.Println("UploadImage")
	height := int(img.Rows)
	width := int(img.Cols)
	mType := gocv.MatType(img.Type)
	mat, err := gocv.NewMatFromBytes(height, width, mType, img.Image)
	gocv.Resize(mat, &mat,image.Point{416, 416}, 0, 0, gocv.InterpolationNearestNeighbor)
	if err != nil {
		log.Fatal(err)
	}
	//gocv.IMWrite("recv.jpg", mat)
	//sdl.Show(comm.window, &mat)
	//if err != nil {
	//	log.Fatal(err)
	//}
	mImg, err := mat.ToImage()
	if err != nil {
		log.Fatal(err)
	}
	comm.yo.Detect(mImg)
	log.Println("received image")
	resp := pb.ImageResponse{Success: true}
	//err = mat.Close()
	if err != nil {
		log.Fatal(err)
	}
	return &resp, nil
}

func newServer() *clientComm {
	yo := yo.NewTinyYolo()
	//s := &clientComm{yo: yo, window: gocv.NewWindow("ed")}
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

package main

import (
	"context"
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	pb "github.com/bosdhill/iot_detect_2020/edge/interfaces"
	"log"
	"net"
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
}

func (comm *clientComm) UploadImage(ctx context.Context, img *pb.Image) (*pb.ImageResponse, error) {
	log.Println("UploadImage")
	_, err := gocv.NewMatFromBytes(int(img.Rows), int(img.Cols), gocv.MatType(int(img.Type)), img.Image)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("received image")
	resp := pb.ImageResponse{Success: true}
	return &resp, nil
}

func newServer() *clientComm {
	s := &clientComm{}
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

package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	piServerAddr       = flag.String("pi_server_addr", "192.168.1.121:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

type edgeComm struct {
	client pb.UploaderClient
}

func NewEdgeComm() (*edgeComm, error) {
	log.Println("NewEdgeComm")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	conn, err := grpc.DialContext(ctx, *serverAddr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	//defer conn.Close()
	client := pb.NewUploaderClient(conn)
	return &edgeComm{client}, nil
}

func imgToUploadReq(img gocv.Mat) *pb.Image {
	bImg := img.ToBytes()
	rows := int32(img.Rows())
	cols := int32(img.Cols())
	mType := img.Type()
	return &pb.Image{Image: bImg, Rows: rows, Cols: cols, Type: int32(mType)}
}

// TODO batch image frames when uploading
// FIXME message size limit capped at 4 MB -- fails with larger images
// FIXME shouldn't timeout with streaming rpc
func (e *edgeComm) UploadImage(c chan gocv.Mat) {
	log.Printf("UploadImage")
	// TODO timeout should be twice FPS * number of Frames per video
	//ctx, _ := context.WithTimeout(context.Background(), 0)
	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()
	stream, err := e.client.UploadImage(ctx)
	if err != nil {
		log.Fatalf("%v.UploadImage(_) = _, %v: ", e.client, err)
	}
	for img := range c {
		req := imgToUploadReq(img)
		if err := stream.Send(req); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, req, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("ImageResponse: %v", reply.String())
}

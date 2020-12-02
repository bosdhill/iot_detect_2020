package main

import (
	"context"
	"flag"
	"log"
	"math"
	"time"

	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

// EdgeComm contains an Edge stub
type EdgeComm struct {
	client pb.UploaderClient
}

// NewEdgeComm returns a new edge communication component
func NewEdgeComm(addr string) (*EdgeComm, error) {
	log.Println("NewEdgeComm")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}

	client := pb.NewUploaderClient(conn)
	return &EdgeComm{client}, nil
}

func imgToUploadReq(img gocv.Mat) *pb.UploadImageRequest {
	bImg := img.ToBytes()
	rows := int32(img.Rows())
	cols := int32(img.Cols())
	mType := img.Type()
	log.Println("dimensions:", rows, cols)
	return &pb.UploadImageRequest{Image: &pb.Image{Image: bImg, Rows: rows, Cols: cols, Type: int32(mType)}}
}

// UploadImage streams image frames to the Edge
func (e *EdgeComm) UploadImage(c chan gocv.Mat) {
	log.Printf("UploadImage")
	ctx, _ := context.WithCancel(context.Background())
	stream, err := e.client.UploadImage(ctx)
	if err != nil {
		log.Fatalf("UploadImage: could not upload image with error = %v", err)
	}
	for img := range c {
		req := imgToUploadReq(img)
		if err := stream.Send(req); err != nil {
			log.Fatalf("UploadImage: Send: could not send to stream with error = %v and dimensions = %v, %v",
				err, req.Image.GetCols(), req.Image.GetRows())
		}
		// prevent memory leak
		if err := img.Close(); err != nil {
			log.Fatalf("UploadImage: could not close img with error = %s", err)
		}
	}
	close(c)
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("UploadImage: could not CloseAndRecv() got error %v, want %v", err, nil)
	}
	log.Printf("UploadImageResponse: %v", reply.String())
}

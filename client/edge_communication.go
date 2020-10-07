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
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		log.Fatalf("Error while dialing. Err: %v", err)
	}
	//defer conn.Close()
	client := pb.NewUploaderClient(conn)
	return &EdgeComm{client}, nil
}

func imgToUploadReq(img gocv.Mat) *pb.Image {
	bImg := img.ToBytes()
	rows := int32(img.Rows())
	cols := int32(img.Cols())
	mType := img.Type()
	return &pb.Image{Image: bImg, Rows: rows, Cols: cols, Type: int32(mType)}
}

// UploadImage streams image frames to the Edge
// TODO batch image frames when uploading
// FIXME message size limit capped at 4 MB -- fails with larger images
func (e *EdgeComm) UploadImage(c chan gocv.Mat) {
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
		// prevent memory leak
		if err := img.Close(); err != nil {
			log.Fatalf("UploadImage: could not close img with error = %s", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("ImageResponse: %v", reply.String())
}

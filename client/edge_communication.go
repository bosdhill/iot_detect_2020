package main

import (
	"context"
	"flag"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	"log"
	"time"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
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

// TODO batch image frames when uploading
// FIXME message size limit capped at 4 MB -- fails with larger images
// FIXME shouldn't timeout with streaming rpc
func (e *edgeComm) UploadImage(c chan gocv.Mat) {
	log.Printf("UploadImage")
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	//defer cancel()
	for img := range c {
		bImg := img.ToBytes()
		rows := int32(img.Rows())
		cols := int32(img.Cols())
		mType := img.Type()
		req := 	&pb.Image{Image:bImg, Rows:rows, Cols:cols, Type:int32(mType)}
		resp, err := e.client.UploadImage(ctx, req)
		if err != nil {
			log.Fatalf("%v.UploadImage(_) = _, %v: ", e.client, err)
		}
		if resp.Success {
			log.Println("Success")
		}
	}
	log.Println("done uploading images")
}


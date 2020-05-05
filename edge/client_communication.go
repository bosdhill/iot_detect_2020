package main

import (
	"flag"
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"time"

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
	ds *dataStore
	lis net.Listener
}

func uploadReqToImg(req *pb.Image) gocv.Mat {
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
	return mat
}

// TODO find a way to annotate image frames after object detection
func (comm *clientComm) UploadImage(stream pb.Uploader_UploadImageServer) error {
	log.Println("UploadImage")
	count := 0
	sec := time.Duration(0)
	resCh := make(chan DetectionResult)
	iCh := make(chan *gocv.Mat)
	go caffeWorker(iCh, resCh)
	go comm.ds.InsertWorker(resCh)
	for {
		req, err := stream.Recv()
		count++
		log.Println("received image from stream", count)
		if err == io.EOF {
			log.Println("EOF")
			log.Println("AVG", sec / time.Duration(count))
			return stream.SendAndClose(&pb.ImageResponse{Success: true})
		}
		if err != nil {
			log.Println("err=", err)
			return err
		}
		img := uploadReqToImg(req)
		//t := time.Now()
		iCh <- &img
		//res := <- resCh
		//e := time.Since(t)
		//log.Println("Detected", res, e)
		//sec += e
		//log.Println(res)
		//for label, box := range res.detections {
		//	gocv.Rectangle(&img, image.Rect(box.topleft.X, box.topleft.Y, box.bottomright.X, box.bottomright.Y), color.RGBA{230, 25, 75, 0}, 1)
		//	gocv.PutText(&img, box.label, image.Point{box.topleft.X, box.topleft.Y - 5}, gocv.FontHersheySimplex, 0.5, color.RGBA{230, 25, 75, 0}, 1)
		//}
		//gocv.IMWrite("detect.jpg", img)
	}
}

func NewClientCommunication(ds *dataStore) (*clientComm, error) {
	log.Println("ServeClient")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		return nil, err
	}
	cComm := &clientComm{ds: ds, lis: lis}
	return cComm, nil
}

func (comm *clientComm) ServeClient() error {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterUploaderServer(grpcServer, comm)
	err := grpcServer.Serve(comm.lis)
	if err != nil {
		return err
	}
	return nil
}

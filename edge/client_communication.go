package main

import (
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"image"
	"io"
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

// FIXME resize() needs to return a matrix that is resized with aspect ratio and borders filled with black.
//  Currently CopyMakeBorder crashes with:
//  (-215:Assertion failed) top >= 0 && bottom >= 0 && left >= 0 && right >= 0 && _src.dims() <= 2 in function 'copyMakeBorder'
//func resize(mat *gocv.Mat, height int, width int) {
//	//border_v = 0
//	//border_h = 0
//	//if (IMG_COL/IMG_ROW) >= (img.shape[0]/img.shape[1]):
//	//border_v = int((((IMG_COL/IMG_ROW)*img.shape[1])-img.shape[0])/2)
//	//else:
//	//border_h = int((((IMG_ROW/IMG_COL)*img.shape[0])-img.shape[1])/2)
//	//img = cv2.copyMakeBorder(img, border_v, border_v, border_h, border_h, cv2.BORDER_CONSTANT, 0)
//	//img = cv2.resize(img, (IMG_ROW, IMG_COL))
//	border_v := 0
//	border_h := 0
//	if 1 >= width/height {
//		border_v = int((height - width)/2)
//	} else {
//		border_v = int((width - height)/2)
//	}
//	black := color.RGBA{R: 0,G: 0, B: 0, A: 0}
//	p := image.Point{X: 416, Y: 416}
//	gocv.CopyMakeBorder(*mat, mat, border_v, border_v, border_h, border_h, gocv.BorderConstant, black)
//	gocv.Resize(*mat, mat, p, 0, 0, gocv.InterpolationNearestNeighbor)
//}

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

func resizeToImage(img *gocv.Mat, height int, width int) image.Image {
	p := image.Point{X: 416, Y: 416}
	gocv.Resize(*img, img, p, 0, 0, gocv.InterpolationNearestNeighbor)
	//gocv.IMWrite("recv.jpg", *img)
	mImg, err := img.ToImage()
	if err != nil {
		log.Fatal(err)
	}
	return mImg
}

// TODO find a way to annotate image frames after object detection
func (comm *clientComm) UploadImage(stream pb.Uploader_UploadImageServer) (error) {
	log.Println("UploadImage")
	count := 0
	resCh := make(chan DetectionResult)
	iCh := make(chan *gocv.Mat)
	go caffeWorker(iCh, resCh)
	for {
		req, err := stream.Recv()
		count++
		log.Println("received image from stream", count)
		if err == io.EOF {
			log.Println("EOF")
			return stream.SendAndClose(&pb.ImageResponse{Success: true})
		}
		if err != nil {
			log.Println("err=", err)
			return err
		}
		img := uploadReqToImg(req)
		iCh <- &img
		res := <- resCh
		log.Println("Detected", res)
	}
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

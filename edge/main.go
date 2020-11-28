package main

import (
	"context"
	"flag"
	"github.com/bosdhill/iot_detect_2020/edge/communication"
	"github.com/bosdhill/iot_detect_2020/edge/datastore"
	"github.com/bosdhill/iot_detect_2020/edge/detection"
	"github.com/bosdhill/iot_detect_2020/edge/eventondetect"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
)

const (
	defaultBatchSize = 20
	credentials      = "credentials.env"
	defaultUploadTTL = 30
	defaultDeleteTTL = 60 * 60 * 24 // 24 hour TTL
)

var (
	cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
	serverAddr    = flag.String("server-addr", "192.168.1.121:10000", "The edge server address in the format of host:port")
	appServerAddr = flag.String("app-server-addr", "localhost:4200", "The app server address in the format of host:port")
	withCuda      = flag.Bool("with-cuda", false, "Determines whether cuda is enabled or not")
	matprofile    = flag.Bool("matprofile", false, "displays profile count of gocv.Mat")
	uploadTTL     = flag.Int64("uploadTTL", defaultUploadTTL, "TTL for local mongodb instance")
	deleteTTL     = flag.Int64("deleteTTL", defaultDeleteTTL, "TTL for local mongodb instance")
	batchSize     = flag.Int64("batchsize", defaultBatchSize, "Batchsize for cloud upload")
	withCloud     = flag.Bool("with-cloud", true, "enable cloud backups")
	proto         = "detection/model/tiny_yolo_deploy.prototxt"
	model         = "detection/model/tiny_yolo.caffemodel"
)

func getEnvVars() {
	err := godotenv.Load(credentials)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	getEnvVars()
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)

		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()

		defer pprof.StopCPUProfile()
	}

	ctx, _ := context.WithCancel(context.Background())

	mongoUri := os.Getenv("MONGO_LOCAL_URI")

	ds, err := datastore.New(ctx, mongoUri)
	if err != nil {
		panic(err)
	}

	eod, err := eventondetect.New(ctx, *appServerAddr)
	if err != nil {
		panic(err)
	}

	_, err = eod.RegisterEventFilters(detection.ClassNames)
	if err != nil {
		panic(err)
	}

	od, err := detection.New(ctx, eod, *withCuda, proto, model)
	if err != nil {
		panic(err)
	}

	clientComm, err := communication.NewClientCommunication(ctx, *serverAddr, ds, od)
	if err != nil {
		panic(err)
	}

	if *withCloud {
		mongoAtlasUri := os.Getenv("MONGO_ATLAS_URI")

		cloudComm, err := communication.NewCloudCommunication(ctx, ds, mongoAtlasUri)
		if err != nil {
			panic(err)
		}

		cloudComm.CloudUpload(*batchSize, *uploadTTL, *deleteTTL)
	}

	err = clientComm.ServeClient()
	if err != nil {
		panic(err)
	}
}

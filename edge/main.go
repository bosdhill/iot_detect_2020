package main

import (
	"context"
	"flag"
	aod "github.com/bosdhill/iot_detect_2020/edge/actionondetect"
	comm "github.com/bosdhill/iot_detect_2020/edge/communication"
	ds "github.com/bosdhill/iot_detect_2020/edge/datastore"
	od "github.com/bosdhill/iot_detect_2020/edge/detection"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
)

const (
	//TTL       = 30 * 60 * 60 // 30 minute TTL
	TTL         = 30 // 30 second TTL
	BatchSize   = 20
	Credentials = "credentials.env"
)

var (
	cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
	serverAddr    = flag.String("server-addr", "192.168.1.121:10000", "The edge server address in the format of host:port")
	appServerAddr = flag.String("app-server-addr", "localhost:4200", "The app server address in the format of host:port")
	withCuda      = flag.Bool("with-cuda", false, "Determines whether cuda is enabled or not")
	matprofile    = flag.Bool("matprofile", false, "displays profile count of gocv.Mat")
	ttl           = flag.Int64("ttl", TTL, "TTL for local mongodb instance")
	batchSize     = flag.Int64("batchsize", BatchSize, "Batchsize for cloud upload")
	withCloud 	  = flag.Bool("with-cloud", true, "enable cloud backups")
	proto         = "detection/model/tiny_yolo_deploy.prototxt"
	model         = "detection/model/tiny_yolo.caffemodel"
)

func getEnvVars() {
	err := godotenv.Load(Credentials)
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

	ds, err := ds.NewMongoDataStore(ctx, mongoUri)
	if err != nil {
		panic(err)
	}

	aod, err := aod.NewActionOnDetect(ctx, *appServerAddr)
	if err != nil {
		panic(err)
	}

	_, err = aod.RegisterEvents(od.ClassNames)
	if err != nil {
		panic(err)
	}

	od, err := od.NewObjectDetection(ctx, aod, *withCuda, proto, model)
	if err != nil {
		panic(err)
	}

	clientComm, err := comm.NewClientCommunication(ctx, *serverAddr, ds, od)
	if err != nil {
		panic(err)
	}

	if *withCloud {
		mongoAtlasUri := os.Getenv("MONGO_ATLAS_URI")

		cloudComm, err := comm.NewCloudCommunication(ctx, ds, mongoAtlasUri)
		if err != nil {
			panic(err)
		}

		go cloudComm.CloudInsert(*batchSize, *ttl)
	}

	err = clientComm.ServeClient()
	if err != nil {
		panic(err)
	}
}

# iot_detect_2020
go version

# Raspberry Pi set up (Linux)
1. Install the latest version of go and set your GOPATH and GOROOT in .bashrc. Clone this repo into $GOPATH/src/github.com/bosdhill/
2. add 
```
export GOBIN=/home/pi/go/bin
export GO111MODULE="on"
```
to your .bashrc
3. cd to`iot_detect_2020` and install [gocv](https://gocv.io/getting-started/linux/) 
4. cd to `iot_detect_2020` and run
```
go build ./...
```
5. OR go to `$GOPATH/pkg/mod/gocv.io/x/gocv@vX.YY.Z` if you tried to build first and run `make install`.

# Mac set up
```
go get -u -d gocv.io/x/gocv # may need to symlink /usr/local/lib/pkgconfig/opencv4.pc
go get google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces

go build ./...
```

```
which protoc
which protoc-gen-go

export GOPATH=$HOME/golang
export GOBIN=$GOPATH/bin
export GO111MODULE="on" 
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
```

to generate interface.pb.go in each module
```
cd <module>
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces
```

# build

build modules with
```
cd <module>
go build
```

# coco speedup
AVG detect time for mp4 = 94.796409ms
AVG python detect time for mp4 = 10899154700380465ms
~15% speedup on AVG

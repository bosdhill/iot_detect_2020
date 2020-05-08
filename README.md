# iot_detect_2020
go version

# set up
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

# iot_detect_2020


## Preqrequisites
- `edge/model/tiny_yolo.caffemodel`
- OpenCV 4.5.0
- `edge/credentials.env` with `MONGO_URI` and optionally `MONGO_ATLAS_URI` 

# Nvidia Jetson Nano setup
Install go

``` sh
sudo snap install --classic go
go version # should be go1.15.3
```

Add the environent variables `GOROOT` and `GOPATH` output by `go env` to your `~/.bashrc` file

``` sh
go env
echo GOPATH="$HOME/go" >> ~/.bashrc
echo GOROOT="/snap/go/6640" >> ~/.bashrc
```

get and install gocv along with opencv

``` sh
go get -u -d gocv.io/x/gocv
cd $GOPATH/src/gocv.io/x/gocv
```
verify CUDA

``` sh
cat /usr/local/cuda/version.txt # CUDA Version 10.2.89
```
install gocv with cuda support

``` sh
make install_cuda
```


# Raspberry Pi set up (Linux)

Install the latest version of go and set your GOPATH and GOROOT in .bashrc. Clone this repo into $GOPATH/src/github.com/bosdhill/

``` sh
cd $HOME
file='go1.14.2.linux-armv6l.tar.gz'
wget "https://dl.google.com/go/$file"
sudo tar -C /usr/local -xvf "$file"
cat >> ~/.bashrc << 'EOF'
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin
EOF
source ~/.bashrc
```

Add the following to your .bashrc

```
export GOBIN=/home/pi/go/bin
export GO111MODULE="on"
```

cd to`iot_detect_2020` and install [gocv](https://gocv.io/getting-started/linux/)
cd to `iot_detect_2020` and run with

``` sh
go build ./...
```

# Mac set up
``` sh
brew install opencv@4
brew install pkgconfig
brew install golang # follow set up https://medium.com/@jimkang/install-go-on-mac-with-homebrew-5fa421fc55f5
go get -u -d gocv.io/x/gocv # may need to symlink /usr/local/lib/pkgconfig/opencv4.pc
go get google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces

go build ./...
```


# Installing this repo
``` sh
mkdir -p $GOPATH/src/github.com/bosdhill/
cd $GOPATH/src/github.com/bosdhill/
git clone git@github.com:bosdhill/iot_detect_2020.git
```

# Protobuf
``` sh
cd $GOPATH/src/github.com/bosdhill
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces
```

# Building

build modules with
``` sh
cd <module>
go build
```

## CPU profiling
``` sh
go build -x -tags matprofile
./edge --cpuprofile=edge.prof
go tool pprof http://localhost:6060/debug/pprof/heap + top and web or
go tool pprof -png http://localhost:6060/debug/pprof/heap > out.png
```

## Gocv profiling
Since gocv.Mat is allocated with C code, the Go garbage collector does not handle its clean up. Every Mat needs to be closed. 
In order to detect these gocv.Mat memory leaks, build with:
```sh
go build -tags matprofile
```
and run with
```sh
./edge --matprofile true
```
More [here](https://gocv.io/blog/2018-11-28-opencv-4-support-and-custom-profiling/)

# iot_detect_2020


## Preqrequisites
- Need to install sqlite3 on raspberry pi and nvidia jetson
- Need to have `edge/model/tiny_yolo.caffemodel`

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

[Ref](https://github.com/hybridgroup/gocv#ubuntulinux)

# Raspberry Pi set up (Linux)

1. Install the latest version of go and set your GOPATH and GOROOT in .bashrc. Clone this repo into $GOPATH/src/github.com/bosdhill/

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

2. add

```
export GOBIN=/home/pi/go/bin
export GO111MODULE="on"
```

to your .bashrc
3. cd to`iot_detect_2020` and install [gocv](https://gocv.io/getting-started/linux/)
4. cd to `iot_detect_2020` and run

``` sh
go build ./...
```
5. OR go to `$GOPATH/pkg/mod/gocv.io/x/gocv@vX.YY.Z` if you tried to build first and run `make install`.

# Mac set up
``` sh
brew install opencv4
brew install pkgconfig
go get -u -d gocv.io/x/gocv # may need to symlink /usr/local/lib/pkgconfig/opencv4.pc
go get google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces

go build ./...
```

gocv issues:

``` sh
may need to symlink /usr/local/lib/pkgconfig/opencv4.pc

and add this to your ~/.bash_profile:
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig/opencv4.pc:$PATH"
```

``` sh
which protoc
which protoc-gen-go

export GOPATH=$HOME/golang
export GOBIN=$GOPATH/bin
export GO111MODULE="on"
export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
```

to generate interface.pb.go in each module
``` sh
cd <module>
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces
```

# build

build modules with
``` sh
cd <module>
go build
```


# coco speedup
AVG detect time for mp4 = 94.796409ms
AVG python detect time for mp4 = 10899154700380465ms
~15% speedup on AVG

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


# Notes on gocv and opencv

Ensure you uninstall prior your prior opencv versions, as this will cause symlinking issues
with the package config files of opencv4
``` sh
brew uninstall opencv@2
brew install opencv
brew doctor
```
and ensure your gocv module version matches up with your opencv version.

# Installing this repo
``` sh
mkdir -p $GOPATH/src/github.com/bosdhill/
cd $GOPATH/src/github.com/bosdhill/
git clone git@github.com:bosdhill/iot_detect_2020.git
```


# MongoDB

Set up mongodb data directory
```sh
$ cd $GOPATH/src/github.com/bosdhill/iot_detect_2020
$ sudo chown -R `id -un` edge/datastore/mongodata
```

 
Run mongodb in the background
```sh

$ mongod --dbpath=$GOPATH/src/github.com/bosdhill/iot_detect_2020/edge/datastore/mongodata
```

To connect to mongodb

```sh
mongo
```
# iot_detect_2020
go version

# Raspberry Pi set up (Linux)
1. Install the latest version of go and set your GOPATH and GOROOT in .bashrc. Clone this repo into $GOPATH/src/github.com/bosdhill/
```
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
```
go build ./...
```
5. OR go to `$GOPATH/pkg/mod/gocv.io/x/gocv@vX.YY.Z` if you tried to build first and run `make install`.

# Mac set up
```
brew install opencv4
brew install pkgconfig
go get -u -d gocv.io/x/gocv # may need to symlink /usr/local/lib/pkgconfig/opencv4.pc
go get google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I interfaces/ interfaces/interfaces.proto --go_out=plugins=grpc:interfaces

go build ./...
```

gocv issues:
```
may need to symlink /usr/local/lib/pkgconfig/opencv4.pc

and add this to your ~/.bash_profile:
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig/opencv4.pc:$PATH"
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


## Preqrequisites
- Need to install sqlite3 on raspberry pi
# coco speedup
AVG detect time for mp4 = 94.796409ms
AVG python detect time for mp4 = 10899154700380465ms
~15% speedup on AVG

## profiling
`go build -x -tags matprofile`
`./edge --cpuprofile=edge.prof`
`go tool pprof http://localhost:6060/debug/pprof/heap` + `top` and `web` or
`go tool pprof -png http://localhost:6060/debug/pprof/heap > out.png`


# Notes on gocv and opencv

Ensure you uninstall prior your prior opencv versions, as this will cause symlinking issues
with the package config files of opencv4:
```
brew uninstall opencv@2
brew install opencv
brew doctor
```
and ensure your gocv module version matches up with your opencv version.
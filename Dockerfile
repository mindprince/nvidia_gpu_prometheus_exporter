FROM golang:1.10
ENV IPATH=github.com/mindprince/nvidia_gpu_prometheus_exporter
RUN go get -u github.com/golang/dep/cmd/dep
WORKDIR $GOPATH/src/$IPATH

ADD Gopkg.* ./
RUN dep ensure --vendor-only

ADD . .
RUN go install ./...

ENTRYPOINT [ "/go/bin/nvidia_gpu_prometheus_exporter" ]

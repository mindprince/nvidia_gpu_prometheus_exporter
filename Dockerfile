FROM golang as build
RUN go get github.com/mindprince/nvidia_gpu_prometheus_exporter

FROM nvidia/cuda:9.0-base

RUN mkdir /exporter
WORKDIR /exporter

COPY --from=build /go/bin/nvidia_gpu_prometheus_exporter .

CMD ./nvidia_gpu_prometheus_exporter
EXPOSE 9445

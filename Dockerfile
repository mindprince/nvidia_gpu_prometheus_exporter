FROM golang as build
RUN go get github.com/mindprince/nvidia_gpu_prometheus_exporter

FROM ubuntu:18.04
COPY --from=build /go/bin/nvidia_gpu_prometheus_exporter /
CMD /nvidia_gpu_prometheus_exporter
ENV NVIDIA_VISIBLE_DEVICES=all
EXPOSE 9445

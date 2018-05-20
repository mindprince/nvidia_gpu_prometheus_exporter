PKG=github.com/mindprince/nvidia_gpu_prometheus_exporter

.PHONY: build
build:
	docker run -v $(shell pwd):/go/src/$(PKG) --workdir=/go/src/$(PKG) golang:1.10 go build

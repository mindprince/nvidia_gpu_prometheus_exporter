NVIDIA GPU Prometheus Exporter
------------------------------

This is a [Prometheus Exporter](https://prometheus.io/docs/instrumenting/exporters/) for
exporting NVIDIA GPU metrics. It uses the [Go bindings](https://github.com/mindprince/gonvml)
for [NVIDIA Management Library](https://developer.nvidia.com/nvidia-management-library-nvml)
(NVML) which is a C-based API that can be used for monitoring NVIDIA GPU devices.
Unlike some other similar exporters, it does not call the
[`nvidia-smi`](https://developer.nvidia.com/nvidia-system-management-interface) binary.

## Building

The repository includes `nvml.h`, so there are no special requirements from the
build environment. `go get` should be able to build the exporter binary.

```
go get github.com/mindprince/nvidia_gpu_prometheus_exporter
```

## Running

The exporter requires the following:
- access to NVML library (`libnvidia-ml.so.1`).
- access to the GPU devices.

To make sure that the exporter can access the NVML libraries, either add them
to the search path for shared libraries. Or set `LD_LIBRARY_PATH` to point to
their location.

By default the metrics are exposed on port `9445`. This can be updated using
the `-web.listen-address` flag.

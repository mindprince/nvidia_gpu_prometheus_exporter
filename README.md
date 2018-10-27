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

## Running inside a container

There's a docker image available on Docker Hub at
[mindprince/nvidia_gpu_prometheus_exporter](https://hub.docker.com/r/mindprince/nvidia_gpu_prometheus_exporter/)

If you are running the exporter inside a container, you will need to do the
following to give the container access to NVML library:
```
-e LD_LIBRARY_PATH=<path-where-nvml-is-present>
--volume <above-path>:<above-path>
```

And you will need to do one of the following to give it access to the GPU
devices:
- Run with `--privileged`
- If you are on docker v17.04.0-ce or above, run with `--device-cgroup-rule 'c 195:* mrw'`
- Run with `--device /dev/nvidiactl:/dev/nvidiactl /dev/nvidia0:/dev/nvidia0 /dev/nvidia1:/dev/nvidia1 <and-so-on-for-all-nvidia-devices>`

If you don't want to do the above, you can run it using nvidia-docker.

## Running using [nvidia-docker](https://github.com/NVIDIA/nvidia-docker)

```
nvidia-docker run -p 9445:9445 -ti mindprince/nvidia_gpu_prometheus_exporter:0.1
```

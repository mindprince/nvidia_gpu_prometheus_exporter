package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/mindprince/gonvml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "nvidia_gpu"
)

var (
	addr = flag.String("web.listen-address", ":9445", "Address to listen on for web interface and telemetry.")
)

type Collector struct {
	sync.Mutex
	numDevices  prometheus.Gauge
	usedMemory  *prometheus.GaugeVec
	totalMemory *prometheus.GaugeVec
}

func NewCollector() *Collector {
	return &Collector{
		numDevices: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "num_devices",
				Help:      "Number of NVIDIA GPU devices",
			},
		),
		usedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_used_bytes",
				Help:      "Memory used by the GPU device in bytes",
			},
			[]string{"uuid", "name"},
		),
		totalMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_total_bytes",
				Help:      "Total memory of the GPU device in bytes",
			},
			[]string{"uuid", "name"},
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numDevices.Desc()
	c.usedMemory.Describe(ch)
	c.totalMemory.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Only one Collect call in progress at a time.
	c.Lock()
	defer c.Unlock()

	numDevices, err := gonvml.DeviceCount()
	if err != nil {
		log.Printf("DeviceCount() error: %v", err)
		return
	} else {
		c.numDevices.Set(float64(numDevices))
		ch <- c.numDevices
	}

	for i := 0; i < int(numDevices); i++ {
		dev, err := gonvml.DeviceHandleByIndex(uint(i))
		if err != nil {
			log.Printf("DeviceHandleByIndex(%d) error: %v", i, err)
			continue
		}

		uuid, err := dev.UUID()
		if err != nil {
			log.Printf("UUID() error: %v", err)
			continue
		}

		name, err := dev.Name()
		if err != nil {
			log.Printf("Name() error: %v", err)
			continue
		}

		totalMemory, usedMemory, err := dev.MemoryInfo()
		if err != nil {
			log.Printf("MemoryInfo() error: %v", err)
		} else {
			c.usedMemory.WithLabelValues(uuid, name).Set(float64(usedMemory))
			c.totalMemory.WithLabelValues(uuid, name).Set(float64(totalMemory))
		}
	}
	c.usedMemory.Collect(ch)
	c.totalMemory.Collect(ch)
}

func main() {
	err := gonvml.Initialize()
	if err != nil {
		log.Fatalf("Couldn't initialize gonvml: %v", err)
	}
	defer gonvml.Shutdown()

	collector := NewCollector()

	prometheus.MustRegister(collector)

	http.ListenAndServe(*addr, promhttp.Handler())
}

package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
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

	labels = []string{"minor_number", "uuid", "name"}
	
	isFanSpeedEnabled = true
)

type Collector struct {
	sync.Mutex
	numDevices  prometheus.Gauge
	usedMemory  *prometheus.GaugeVec
	totalMemory *prometheus.GaugeVec
	dutyCycle   *prometheus.GaugeVec
	powerUsage  *prometheus.GaugeVec
	temperature *prometheus.GaugeVec
	fanSpeed    *prometheus.GaugeVec
}

func NewCollector() *Collector {
	return &Collector{
		numDevices: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "num_devices",
				Help:      "Number of GPU devices",
			},
		),
		usedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_used_bytes",
				Help:      "Memory used by the GPU device in bytes",
			},
			labels,
		),
		totalMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_total_bytes",
				Help:      "Total memory of the GPU device in bytes",
			},
			labels,
		),
		dutyCycle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "duty_cycle",
				Help:      "Percent of time over the past sample period during which one or more kernels were executing on the GPU device",
			},
			labels,
		),
		powerUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "power_usage_milliwatts",
				Help:      "Power usage of the GPU device in milliwatts",
			},
			labels,
		),
		temperature: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "temperature_celsius",
				Help:      "Temperature of the GPU device in celsius",
			},
			labels,
		),
		fanSpeed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "fanspeed_percent",
				Help:      "Fanspeed of the GPU device as a percent of its maximum",
			},
			labels,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numDevices.Desc()
	c.usedMemory.Describe(ch)
	c.totalMemory.Describe(ch)
	c.dutyCycle.Describe(ch)
	c.powerUsage.Describe(ch)
	c.temperature.Describe(ch)
	c.fanSpeed.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Only one Collect call in progress at a time.
	c.Lock()
	defer c.Unlock()

	c.usedMemory.Reset()
	c.totalMemory.Reset()
	c.dutyCycle.Reset()
	c.powerUsage.Reset()
	c.temperature.Reset()
	c.fanSpeed.Reset()

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

		minorNumber, err := dev.MinorNumber()
		if err != nil {
			log.Printf("MinorNumber() error: %v", err)
			continue
		}
		minor := strconv.Itoa(int(minorNumber))

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
			c.usedMemory.WithLabelValues(minor, uuid, name).Set(float64(usedMemory))
			c.totalMemory.WithLabelValues(minor, uuid, name).Set(float64(totalMemory))
		}

		dutyCycle, _, err := dev.UtilizationRates()
		if err != nil {
			log.Printf("UtilizationRates() error: %v", err)
		} else {
			c.dutyCycle.WithLabelValues(minor, uuid, name).Set(float64(dutyCycle))
		}

		powerUsage, err := dev.PowerUsage()
		if err != nil {
			log.Printf("PowerUsage() error: %v", err)
		} else {
			c.powerUsage.WithLabelValues(minor, uuid, name).Set(float64(powerUsage))
		}

		temperature, err := dev.Temperature()
		if err != nil {
			log.Printf("Temperature() error: %v", err)
		} else {
			c.temperature.WithLabelValues(minor, uuid, name).Set(float64(temperature))
		}

		if isFanSpeedEnabled {
			fanSpeed, err := dev.FanSpeed()
			if err != nil {
				log.Printf("FanSpeed() error: %v", err)
				isFanSpeedEnabled = false
			} else {
				c.fanSpeed.WithLabelValues(minor, uuid, name).Set(float64(fanSpeed))
			}
		}
	}
	c.usedMemory.Collect(ch)
	c.totalMemory.Collect(ch)
	c.dutyCycle.Collect(ch)
	c.powerUsage.Collect(ch)
	c.temperature.Collect(ch)
	c.fanSpeed.Collect(ch)
}

func main() {
	flag.Parse()

	if err := gonvml.Initialize(); err != nil {
		log.Fatalf("Couldn't initialize gonvml: %v. Make sure NVML is in the shared library search path.", err)
	}
	defer gonvml.Shutdown()

	if driverVersion, err := gonvml.SystemDriverVersion(); err != nil {
		log.Printf("SystemDriverVersion() error: %v", err)
	} else {
		log.Printf("SystemDriverVersion(): %v", driverVersion)
	}

	prometheus.MustRegister(NewCollector())

	// Serve on all paths under addr
	log.Fatalf("ListenAndServe error: %v", http.ListenAndServe(*addr, promhttp.Handler()))
}

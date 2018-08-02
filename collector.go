package main

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "alicloud_redis"

type redisCollector struct {
	mutex      sync.Mutex
	instanceID string

	memoryUsage      *prometheus.Desc
	connectionUsage  *prometheus.Desc
	intranetInRatio  *prometheus.Desc
	intranetOutRatio *prometheus.Desc
	intranetIn       *prometheus.Desc
	intranetOut      *prometheus.Desc
	failedCount      *prometheus.Desc
	cpuUsage         *prometheus.Desc
	usedMemory       *prometheus.Desc
}

func NewRedisCollector(instanceID string) (prometheus.Collector, error) {
	return &redisCollector{
		instanceID: instanceID,
		memoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "usage"),
			"Used capacity percentage",
			[]string{"instanceId"},
			nil,
		),
		connectionUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "connection", "usage"),
			"Percentage of used connections",
			[]string{"instanceId"},
			nil,
		),
		intranetInRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "intranet", "in_ratio"),
			"Write network bandwidth usage",
			[]string{"instanceId"},
			nil,
		),
		intranetOutRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "intranet", "out_ratio"),
			"Read network bandwidth usage",
			[]string{"instanceId"},
			nil,
		),
		intranetIn: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "intranet", "in"),
			"Write network rate",
			[]string{"instanceId"},
			nil,
		),
		intranetOut: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "intranet", "out"),
			"Read network rate",
			[]string{"instanceId"},
			nil,
		),
		failedCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "connection", "failed_count"),
			"KVSTORE failures",
			[]string{"instanceId"},
			nil,
		),
		cpuUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "cpu", "usage"),
			"CPU usage",
			[]string{"instanceId"},
			nil,
		),
		usedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "used"),
			"Memory usage",
			[]string{"instanceId"},
			nil,
		),
	}, nil
}

func (c *redisCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.memoryUsage
	ch <- c.connectionUsage
	ch <- c.intranetInRatio
	ch <- c.intranetOutRatio
	ch <- c.intranetIn
	ch <- c.intranetOut
	ch <- c.failedCount
	ch <- c.cpuUsage
	ch <- c.usedMemory
}

func (c *redisCollector) collect(ch chan<- prometheus.Metric) error {
	checkInstanceList(c.instanceID)
	var i = readCache(c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.memoryUsage, prometheus.GaugeValue, i.MemoryUsage, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.connectionUsage, prometheus.GaugeValue, i.ConnectionUsage, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.intranetInRatio, prometheus.GaugeValue, i.IntranetInRatio, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.intranetOutRatio, prometheus.GaugeValue, i.IntranetOutRatio, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.intranetIn, prometheus.GaugeValue, i.IntranetIn, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.intranetOut, prometheus.GaugeValue, i.IntranetOut, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.failedCount, prometheus.GaugeValue, i.FailedCount, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.cpuUsage, prometheus.GaugeValue, i.CpuUsage, c.instanceID)

	ch <- prometheus.MustNewConstMetric(c.usedMemory, prometheus.GaugeValue, fixDecimal(i.UsedMemory/1024/1024), c.instanceID)

	return nil
}

func (c *redisCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()
	if err := c.collect(ch); err != nil {
		log.Errorf("Error: %s", err)
	}
	return
}

func fixDecimal(x float64) float64 {
	str := decimal.NewFromFloat(x).StringFixed(2)
	y, _ := decimal.NewFromString(str)
	z, _ := y.Float64()
	return z
}

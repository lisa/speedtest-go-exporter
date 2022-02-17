package exporter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	myspeedtest "github.com/lisa/speedtest-go-exporter/pkg/speedtest"
	"github.com/lisa/speedtest-go-exporter/pkg/version"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	klog "k8s.io/klog/v2"
)

const (
	collectorNamespace = "speedtest_"
)

type speedtestCollector struct {
	timeout time.Duration // how long to allow the speedtest to run. any longer and metrics from it are discarded

	uploadMbps     *prometheus.Desc
	downloadMbps   *prometheus.Desc
	pingLatencyMs  *prometheus.Desc
	testDurationMs *prometheus.Desc
	testErrors     *prometheus.Desc
	testTimeouts   *prometheus.Desc

	timeouts int64
	errors   int64
}

func newSpeedtestCollector(timeout time.Duration) *speedtestCollector {
	return &speedtestCollector{
		timeout:  timeout,
		timeouts: 0,
		errors:   0,

		uploadMbps: prometheus.NewDesc(
			collectorNamespace+"upload_mbps",
			"Speedtest upload measurement in megabits per second",
			[]string{},
			nil,
		),
		downloadMbps: prometheus.NewDesc(
			collectorNamespace+"download_mbps",
			"Speedtest download measurement in megabits per second",
			[]string{},
			nil,
		),
		pingLatencyMs: prometheus.NewDesc(
			collectorNamespace+"ping_latency_ms",
			"Speedtest ping measurement in ms",
			[]string{},
			nil,
		),
		testDurationMs: prometheus.NewDesc(
			collectorNamespace+"duration_ms",
			"Speedtest duration in ms",
			[]string{},
			nil,
		),
		testErrors: prometheus.NewDesc(
			collectorNamespace+"test_errors",
			"How many times invalid test results have been encountered",
			[]string{},
			nil,
		),
		testTimeouts: prometheus.NewDesc(
			collectorNamespace+"test_timeouts",
			"how many times the speedtest exceeded the timeout",
			[]string{},
			nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *speedtestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.downloadMbps
	ch <- c.uploadMbps
	ch <- c.pingLatencyMs
	ch <- c.testDurationMs
	ch <- c.testErrors
	ch <- c.testTimeouts
}

// Collect implements prometheus.Collector
func (c *speedtestCollector) Collect(ch chan<- prometheus.Metric) {
	// do the work of collecting and measuring
	klog.Infof("Processing...")
	timeoutChan := make(chan error, 1)
	testRun := myspeedtest.Speedtest{}
	var err error
	// Timebox the test so it doesn't run for too long.
	// This should be less than the scrape interval.
	go func() {
		err = testRun.RunTest()
		timeoutChan <- err
	}()
	select {
	case res := <-timeoutChan:
		if res != nil {
			c.errors++
			ch <- prometheus.MustNewConstMetric(
				c.testErrors,
				prometheus.CounterValue,
				float64(c.errors),
			)
			klog.Warningf("Skipping adding results due to an observed error (test time %s): %s", testRun.TestDuration, err)
			return
		}
	case <-time.After(c.timeout):
		c.timeouts++
		ch <- prometheus.MustNewConstMetric(
			c.testTimeouts,
			prometheus.CounterValue,
			float64(c.timeouts),
		)
		klog.Warningf("Timed out (%s). Recording nothing.", c.timeout)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.uploadMbps,
		prometheus.GaugeValue,
		testRun.Result.UploadMbps,
	)
	ch <- prometheus.MustNewConstMetric(
		c.downloadMbps,
		prometheus.GaugeValue,
		testRun.Result.DownloadMbps,
	)
	ch <- prometheus.MustNewConstMetric(
		c.pingLatencyMs,
		prometheus.GaugeValue,
		float64(testRun.Result.PingLatencyMs/time.Millisecond), // convert to ms for metric
	)
	ch <- prometheus.MustNewConstMetric(
		c.testDurationMs,
		prometheus.GaugeValue,
		float64(testRun.TestDuration/time.Millisecond), // convert to ms for metric
	)
	ch <- prometheus.MustNewConstMetric(
		c.testErrors,
		prometheus.CounterValue,
		float64(c.errors),
	)
	ch <- prometheus.MustNewConstMetric(
		c.testTimeouts,
		prometheus.CounterValue,
		float64(c.timeouts),
	)
	klog.Infof("teststarts='%s'; pingLatency='%s'; UploadMbps='%f'; DownloadMbps='%f'; testDuration='%s'",
		testRun.TestStartTime.Format(time.RFC3339),
		testRun.Result.PingLatencyMs,
		testRun.Result.UploadMbps,
		testRun.Result.DownloadMbps,
		testRun.TestDuration)
}

type MetricsServer struct {
	MetricsTimeout time.Duration
	ListenPort     int
	MetricsPath    string
}

func (m *MetricsServer) Listen() {
	prometheus.MustRegister(newSpeedtestCollector(m.MetricsTimeout))
	klog.Infof("Version %s, githash %s", version.Version, version.GitHash[0:8])
	klog.Infof("Starting listener on :%d %s with test timeout %s...", m.ListenPort, m.MetricsPath, m.MetricsTimeout)
	var srv http.Server

	http.Handle(m.MetricsPath, promhttp.Handler())

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			klog.Infof("HTTP Server Shut down: %v", err)
		}
		close(idleConnsClosed)
	}()

	srv.Addr = fmt.Sprintf(":%d", m.ListenPort)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
	<head><title>Speedtest Exporter</title></head>
		<body>
			<p><a href='/metrics'>Metrics</a></p>
		</body>
</html>
`))
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		klog.Infof("HTTP Server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

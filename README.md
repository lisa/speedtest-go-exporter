# speedtest-go-exporter

Based on https://github.com/showwin/speedtest-go

# Exported Metrics

* `speedtest_upload_mbps` GAUGE Speedtest upload measurement in megabits per second
* `speedtest_download_mbps` GAUGE Speedtest download measurement in megabits per second
* `speedtest_ping_latency_ms` GAUGE Speedtest ping measurement in ms
* `speedtest_duration_ms` GAUGE Speedtest duration in ms
* `speedtest_test_errors` COUNTER How many times invalid test results have been encountered (this will correlate to gaps in the data)
* `speedtest_test_timeouts` COUNTER how many times the speedtest exceeded the timeout (default 2 minutes; see -t option)

# Usage

```
./speedtest-exporter exporter --help
Runs the exporter

Usage:
  speedtest-exporter exporter [flags]

Flags:
  -h, --help             help for exporter
  -p, --path string      Metrics path to listen on (default "/metrics")
  -P, --port int         Port to listen for metrics on (default 8080)
  -t, --timeoutSec int   Timeout in seconds for each speedtest. Tests exceeding this timeout will be terminated with no results recorded. (default 120)
```

## Execute on the Commandline


## Kubernetes

Note: This example uses the `latest` tag of `quay.io/lisa/speedtest-exporter`. Some users may wish to pin to a specific version, which can be done by browsing the [available tags](https://quay.io/repository/lisa/speedtest-exporter?tab=tags) at quay.io. Additionally, the following Kubernetes YAML files can be found in [/manifests](/manifests), and used as the basis of more complex uses of the exporter.

Create a `Deployment`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: speedtest-exporter
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  selector:
    matchLabels:
      app: speedtest-exporter
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: speedtest-exporter
    spec:
      containers:
      - command:
        - /speedtest-exporter
        - exporter
        image: quay.io/lisa/speedtest-exporter:latest
        imagePullPolicy: IfNotPresent
        name: speedtest-exporter
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
```

Create a `Service`:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: speedtest-exporter
  name: speedtest-exporter
spec:
  ports:
  - name: metrics
    port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: speedtest-exporter
  sessionAffinity: None
  type: ClusterIP
```

## Prometheus 

If using [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator), create a `ServiceMonitor`, perhaps tailoring for your own prometheus-operator installation which may restrict namespaces, or require other work to consume this `ServiceMonitor`. Configure Prometheus to use a scrape timeout that matches the exporter's timeout (default 120 seconds, specified at runtime with `speedtest-exporter exporter -t <timeout>`).

Graphing with Grafana or similar tools will require setting the "query options" minimum interval to match the Prometheus scrape interval.


```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: speedtest-exporter
  name: speedtest-exporter
spec:
  endpoints:
  - interval: 10m
    port: metrics
    scrapeTimeout: 2m
  jobLabel: speedtest-exporter
  selector:
    matchLabels:
      app: speedtest-exporter
```

# Development

Clone the repository and compile:

```shell
git clone https://github.com/lisa/speedtest-go-exporter.git
cd speedtest-go-exporter
make build
```

Run the resulting `speedtest-exporter` binary, refer to above Usage.

## Container Images

The [Makefile](/Makefile) will build for `arm64` and `amd64` Linux targets and tag for `quay.io/lisa/speedtest-exporter` using today's date as the version/tag with revision 1 (ex 2022.02.1).

To build separate container images for `amd64`, `arm64` and `s390x` tagged `docker.io/thedoh/speedtest-exporter:1.0.5`:

```shell
make ARCHES="amd64 arm64 s390x" REGISTRY="docker.io" IMG="thedoh/speedtest-exporter" REVISION=5 VERSION=1.0 docker-build
```

These individual images can be "bundled" with an [image manifest](https://github.com/opencontainers/image-spec/blob/main/manifest.md):

```shell
make ARCHES="amd64 arm64 s390x" REGISTRY="docker.io" IMG="thedoh/speedtest-exporter" REVISION=5 VERSION=1.0 docker-build docker-multiarch
```

Pushed to the registry:

```shell
make ARCHES="amd64 arm64 s390x" REGISTRY="docker.io" IMG="thedoh/speedtest-exporter" REVISION=5 VERSION=1.0 docker-build docker-multiarch docker-push
```

**Note** that there is not currently support for individually pushing the arch-specific images (though they are pushed in the `docker-multiarch` make target).

## Roadmap

Refer to the [ROADMAP.md](/ROADMAP.md) for details.

# Grafana Dashboard

A Grafana dashboard is included in this repository at [/grafana-dashboard.json](/grafana-dashboard.json).

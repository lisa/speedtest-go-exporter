package speedtest

import (
	"fmt"
	"strings"
	"time"

	gospeedtest "github.com/showwin/speedtest-go/speedtest"
	klog "k8s.io/klog/v2"
)

// SpeedtestResult encapsulates the results strictly required for the speedtest.
// Duration is omitted here since it's not strictly needed (see parent object)
type SpeedtestResult struct {
	UploadMbps    float64       // Upload bandwidth in Mbps
	DownloadMbps  float64       // Download bandwidth in Mbps
	PingLatencyMs time.Duration // Ping to test server in ms
}

type Speedtest struct {
	TestStartTime time.Time        // When did the test start?
	TestDuration  time.Duration    // How long did the test take?
	Result        *SpeedtestResult // Results from interacting with the remote server
}

func elapsed(start time.Time) func() time.Duration {
	return func() time.Duration {
		return time.Since(start)
	}
}

// RunTest will run the test and return an error should one occur.
func (g *Speedtest) RunTest() error {
	user, err := gospeedtest.FetchUserInfo()
	if err != nil {
		return err
	}
	serverList, err := gospeedtest.FetchServers(user)
	if err != nil {
		return err
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		return err
	}
	// upstream String() includes \n, and it'd be better for output here if there weren't any \n
	klog.V(2).Infof("Using server %s (hostname %s)", strings.ReplaceAll(targets[0].String(), "\n", ""), targets[0].Host)
	g.TestStartTime = time.Now()
	defer func() {
		g.TestDuration = elapsed(g.TestStartTime)()
	}()
	// Only use the first resulting server since there's no support in the
	// exporter for more
	targets[0].PingTest()
	targets[0].DownloadTest(false)
	targets[0].UploadTest(false)
	if !targets[0].CheckResultValid() || targets[0].ULSpeed == 0.0 || targets[0].DLSpeed == 0.0 {
		klog.Warningf("Test wasn't valid. UL speed %f, DL speed %f", targets[0].ULSpeed, targets[0].DLSpeed)
		return fmt.Errorf("Obtained invalid test results. Recommend not using")
	}
	g.Result = &SpeedtestResult{
		UploadMbps:    targets[0].ULSpeed,
		DownloadMbps:  targets[0].DLSpeed,
		PingLatencyMs: targets[0].Latency,
	}
	return nil
}

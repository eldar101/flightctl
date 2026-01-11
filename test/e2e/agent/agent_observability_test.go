package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/flightctl/flightctl/test/harness/e2e"
	"github.com/flightctl/flightctl/test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const (
	telemetryGatewayNamespace   = "flightctl-external"
	telemetryGatewayConfigMap   = "flightctl-telemetry-gateway-config"
	telemetryGatewayServiceName = "svc/flightctl-telemetry-gateway"
	telemetryGatewayMetricsPort = 9464
	prometheusServiceName       = "svc/flightctl-prometheus"
	prometheusService           = "flightctl-prometheus"
	prometheusPort              = 9090
	metricsEndpointPath         = "/metrics"
	promQueryEndpointPath       = "/api/v1/query"
	telemetryGatewayConfigPath  = "jsonpath={.data.config\\.yaml}"
)

type promQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

var _ = Describe("Device observability", func() {
	Context("telemetry gateway metrics", func() {
		It("should export device host metrics via the telemetry gateway", Label("85040"), func() {
			// Get harness directly - no shared package-level variable
			harness := e2e.GetWorkerHarness()

			ctxStr, err := e2e.GetContext()
			if err != nil || (ctxStr != util.KIND && ctxStr != util.OCP) {
				Skip("Kubernetes context required for telemetry gateway metrics")
			}

			By("verifying telemetry gateway configuration exports Prometheus metrics")
			cfg, err := getTelemetryGatewayConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg).To(ContainSubstring("prometheus"))
			Expect(cfg).To(ContainSubstring("listen"))
			Expect(cfg).To(ContainSubstring("logLevel"))
			Expect(cfg).To(ContainSubstring("tls:"))
			Expect(cfg).To(ContainSubstring("certFile"))
			Expect(cfg).To(ContainSubstring("keyFile"))
			Expect(cfg).To(ContainSubstring("caCert"))

			if !strings.Contains(cfg, "forward:") {
				Skip("telemetry gateway forward configuration is required for this test case")
			}
			Expect(cfg).To(ContainSubstring("endpoint"))
			Expect(cfg).To(ContainSubstring("insecureSkipTlsVerify"))
			Expect(cfg).To(ContainSubstring("caFile"))
			Expect(cfg).To(ContainSubstring("certFile"))
			Expect(cfg).To(ContainSubstring("keyFile"))

			By("enrolling a device and updating to the v10 image with OTEL collector")
			deviceId, _ := harness.EnrollAndWaitForOnlineStatus()
			nextRenderedVersion, err := harness.PrepareNextDeviceVersion(deviceId)
			Expect(err).ToNot(HaveOccurred())
			_, _, err = harness.WaitForBootstrapAndUpdateToVersion(deviceId, util.DeviceTags.V10)
			Expect(err).ToNot(HaveOccurred())
			err = harness.WaitForDeviceNewRenderedVersion(deviceId, nextRenderedVersion)
			Expect(err).ToNot(HaveOccurred())

			By("waiting for otelcol to be running on the device")
			Eventually(func() string {
				stdout, err := harness.VM.RunSSH([]string{"sudo", "systemctl", "is-active", "otelcol"}, nil)
				if err != nil {
					return ""
				}
				return strings.TrimSpace(stdout.String())
			}, TIMEOUT, POLLING).Should(Equal("active"))

			By("port-forwarding telemetry gateway metrics")
			pfCtx, cancel := context.WithCancel(context.Background())
			cmd, done, err := startPortForward(pfCtx, telemetryGatewayNamespace, telemetryGatewayServiceName, telemetryGatewayMetricsPort, telemetryGatewayMetricsPort)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				cancel()
				if cmd != nil && cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				select {
				case <-done:
				case <-time.After(5 * time.Second):
				}
			}()

			By("verifying telemetry gateway metrics include device host metrics")
			metricsURL := fmt.Sprintf("http://127.0.0.1:%d%s", telemetryGatewayMetricsPort, metricsEndpointPath)
			Eventually(func() int {
				body := metricsBody(metricsURL)()
				if body == "" {
					return 0
				}
				return len(strings.Split(strings.TrimSpace(body), "\n"))
			}, TIMEOUT, POLLING).Should(BeNumerically(">=", 10))

			states := []string{"idle", "interrupt", "nice"}
			requiredLabels := []string{"org_id", "otel_scope_schema_url", "otel_scope_version"}
			requiredNonEmpty := []string{"org_id", "otel_scope_version"}
			for _, state := range states {
				exact := map[string]string{
					"cpu":             "cpu0",
					"device_id":       deviceId,
					"otel_scope_name": "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper",
					"state":           state,
				}
				Eventually(func() bool {
					body := metricsBody(metricsURL)()
					if body == "" {
						return false
					}
					families, err := parsePrometheusMetrics(body)
					if err != nil {
						return false
					}
					family, ok := families["system_cpu_time_seconds_total"]
					if !ok {
						return false
					}
					return metricFamilyHasLabels(family, exact, requiredLabels, requiredNonEmpty)
				}, TIMEOUT, POLLING).Should(BeTrue())
			}

			By("verifying Prometheus queries return device metrics")
			err = verifyServiceExists(util.E2E_NAMESPACE, prometheusService)
			Expect(err).ToNot(HaveOccurred())

			promCtx, promCancel := context.WithCancel(context.Background())
			promCmd, promDone, err := startPortForward(promCtx, util.E2E_NAMESPACE, prometheusServiceName, prometheusPort, prometheusPort)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				promCancel()
				if promCmd != nil && promCmd.Process != nil {
					_ = promCmd.Process.Kill()
				}
				select {
				case <-promDone:
				case <-time.After(5 * time.Second):
				}
			}()

			promURL := fmt.Sprintf("http://127.0.0.1:%d", prometheusPort)
			queryAll := fmt.Sprintf(`{device_id="%s"}`, deviceId)
			Eventually(func() int {
				resp, err := promQuery(promURL, queryAll)
				if err != nil {
					return 0
				}
				return len(resp.Data.Result)
			}, TIMEOUT, POLLING).Should(BeNumerically(">", 0))

			queryCount := fmt.Sprintf(`count({device_id="%s"})`, deviceId)
			Eventually(func() float64 {
				resp, err := promQuery(promURL, queryCount)
				if err != nil {
					return 0
				}
				if len(resp.Data.Result) == 0 || len(resp.Data.Result[0].Value) < 2 {
					return 0
				}
				valueStr, ok := resp.Data.Result[0].Value[1].(string)
				if !ok {
					return 0
				}
				val, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					return 0
				}
				return val
			}, TIMEOUT, POLLING).Should(BeNumerically(">", 0))
		})
	})
})

func startPortForward(ctx context.Context, namespace, target string, localPort, remotePort int) (*exec.Cmd, <-chan error, error) {
	// #nosec G204 -- command args are fixed and controlled in test.
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward",
		"-n", namespace,
		target,
		fmt.Sprintf("%d:%d", localPort, remotePort),
		"--address", "127.0.0.1",
	)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	return cmd, done, nil
}

func fetchMetrics(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func metricsBody(url string) func() string {
	return func() string {
		body, err := fetchMetrics(url)
		if err != nil {
			return ""
		}
		return body
	}
}

func parsePrometheusMetrics(body string) (map[string]*dto.MetricFamily, error) {
	parser := expfmt.TextParser{}
	return parser.TextToMetricFamilies(strings.NewReader(body))
}

func metricFamilyHasLabels(family *dto.MetricFamily, exact map[string]string, required, requiredNonEmpty []string) bool {
	if family == nil {
		return false
	}
	for _, metric := range family.GetMetric() {
		if metricMatchesLabels(metric, exact, required, requiredNonEmpty) {
			return true
		}
	}
	return false
}

func metricMatchesLabels(metric *dto.Metric, exact map[string]string, required, requiredNonEmpty []string) bool {
	labelValues := map[string]string{}
	for _, label := range metric.GetLabel() {
		labelValues[label.GetName()] = label.GetValue()
	}
	for name, value := range exact {
		if labelValues[name] != value {
			return false
		}
	}
	for _, name := range required {
		if _, ok := labelValues[name]; !ok {
			return false
		}
	}
	for _, name := range requiredNonEmpty {
		if labelValues[name] == "" {
			return false
		}
	}
	return true
}

func promQuery(baseURL, query string) (promQueryResponse, error) {
	var parsed promQueryResponse
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, baseURL+promQueryEndpointPath, nil)
	if err != nil {
		return parsed, err
	}
	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return parsed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parsed, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return parsed, err
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return parsed, err
	}
	if parsed.Status != "success" {
		return parsed, fmt.Errorf("prometheus query failed: %s", parsed.Status)
	}

	return parsed, nil
}

func getTelemetryGatewayConfig() (string, error) {
	// #nosec G204 -- command args are fixed and controlled in test.
	out, err := exec.Command("kubectl", "get", "configmap", telemetryGatewayConfigMap,
		"-n", telemetryGatewayNamespace,
		"-o", telemetryGatewayConfigPath,
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kubectl get configmap: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func verifyServiceExists(namespace, name string) error {
	// #nosec G204 -- command args are fixed and controlled in test.
	out, err := exec.Command("kubectl", "get", "svc", "-n", namespace, name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl get svc %s/%s: %w: %s", namespace, name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

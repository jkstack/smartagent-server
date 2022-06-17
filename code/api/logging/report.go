package logging

import (
	"server/code/client"

	"github.com/jkstack/anet"
	"github.com/prometheus/client_golang/prometheus"
)

func setValue(vec *prometheus.GaugeVec, id, tag string, n float64) {
	vec.With(prometheus.Labels{
		"id":         id,
		"agent_type": "god",
		"tag":        tag,
	}).Set(n)
}

func (h *Handler) onReport(cli *client.Client, data anet.LoggingReport) {
	setValue(h.stK8s, cli.ID(), "count", float64(data.CountK8s))
	setValue(h.stDocker, cli.ID(), "count", float64(data.CountDocker))
	setValue(h.stFile, cli.ID(), "count", float64(data.CountFile))
	h.onReportAgent(cli.ID(), data.AgentInfo)
	h.onReportReporter(cli.ID(), data.Info)
	h.onReportK8s(cli.ID(), data.K8s)
	h.onReportDocker(cli.ID(), data.Docker)
	h.onReportFile(cli.ID(), data.File)
}

func (h *Handler) onReportAgent(id string, info anet.LoggingReportAgentInfo) {
	h.stAgentVersion.With(prometheus.Labels{
		"id":         id,
		"agent_type": "god",
		"go_version": info.GoVersion,
	}).Set(1)
	setValue(h.stAgent, id, "threads", float64(info.Threads))
	setValue(h.stAgent, id, "routines", float64(info.Routines))
	setValue(h.stAgent, id, "startup", float64(info.Startup))
	setValue(h.stAgent, id, "heap_in_use", float64(info.HeapInuse))
	setValue(h.stAgent, id, "gc_0", info.GC["0"])
	setValue(h.stAgent, id, "gc_0.25", info.GC["25"])
	setValue(h.stAgent, id, "gc_0.5", info.GC["50"])
	setValue(h.stAgent, id, "gc_0.75", info.GC["75"])
	setValue(h.stAgent, id, "gc_1", info.GC["100"])
	setValue(h.stAgent, id, "in_packets", float64(info.InPackets))
	setValue(h.stAgent, id, "in_bytes", float64(info.InBytes))
	setValue(h.stAgent, id, "out_packets", float64(info.OutPackets))
	setValue(h.stAgent, id, "out_bytes", float64(info.OutBytes))
}

func (h *Handler) onReportReporter(id string, info anet.LoggingReportInfo) {
	setValue(h.stReporter, id, "qps", info.QPS)
	setValue(h.stReporter, id, "avg_cost", info.AvgCost)
	setValue(h.stReporter, id, "p_0", float64(info.P0))
	setValue(h.stReporter, id, "p_50", float64(info.P50))
	setValue(h.stReporter, id, "p_90", float64(info.P90))
	setValue(h.stReporter, id, "p_99", float64(info.P99))
	setValue(h.stReporter, id, "p_100", float64(info.P100))
	setValue(h.stReporter, id, "count", float64(info.Count))
	setValue(h.stReporter, id, "bytes", float64(info.Bytes))
	setValue(h.stReporter, id, "http_error_count", float64(info.HttpErr))
	setValue(h.stReporter, id, "service_error_count", float64(info.SrvErr))
}

func (h *Handler) onReportK8s(id string, info anet.LoggingReportK8sData) {
	setValue(h.stK8s, id, "qps", info.QPS)
	setValue(h.stK8s, id, "avg_cost", info.AvgCost)
	setValue(h.stK8s, id, "p_0", float64(info.P0))
	setValue(h.stK8s, id, "p_50", float64(info.P50))
	setValue(h.stK8s, id, "p_90", float64(info.P90))
	setValue(h.stK8s, id, "p_99", float64(info.P99))
	setValue(h.stK8s, id, "p_100", float64(info.P100))
	setValue(h.stK8s, id, "count_service", float64(info.CountService))
	setValue(h.stK8s, id, "count_pod", float64(info.CountPod))
	setValue(h.stK8s, id, "count_container", float64(info.CountContainer))
}

func (h *Handler) onReportDocker(id string, info anet.LoggingReportDockerData) {
	setValue(h.stDocker, id, "count_container", float64(info.Count))
}

func (h *Handler) onReportFile(id string, info anet.LoggingReportFileData) {
	setValue(h.stFile, id, "count_files", float64(info.Count))
}

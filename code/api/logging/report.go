package logging

import (
	"server/code/client"

	"github.com/jkstack/anet"
	"github.com/prometheus/client_golang/prometheus"
)

func setValue(vec *prometheus.GaugeVec, id, tag string, n float64) {
	vec.With(prometheus.Labels{
		"id":  id,
		"tag": tag,
	}).Set(n)
}

func (h *Handler) onReport(cli *client.Client, data anet.LoggingReport) {
	setValue(h.stK8s, cli.ID(), "k8s_task_count", float64(data.CountK8s))
	setValue(h.stDocker, cli.ID(), "docker_task_count", float64(data.CountDocker))
	setValue(h.stFile, cli.ID(), "file_task_count", float64(data.CountFile))
	h.onReportAgent(cli.ID(), data.AgentInfo)
}

func (h *Handler) onReportAgent(id string, info anet.LoggingReportAgentInfo) {
	h.stAgentVersion.With(prometheus.Labels{
		"id":         id,
		"go_version": info.GoVersion,
	}).Set(1)
	setValue(h.stAgent, id, "threads", float64(info.Threads))
	setValue(h.stAgent, id, "routines", float64(info.Routines))
	setValue(h.stAgent, id, "startup", float64(info.Startup))
	setValue(h.stAgent, id, "heap_in_use", float64(info.HeapInuse))
	setValue(h.stAgent, id, "gc_0", info.GC["0"])
	setValue(h.stAgent, id, "gc_25", info.GC["0.25"])
	setValue(h.stAgent, id, "gc_50", info.GC["0.5"])
	setValue(h.stAgent, id, "gc_75", info.GC["0.75"])
	setValue(h.stAgent, id, "gc_100", info.GC["1"])
	setValue(h.stAgent, id, "in_packets", float64(info.InPackets))
	setValue(h.stAgent, id, "in_bytes", float64(info.InBytes))
	setValue(h.stAgent, id, "out_packets", float64(info.OutPackets))
	setValue(h.stAgent, id, "out_bytes", float64(info.OutBytes))
}

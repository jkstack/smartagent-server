package agent

import (
	"github.com/jkstack/anet"
	"github.com/lwch/logging"
	"github.com/prometheus/client_golang/prometheus"
)

func setValue(vec *prometheus.GaugeVec, id, tag string, n float64) {
	vec.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"tag":        tag,
	}).Set(n)
}

func (h *Handler) basicInfo(id string, info *anet.AgentInfo) {
	h.stAgentVersion.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"version":    info.Version,
		"go_version": info.GoVersion,
	}).Set(1)
	logging.Info("gc of agent [%s]: %v", id, info.GC)
	setValue(h.stAgentInfo, id, "cpu_usage", float64(info.CpuUsage))
	setValue(h.stAgentInfo, id, "memory_usage", float64(info.MemoryUsage))
	setValue(h.stAgentInfo, id, "threads", float64(info.Threads))
	setValue(h.stAgentInfo, id, "routines", float64(info.Routines))
	setValue(h.stAgentInfo, id, "startup", float64(info.Startup))
	setValue(h.stAgentInfo, id, "heap_in_use", float64(info.HeapInuse))
	setValue(h.stAgentInfo, id, "gc_0", info.GC["0"])
	setValue(h.stAgentInfo, id, "gc_0.25", info.GC["25"])
	setValue(h.stAgentInfo, id, "gc_0.5", info.GC["50"])
	setValue(h.stAgentInfo, id, "gc_0.75", info.GC["75"])
	setValue(h.stAgentInfo, id, "gc_1", info.GC["100"])
	setValue(h.stAgentInfo, id, "in_packets", float64(info.InPackets))
	setValue(h.stAgentInfo, id, "in_bytes", float64(info.InBytes))
	setValue(h.stAgentInfo, id, "out_packets", float64(info.OutPackets))
	setValue(h.stAgentInfo, id, "out_bytes", float64(info.OutBytes))
}

func setPluginValue(vec *prometheus.GaugeVec, id, name, tag string, n float64) {
	vec.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"name":       name,
		"tag":        tag,
	}).Set(n)
}

func (h *Handler) pluginInfo(id string, info *anet.AgentInfo) {
	setValue(h.stAgentInfo, id, "plugin_execd", float64(info.PluginExecd))
	setValue(h.stAgentInfo, id, "plugin_running", float64(info.PluginRunning))
	for k, v := range info.PluginUseCount {
		setPluginValue(h.stPlugin, id, k, "use", float64(v))
	}
	for k, v := range info.PluginOutPackets {
		setPluginValue(h.stPlugin, id, k, "out_packets", float64(v))
	}
	for k, v := range info.PluginOutBytes {
		setPluginValue(h.stPlugin, id, k, "out_bytes", float64(v))
	}
}

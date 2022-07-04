package agent

import (
	"github.com/jkstack/anet"
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
	h.stAgent.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"version":    info.Version,
	}).Set(1)
	h.stAgent.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"go_version": info.GoVersion,
	}).Set(1)
	setValue(h.stAgent, id, "cpu_usage", float64(info.CpuUsage))
	setValue(h.stAgent, id, "memory_usage", float64(info.MemoryUsage))
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

func setPluginValue(vec *prometheus.GaugeVec, id, name, tag string, n float64) {
	vec.With(prometheus.Labels{
		"id":         id,
		"agent_type": "smart",
		"name":       name,
		"tag":        tag,
	}).Set(n)
}

func (h *Handler) pluginInfo(id string, info *anet.AgentInfo) {
	setValue(h.stAgent, id, "plugin_execd", float64(info.PluginExecd))
	setValue(h.stAgent, id, "plugin_running", float64(info.PluginRunning))
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

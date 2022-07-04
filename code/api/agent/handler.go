package agent

import (
	"server/code/client"
	"server/code/conf"

	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
	"github.com/prometheus/client_golang/prometheus"
)

// Handler server handler
type Handler struct {
	cfg      *conf.Configure
	stAgent  *prometheus.GaugeVec
	stPlugin *prometheus.GaugeVec
}

// New new cmd handler
func New() *Handler {
	return &Handler{}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	h.stAgent = stats.RawVec("agent_info", []string{"id", "agent_type", "tag"})
	h.stPlugin = stats.RawVec("plugin_info", []string{"id", "agent_type", "name", "tag"})
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/agent/exists":    h.exists,
		"/agent/sniffer":   h.sniffer,
		"/agent/install":   h.install,
		"/agent/config":    h.config,
		"/agent/uninstall": h.uninstall,
		"/agent/restart":   h.restart,
		"/agent/start":     h.start,
		"/agent/stop":      h.stop,
		"/agent/upgrade":   h.upgrade,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

func (h *Handler) OnMessage(cli *client.Client, msg *anet.Msg) {
	if msg.Type != anet.TypeAgentInfo {
		return
	}
	h.basicInfo(cli.ID(), msg.AgentInfo)
	h.pluginInfo(cli.ID(), msg.AgentInfo)
}

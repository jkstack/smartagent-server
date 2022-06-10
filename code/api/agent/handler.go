package agent

import (
	"server/code/client"
	"server/code/conf"

	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
)

// Handler server handler
type Handler struct {
	cfg *conf.Configure
}

// New new cmd handler
func New() *Handler {
	return &Handler{}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
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

func (h *Handler) OnMessage(*client.Client, *anet.Msg) {
}

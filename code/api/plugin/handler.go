package plugin

import (
	"server/code/client"
	"server/code/conf"

	"github.com/lwch/api"
)

// Handler handler
type Handler struct {
	cfg *conf.Configure
}

// New new handler
func New() *Handler {
	return &Handler{}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure) {
	h.cfg = cfg
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/file/plugin/":     h.serveFile,
		"/plugin/reload":    h.reload,
		"/plugin/install":   h.install,
		"/plugin/list":      h.list,
		"/plugin/uninstall": h.uninstall,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

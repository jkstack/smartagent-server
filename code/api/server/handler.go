package server

import (
	"server/code/client"
	"server/code/conf"

	"github.com/lwch/api"
)

// Handler server handler
type Handler struct {
	cfg     *conf.Configure
	version string
}

// New new cmd handler
func New(version string) *Handler {
	return &Handler{version: version}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure) {
	h.cfg = cfg
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/server/info": h.info,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

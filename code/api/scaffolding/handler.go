package scaffolding

import (
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/lwch/api"
)

// Handler server handler
type Handler struct {
	sync.RWMutex
	cfg *conf.Configure
}

// New new cmd handler
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
		"/scaffolding/foo": h.foo,
	}
}

func (h *Handler) OnConnect(cli *client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

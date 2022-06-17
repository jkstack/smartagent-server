package scaffolding

import (
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/stat"
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
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
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

func (h *Handler) OnMessage(*client.Client, *anet.Msg) {
}

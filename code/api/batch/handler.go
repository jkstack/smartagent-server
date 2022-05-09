package batch

import (
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/lwch/api"
)

// Handler agent handler
type Handler struct {
	sync.RWMutex
}

// New create handler
func New() *Handler {
	return &Handler{}
}

// Init init module
func (h *Handler) Init(*conf.Configure) {
}

// HandleFuncs list funcs
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/batch/generate": h.generate,
		"/batch/verify":   h.verify,
		"/batch/commit":   h.commit,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

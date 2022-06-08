package hm

import (
	lapi "server/code/api"
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
)

// Handler cmd handler
type Handler struct {
	sync.RWMutex
	cfg          *conf.Configure
	stUsage      *stat.Counter
	stTotalTasks *stat.Counter
}

// New new cmd handler
func New() *Handler {
	return &Handler{}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	h.stUsage = stats.NewCounter("plugin_count_hm")
	h.stTotalTasks = stats.NewCounter(lapi.TotalTasksLabel)
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/hm/static": h.static,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

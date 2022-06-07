package cmd

import (
	"server/code/client"
	"server/code/conf"
	"sync"
	"time"

	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
)

const clearTimeout = 30 * time.Minute

// Handler cmd handler
type Handler struct {
	sync.RWMutex
	cfg     *conf.Configure
	clients map[string]*cmdClient // cid => client
	stUsage *stat.Counter
}

// New new cmd handler
func New() *Handler {
	return &Handler{
		clients: make(map[string]*cmdClient),
	}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	h.stUsage = stats.NewCounter("plugin_count_exec")
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/cmd/run":      h.run,
		"/cmd/ps":       h.ps,
		"/cmd/pty":      h.pty,
		"/cmd/kill":     h.kill,
		"/cmd/status":   h.status,
		"/cmd/sync_run": h.syncRun,
		"/cmd/channel/": h.channel,
	}
}

func (h *Handler) cli(cli *client.Client) *cmdClient {
	h.Lock()
	defer h.Unlock()
	if cli, ok := h.clients[cli.ID()]; ok {
		return cli
	}
	c := newClient(cli)
	h.clients[cli.ID()] = c
	return c
}

func (h *Handler) cliNew(cli *client.Client) *cmdClient {
	c := newClient(cli)
	h.Lock()
	h.clients[cli.ID()] = c
	h.Unlock()
	return c
}

func (h *Handler) cliFrom(id string) *cmdClient {
	h.RLock()
	defer h.RUnlock()
	return h.clients[id]
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(id string) {
	h.RLock()
	cli := h.clients[id]
	h.RUnlock()
	if cli != nil {
		cli.close()
		h.Lock()
		delete(h.clients, id)
		h.Unlock()
	}
}

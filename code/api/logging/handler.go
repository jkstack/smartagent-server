package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

// Handler server handler
type Handler struct {
	sync.RWMutex
	cfg  *conf.Configure
	data map[int64]*context
}

// New new cmd handler
func New() *Handler {
	return &Handler{
		data: make(map[int64]*context),
	}
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	runtime.Assert(h.loadConfig(filepath.Join(h.cfg.DataDir, "logging")))
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/logging/config": h.config,
		"/logging/start":  h.start,
		"/logging/stop":   h.stop,
		"/logging/remove": h.remove,
	}
}

func (h *Handler) OnConnect(cli *client.Client) {
	var send []*context
	h.RLock()
	for _, ctx := range h.data {
		if ctx.in(cli.ID()) {
			send = append(send, ctx)
		}
	}
	h.RUnlock()
	// TODO: collect not running
	for _, ctx := range send {
		ctx.reSend(cli, h.cfg.LoggingReport)
	}
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

func (h *Handler) loadConfig(dir string) error {
	os.MkdirAll(dir, 0755)
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	load := func(dir string) error {
		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer f.Close()
		var ctx context
		err = json.NewDecoder(f).Decode(&ctx)
		if err != nil {
			return err
		}
		h.Lock()
		h.data[ctx.ID] = &ctx
		h.Unlock()
		return nil
	}
	for _, file := range files {
		err = load(file)
		if err != nil {
			return err
		}
	}
	return nil
}

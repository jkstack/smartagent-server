package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	lapi "server/code/api"
	"server/code/client"
	"server/code/conf"
	"sync"

	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
	"github.com/lwch/runtime"
	"github.com/prometheus/client_golang/prometheus"
)

// Handler server handler
type Handler struct {
	sync.RWMutex
	cfg            *conf.Configure
	data           map[int64]*context
	stTotalTasks   *stat.Counter
	stK8s          *prometheus.GaugeVec
	stDocker       *prometheus.GaugeVec
	stFile         *prometheus.GaugeVec
	stAgentVersion *prometheus.GaugeVec
	stAgent        *prometheus.GaugeVec
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
	h.stTotalTasks = stats.NewCounter(lapi.TotalTasksLabel)
	h.stK8s = stats.RawVec("agent_logging_k8s_info", []string{"id", "tag"})
	h.stDocker = stats.RawVec("agent_logging_docker_info", []string{"id", "tag"})
	h.stFile = stats.RawVec("agent_logging_file_info", []string{"id", "tag"})
	h.stAgentVersion = stats.RawVec("agent_version", []string{"id", "go_version"})
	h.stAgent = stats.RawVec("agent_info", []string{"id", "tag"})
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

func (h *Handler) OnMessage(cli *client.Client, msg *anet.Msg) {
	if msg.Type != anet.TypeLoggingReport {
		return
	}
	h.onReport(cli, *msg.LoggingReport)
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
		ctx.Args.parent = &ctx
		ctx.parent = h
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

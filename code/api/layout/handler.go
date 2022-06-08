package layout

import (
	lapi "server/code/api"
	"server/code/api/file"
	"server/code/client"
	"server/code/conf"
	"sync"
	"time"

	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
)

type taskHandler interface {
	Name() string
	Clone(Task, taskInfo) taskHandler
	Check(Task) error
	WantRun(map[string]string) bool
	Run(id, dir, user, pass string, args map[string]string) error
}

// Handler server handler
type Handler struct {
	sync.RWMutex
	cfg          *conf.Configure
	runners      map[string]*runner
	handlers     map[string]taskHandler
	idx          uint32
	fh           *file.Handler
	stExecUsage  *stat.Counter
	stFileUsage  *stat.Counter
	stTotalTasks *stat.Counter
}

// New new cmd handler
func New(fh *file.Handler) *Handler {
	h := &Handler{
		runners:  make(map[string]*runner),
		handlers: make(map[string]taskHandler),
		fh:       fh,
	}
	h.handlers["exec"] = &execHandler{}
	h.handlers["file"] = &fileHandler{}
	go h.clear()
	return h
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	h.stExecUsage = stats.NewCounter("plugin_count_exec")
	h.stFileUsage = stats.NewCounter("plugin_count_file")
	h.stTotalTasks = stats.NewCounter(lapi.TotalTasksLabel)
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/layout/run":    h.run,
		"/layout/status": h.status,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

func (h *Handler) clear() {
	run := func() {
		remove := make([]string, 0, len(h.runners))
		h.RLock()
		for id, r := range h.runners {
			if time.Since(r.deadline).Hours() > 1 {
				remove = append(remove, id)
			}
		}
		h.RUnlock()

		h.Lock()
		for _, id := range remove {
			delete(h.runners, id)
		}
		h.Unlock()
	}
	for {
		run()
		time.Sleep(time.Minute)
	}
}

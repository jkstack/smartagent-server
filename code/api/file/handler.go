package file

import (
	"os"
	lapi "server/code/api"
	"server/code/client"
	"server/code/conf"
	"sync"
	"time"

	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/api"
	"github.com/lwch/logging"
)

// Handler cmd handler
type Handler struct {
	sync.RWMutex
	cfg          *conf.Configure
	uploadCache  map[string]*uploadInfo
	stUsage      *stat.Counter
	stTotalTasks *stat.Counter
}

// New new cmd handler
func New() *Handler {
	h := &Handler{
		uploadCache: make(map[string]*uploadInfo),
	}
	go h.clear()
	return h
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure, stats *stat.Mgr) {
	h.cfg = cfg
	h.stUsage = stats.NewCounter("plugin_count_file")
	h.stTotalTasks = stats.NewCounter(lapi.TotalTasksLabel)
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/file/ls":          h.ls,
		"/file/download":    h.download,
		"/file/upload":      h.upload,
		"/file/upload/":     h.uploadHandle,
		"/file/upload_from": h.uploadFrom,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

func (h *Handler) clear() {
	run := func() {
		list := make([]string, 0, len(h.uploadCache))
		h.RLock()
		for id, cache := range h.uploadCache {
			if time.Now().After(cache.timeout) {
				list = append(list, id)
			}
		}
		h.RUnlock()
		if len(list) > 0 {
			logging.Info("collected %d files to remove from upload cache", len(list))
		}
		var cnt int
		for _, id := range list {
			if h.RemoveUploadCache(id) {
				cnt++
			}
		}
		if len(list) > 0 {
			logging.Info("%d files removed from upload cache", cnt)
		}
	}
	for {
		run()
		time.Sleep(time.Minute)
	}
}

func (h *Handler) LogUploadCache(taskID, dir, token string,
	deadline time.Time, rm bool) {
	h.Lock()
	h.uploadCache[taskID] = &uploadInfo{
		token:   token,
		dir:     dir,
		rm:      rm,
		timeout: deadline,
	}
	h.Unlock()
	logging.Info("log cache: %s", taskID)
}

func (h *Handler) RemoveUploadCache(id string) bool {
	h.Lock()
	cache := h.uploadCache[id]
	delete(h.uploadCache, id)
	h.Unlock()
	if cache == nil {
		return false
	}
	logging.Info("removed cache: %s", id)
	if cache.rm {
		os.Remove(cache.dir)
	}
	return true
}

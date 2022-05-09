package install

import (
	"server/code/client"
	"server/code/conf"
	"server/code/utils"
	"sync"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
)

type Action struct {
	Action string `json:"action"`
	Time   int64  `json:"time"`
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Msg    string `json:"msg"`
}

type Info struct {
	updated time.Time
	Done    bool     `json:"done"`
	Actions []Action `json:"actions"`
}

// Handler cmd handler
type Handler struct {
	sync.RWMutex
	cfg  *conf.Configure
	data map[string]*Info
}

// New new cmd handler
func New() *Handler {
	h := &Handler{data: make(map[string]*Info)}
	go h.clear()
	return h
}

// Init init handler
func (h *Handler) Init(cfg *conf.Configure) {
	h.cfg = cfg
}

// HandleFuncs get handle functions
func (h *Handler) HandleFuncs() map[string]func(*client.Clients, *api.Context) {
	return map[string]func(*client.Clients, *api.Context){
		"/install/run":    h.run,
		"/install/status": h.status,
	}
}

func (h *Handler) OnConnect(*client.Client) {
}

// OnClose agent on close
func (h *Handler) OnClose(string) {
}

func (h *Handler) loop(cli *client.Client, taskID string) {
	ch := cli.ChanRead(taskID)
	h.RLock()
	info := h.data[taskID]
	h.RUnlock()
	for {
		msg := <-ch
		switch msg.Type {
		case anet.TypeInstallRep:
			rep := msg.InstallRep
			add := func(act string) {
				info.updated = time.Now()
				info.Actions = append(info.Actions, Action{
					Action: act,
					Time:   rep.Time,
					Name:   rep.Name,
					OK:     rep.OK,
					Msg:    rep.Msg,
				})
			}
			switch rep.Action {
			case anet.InstallActionDownload:
				add("download")
			case anet.InstallActionInstall:
				add("install")
			case anet.InstallActionDone:
				add("done")
				info.Done = true
			}
		case anet.TypeError:
			info.updated = time.Now()
			info.Actions = append(info.Actions, Action{
				Action: "done",
				Time:   time.Now().Unix(),
				OK:     false,
				Msg:    msg.ErrorMsg,
			})
			info.Done = true
		}
	}
}

func (h *Handler) clear() {
	for {
		time.Sleep(time.Minute)
		h.remove()
	}
}

func (h *Handler) remove() {
	defer utils.Recover("remove")
	remove := make([]string, 0, len(h.data))
	h.RLock()
	for id, info := range h.data {
		if time.Since(info.updated).Hours() > 24 {
			remove = append(remove, id)
		}
	}
	h.RUnlock()

	h.Lock()
	for _, id := range remove {
		delete(h.data, id)
	}
	h.Unlock()
}

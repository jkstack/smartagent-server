package layout

import (
	"fmt"
	"server/code/client"
	"server/code/conf"
	"sync"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const (
	runSeq = iota + 1
	runPar
	runEvo
)

type status struct {
	ID       string `json:"id"`
	Created  int64  `json:"created"`
	Finished int64  `json:"finished"`
	OK       bool   `json:"ok"`
	Msg      string `json:"msg"`
}

type runner struct {
	sync.RWMutex
	parent   *Handler
	id       string
	hosts    []string
	flags    byte
	user     string
	pass     string
	args     []map[string]string
	deadline time.Time
	tasks    []taskHandler
	clients  *client.Clients
	// status
	done     bool
	created  int64
	finished int64
	nodes    map[string]*status
}

func newRunner(parent *Handler, clients *client.Clients,
	idx uint32, hosts []string,
	mode string,
	errContinue bool,
	user, pass string) *runner {
	rand, err := runtime.UUID(8, "0123456789abcdef")
	runtime.Assert(err)
	taskID := fmt.Sprintf("yaml-%s-%06d-%s",
		time.Now().Format("20060102"), idx, rand)
	return &runner{
		parent:  parent,
		id:      taskID,
		hosts:   hosts,
		flags:   makeFlags(mode, errContinue),
		user:    user,
		pass:    pass,
		clients: clients,
		nodes:   make(map[string]*status),
	}
}

func makeFlags(mode string, errContinue bool) byte {
	var h byte
	if errContinue {
		h = 1 << 7
	}
	var l byte
	switch mode {
	case "sequence":
		l = runSeq
	case "parallel":
		l = runPar
	case "evenodd":
		l = runEvo
	}
	return h | l
}

func (r *runner) runMode() byte {
	return r.flags & 0x7f
}

func (r *runner) errContinue() bool {
	return (r.flags >> 7) > 0
}

func (r *runner) runHost(id, dir string, args map[string]string) error {
	status := &status{
		ID:      id,
		Created: time.Now().Unix(),
	}
	r.Lock()
	r.nodes[id] = status
	r.Unlock()
	for _, t := range r.tasks {
		if !t.WantRun(args) {
			continue
		}
		err := t.Run(id, dir, r.user, r.pass, args)
		if err != nil {
			status.Finished = time.Now().Unix()
			status.OK = false
			status.Msg = fmt.Sprintf("run task [%s] failed: %v", t.Name(), err.Error())
			logging.Error("run task [%s] on agent [%s] failed: %v", t.Name(), id, err)
			return err
		}
	}
	status.Finished = time.Now().Unix()
	status.OK = true
	return nil
}

func (r *runner) run(dir string) {
	defer func() {
		r.done = true
		r.finished = time.Now().Unix()
		logging.Info("task [%s] run done", r.id)
	}()
	r.created = time.Now().Unix()
	r.args = make([]map[string]string, len(r.hosts))
	for i := 0; i < len(r.hosts); i++ {
		r.args[i] = make(map[string]string)
	}
	switch r.runMode() {
	case runSeq:
		for i, host := range r.hosts {
			args := r.args[i]
			initArgs(host, i, args)
			err := r.runHost(host, dir, args)
			if err != nil && !r.errContinue() {
				break
			}
		}
	case runPar:
		for i, host := range r.hosts {
			args := r.args[i]
			initArgs(host, i, args)
			go r.runHost(host, dir, args)
		}
	case runEvo:
		var wg sync.WaitGroup
		var err error
		wg.Add((len(r.hosts) + 1) >> 1)
		for i := 0; i < len(r.hosts); i += 2 {
			args := r.args[i]
			initArgs(r.hosts[i], i, args)
			go func(id string, args map[string]string) {
				defer wg.Done()
				er := r.runHost(id, dir, args)
				if er != nil {
					err = er
				}
			}(r.hosts[i], args)
		}
		wg.Wait()
		if err != nil && !r.errContinue() {
			break
		}
		for i := 1; i < len(r.hosts); i += 2 {
			args := r.args[i]
			initArgs(r.hosts[i], i, args)
			go r.runHost(r.hosts[i], dir, args)
		}
	}
}

func initArgs(agentID string, idx int, args map[string]string) {
	args["ID"] = agentID
	args["IDX"] = fmt.Sprintf("%d", idx)
	if idx%2 == 0 {
		args["EVEN"] = "1"
		args["ODD"] = "0"
	} else {
		args["EVEN"] = "0"
		args["ODD"] = "1"
	}
}

func (r *runner) GetClient(id string) *client.Client {
	return r.clients.Get(id)
}

func (r *runner) GetPlugin(name, os, arch string) *conf.PluginInfo {
	return r.parent.cfg.GetPlugin(name, os, arch)
}

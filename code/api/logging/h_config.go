package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type configArgs struct {
	parent   *context
	Exclude  string        `json:"exclude"`
	Batch    int           `json:"batch"`
	Buffer   int           `json:"buffer"`
	Interval int           `json:"interval"`
	K8s      *k8sConfig    `json:"k8s,omitempty"`
	Docker   *dockerConfig `json:"docker,omitempty"`
	File     *fileConfig   `json:"file,omitempty"`
}

type context struct {
	parent  *Handler
	ID      int64      `json:"id"`
	Args    configArgs `json:"args"`
	Targets []string   `json:"cids"`
	Started bool       `json:"started"`
}

func (ctx *context) in(id string) bool {
	for _, cid := range ctx.Targets {
		if cid == id {
			return true
		}
	}
	return false
}

func (h *Handler) config(clients *client.Clients, ctx *api.Context) {
	t := ctx.XStr("type")
	var rt context
	rt.Args.parent = &rt
	rt.ID = ctx.XInt64("pid")
	rt.Args.Exclude = ctx.OStr("exclude", "")
	rt.Args.Batch = ctx.OInt("batch", 1000)
	rt.Args.Buffer = ctx.OInt("buffer", 4096)
	rt.Args.Interval = ctx.OInt("interval", 30)

	var err error

	if len(rt.Args.Exclude) > 0 {
		_, err = regexp.Compile(rt.Args.Exclude)
		if err != nil {
			lapi.BadParamErr(fmt.Sprintf("exclude: %v", err))
			return
		}
	}

	switch t {
	case "k8s":
		rt.Args.K8s = new(k8sConfig)
		err = rt.Args.K8s.build(ctx)
	case "docker":
		rt.Args.Docker = new(dockerConfig)
		err = rt.Args.Docker.build(ctx)
	case "logtail":
		rt.Args.File = new(fileConfig)
		err = rt.Args.File.build(ctx)
	default:
		lapi.BadParamErr("type")
		return
	}
	runtime.Assert(err)

	switch {
	case rt.Args.K8s != nil:
		var cid string
		cid, err = rt.Args.sendK8s(clients, rt.ID, h.cfg.LoggingReport)
		if err == errNoCollector {
			ctx.ERR(1, err.Error())
			return
		}
		rt.Targets = []string{cid}
	default:
		ids := ctx.XCsv("ids")
		var clis []*client.Client
		for _, id := range ids {
			cli := clients.Get(id)
			if cli == nil {
				ctx.NotFound(fmt.Sprintf("agent: %s", id))
				return
			}
			clis = append(clis, cli)
		}
		err = rt.Args.sendTargets(clis, rt.ID, h.cfg.LoggingReport)
		rt.Targets = ids
	}
	runtime.Assert(err)

	dir := filepath.Join(h.cfg.DataDir, "logging", fmt.Sprintf("%d.json", rt.ID))
	err = saveConfig(dir, rt)
	runtime.Assert(err)

	h.Lock()
	h.data[rt.ID] = &rt
	h.Unlock()

	ctx.OK(nil)
}

func saveConfig(dir string, rt context) error {
	os.MkdirAll(filepath.Dir(dir), 0755)
	f, err := os.Create(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(rt)
}

func (ctx *context) reSend(cli *client.Client, report string) {
	err := ctx.Args.sendTo(cli, ctx.ID, report)
	if err != nil {
		logging.Error("send logging config of project %d to client [%s]: %v",
			ctx.ID, cli.ID())
		return
	}
	if ctx.Started {
		taskID, err := cli.SendLoggingStart(ctx.ID)
		if err != nil {
			logging.Error("send logging start of project %d: %v",
				ctx.ID, cli.ID())
			return
		}
		defer cli.ChanClose(taskID)
		var msg *anet.Msg
		select {
		case msg = <-cli.ChanRead(taskID):
		case <-time.After(api.RequestTimeout):
			logging.Error("wait logging start status of project %d: %v",
				ctx.ID, cli.ID())
			return
		}

		switch {
		case msg.Type == anet.TypeError:
			logging.Error("get logging start status of project %d: %v",
				ctx.ID, cli.ID())
			return
		case msg.Type != anet.TypeLoggingStatusRep:
			logging.Error("get logging start status of project %d: %v",
				ctx.ID, cli.ID())
			return
		}

		if !msg.LoggingStatusRep.OK {
			logging.Error("get logging start status of project %d: %v",
				ctx.ID, cli.ID())
			return
		}
	}
}

func (args *configArgs) sendTo(cli *client.Client, pid int64, report string) error {
	switch {
	case args.K8s != nil:
		_, err := cli.SendLoggingConfigK8s(pid, args.Exclude,
			args.Batch, args.Buffer, args.Interval, report,
			args.K8s.Namespace, args.K8s.Names, args.K8s.Dir, args.K8s.Api, args.K8s.Token)
		args.parent.parent.stTotalTasks.Inc()
		return err
	case args.Docker != nil:
		_, err := cli.SendLoggingConfigDocker(pid, args.Exclude,
			args.Batch, args.Buffer, args.Interval, report,
			args.Docker.ContainerName, args.Docker.ContainerTag, args.Docker.Dir)
		args.parent.parent.stTotalTasks.Inc()
		return err
	case args.File != nil:
		_, err := cli.SendLoggingConfigFile(pid, args.Exclude,
			args.Batch, args.Buffer, args.Interval, report,
			args.File.Dir)
		args.parent.parent.stTotalTasks.Inc()
		return err
	default:
		return errors.New("unsupported")
	}
}

func (args *configArgs) sendK8s(clients *client.Clients, pid int64, report string) (string, error) {
	var cli *client.Client
	clis := clients.Prefix(args.K8s.Namespace + "-k8s-")
	if len(clis) == 0 {
		clis = clients.Prefix("k8s-")
		if len(clis) == 0 {
			return "", errNoCollector
		}
	}
	cli = clis[int(pid)%len(clis)]
	err := args.sendTo(cli, pid, report)
	if err != nil {
		return "", err
	}
	return cli.ID(), nil
}

func (args *configArgs) sendTargets(targets []*client.Client, pid int64, report string) error {
	for _, cli := range targets {
		err := args.sendTo(cli, pid, report)
		if err != nil {
			logging.Error("send logging config to %s: %v", cli.ID(), err)
			return err
		}
	}
	return nil
}

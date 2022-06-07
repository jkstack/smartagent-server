package cmd

import (
	"fmt"
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (h *Handler) run(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	cmd := ctx.XStr("cmd")
	args := ctx.OCsv("args", []string{})
	timeout := ctx.OInt("timeout", 3600)
	auth := ctx.OStr("auth", "")
	user := ctx.OStr("user", "")
	pass := ctx.OStr("pass", "")
	workdir := ctx.OStr("workdir", "")
	env := ctx.OCsv("env", []string{})

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("exec", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("exec")
	}

	if timeout <= 0 {
		timeout = 3600
	}

	runCli := h.cliFrom(id)
	if runCli == nil {
		runCli = h.cliNew(cli)
	}

	taskID, err := cli.SendExec(p, cmd, args, timeout, auth, user, pass, workdir, env)
	runtime.Assert(err)

	h.stUsage.Inc()

	logging.Info("run [%s] on %s, task_id=%s, plugin.version=%s", cmd, id, taskID, p.Version)

	var msg *anet.Msg
	select {
	case msg = <-cli.ChanRead(taskID):
	case <-time.After(api.RequestTimeout):
		ctx.Timeout()
	}

	switch {
	case msg.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, msg.ErrorMsg)
		return
	case msg.Type != anet.TypeExecd:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", msg.Type))
		return
	}

	if !msg.Execd.OK {
		ctx.ERR(1, msg.Execd.Msg)
		return
	}

	process, err := newProcess(runCli, h.cfg, msg.Execd.Pid, cmd, taskID)
	runtime.Assert(err)
	go process.recv()
	runCli.addProcess(process)

	ctx.OK(map[string]interface{}{
		"channel_id": taskID,
		"pid":        msg.Execd.Pid,
	})
}

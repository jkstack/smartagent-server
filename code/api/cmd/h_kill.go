package cmd

import (
	lapi "server/code/api"
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/logging"
)

func (h *Handler) kill(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	pid := ctx.XInt("pid")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	runCli := h.cli(cli)
	if cli == nil {
		ctx.NotFound("client")
	}

	runCli.RLock()
	process := runCli.process[pid]
	runCli.RUnlock()

	if process == nil {
		ctx.NotFound("process")
	}

	p := h.cfg.GetPlugin("exec", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("exec")
	}

	taskID := process.sendKill(p)

	h.stTotalTasks.Inc()

	logging.Info("kill [%d] on %s, task_id=%s, plugin.version=%s", pid, id, taskID, p.Version)

	ctx.OK(nil)
}

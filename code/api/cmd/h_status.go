package cmd

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) status(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	pid := ctx.XInt("pid")

	cli := h.cliFrom(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	cli.RLock()
	p := cli.process[pid]
	cli.RUnlock()

	if p == nil {
		ctx.NotFound("process")
	}

	ctx.OK(map[string]interface{}{
		"id":      p.id,
		"channel": p.taskID,
		"created": p.created.Unix(),
		"running": p.running,
		"code":    p.code,
	})
}

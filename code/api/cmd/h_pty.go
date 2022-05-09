package cmd

import (
	"io"
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) pty(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	pid := ctx.XInt("pid")

	cli := h.cliFrom(id)
	if cli == nil {
		ctx.HTTPNotFound("client")
		return
	}

	cli.RLock()
	p := cli.process[pid]
	cli.RUnlock()

	if p == nil {
		ctx.HTTPNotFound("process")
	}

	data, err := p.read()
	if err != nil && err != io.EOF {
		panic(err)
	}

	ctx.Body(data)
}

package host

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) info(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	ctx.OK(map[string]interface{}{
		"hostname": cli.HostName(),
		"os":       cli.OS(),
	})
}

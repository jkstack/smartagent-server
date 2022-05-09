package agent

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) exists(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	if clients.Get(id) == nil {
		ctx.NotFound("id")
	}
	ctx.OK(nil)
}

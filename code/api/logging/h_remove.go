package logging

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) remove(clients *client.Clients, ctx *api.Context) {
	ctx.OK(nil)
}

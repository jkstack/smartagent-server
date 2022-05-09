package plugin

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) reload(clients *client.Clients, ctx *api.Context) {
	h.cfg.LoadPlugin()
	ctx.OK(nil)
}

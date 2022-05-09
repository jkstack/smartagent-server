package host

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) search(clients *client.Clients, ctx *api.Context) {
	keyword := ctx.XStr("keyword")

	var id string
	clients.Range(func(c *client.Client) bool {
		if c.IP() == keyword || c.Mac() == keyword {
			id = c.ID()
			return false
		}
		return true
	})

	ctx.OK(id)
}

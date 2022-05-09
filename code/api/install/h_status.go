package install

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) status(clients *client.Clients, ctx *api.Context) {
	taskID := ctx.XStr("task_id")

	h.RLock()
	info := h.data[taskID]
	h.RUnlock()

	if info == nil {
		ctx.NotFound("task")
		return
	}

	ctx.OK(*info)
}

package cmd

import (
	"server/code/client"
	"sort"
	"time"

	"github.com/lwch/api"
)

func (h *Handler) ps(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")

	type item struct {
		ID      int    `json:"id"`
		Channel string `json:"channel"`
		Name    string `json:"name"`
		Start   int64  `json:"start_time"`
		Up      int64  `json:"up_time"`
	}

	cli := h.cliFrom(id)
	if cli == nil {
		ctx.OK([]item{})
		return
	}

	var list []item

	cli.RLock()
	for _, p := range cli.process {
		if !p.running {
			continue
		}
		list = append(list, item{
			ID:      p.id,
			Channel: p.taskID,
			Name:    p.cmd,
			Start:   p.created.Unix(),
			Up:      int64(time.Since(p.created).Seconds()),
		})
	}
	cli.RUnlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].Start < list[j].Start
	})

	ctx.OK(list)
}

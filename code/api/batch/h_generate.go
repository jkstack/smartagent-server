package batch

import (
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) generate(clients *client.Clients, ctx *api.Context) {
	cnt := ctx.XInt("count")
	ret := make(map[string]bool, cnt)
	for i := 0; i < cnt; i++ {
		var err error
		for j := 0; j < 10; j++ {
			var id string
			id, err = runtime.UUID(16, "0123456789abcdef")
			if err != nil {
				continue
			}
			id = "agent-" + id
			if clients.Get(id) == nil && !ret[id] {
				ret[id] = true
				break
			}
		}
		runtime.Assert(err)
	}
	list := make([]string, 0, len(ret))
	for id := range ret {
		list = append(list, id)
	}
	ctx.OK(list)
}

package layout

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) status(clients *client.Clients, ctx *api.Context) {
	taskID := ctx.XStr("task_id")

	h.RLock()
	r := h.runners[taskID]
	h.RUnlock()

	if r == nil {
		ctx.NotFound("task")
		return
	}

	var ret struct {
		Done        bool     `json:"done"`
		Created     int64    `json:"created"`
		Finished    int64    `json:"finished"`
		TotalCnt    int      `json:"total_count"`
		FinishedCnt int      `json:"finished_count"`
		Nodes       []status `json:"nodes"`
	}

	ret.Done = r.done
	ret.Created = r.created
	ret.Finished = r.finished
	ret.TotalCnt = len(r.hosts)

	var finished int
	for _, id := range r.hosts {
		r.RLock()
		st := r.nodes[id]
		r.RUnlock()
		if st == nil {
			continue
		}
		if st.Finished > 0 {
			finished++
		}
		ret.Nodes = append(ret.Nodes, *st)
	}
	ret.FinishedCnt = finished

	ctx.OK(ret)
}

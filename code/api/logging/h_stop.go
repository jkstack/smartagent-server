package logging

import (
	"fmt"
	"net/http"
	"path/filepath"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) stop(clients *client.Clients, ctx *api.Context) {
	pid := ctx.XInt64("pid")

	h.RLock()
	t := h.data[pid]
	h.RUnlock()

	if t == nil {
		ctx.NotFound("project")
		return
	}

	if !t.Started {
		ctx.ERR(1, "project is not started")
		return
	}

	for _, cid := range t.Targets {
		cli := clients.Get(cid)
		if cli == nil {
			ctx.NotFound("agent")
			return
		}

		taskID, err := cli.SendLoggingStop(t.ID)
		runtime.Assert(err)
		defer cli.ChanClose(taskID)

		h.stTotalTasks.Inc()

		var msg *anet.Msg
		select {
		case msg = <-cli.ChanRead(taskID):
		case <-time.After(api.RequestTimeout):
			ctx.Timeout()
		}

		switch {
		case msg.Type == anet.TypeError:
			ctx.ERR(http.StatusServiceUnavailable, msg.ErrorMsg)
			return
		case msg.Type != anet.TypeLoggingStatusRep:
			ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", msg.Type))
			return
		}

		if !msg.LoggingStatusRep.OK {
			if msg.LoggingStatusRep.Msg != anet.LoggingNotRunningMsg {
				ctx.ERR(http.StatusInternalServerError, msg.LoggingStatusRep.Msg)
				return
			}
		}
	}

	t.Started = false
	dir := filepath.Join(h.cfg.DataDir, "logging", fmt.Sprintf("%d.json", t.ID))
	saveConfig(dir, *t)

	ctx.OK(nil)
}

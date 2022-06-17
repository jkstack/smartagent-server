package scaffolding

import (
	"fmt"
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) foo(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("scaffolding", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("scaffolding")
	}

	taskID, err := cli.SendFoo(p)
	runtime.Assert(err)

	defer cli.ChanClose(taskID)

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
	case msg.Type != anet.TypeBar:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", msg.Type))
		return
	}

	ctx.OK(nil)
}

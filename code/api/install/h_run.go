package install

import (
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) run(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	uri := ctx.OStr("uri", "")
	url := ctx.OStr("url", "")
	dir := ctx.OStr("dir", "")
	timeout := ctx.OInt("timeout", 3600)
	auth := ctx.OStr("auth", "")
	user := ctx.OStr("user", "")
	pass := ctx.OStr("pass", "")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("install", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("install")
	}

	if len(uri) == 0 && len(url) == 0 {
		ctx.ERR(http.StatusBadRequest, "missing url/uri param")
		return
	}

	taskID, err := cli.SendInstall(p, uri, url, dir, timeout, auth, user, pass)
	runtime.Assert(err)

	h.stUsage.Inc()

	info := &Info{updated: time.Now()}
	h.Lock()
	h.data[taskID] = info
	h.Unlock()
	go h.loop(cli, taskID)

	ctx.OK(taskID)
}

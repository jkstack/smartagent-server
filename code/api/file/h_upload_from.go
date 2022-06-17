package file

import (
	"fmt"
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

func (h *Handler) uploadFrom(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	dir := ctx.XStr("dir")
	name := ctx.XStr("name")
	uri := ctx.XStr("uri")
	auth := ctx.OStr("auth", "")
	user := ctx.OStr("user", "")
	pass := ctx.OStr("pass", "")
	mod := ctx.OInt("mod", 0644)
	ownUser := ctx.OStr("own_user", "")
	ownGroup := ctx.OStr("own_group", "")
	timeout := ctx.OInt("timeout", 60)

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("file", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("file")
	}

	taskID, err := cli.SendUpload(p, client.UploadContext{
		Dir:      dir,
		Name:     name,
		Auth:     auth,
		User:     user,
		Pass:     pass,
		Mod:      mod,
		OwnUser:  ownUser,
		OwnGroup: ownGroup,
		Md5Check: false,
		Uri:      uri,
	}, "")
	runtime.Assert(err)
	defer cli.ChanClose(taskID)

	h.stUsage.Inc()
	h.stTotalTasks.Inc()

	logging.Info("upload [%s] to %s on %s, task_id=%s, plugin.version=%s",
		name, dir, id, taskID, p.Version)

	var rep *anet.Msg
	select {
	case rep = <-cli.ChanRead(taskID):
	case <-time.After(time.Duration(timeout) * time.Second):
		ctx.Timeout()
	}

	switch {
	case rep.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, rep.ErrorMsg)
		return
	case rep.Type != anet.TypeUploadRep:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", rep.Type))
		return
	}

	if !rep.UploadRep.OK {
		ctx.ERR(1, rep.UploadRep.ErrMsg)
		return
	}

	ctx.OK(map[string]interface{}{
		"dir": rep.UploadRep.Dir,
	})
}

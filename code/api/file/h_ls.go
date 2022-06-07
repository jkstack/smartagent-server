package file

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

func (h *Handler) ls(clients *client.Clients, ctx *api.Context) {
	id := ctx.XStr("id")
	dir := ctx.XStr("dir")

	cli := clients.Get(id)
	if cli == nil {
		ctx.NotFound("client")
	}

	p := h.cfg.GetPlugin("file", cli.OS(), cli.Arch())
	if p == nil {
		lapi.PluginNotInstalledErr("file")
	}

	taskID, err := cli.SendLS(p, dir)
	runtime.Assert(err)
	defer cli.ChanClose(taskID)

	h.stUsage.Inc()

	var msg *anet.Msg
	select {
	case msg = <-cli.ChanRead(taskID):
	case <-time.After(lapi.RequestTimeout):
		ctx.Timeout()
	}

	switch {
	case msg.Type == anet.TypeError:
		ctx.ERR(http.StatusServiceUnavailable, msg.ErrorMsg)
		return
	case msg.Type != anet.TypeLsRep:
		ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("invalid message type: %d", msg.Type))
		return
	}

	if !msg.LSRep.OK {
		if msg.LSRep.ErrMsg == "directory not found" {
			ctx.NotFound("directory")
		}
		ctx.ERR(http.StatusInternalServerError, msg.LSRep.ErrMsg)
		return
	}

	type item struct {
		Name    string `json:"name"`
		Auth    uint32 `json:"auth"`
		User    string `json:"user"`
		Group   string `json:"group"`
		Size    uint64 `json:"size"`
		ModTime int64  `json:"mod_time"`
		IsDir   bool   `json:"is_dir"`
		IsLink  bool   `json:"is_link"`
	}
	items := make([]item, len(msg.LSRep.Files))
	for i, file := range msg.LSRep.Files {
		items[i] = item{
			Name:    file.Name,
			Auth:    uint32(file.Mod),
			User:    file.User,
			Group:   file.Group,
			Size:    file.Size,
			ModTime: file.ModTime.Unix(),
			IsDir:   file.Mod.IsDir(),
			IsLink:  file.IsLink,
		}
	}
	ctx.OK(map[string]interface{}{
		"dir":   msg.LSRep.Dir,
		"files": items,
	})
}

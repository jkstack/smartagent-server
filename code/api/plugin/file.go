package plugin

import (
	"server/code/client"
	"strings"

	"github.com/lwch/api"
	"github.com/lwch/logging"
)

func (h *Handler) serveFile(clients *client.Clients, ctx *api.Context) {
	uri := strings.TrimPrefix(ctx.URI(), "/file/plugin/")
	tmp := strings.SplitN(uri, "/", 2)
	logging.Info("download plugin %s, md5=%s", tmp[0], tmp[1])
	info := h.cfg.PluginByMD5(tmp[0], tmp[1])
	if info == nil {
		ctx.NotFound("plugin")
	}
	ctx.ServeFile(info.Dir)
}

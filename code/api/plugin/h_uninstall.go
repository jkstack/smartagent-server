package plugin

import (
	"os"
	"path"
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) uninstall(clients *client.Clients, ctx *api.Context) {
	name := ctx.XStr("name")
	version := ctx.XStr("version")

	runtime.Assert(os.RemoveAll(path.Join(h.cfg.PluginDir, name, version)))

	h.cfg.LoadPlugin()

	ctx.OK(version)
}

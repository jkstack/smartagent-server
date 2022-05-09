package server

import (
	"os"
	"path"
	"server/code/client"

	"github.com/lwch/api"
)

func created(dir string) int64 {
	fi, err := os.Stat(path.Join(dir, ".version"))
	if err != nil {
		return 0
	}
	return fi.ModTime().Unix()
}

func (h *Handler) info(clients *client.Clients, ctx *api.Context) {
	ctx.OK(map[string]interface{}{
		"clients": clients.Size(),
		"plugins": h.cfg.PluginCount(),
		"version": h.version,
		"created": created(h.cfg.WorkDir),
	})
}

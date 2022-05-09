package plugin

import (
	"os"
	"path"
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/logging"
)

func (h *Handler) list(clients *client.Clients, ctx *api.Context) {
	type item struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Created int64  `json:"created"`
	}
	var list []item
	h.cfg.RangePlugin(func(name, version string) {
		dir := path.Join(h.cfg.PluginDir, name, version)
		logging.Info("plugin.dir=%s", dir)
		fi, _ := os.Stat(dir)
		list = append(list, item{
			Name:    name,
			Version: version,
			Created: fi.ModTime().Unix(),
		})
	})
	ctx.OK(list)
}

package host

import (
	"server/code/client"

	"github.com/lwch/api"
)

func (h *Handler) list(clients *client.Clients, ctx *api.Context) {
	ids := make(map[string]bool)
	for _, id := range ctx.OCsv("ids", []string{}) {
		ids[id] = true
	}
	type cli struct {
		ID       string `json:"id"`
		IP       string `json:"ip"`
		MAC      string `json:"mac"`
		OS       string `json:"os"`
		Platform string `json:"platform"`
		Arch     string `json:"arch"`
		Version  string `json:"version"`
	}
	var list []cli
	clients.Range(func(c *client.Client) bool {
		if len(ids) > 0 && !ids[c.ID()] {
			return true
		}
		list = append(list, cli{
			ID:       c.ID(),
			IP:       c.IP(),
			MAC:      c.Mac(),
			OS:       c.OS(),
			Platform: c.Platform(),
			Arch:     c.Arch(),
			Version:  c.Version(),
		})
		return true
	})
	ctx.OK(list)
}

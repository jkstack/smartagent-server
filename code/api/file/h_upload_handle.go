package file

import (
	"server/code/client"
	"strings"

	"github.com/lwch/api"
)

func (h *Handler) uploadHandle(clients *client.Clients, ctx *api.Context) {
	id := strings.TrimPrefix(ctx.URI(), "/file/upload/")
	h.RLock()
	cache := h.uploadCache[id]
	h.RUnlock()
	if cache == nil {
		ctx.HTTPNotFound("file")
		return
	}
	if ctx.Token() != cache.token {
		ctx.HTTPForbidden("access denied")
		return
	}
	ctx.ServeFile(cache.dir)
	h.RemoveUploadCache(id)
}

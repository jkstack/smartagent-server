package agent

import (
	"net"
	"server/code/client"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) sniffer(clients *client.Clients, ctx *api.Context) {
	addr := ctx.XStr("addr")
	remote, err := net.ResolveTCPAddr("tcp", addr)
	runtime.Assert(err)
	c, err := net.DialTCP("tcp", nil, remote)
	runtime.Assert(err)
	defer c.Close()
	ctx.OK(c.LocalAddr().(*net.TCPAddr).IP.String())
}

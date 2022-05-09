package agent

import (
	"fmt"
	"server/code/client"
	"server/code/sshcli"
	"server/code/utils"

	"github.com/lwch/api"
)

func (h *Handler) stop(clients *client.Clients, ctx *api.Context) {
	addr := ctx.XStr("addr")
	user := ctx.XStr("user")
	pass := utils.DecryptPass(ctx.XStr("pass"))

	cli, err := sshcli.New(addr, user, pass)
	if err != nil {
		ctx.ERR(1, fmt.Sprintf("ssh连接失败：%v", err))
		return
	}

	str, err := service(cli, pass, "smartagent", "stop")
	if err != nil {
		ctx.ERR(2, fmt.Sprintf("停止失败：%s", str))
		return
	}

	ctx.OK(nil)

}

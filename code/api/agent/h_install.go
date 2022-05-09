package agent

import (
	"fmt"
	"io"
	"server/code/agent"
	"server/code/client"
	"server/code/sshcli"
	"server/code/utils"
	"strings"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) install(clients *client.Clients, ctx *api.Context) {
	addr := ctx.XStr("addr")
	user := ctx.XStr("user")
	pass := utils.DecryptPass(ctx.XStr("pass"))
	f, _, err := ctx.File("file")
	runtime.Assert(err)
	defer f.Close()

	cli, err := sshcli.New(addr, user, pass)
	if err != nil {
		ctx.ERR(1, fmt.Sprintf("ssh连接失败：%v", err))
		return
	}
	defer cli.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		ctx.ERR(3, fmt.Sprintf("文件上传失败：%v", err))
		return
	}
	err = agent.Extract(cli, pass, data)
	switch {
	case err == nil:
		ctx.OK(nil)
	case err.Error() == "无法重复安装":
		ctx.ERR(2, err.Error())
	case strings.Contains(err.Error(), "文件上传失败："):
		ctx.ERR(3, err.Error())
	default:
		ctx.ERR(4, err.Error())
	}
}

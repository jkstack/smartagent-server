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

func (h *Handler) upgrade(clients *client.Clients, ctx *api.Context) {
	addr := ctx.XStr("addr")
	user := ctx.XStr("user")
	pass := utils.DecryptPass(ctx.XStr("pass"))
	restart := ctx.XBool("restart")
	f, _, err := ctx.File("file")
	runtime.Assert(err)
	defer f.Close()

	cli, err := sshcli.New(addr, user, pass)
	runtime.Assert(err)
	defer cli.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		ctx.ERR(1, fmt.Sprintf("文件上传失败：%v", err))
		return
	}
	version, err := agent.Upgrade(cli, pass, data, restart)
	switch {
	case err == nil:
		ctx.OK(version)
	case strings.Contains(err.Error(), "文件上传失败："):
		ctx.ERR(1, err.Error())
	case strings.Contains(err.Error(), "安装包解压失败："):
		ctx.ERR(2, err.Error())
	case strings.Contains(err.Error(), "生成配置文件失败："),
		strings.Contains(err.Error(), "移动配置文件失败："):
		ctx.ERR(3, err.Error())
	case strings.Contains(err.Error(), "修改安装目录失败："),
		strings.Contains(err.Error(), "修改启动脚本中的安装目录失败："):
		ctx.ERR(4, err.Error())
	default:
		ctx.ERR(5, err.Error())
	}
}

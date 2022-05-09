package agent

import (
	"server/code/agent"
	"server/code/client"
	"server/code/sshcli"
	"server/code/utils"
	"strings"

	"github.com/lwch/api"
	"github.com/lwch/runtime"
)

func (h *Handler) config(clients *client.Clients, ctx *api.Context) {
	addr := ctx.XStr("addr")
	user := ctx.XStr("user")
	pass := ctx.XStr("pass")

	var cfg agent.Config
	cfg.ID = ctx.XStr("id")
	cfg.Server = ctx.XStr("server")
	cfg.User = ctx.OStr("owner", "nobody")
	cfg.PluginDir = ctx.OStr("plugin_dir", agent.DefaultInstallDir+"/plugin")
	cfg.LogDir = ctx.OStr("log_dir", agent.DefaultInstallDir+"/logs")
	cfg.LogSize = utils.Bytes(ctx.OInt("log_size", 50*1024*1024))
	cfg.LogRotate = ctx.OInt("log_rotate", 7)
	cfg.CPU = uint32(ctx.OInt("cpu_limit", 1))
	cfg.Memory = utils.Bytes(ctx.OInt("memory_limit", 100*1024*1024))

	cli, err := sshcli.New(addr, user, pass)
	runtime.Assert(err)
	defer cli.Close()

	content, err := agent.Install(cli, pass, cfg)

	switch {
	case err == nil:
		ctx.OK(content)
	case strings.Contains(err.Error(), "生成配置文件失败："),
		strings.Contains(err.Error(), "移动配置文件失败："):
		ctx.ERR(1, err.Error())
	case strings.Contains(err.Error(), "目录创建失败："),
		strings.Contains(err.Error(), "目录属主修改失败："):
		ctx.ERR(2, err.Error())
	case strings.Contains(err.Error(), "启动失败："):
		ctx.ERR(3, err.Error())
	}
}

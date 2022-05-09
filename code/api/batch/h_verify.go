package batch

import (
	"context"
	"os"
	"server/code/client"
	"server/code/sshcli"
	"server/code/utils"
	"sync"

	"github.com/lwch/api"
)

func (h *Handler) verify(clients *client.Clients, ctx *api.Context) {
	action := ctx.XStr("action")
	addr := ctx.XCsv("addr")
	user := ctx.XCsv("user")
	pass := ctx.XCsv("pass")

	want := true
	if action == "install" {
		want = false
	}

	var wg sync.WaitGroup
	errors := make([]string, len(addr))
	wg.Add(len(addr))
	for i := range addr {
		go func(i int) {
			defer wg.Done()
			pass := utils.DecryptPass(pass[i])
			cli, err := sshcli.New(addr[i], user[i], pass)
			if err != nil {
				errors[i] = "ssh连接失败"
				return
			}
			defer cli.Close()
			cli.Do(context.Background(), "sudo -k", "")
			_, err = cli.Do(context.Background(), "sudo id", pass+"\n")
			if err != nil {
				errors[i] = "没有管理员权限"
				return
			}
			_, err = cli.Stat("/opt/smartagent/.success")
			installed := !os.IsNotExist(err)
			if want && !installed {
				errors[i] = "未安装"
				return
			}
			if !want && installed {
				errors[i] = "已安装"
				return
			}
			if diskLeft(cli, "/opt/smartagent") <= 100*1024*1024 {
				errors[i] = "磁盘空间不足"
				return
			}
			if memLeft(cli) <= 100*1024*1024 {
				errors[i] = "内存不足"
				return
			}
		}(i)
	}

	wg.Wait()
	type item struct {
		Addr string `json:"addr"`
		Msg  string `json:"msg"`
	}
	ret := make([]item, 0, len(errors))
	for i, err := range errors {
		if len(err) == 0 {
			continue
		}
		ret = append(ret, item{
			Addr: addr[i],
			Msg:  err,
		})
	}
	ctx.OK(ret)
}

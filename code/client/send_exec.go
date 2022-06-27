package client

import (
	"encoding/base64"
	"server/code/conf"
	"server/code/utils"

	"github.com/jkstack/anet"
)

func (cli *Client) SendExec(p *conf.PluginInfo, cmd string, args []string, timeout int,
	auth, user, pass, workdir string, env []string, deferRM string) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	if len(pass) > 0 {
		enc, err := anet.Encrypt([]byte(pass))
		if err != nil {
			return "", err
		}
		pass = "$1$" + base64.StdEncoding.EncodeToString(enc)
	}
	var msg anet.Msg
	msg.Type = anet.TypeExec
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.Exec = &anet.ExecPayload{
		Cmd:     cmd,
		Args:    args,
		Timeout: timeout,
		Auth:    auth,
		User:    user,
		Pass:    pass,
		WorkDir: workdir,
		Env:     env,
		DeferRM: deferRM,
	}
	ch := make(chan *anet.Msg, 10)
	cli.Lock()
	cli.taskRead[id] = ch
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}

func (cli *Client) SendKill(p *conf.PluginInfo, pid int) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeExecKill
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.ExecKill = &anet.ExecKill{
		Pid: pid,
	}
	cli.chWrite <- &msg
	return id, nil
}

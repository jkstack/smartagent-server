package client

import (
	"server/code/conf"
	"server/code/utils"

	"github.com/jkstack/anet"
)

func (cli *Client) SendInstall(p *conf.PluginInfo, uri, url, dir string,
	timeout int, auth, user, pass string) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeInstallArgs
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	msg.InstallArgs = &anet.InstallArgs{
		URI:     uri,
		URL:     url,
		Dir:     dir,
		Timeout: timeout,
		Auth:    auth,
		User:    user,
		Pass:    pass,
	}
	cli.Lock()
	cli.taskRead[id] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}

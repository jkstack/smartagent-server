package client

import (
	"server/code/conf"
	"server/code/utils"

	"github.com/jkstack/anet"
)

func (cli *Client) SendHMStatic(p *conf.PluginInfo) (string, error) {
	id, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeHMStaticReq
	msg.TaskID = id
	msg.Plugin = fillPlugin(p)
	cli.Lock()
	cli.taskRead[id] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return id, nil
}

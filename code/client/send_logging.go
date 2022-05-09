package client

import (
	"server/code/utils"

	"github.com/jkstack/anet"
)

func (cli *Client) SendLoggingConfigK8s(pid int64, ext string, batch, buffer, interval int, report string,
	ns string, names []string, dir, api, token string) (string, error) {
	taskID, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeLoggingConfig
	msg.TaskID = taskID
	msg.LoggingConfig = &anet.LoggingConfig{
		Pid:      pid,
		T:        anet.LoggingTypeK8s,
		Exclude:  ext,
		Batch:    batch,
		Buffer:   buffer,
		Interval: interval,
		Report:   report,
		K8s: &anet.LoggingConfigK8s{
			Namespace: ns,
			Names:     names,
			Dir:       dir,
			Api:       api,
			Token:     token,
		},
	}
	cli.chWrite <- &msg
	return taskID, nil
}

func (cli *Client) SendLoggingConfigFile(pid int64, ext string, batch, buffer, interval int, report string,
	dir string) (string, error) {
	taskID, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeLoggingConfig
	msg.TaskID = taskID
	msg.LoggingConfig = &anet.LoggingConfig{
		Pid:      pid,
		T:        anet.LoggingTypeFile,
		Exclude:  ext,
		Batch:    batch,
		Buffer:   buffer,
		Interval: interval,
		Report:   report,
		File: &anet.LoggingConfigFile{
			Dir: dir,
		},
	}
	cli.chWrite <- &msg
	return taskID, nil
}

func (cli *Client) SendLoggingStart(id int64) (string, error) {
	taskID, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeLoggingStatusReq
	msg.TaskID = taskID
	msg.LoggingStatusReq = &anet.LoggingStatusReq{
		ID:      id,
		Running: true,
	}
	cli.Lock()
	cli.taskRead[taskID] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return taskID, nil
}

func (cli *Client) SendLoggingStop(id int64) (string, error) {
	taskID, err := utils.TaskID()
	if err != nil {
		return "", err
	}
	var msg anet.Msg
	msg.Type = anet.TypeLoggingStatusReq
	msg.TaskID = taskID
	msg.LoggingStatusReq = &anet.LoggingStatusReq{
		ID:      id,
		Running: false,
	}
	cli.Lock()
	cli.taskRead[taskID] = make(chan *anet.Msg)
	cli.Unlock()
	cli.chWrite <- &msg
	return taskID, nil
}

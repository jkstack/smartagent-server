package layout

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"server/code/client"
	"strings"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

type execHandler struct {
	taskInfo
	cmd    string
	output string
}

func (h *execHandler) Check(t Task) error {
	if len(t.Cmd) == 0 {
		return fmt.Errorf("missing cmd for [%s] task", t.Name)
	}
	return nil
}

func (h *execHandler) Clone(t Task, info taskInfo) taskHandler {
	return &execHandler{
		taskInfo: info,
		cmd:      t.Cmd,
		output:   t.Output,
	}
}

func (h *execHandler) Run(id, dir, user, pass string, args map[string]string) error {
	deadline := h.deadline()
	args["DEADLINE"] = fmt.Sprintf("%d", deadline.Unix())
	logging.Info("exec [%s] on agent [%s]", h.name, id)
	cli := h.parent.GetClient(id)
	if cli == nil {
		return errClientNotfound(id)
	}
	p := h.parent.GetPlugin("exec", cli.OS(), cli.Arch())
	if p == nil {
		return errPluginNotInstalled("exec")
	}
	timeout := int(deadline.Sub(time.Now()).Seconds() + .5)
	taskID, err := cli.SendExec(p, h.cmd, nil, timeout, h.auth,
		h.parent.user, h.parent.pass, "", nil)
	if err != nil {
		return fmt.Errorf("send exec [%s] on agent [%s]: %v", h.name, id, err)
	}
	defer cli.ChanClose(taskID)

	h.parent.parent.stExecUsage.Inc()
	h.parent.parent.stTotalTasks.Inc()

	return h.read(cli, h.name, taskID, time.After(time.Duration(timeout)*time.Second), args)
}

func (h *execHandler) read(cli *client.Client,
	taskName, taskID string,
	deadline <-chan time.Time, args map[string]string) error {
	pid, err := h.readAck(cli, taskID, deadline)
	if err != nil {
		return err
	}
	logging.Info("exec [%s] on agent [%s] successed, pid=%d", taskName, cli.ID(), pid)
	return h.readBody(cli, taskName, taskID, deadline, args)
}

func (h *execHandler) readAck(cli *client.Client,
	taskID string, deadline <-chan time.Time) (int, error) {
	var msg *anet.Msg
	select {
	case msg = <-cli.ChanRead(taskID):
	case <-deadline:
		return 0, errTimeout
	}
	switch {
	case msg.Type == anet.TypeError:
		return 0, errors.New(msg.ErrorMsg)
	case msg.Type != anet.TypeExecd:
		return 0, fmt.Errorf("invalid message type: %d", msg.Type)
	}
	if !msg.Execd.OK {
		return 0, errors.New(msg.Execd.Msg)
	}
	return msg.Execd.Pid, nil
}

func (h *execHandler) readBody(cli *client.Client,
	taskName, taskID string,
	deadline <-chan time.Time, args map[string]string) error {
	ch := cli.ChanRead(taskID)
	var cache bytes.Buffer
	for {
		var msg *anet.Msg
		select {
		case msg = <-ch:
		case <-deadline:
			return errTimeout
		}
		if msg == nil {
			return nil
		}
		switch msg.Type {
		case anet.TypeError:
			return errors.New(msg.ErrorMsg)
		case anet.TypeExecData:
			data, _ := base64.StdEncoding.DecodeString(msg.ExecData.Data)
			logging.Debug("exec [%s] on agent [%s] recved %d bytes", taskName, cli.ID(), len(data))
			_, err := cache.Write(data)
			if err != nil {
				return fmt.Errorf("write cache: %v", err)
			}
		case anet.TypeExecDone:
			logging.Info("exec [%s] on agent [%s] run done, code=%d",
				taskName, cli.ID(), msg.ExecDone.Code)
			if msg.ExecDone.Code != 0 {
				logging.Info("exec [%s] on agent [%s] failed: %s",
					taskName, cli.ID(), cache.String())
				return fmt.Errorf("code=%d", msg.ExecDone.Code)
			}
			if len(h.output) > 0 {
				str := cache.String()
				str = strings.TrimLeft(str, "Password: \r\n")
				if strings.HasPrefix(str, "[sudo]") {
					tmp := strings.SplitN(str, "\n", 2)
					str = tmp[1]
				}
				str = strings.TrimRight(str, "\r\n")
				args[h.output] = str
				if len(args[h.output]) < 1024 {
					logging.Info("exec [%s] on agent [%s] log variable [%s] value:\n%s",
						taskName, cli.ID(), h.output, args[h.output])
				}
			}
			return nil
		}
	}
}

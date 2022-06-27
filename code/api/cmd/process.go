package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"server/code/conf"
	"time"

	"github.com/jkstack/anet"
	"github.com/lwch/l2cache"
	"github.com/lwch/logging"
)

const memoryCacheSize = 102400

var callbackCli = &http.Client{Timeout: time.Minute}

type process struct {
	parent   *cmdClient
	id       int
	cmd      string
	callback string
	created  time.Time
	updated  time.Time
	taskID   string
	running  bool
	code     int
	cache    *l2cache.Cache
	ctx      context.Context
	cancel   context.CancelFunc
}

func newProcess(parent *cmdClient, cfg *conf.Configure, id int, cmd, taskID, callback string) (*process, error) {
	cache, err := l2cache.New(memoryCacheSize, path.Join(cfg.CacheDir, "cmd"))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &process{
		parent:   parent,
		id:       id,
		cmd:      cmd,
		callback: callback,
		created:  time.Now(),
		updated:  time.Now(),
		taskID:   taskID,
		running:  true,
		code:     -65535,
		cache:    cache,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (p *process) recv() {
	defer p.cancel()
	defer p.cb()
	cli := p.parent.cli
	ch := cli.ChanRead(p.taskID)
	defer cli.ChanClose(p.taskID)
	for {
		var msg *anet.Msg
		select {
		case msg = <-ch:
		case <-p.ctx.Done():
			return
		}
		if msg == nil {
			return
		}
		p.updated = time.Now()
		switch msg.Type {
		case anet.TypeError:
			p.running = false
			p.code = -255
			p.cache.Write([]byte(msg.ErrorMsg))
			logging.Error("task %s on [%s] run failed: %s", p.taskID, cli.ID(), msg.ErrorMsg)
			return
		case anet.TypeExecData:
			data, _ := base64.StdEncoding.DecodeString(msg.ExecData.Data)
			p.cache.Write(data)
			logging.Debug("task %s on [%s] recved %d bytes", p.taskID, cli.ID(), len(data))
		case anet.TypeExecDone:
			p.running = false
			p.code = msg.ExecDone.Code
			logging.Info("task %s on [%s] run done, code=%d", p.taskID, cli.ID(), msg.ExecDone.Code)
			return
		}
	}
}

func (p *process) close() {
	p.cache.Close()
	p.cancel()
}

func (p *process) read() ([]byte, error) {
	return ioutil.ReadAll(p.cache)
}

func (p *process) sendKill(plugin *conf.PluginInfo) string {
	id, _ := p.parent.cli.SendKill(plugin, p.id)
	return id
}

func (p *process) wait() {
	<-p.ctx.Done()
}

func (p *process) cb() {
	if len(p.callback) == 0 {
		return
	}
	u, err := url.Parse(p.callback)
	if err != nil {
		logging.Error("parse for callback url [%s]: %v", p.callback, err)
		return
	}
	args := u.Query()
	args.Set("agent_id", p.parent.cli.ID())
	args.Set("pid", fmt.Sprintf("%d", p.id))
	u.RawQuery = args.Encode()
	resp, err := callbackCli.Get(u.String())
	if err != nil {
		logging.Error("callback to [%s]: %v", p.callback, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		logging.Error("callback to [%s] is not http200: %d\n%s", resp.StatusCode, string(data))
		return
	}
}

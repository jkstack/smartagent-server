package client

import (
	"context"
	"fmt"
	"server/code/conf"
	"server/code/utils"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

const channelBuffer = 10000

type Client struct {
	sync.RWMutex
	info   anet.ComePayload
	remote *websocket.Conn
	// runtime
	chRead   chan *anet.Msg
	chWrite  chan *anet.Msg
	taskRead map[string]chan *anet.Msg
}

func (cli *Client) close() {
	logging.Info("client %s connection closed", cli.info.ID)
	if cli.remote != nil {
		cli.remote.Close()
	}
	close(cli.chRead)
	close(cli.chWrite)
}

func (cli *Client) remoteAddr() string {
	return cli.remote.RemoteAddr().String()
}

func (cli *Client) read(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		utils.Recover(fmt.Sprintf("cli.read(%s)", cli.remoteAddr()))
		cancel()
	}()
	cli.remote.SetReadDeadline(time.Time{})
	send := func(taskID string, ch chan *anet.Msg, msg *anet.Msg) {
		defer func() {
			if err := recover(); err != nil {
				logging.Error("write to channel %s timeout", taskID)
			}
		}()
		select {
		case <-ctx.Done():
			return
		case ch <- msg:
		case <-time.After(10 * time.Second):
			return
		}
	}
	for {
		var msg anet.Msg
		err := cli.remote.ReadJSON(&msg)
		if err != nil {
			logging.Error("cli.read(%s): %v", cli.remoteAddr(), err)
			return
		}
		ch := cli.chRead
		if len(msg.TaskID) > 0 {
			cli.RLock()
			ch = cli.taskRead[msg.TaskID]
			cli.RUnlock()
			if ch == nil {
				// logging.Error("response channel %s not found", msg.TaskID)
				continue
			}
		}
		send(msg.TaskID, ch, &msg)
	}
}

func (cli *Client) write(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		utils.Recover(fmt.Sprintf("cli.write(%s)", cli.remoteAddr()))
		cancel()
	}()
	send := func(msg *anet.Msg, i int) bool {
		err := cli.remote.WriteJSON(msg)
		if err != nil {
			logging.Error("cli.write(%s) %d times: %v", cli.remoteAddr(), i, err)
			return false
		}
		return true
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-cli.chWrite:
			if msg == nil {
				return
			}
			if msg.Important {
				for i := 0; i < 10; i++ {
					if send(msg, i+1) {
						break
					}
				}
				continue
			}
			send(msg, 1)
		}
	}
}

func (cli *Client) ID() string {
	return cli.info.ID
}

func (cli *Client) OS() string {
	return cli.info.OS
}

func (cli *Client) Platform() string {
	return cli.info.Platform
}

func (cli *Client) Arch() string {
	return cli.info.Arch
}

func (cli *Client) Version() string {
	return cli.info.Version
}

func (cli *Client) IP() string {
	return cli.info.IP.String()
}

func (cli *Client) Mac() string {
	return cli.info.MAC
}

func (cli *Client) HostName() string {
	return cli.info.HostName
}

func (cli *Client) ChanRead(id string) <-chan *anet.Msg {
	cli.RLock()
	defer cli.RUnlock()
	return cli.taskRead[id]
}

func (cli *Client) ChanClose(id string) {
	cli.Lock()
	defer cli.Unlock()
	if ch := cli.taskRead[id]; ch != nil {
		close(ch)
		delete(cli.taskRead, id)
	}
}

func fillPlugin(p *conf.PluginInfo) *anet.PluginInfo {
	return &anet.PluginInfo{
		Name:    p.Name,
		Version: p.Version.String(),
		MD5:     p.MD5,
		URI:     p.URI,
	}
}

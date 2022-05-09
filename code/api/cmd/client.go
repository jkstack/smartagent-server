package cmd

import (
	"server/code/client"
	"sync"
	"time"
)

type cmdClient struct {
	sync.RWMutex
	cli     *client.Client
	process map[int]*process // pid => process
}

func newClient(c *client.Client) *cmdClient {
	cli := &cmdClient{
		cli:     c,
		process: make(map[int]*process),
	}
	go cli.clear()
	return cli
}

func (cli *cmdClient) clear() {
	run := func() {
		var list []*process
		cli.Lock()
		for _, p := range cli.process {
			if time.Since(p.updated) > clearTimeout {
				list = append(list, p)
			}
		}
		cli.Unlock()

		for _, p := range list {
			p.close()
			cli.Lock()
			delete(cli.process, p.id)
			cli.Unlock()
		}
	}
	for {
		run()
		time.Sleep(time.Minute)
	}
}

func (cli *cmdClient) addProcess(p *process) {
	cli.Lock()
	cli.process[p.id] = p
	cli.Unlock()
}

func (cli *cmdClient) close() {
	list := make([]*process, 0, len(cli.process))
	cli.RLock()
	for _, p := range cli.process {
		list = append(list, p)
	}
	cli.RUnlock()
	for _, p := range list {
		p.close()
	}
	cli.Lock()
	cli.process = make(map[int]*process)
	cli.Unlock()
}

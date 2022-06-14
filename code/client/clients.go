package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jkstack/anet"
	"github.com/jkstack/jkframe/stat"
	"github.com/lwch/logging"
)

// Clients clients
type Clients struct {
	sync.RWMutex
	data         map[string]*Client
	stInPackets  *stat.Counter
	stOutPackets *stat.Counter
	stInBytes    *stat.Counter
	stOutBytes   *stat.Counter
}

func NewClients(stats *stat.Mgr) *Clients {
	clients := &Clients{
		data:         make(map[string]*Client),
		stInPackets:  stats.NewCounter("in_packets"),
		stOutPackets: stats.NewCounter("out_packets"),
		stInBytes:    stats.NewCounter("in_bytes"),
		stOutBytes:   stats.NewCounter("out_bytes"),
	}
	go clients.print()
	return clients
}

// New new client
func (cs *Clients) New(conn *websocket.Conn, come *anet.ComePayload, cancel context.CancelFunc) *Client {
	t := "smart"
	if come.Name == "godagent" {
		t = "god"
	}
	cli := &Client{
		t:        t,
		parent:   cs,
		info:     *come,
		remote:   conn,
		chRead:   make(chan *anet.Msg, channelBuffer),
		chWrite:  make(chan *anet.Msg, channelBuffer),
		taskRead: make(map[string]chan *anet.Msg, channelBuffer/10),
	}
	ctx, cancel := context.WithCancel(context.Background())
	go cli.read(ctx, cancel)
	go cli.write(ctx, cancel)
	go func() {
		<-ctx.Done()

		cli.close()
		cs.Lock()
		delete(cs.data, come.ID)
		cs.Unlock()

		cancel()
	}()
	return cli
}

// Add add client
func (cs *Clients) Add(cli *Client) {
	cs.Lock()
	defer cs.Unlock()
	if old := cs.data[cli.info.ID]; old != nil {
		old.close()
	}
	cs.data[cli.info.ID] = cli
}

// Get get client by id
func (cs *Clients) Get(id string) *Client {
	cs.RLock()
	defer cs.RUnlock()
	return cs.data[id]
}

// Range list clients
func (cs *Clients) Range(cb func(*Client) bool) {
	cs.RLock()
	defer cs.RUnlock()
	for _, c := range cs.data {
		next := cb(c)
		if !next {
			return
		}
	}
}

// Size get clients count
func (cs *Clients) Size() int {
	return len(cs.data)
}

func (cs *Clients) print() {
	var logs []string
	cs.RLock()
	for _, cli := range cs.data {
		if len(cli.chWrite) > 0 || len(cli.chRead) > 0 {
			logs = append(logs, fmt.Sprintf("client %s: write chan=%d, read chan=%d",
				cli.ID(), len(cli.chWrite), len(cli.chRead)))
		}
	}
	cs.RUnlock()
	logging.Info(strings.Join(logs, "\n"))
}

func (cs *Clients) Prefix(str string) []*Client {
	var ret []*Client
	cs.RLock()
	for id, cli := range cs.data {
		if strings.HasPrefix(id, str) {
			ret = append(ret, cli)
		}
	}
	cs.RUnlock()
	return ret
}

func (cs *Clients) Contains(str string) []*Client {
	var ret []*Client
	cs.RLock()
	for id, cli := range cs.data {
		if strings.Contains(id, str) {
			ret = append(ret, cli)
		}
	}
	cs.RUnlock()
	return ret
}

func (cs *Clients) All() []*Client {
	var ret []*Client
	cs.RLock()
	for _, cli := range cs.data {
		ret = append(ret, cli)
	}
	cs.RUnlock()
	return ret
}

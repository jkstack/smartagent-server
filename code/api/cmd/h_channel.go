package cmd

import (
	"io"
	"net/http"
	"server/code/client"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
}

func (h *Handler) channel(clients *client.Clients, ctx *api.Context) {
	uri := strings.TrimPrefix(ctx.URI(), "/cmd/channel/")

	tmp := strings.SplitN(uri, "/", 2)
	cli := h.cliFrom(tmp[0])
	if cli == nil {
		ctx.HTTPNotFound("client")
		return
	}

	pid, err := strconv.ParseInt(tmp[1], 10, 64)
	runtime.Assert(err)

	cli.RLock()
	p := cli.process[int(pid)]
	cli.RUnlock()

	if p == nil {
		ctx.HTTPNotFound("process")
		return
	}

	var conn *websocket.Conn

	ctx.RawCallback(func(w http.ResponseWriter, r *http.Request) {
		var err error
		conn, err = upgrader.Upgrade(w, r, nil)
		runtime.Assert(err)
	})

	for {
		data, err := p.read()
		if err != nil {
			if err == io.EOF {
				return
			}
			logging.Error("chan read: %v", err)
			panic(err)
		}
		if len(data) == 0 {
			time.Sleep(time.Second)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			logging.Error("write message: %v", err)
			panic(err)
		}
	}
}

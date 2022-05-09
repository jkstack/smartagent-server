package batch

import (
	"server/code/client"

	"github.com/gorilla/websocket"
	"github.com/lwch/api"
)

var upgrader = websocket.Upgrader{EnableCompression: true}

func (h *Handler) commit(clients *client.Clients, ctx *api.Context) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		logging.Error("batch commit: %v", err)
	//	}
	//}()
	//conn, err := ctx.Upgrade()
	//if err != nil {
	//	ctx.HTTPError(http.StatusInternalServerError, err.Error())
	//	return
	//}
	//defer conn.Close()
	//
	//_, data, err := conn.ReadMessage()
	//runtime.Assert(err)
	//var req bdata.Request
	//runtime.Assert(proto.Unmarshal(data, &req))
	//switch req.GetAct() {
	//case bdata.Request_install:
	//	h.handleInstall(conn, req)
	//case bdata.Request_upgrade:
	//	h.handleUpgrade(conn, req)
	//case bdata.Request_uninstall:
	//	h.handleUninstall(conn, req)
	//}
}

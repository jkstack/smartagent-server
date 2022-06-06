package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server/code/client"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jkstack/anet"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
}

func (app *App) agent(w http.ResponseWriter, r *http.Request,
	onConnect chan *client.Client,
	onClose chan string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Error("upgrade websocket: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	come, err := app.waitCome(conn)
	if err != nil {
		logging.Error("wait come message(%s): %v", conn.RemoteAddr().String(), err)
		return
	}
	if app.handshake(conn, come) {
		app.stAgentCount.Inc()
		logging.Info("client %s connection on, os=%s, arch=%s, mac=%s",
			come.ID, come.OS, come.Arch, come.MAC)
		cli := app.clients.New(conn, come, onClose)
		app.clients.Add(cli)
		onConnect <- cli
	}
}

func (app *App) waitCome(conn *websocket.Conn) (*anet.ComePayload, error) {
	conn.SetReadDeadline(time.Now().Add(time.Minute))
	var msg anet.Msg
	err := conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	if msg.Type != anet.TypeCome {
		return nil, errors.New("invalid come message")
	}
	return msg.Come, nil
}

func (app *App) handshake(conn *websocket.Conn, come *anet.ComePayload) (ok bool) {
	var errMsg string
	defer func() {
		var rep anet.Msg
		rep.Type = anet.TypeHandshake
		rep.Important = true
		if ok {
			rep.Handshake = &anet.HandshakePayload{
				OK: true,
				ID: come.ID,
				// TODO: redirect
			}
		} else {
			rep.Handshake = &anet.HandshakePayload{
				OK:  false,
				Msg: errMsg,
			}
		}
		data, err := json.Marshal(rep)
		if err != nil {
			logging.Error("build handshake message: %v", err)
			return
		}
		conn.WriteMessage(websocket.TextMessage, data)
	}()
	app.connectLock.Lock()
	defer app.connectLock.Unlock()
	if len(come.ID) == 0 {
		id, err := runtime.UUID(16, "0123456789abcdef")
		if err != nil {
			errMsg = fmt.Sprintf("generate agent id: %v", err)
			logging.Error(errMsg)
			return false
		}
		come.ID = fmt.Sprintf("agent-%s-%s", time.Now().Format("20060102"), id)
	} else if app.clients.Get(come.ID) != nil {
		errMsg = "agent id conflict"
		return false
	}
	return true
}

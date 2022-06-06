package app

import (
	"fmt"
	"net/http"
	lapi "server/code/api"
	"server/code/client"
	"strings"

	"github.com/lwch/api"
	"github.com/lwch/logging"
)

func (app *App) reg(uri string, cb func(*client.Clients, *api.Context)) {
	http.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if app.blocked {
			http.Error(w, "raise limit", http.StatusServiceUnavailable)
			return
		}
		statName := strings.ReplaceAll(uri, "/", "_")
		counter := app.stats.NewCounter("api_counter" + statName)
		counter.Inc()
		tick := app.stats.NewTick("api_pref" + statName)
		defer tick.Close()
		ctx := api.NewContext(w, r)
		defer func() {
			if err := recover(); err != nil {
				switch err := err.(type) {
				case api.MissingParam:
					ctx.ERR(http.StatusBadRequest, err.Error())
				case api.BadParam:
					ctx.ERR(http.StatusBadRequest, err.Error())
				case api.NotFound:
					ctx.ERR(http.StatusNotFound, err.Error())
				case api.Timeout:
					ctx.ERR(http.StatusBadGateway, err.Error())
				case lapi.PluginNotInstalled:
					ctx.ERR(http.StatusNotAcceptable, err.Error())
				case lapi.Notfound:
					ctx.ERR(http.StatusNotFound, err.Error())
				default:
					ctx.ERR(http.StatusInternalServerError, fmt.Sprintf("%v", err))
					logging.Error("err: %v", err)
				}
			}
		}()
		cb(app.clients, ctx)
	})
}

package app

import (
	"fmt"
	"net/http"
	"os"
	"server/code/api/agent"
	"server/code/api/cmd"
	"server/code/api/file"
	"server/code/api/hm"
	"server/code/api/host"
	"server/code/api/install"
	"server/code/api/layout"
	apilogging "server/code/api/logging"
	"server/code/api/plugin"
	"server/code/api/server"
	"server/code/client"
	"server/code/conf"
	"server/code/utils"
	"sync"
	"time"

	"github.com/jkstack/jkframe/stat"
	"github.com/kardianos/service"
	"github.com/lwch/api"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"github.com/shirou/gopsutil/v3/disk"
)

type handler interface {
	Init(*conf.Configure)
	HandleFuncs() map[string]func(*client.Clients, *api.Context)
	OnConnect(*client.Client)
	OnClose(string)
}

// App app
type App struct {
	cfg         *conf.Configure
	clients     *client.Clients
	version     string
	blocked     bool
	connectLock sync.Mutex
	stats       *stat.Mgr
}

// New new app
func New(cfg *conf.Configure, version string) *App {
	app := &App{
		cfg:     cfg,
		clients: client.NewClients(),
		version: version,
		blocked: false,
		stats:   stat.New(5 * time.Second),
	}
	go app.limit()
	return app
}

// Start start app
func (app *App) Start(s service.Service) error {
	go func() {
		logging.SetSizeRotate(logging.SizeRotateConfig{
			Dir:         app.cfg.LogDir,
			Name:        "smartagent-server",
			Size:        int64(app.cfg.LogSize.Bytes()),
			Rotate:      app.cfg.LogRotate,
			WriteStdout: true,
			WriteFile:   true,
		})
		defer logging.Flush()

		defer utils.Recover("service")

		os.RemoveAll(app.cfg.CacheDir)

		var mods []handler
		mods = append(mods, plugin.New())
		mods = append(mods, cmd.New())
		fh := file.New()
		mods = append(mods, fh)
		mods = append(mods, hm.New())
		mods = append(mods, host.New())
		mods = append(mods, server.New(app.version))
		mods = append(mods, agent.New())
		mods = append(mods, install.New())
		mods = append(mods, layout.New(fh))
		mods = append(mods, apilogging.New())

		for _, mod := range mods {
			mod.Init(app.cfg)
			for uri, cb := range mod.HandleFuncs() {
				app.reg(uri, cb)
			}
		}

		http.HandleFunc("/metrics", app.stats.ServeHTTP)
		http.HandleFunc("/ws/agent", func(w http.ResponseWriter, r *http.Request) {
			onConnect := make(chan *client.Client)
			onClose := make(chan string)
			go func() {
				cli := <-onConnect
				for _, mod := range mods {
					mod.OnConnect(cli)
				}
			}()
			app.agent(w, r, onConnect, onClose)
			id := <-onClose
			logging.Info("client %s connection closed", id)
			for _, mod := range mods {
				mod.OnClose(id)
			}
		})

		logging.Info("http listen on %d", app.cfg.Listen)
		runtime.Assert(http.ListenAndServe(fmt.Sprintf(":%d", app.cfg.Listen), nil))
	}()
	return nil
}

func (app *App) Stop(s service.Service) error {
	return nil
}

func (app *App) limit() {
	for {
		usage, err := disk.Usage(app.cfg.CacheDir)
		if err == nil {
			if usage.UsedPercent > float64(app.cfg.CacheThreshold) {
				app.blocked = true
			} else {
				app.blocked = false
			}
		}
		time.Sleep(time.Second)
	}
}

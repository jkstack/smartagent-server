package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	rt "runtime"
	"server/code/app"
	"server/code/conf"

	_ "net/http/pprof"

	"github.com/kardianos/service"
	"github.com/lwch/runtime"
)

var (
	version      string = "0.0.0"
	gitBranch    string = "<branch>"
	gitHash      string = "<hash>"
	gitReversion string = "0"
	buildTime    string = "0000-00-00 00:00:00"
)

func showVersion() {
	fmt.Printf("程序版本: %s\n代码版本: %s.%s.%s\n时间: %s\ngo版本: %s\n",
		version,
		gitBranch, gitHash, gitReversion,
		buildTime,
		rt.Version())
}

func main() {
	cf := flag.String("conf", "", "配置文件所在路径")
	ver := flag.Bool("version", false, "查看版本号")
	act := flag.String("action", "", "install或uninstall")
	flag.Parse()

	if *ver {
		showVersion()
		return
	}

	if len(*cf) == 0 {
		fmt.Println("缺少-conf参数")
		os.Exit(1)
	}

	var user string
	var depends []string
	if rt.GOOS != "windows" {
		user = "root"
		depends = append(depends, "After=network.target")
	}

	dir, err := filepath.Abs(*cf)
	runtime.Assert(err)

	opt := make(service.KeyValue)
	opt["LimitNOFILE"] = 65535

	appCfg := &service.Config{
		Name:         "smartagent-server",
		DisplayName:  "smartagent-server",
		Description:  "smartagent server",
		UserName:     user,
		Arguments:    []string{"-conf", dir},
		Dependencies: depends,
		Option:       opt,
	}

	dir, err = os.Executable()
	runtime.Assert(err)

	cfg := conf.Load(*cf, filepath.Join(filepath.Dir(dir), "/../"))

	app := app.New(cfg, version)
	sv, err := service.New(app, appCfg)
	runtime.Assert(err)

	switch *act {
	case "install":
		runtime.Assert(sv.Install())
	case "uninstall":
		runtime.Assert(sv.Uninstall())
	default:
		runtime.Assert(sv.Run())
	}
}

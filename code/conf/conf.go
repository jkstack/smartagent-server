package conf

import (
	"os"
	"path/filepath"
	"server/code/utils"
	"sync"

	"github.com/lwch/kvconf"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

const (
	defaultCacheDir       = "/opt/smartagent-server/cache"
	defaultCacheThreshold = 80
	defaultDataDir        = "/opt/smartagent-server/data"
	defaultPluginDir      = "/opt/smartagent-server/plugins"
	defaultLogDir         = "/opt/smartagent-server/logs"
	defaultLogSize        = utils.Bytes(50 * 1024 * 1024)
	defaultLogRotate      = 7
)

type Configure struct {
	sync.RWMutex
	Listen         uint16      `kv:"listen"`
	CacheDir       string      `kv:"cache_dir"`
	CacheThreshold uint        `kv:"cache_threshold"`
	DataDir        string      `kv:"data_dir"`
	PluginDir      string      `kv:"plugin_dir"`
	LogDir         string      `kv:"log_dir"`
	LogSize        utils.Bytes `kv:"log_size"`
	LogRotate      int         `kv:"log_rotate"`
	LoggingReport  string      `kv:"logging_report"`
	// runtime
	WorkDir      string
	PluginList   map[string]*PluginMD5List // name => md5 => info
	PluginLatest map[string]utils.Version  // name => version
}

func Load(dir, abs string) *Configure {
	f, err := os.Open(dir)
	runtime.Assert(err)
	defer f.Close()

	var ret Configure
	runtime.Assert(kvconf.NewDecoder(f).Decode(&ret))
	ret.check(abs)

	ret.WorkDir, _ = os.Getwd()
	ret.PluginList = make(map[string]*PluginMD5List)
	ret.PluginLatest = make(map[string]utils.Version)
	ret.LoadPlugin()

	return &ret
}

func (cfg *Configure) GetPlugin(name, os, arch string) *PluginInfo {
	cfg.RLock()
	list := cfg.PluginList[name]
	ver := cfg.PluginLatest[name]
	cfg.RUnlock()
	return list.get(ver, os, arch)
}

func (cfg *Configure) PluginByMD5(name, md5 string) *PluginInfo {
	cfg.RLock()
	list := cfg.PluginList[name]
	cfg.RUnlock()
	return list.by(md5)
}

func (cfg *Configure) PluginCount() int {
	var cnt int
	cfg.RangePlugin(func(name, version string) {
		if version != "0.0.0" {
			cnt++
		}
	})
	return cnt
}

func (cfg *Configure) RangePlugin(cb func(name, version string)) {
	cfg.RLock()
	defer cfg.RUnlock()
	for name, ver := range cfg.PluginLatest {
		cb(name, ver.String())
	}
}

func (cfg *Configure) check(abs string) {
	if cfg.Listen == 0 {
		panic("invalid listen config")
	}
	if len(cfg.CacheDir) == 0 {
		logging.Info("reset conf.cache_dir to default path: %s", defaultCacheDir)
		cfg.CacheDir = defaultCacheDir
	} else if !filepath.IsAbs(cfg.CacheDir) {
		cfg.CacheDir = filepath.Join(abs, cfg.CacheDir)
	}
	if len(cfg.PluginDir) == 0 {
		logging.Info("reset conf.plugin_dir to default path: %s", defaultPluginDir)
		cfg.PluginDir = defaultPluginDir
	} else if !filepath.IsAbs(cfg.PluginDir) {
		cfg.PluginDir = filepath.Join(abs, cfg.PluginDir)
	}
	if len(cfg.LogDir) == 0 {
		logging.Info("reset conf.log_dir to default path: %s", defaultLogDir)
		cfg.LogDir = defaultLogDir
	} else if !filepath.IsAbs(cfg.LogDir) {
		cfg.LogDir = filepath.Join(abs, cfg.LogDir)
	}
	if len(cfg.DataDir) == 0 {
		logging.Info("reset conf.data_dir to default path: %s", defaultDataDir)
		cfg.DataDir = defaultDataDir
	} else if !filepath.IsAbs(cfg.DataDir) {
		cfg.DataDir = filepath.Join(abs, cfg.DataDir)
	}
	if cfg.LogSize == 0 {
		logging.Info("reset conf.log_size to default size: %s", defaultLogSize.String())
		cfg.LogSize = defaultLogSize
	}
	if cfg.LogRotate == 0 {
		logging.Info("reset conf.log_roate to default count: %d", defaultLogRotate)
		cfg.LogRotate = defaultLogRotate
	}
	if cfg.CacheThreshold == 0 || cfg.CacheThreshold > 100 {
		logging.Info("reset conf.cache_threshold to default limit: %d", defaultCacheThreshold)
		cfg.CacheThreshold = defaultCacheThreshold
	}
}
